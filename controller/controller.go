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
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/mtneug/pkg/startstopper"
	"github.com/mtneug/spate/api/types"
)

// Controller monitors Docker Swarm services and scales them if needed.
type Controller struct {
	startstopper.StartStopper

	autoscalers startstopper.Map
	eventQueue  chan types.Event
	eventLoop   startstopper.StartStopper
	changeLoop  startstopper.StartStopper
}

// New creates a new controller.
func New(p time.Duration, m startstopper.Map) (*Controller, error) {
	eq := make(chan types.Event, 20)
	ctrl := &Controller{
		autoscalers: m,
		eventQueue:  eq,
		eventLoop:   newEventLoop(eq, m),
		changeLoop:  newChangeLoop(p, eq, m),
	}
	ctrl.StartStopper = startstopper.NewGo(startstopper.RunnerFunc(ctrl.run))

	return ctrl, nil
}

func (c *Controller) run(ctx context.Context, stopChan <-chan struct{}) error {
	log.Debug("Starting controller")
	defer log.Debug("Controller stopped")

	err := c.changeLoop.Start(ctx)
	if err != nil {
		return err
	}

	err = c.eventLoop.Start(ctx)
	if err != nil {
		_ = c.changeLoop.Stop(ctx)
		return err
	}

	<-stopChan

	var err2 error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		err = c.changeLoop.Stop(ctx)
		wg.Done()
	}()
	go func() {
		err2 = c.eventLoop.Stop(ctx)
		wg.Done()
	}()

	wg.Wait()

	if err != nil {
		return err
	}
	if err2 != nil {
		return err2
	}

	return nil
}