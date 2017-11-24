package apiv1

import (
	"encoding/json"
	"fmt"
)

// EventRegisterArgs is the parameters (json format) of /events/register command
type EventRegisterArgs struct {
	Name      string `json:"name"`
	ProjectID string `json:"filterProjectID"`
}

// EventUnRegisterArgs is the parameters (json format) of /events/unregister command
type EventUnRegisterArgs struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

// Events Type definitions
const (
	// EventTypePrefix Used as event prefix
	EventTypePrefix = "event:" // following by event type

	// Supported Events type
	EVTAll           = EventTypePrefix + "all"
	EVTServerConfig  = EventTypePrefix + "server-config"        // type EventMsg with Data type apiv1.ServerCfg
	EVTProjectAdd    = EventTypePrefix + "project-add"          // type EventMsg with Data type apiv1.ProjectConfig
	EVTProjectDelete = EventTypePrefix + "project-delete"       // type EventMsg with Data type apiv1.ProjectConfig
	EVTProjectChange = EventTypePrefix + "project-state-change" // type EventMsg with Data type apiv1.ProjectConfig
)

// EVTAllList List of all supported events
var EVTAllList = []string{
	EVTServerConfig,
	EVTProjectAdd,
	EVTProjectDelete,
	EVTProjectChange,
}

// EventMsg Event message send over Websocket, data format depend to Type (see DecodeXXX function)
type EventMsg struct {
	Time          string      `json:"time"`      // Timestamp
	FromSessionID string      `json:"sessionID"` // Session ID of client that emits this event
	Type          string      `json:"type"`      // Data type
	Data          interface{} `json:"data"`      // Data
}

// DecodeServerCfg Helper to decode Data field type ServerCfg
func (e *EventMsg) DecodeServerCfg() (ServerCfg, error) {
	p := ServerCfg{}
	if e.Type != EVTServerConfig {
		return p, fmt.Errorf("Invalid type")
	}
	d, err := json.Marshal(e.Data)
	if err == nil {
		err = json.Unmarshal(d, &p)
	}
	return p, err
}

// DecodeProjectConfig Helper to decode Data field type ProjectConfig
func (e *EventMsg) DecodeProjectConfig() (ProjectConfig, error) {
	var err error
	p := ProjectConfig{}
	switch e.Type {
	case EVTProjectAdd, EVTProjectChange, EVTProjectDelete:
		d := []byte{}
		d, err = json.Marshal(e.Data)
		if err == nil {
			err = json.Unmarshal(d, &p)
		}
	default:
		err = fmt.Errorf("Invalid type")
	}
	return p, err
}
