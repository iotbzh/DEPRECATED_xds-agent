package agent

// IPROJECT interface implementation for native/path mapping projects

// PathMap .
type PathMap struct {
	*Context
	server *XdsServer
	folder *XdsFolderConfig
}

// NewProjectPathMap Create a new instance of PathMap
func NewProjectPathMap(ctx *Context, svr *XdsServer) *PathMap {
	p := PathMap{
		Context: ctx,
		server:  svr,
		folder:  &XdsFolderConfig{},
	}
	return &p
}

// Add a new project
func (p *PathMap) Add(cfg ProjectConfig) (*ProjectConfig, error) {
	var err error

	// SEB TODO: check local/server directory access

	err = p.server.FolderAdd(p.server.ProjectToFolder(cfg), p.folder)
	if err != nil {
		return nil, err
	}

	return p.GetProject(), nil
}

// Delete a project
func (p *PathMap) Delete() error {
	return p.server.FolderDelete(p.folder.ID)
}

// GetProject Get public part of project config
func (p *PathMap) GetProject() *ProjectConfig {
	prj := p.server.FolderToProject(*p.folder)
	prj.ServerID = p.server.ID
	return &prj
}

// UpdateProject Set project config
func (p *PathMap) UpdateProject(prj ProjectConfig) (*ProjectConfig, error) {
	p.folder = p.server.ProjectToFolder(prj)
	np := p.GetProject()
	if err := p.events.Emit(EVTProjectChange, np); err != nil {
		return np, err
	}
	return np, nil
}

// GetServer Get the XdsServer that holds this project
func (p *PathMap) GetServer() *XdsServer {
	return p.server
}

// Sync Force project files synchronization
func (p *PathMap) Sync() error {
	return nil
}

// IsInSync Check if project files are in-sync
func (p *PathMap) IsInSync() (bool, error) {
	return true, nil
}
