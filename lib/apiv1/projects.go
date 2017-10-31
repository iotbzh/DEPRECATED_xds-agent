package apiv1

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
}
