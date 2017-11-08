package apiv1

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

// Events definitions
const (
	// EventTypePrefix Used as event prefix
	EventTypePrefix = "event:" // following by event type

	// Supported Events type
	EVTAll           = EventTypePrefix + "all"
	EVTServerConfig  = EventTypePrefix + "server-config"        // data type apiv1.ServerCfg
	EVTProjectAdd    = EventTypePrefix + "project-add"          // data type apiv1.ProjectConfig
	EVTProjectDelete = EventTypePrefix + "project-delete"       // data type apiv1.ProjectConfig
	EVTProjectChange = EventTypePrefix + "project-state-change" // data type apiv1.ProjectConfig
)

// EventMsg Message send
type EventMsg struct {
	Time string      `json:"time"`
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
