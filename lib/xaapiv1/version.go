package xaapiv1

// VersionData
type VersionData struct {
	ID            string `json:"id"`
	Version       string `json:"version"`
	APIVersion    string `json:"apiVersion"`
	VersionGitTag string `json:"gitTag"`
}

// XDSVersion
type XDSVersion struct {
	Client VersionData   `json:"client"`
	Server []VersionData `json:"servers"`
}
