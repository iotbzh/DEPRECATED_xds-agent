/*
 * Copyright (C) 2017 "IoT.bzh"
 * Author Sebastien Douheret <sebastien@iot.bzh>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package agent

import (
	"fmt"
	"time"

	"github.com/iotbzh/xds-agent/lib/xaapiv1"
)

// EventDef Definition on one event
type EventDef struct {
	sids map[string]int
}

// Events Hold registered events per context
type Events struct {
	*Context
	eventsMap map[string]*EventDef
}

// NewEvents creates an instance of Events
func NewEvents(ctx *Context) *Events {
	evMap := make(map[string]*EventDef)
	for _, ev := range xaapiv1.EVTAllList {
		evMap[ev] = &EventDef{
			sids: make(map[string]int),
		}
	}
	return &Events{
		Context:   ctx,
		eventsMap: evMap,
	}
}

// GetList returns the list of all supported events
func (e *Events) GetList() []string {
	return xaapiv1.EVTAllList
}

// Register Used by a client/session to register to a specific (or all) event(s)
func (e *Events) Register(evName, sessionID string) error {
	evs := xaapiv1.EVTAllList
	if evName != xaapiv1.EVTAll {
		if _, ok := e.eventsMap[evName]; !ok {
			return fmt.Errorf("Unsupported event type name")
		}
		evs = []string{evName}
	}
	for _, ev := range evs {
		e.eventsMap[ev].sids[sessionID]++
	}
	return nil
}

// UnRegister Used by a client/session to unregister event(s)
func (e *Events) UnRegister(evName, sessionID string) error {
	evs := xaapiv1.EVTAllList
	if evName != xaapiv1.EVTAll {
		if _, ok := e.eventsMap[evName]; !ok {
			return fmt.Errorf("Unsupported event type name")
		}
		evs = []string{evName}
	}
	for _, ev := range evs {
		if _, exist := e.eventsMap[ev].sids[sessionID]; exist {
			delete(e.eventsMap[ev].sids, sessionID)
			break
		}
	}
	return nil
}

// Emit Used to manually emit an event
func (e *Events) Emit(evName string, data interface{}, fromSid string) error {
	var firstErr error

	if _, ok := e.eventsMap[evName]; !ok {
		return fmt.Errorf("Unsupported event type")
	}

	e.LogSillyf("Emit Event %s: %v", evName, data)

	firstErr = nil
	evm := e.eventsMap[evName]
	for sid := range evm.sids {
		so := e.webServer.sessions.IOSocketGet(sid)
		if so == nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("IOSocketGet return nil (SID=%v)", sid)
			}
			continue
		}
		msg := xaapiv1.EventMsg{
			Time:          time.Now().String(),
			FromSessionID: fromSid,
			Type:          evName,
			Data:          data,
		}

		if err := (*so).Emit(evName, msg); err != nil {
			e.Log.Errorf("WS Emit %v error : %v", evName, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
