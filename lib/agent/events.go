package agent

import (
	"fmt"
	"time"

	"github.com/iotbzh/xds-agent/lib/apiv1"
)

var _EVTAllList = []string{
	apiv1.EVTServerConfig,
	apiv1.EVTProjectAdd,
	apiv1.EVTProjectDelete,
	apiv1.EVTProjectChange,
}

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
	for _, ev := range _EVTAllList {
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
	return _EVTAllList
}

// Register Used by a client/session to register to a specific (or all) event(s)
func (e *Events) Register(evName, sessionID string) error {
	evs := _EVTAllList
	if evName != apiv1.EVTAll {
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
	evs := _EVTAllList
	if evName != apiv1.EVTAll {
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
func (e *Events) Emit(evName string, data interface{}) error {
	var firstErr error

	if _, ok := e.eventsMap[evName]; !ok {
		return fmt.Errorf("Unsupported event type")
	}

	if e.LogLevelSilly {
		e.Log.Debugf("Emit Event %s: %v", evName, data)
	}

	firstErr = nil
	evm := e.eventsMap[evName]
	for sid := range evm.sids {
		so := e.webServer.sessions.IOSocketGet(sid)
		if so == nil {
			if firstErr == nil {
				firstErr = fmt.Errorf("IOSocketGet return nil")
			}
			continue
		}
		msg := apiv1.EventMsg{
			Time: time.Now().String(),
			Type: evName,
			Data: data,
		}
		if err := (*so).Emit(apiv1.EventTypePrefix+evName, msg); err != nil {
			e.Log.Errorf("WS Emit %v error : %v", apiv1.EventTypePrefix+evName, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
