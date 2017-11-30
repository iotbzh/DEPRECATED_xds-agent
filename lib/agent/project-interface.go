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

import "github.com/iotbzh/xds-agent/lib/xaapiv1"

// IPROJECT Project interface
type IPROJECT interface {
	Add(cfg xaapiv1.ProjectConfig) (*xaapiv1.ProjectConfig, error)          // Add a new project
	Setup(prj xaapiv1.ProjectConfig) (*xaapiv1.ProjectConfig, error) // Local setup of the project
	Delete() error                                                      // Delete a project
	GetProject() *xaapiv1.ProjectConfig                                   // Get project public configuration
	Update(prj xaapiv1.ProjectConfig) (*xaapiv1.ProjectConfig, error)       // Update project configuration
	GetServer() *XdsServer                                              // Get XdsServer that holds this project
	Sync() error                                                        // Force project files synchronization
	IsInSync() (bool, error)                                            // Check if project files are in-sync
}
