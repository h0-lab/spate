// Copyright © 2016 Matthias Neugebauer <mtneug@mailbox.org>
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
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	dockerTypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/mtneug/pkg/startstopper"
	"github.com/mtneug/pkg/ulid"
	"github.com/mtneug/spate/api/types"
	"github.com/mtneug/spate/autoscaler"
	"github.com/mtneug/spate/consts"
	"github.com/mtneug/spate/docker"
)

var serviceListOptions dockerTypes.ServiceListOptions

func init() {
	f := filters.NewArgs()
	f.Add("label", fmt.Sprintf("%s=%s", consts.LabelSpate, consts.Enable))
	serviceListOptions = dockerTypes.ServiceListOptions{Filter: f}
}

type changeLoop struct {
	startstopper.StartStopper

	period      time.Duration
	eventQueue  chan<- types.Event
	autoscalers startstopper.Map

	// stored so that it doesn't need to be reallocated
	seen map[string]bool
}

func newChangeLoop(p time.Duration, eq chan<- types.Event, m startstopper.Map) *changeLoop {
	cl := &changeLoop{
		period:      p,
		eventQueue:  eq,
		autoscalers: m,
		seen:        make(map[string]bool),
	}
	cl.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(cl.run))
	return cl
}

func (cl *changeLoop) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Change detection loop started")
	defer log.Debug("Change detection loop stopped")

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

		ss, present := cl.autoscalers.Get(srv.ID)
		if !present {
			// Add
			cl.eventQueue <- types.Event{
				ID:     ulid.New().String(),
				Type:   types.EventTypeServiceCreated,
				Object: srv,
			}
		} else {
			a, ok := ss.(*autoscaler.Autoscaler)
			if !ok {
				log.
					WithError(errors.New("controller: type assertion failed")).
					Error("Failed to get autoscaler")
				return
			}
			if a.Service.Version.Index < srv.Version.Index {
				// Update
				cl.eventQueue <- types.Event{
					ID:     ulid.New().String(),
					Type:   types.EventTypeServiceUpdated,
					Object: srv,
				}
			}
		}
	}

	cl.autoscalers.ForEach(func(id string, ss startstopper.StartStopper) {
		if !cl.seen[id] {
			// Delete
			a, ok := ss.(*autoscaler.Autoscaler)
			if !ok {
				log.
					WithError(errors.New("controller: type assertion failed")).
					Error("Failed to get autoscaler")
				return
			}
			cl.eventQueue <- types.Event{
				ID:     ulid.New().String(),
				Type:   types.EventTypeServiceDeleted,
				Object: a.Service,
			}
		}
		delete(cl.seen, id)
	})
}