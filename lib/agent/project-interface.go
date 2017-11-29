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
