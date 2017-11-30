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

package xaapiv1

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
	EVTServerConfig  = EventTypePrefix + "server-config"        // type EventMsg with Data type xaapiv1.ServerCfg
	EVTProjectAdd    = EventTypePrefix + "project-add"          // type EventMsg with Data type xaapiv1.ProjectConfig
	EVTProjectDelete = EventTypePrefix + "project-delete"       // type EventMsg with Data type xaapiv1.ProjectConfig
	EVTProjectChange = EventTypePrefix + "project-state-change" // type EventMsg with Data type xaapiv1.ProjectConfig
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
