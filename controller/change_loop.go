// Copyright (c) 2016 Matthias Neugebauer <mtneug@mailbox.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controller

import (
	"context"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/mtneug/pkg/startstopper"
	"github.com/mtneug/spate/autoscaler"
	"github.com/mtneug/spate/docker"
	"github.com/mtneug/spate/event"
)

const labelSpate = "de.mtneug.spate"

var serviceListOptions types.ServiceListOptions

func init() {
	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("%s=%s", labelSpate, "enable"))
	serviceListOptions = types.ServiceListOptions{Filters: f}
}

type changeLoop struct {
	startstopper.StartStopper

	period         time.Duration
	eventQueue     chan<- event.Event
	autoscalersMap startstopper.Map

	// stored so that it doesn't need to be reallocated
	seen map[string]bool
}

func newChangeLoop(p time.Duration, eq chan<- event.Event, m startstopper.Map) *changeLoop {
	cl := &changeLoop{
		period:         p,
		eventQueue:     eq,
		autoscalersMap: m,
		seen:           make(map[string]bool),
	}
	cl.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(cl.run))
	return cl
}

func (cl *changeLoop) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Change detection loop started")
	defer log.Debug("Change detection loop stopped")

	cl.tick(ctx)
	for {
		select {
		case <-time.After(cl.period):
			cl.tick(ctx)
		case <-stopChan:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (cl *changeLoop) tick(ctx context.Context) {
	services, err := docker.C.ServiceList(ctx, serviceListOptions)
	if err != nil {
		log.WithError(err).Error("Failed to get list of services")
		return
	}

	for _, srv := range services {
		cl.seen[srv.ID] = true

		ss, present := cl.autoscalersMap.Get(srv.ID)
		if !present {
			// Add
			cl.eventQueue <- event.New(event.TypeServiceCreated, srv)
		} else {
			a := ss.(*autoscaler.Autoscaler)
			a.RLock()
			if a.Service.Version.Index < srv.Version.Index {
				// Update
				cl.eventQueue <- event.New(event.TypeServiceUpdated, srv)
			}
			a.RUnlock()
		}
	}

	cl.autoscalersMap.ForEach(func(id string, ss startstopper.StartStopper) {
		if !cl.seen[id] {
			// Delete
			a := ss.(*autoscaler.Autoscaler)
			a.RLock()
			cl.eventQueue <- event.New(event.TypeServiceDeleted, a.Service)
			a.RUnlock()
		}
		delete(cl.seen, id)
	})
}
