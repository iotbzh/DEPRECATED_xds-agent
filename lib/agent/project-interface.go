package agent

import "github.com/iotbzh/xds-agent/lib/apiv1"

// IPROJECT Project interface
type IPROJECT interface {
	Add(cfg apiv1.ProjectConfig) (*apiv1.ProjectConfig, error)           // Add a new project
	Delete() error                                                             // Delete a project
	GetProject() *apiv1.ProjectConfig                                       // Get project public configuration
	UpdateProject(prj apiv1.ProjectConfig) (*apiv1.ProjectConfig, error) // Update project configuration
	GetServer() *XdsServer                                                     // Get XdsServer that holds this project
	Sync() error                                                               // Force project files synchronization
	IsInSync() (bool, error)                                                   // Check if project files are in-sync
}
