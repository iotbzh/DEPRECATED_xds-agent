package apiv1

// APIConfig parameters (json format) of /config command
type APIConfig struct {
	Servers []ServerCfg `json:"servers"`

	// Not exposed outside in JSON
	Version       string `json:"-"`
	APIVersion    string `json:"-"`
	VersionGitTag string `json:"-"`
}

// ServerCfg .
type ServerCfg struct {
	ID         string `json:"id"`
	URL        string `json:"url"`
	APIURL     string `json:"apiUrl"`
	PartialURL string `json:"partialUrl"`
	ConnRetry  int    `json:"connRetry"`
	Connected  bool   `json:"connected"`
	Disabled   bool   `json:"disabled"`
}
