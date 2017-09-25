package agent

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

type EventCBData map[string]interface{}
type EventCB func(cfg *ProjectConfig, data *EventCBData)

// IPROJECT Project interface
type IPROJECT interface {
	Add(cfg ProjectConfig) (*ProjectConfig, error) // Add a new project
	Delete() error                                 // Delete a project
	GetProject() *ProjectConfig                    // Get project public configuration
	SetProject(prj ProjectConfig) *ProjectConfig   // Set project configuration
	GetServer() *XdsServer                         // Get XdsServer that holds this project
	GetFullPath(dir string) string                 // Get project full path
	Sync() error                                   // Force project files synchronization
	IsInSync() (bool, error)                       // Check if project files are in-sync
}

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
}
