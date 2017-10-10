package agent

import (
	"fmt"
	"time"
)

// Events constants
const (
	// EventTypePrefix Used as event prefix
	EventTypePrefix = "event:" // following by event type

	// Supported Events type
	EVTAll           = "all"
	EVTServerConfig  = "server-config"        // data type ServerCfg
	EVTProjectAdd    = "project-add"          // data type ProjectConfig
	EVTProjectDelete = "project-delete"       // data type ProjectConfig
	EVTProjectChange = "project-state-change" // data type ProjectConfig
)

var _EVTAllList = []string{
	EVTServerConfig,
	EVTProjectAdd,
	EVTProjectDelete,
	EVTProjectChange,
}

// EventMsg Message send
type EventMsg struct {
	Time string      `json:"time"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type EventDef struct {
	sids map[string]int
}

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
	if evName != EVTAll {
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
	if evName != EVTAll {
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
		msg := EventMsg{
			Time: time.Now().String(),
			Type: evName,
			Data: data,
		}
		if err := (*so).Emit(EventTypePrefix+evName, msg); err != nil {
			e.Log.Errorf("WS Emit %v error : %v", EventTypePrefix+evName, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
