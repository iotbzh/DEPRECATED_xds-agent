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

// ProjectType definition
type ProjectType string

const (
	TypePathMap   = "PathMap"
	TypeCloudSync = "CloudSync"
	TypeCifsSmb   = "CIFS"
)

// Project Status definition
const (
	StatusErrorConfig = "ErrorConfig"
	StatusDisable     = "Disable"
	StatusEnable      = "Enable"
	StatusPause       = "Pause"
	StatusSyncing     = "Syncing"
)

// ProjectConfig is the config for one project
type ProjectConfig struct {
	ID         string      `json:"id"`
	ServerID   string      `json:"serverId"`
	Label      string      `json:"label"`
	ClientPath string      `json:"clientPath"`
	ServerPath string      `json:"serverPath"`
	Type       ProjectType `json:"type"`
	Status     string      `json:"status"`
	IsInSync   bool        `json:"isInSync"`
	DefaultSdk string      `json:"defaultSdk"`
	ClientData string      `json:"clientData"` // free form field that can used by client
}

// ProjectConfigUpdatableFields List fields that can be updated using Update function
var ProjectConfigUpdatableFields = []string{
	"Label", "DefaultSdk", "ClientData",
}
