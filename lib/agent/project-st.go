package agent

import "github.com/iotbzh/xds-agent/lib/syncthing"

// SEB TODO

// IPROJECT interface implementation for syncthing projects

// STProject .
type STProject struct {
	*Context
	server *XdsServer
	folder *FolderConfig
}

// NewProjectST Create a new instance of STProject
func NewProjectST(ctx *Context, svr *XdsServer) *STProject {
	p := STProject{
		Context: ctx,
		server:  svr,
		folder:  &FolderConfig{},
	}
	return &p
}

// Add a new project
func (p *STProject) Add(cfg ProjectConfig) (*ProjectConfig, error) {
	var err error

	err = p.server.FolderAdd(p.server.ProjectToFolder(cfg), p.folder)
	if err != nil {
		return nil, err
	}
	svrPrj := p.GetProject()

	// Declare project into local Syncthing
	p.SThg.FolderChange(st.FolderChangeArg{
		ID:           cfg.ID,
		Label:        cfg.Label,
		RelativePath: cfg.ClientPath,
		SyncThingID:  p.server.ServerConfig.Builder.SyncThingID,
	})

	return svrPrj, nil
}

// Delete a project
func (p *STProject) Delete() error {
	return p.server.FolderDelete(p.folder.ID)
}

// GetProject Get public part of project config
func (p *STProject) GetProject() *ProjectConfig {
	prj := p.server.FolderToProject(*p.folder)
	prj.ServerID = p.server.ID
	return &prj
}

// SetProject Set project config
func (p *STProject) SetProject(prj ProjectConfig) *ProjectConfig {
	// SEB TODO
	p.folder = p.server.ProjectToFolder(prj)
	return p.GetProject()
}

// GetServer Get the XdsServer that holds this project
func (p *STProject) GetServer() *XdsServer {
	// SEB TODO
	return p.server
}

// GetFullPath returns the full path of a directory (from server POV)
func (p *STProject) GetFullPath(dir string) string {
	/* SEB
	if &dir == nil {
		return p.folder.DataSTProject.ServerPath
	}
	return filepath.Join(p.folder.DataSTProject.ServerPath, dir)
	*/
	return "SEB TODO"
}

// Sync Force project files synchronization
func (p *STProject) Sync() error {
	// SEB TODO
	return nil
}

// IsInSync Check if project files are in-sync
func (p *STProject) IsInSync() (bool, error) {
	// SEB TODO
	return false, nil
}
