package agent

import (
	"path/filepath"
)

// IPROJECT interface implementation for native/path mapping projects

// PathMap .
type PathMap struct {
	*Context
	server *XdsServer
	folder *FolderConfig
}

// NewProjectPathMap Create a new instance of PathMap
func NewProjectPathMap(ctx *Context, svr *XdsServer) *PathMap {
	p := PathMap{
		Context: ctx,
		server:  svr,
		folder:  &FolderConfig{},
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

// SetProject Set project config
func (p *PathMap) SetProject(prj ProjectConfig) *ProjectConfig {
	p.folder = p.server.ProjectToFolder(prj)
	return p.GetProject()
}

// GetServer Get the XdsServer that holds this project
func (p *PathMap) GetServer() *XdsServer {
	return p.server
}

// GetFullPath returns the full path of a directory (from server POV)
func (p *PathMap) GetFullPath(dir string) string {
	if &dir == nil {
		return p.folder.DataPathMap.ServerPath
	}
	return filepath.Join(p.folder.DataPathMap.ServerPath, dir)
}

// Sync Force project files synchronization
func (p *PathMap) Sync() error {
	return nil
}

// IsInSync Check if project files are in-sync
func (p *PathMap) IsInSync() (bool, error) {
	return true, nil
}
