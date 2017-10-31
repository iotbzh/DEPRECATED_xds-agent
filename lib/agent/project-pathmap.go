package agent

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/iotbzh/xds-agent/lib/apiv1"
	common "github.com/iotbzh/xds-common/golib"
)

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
func (p *PathMap) Add(cfg apiv1.ProjectConfig) (*apiv1.ProjectConfig, error) {
	var err error
	var file *os.File
	errMsg := "ClientPath sanity check error (%d): %v"

	// Sanity check to verify that we have RW permission and path-mapping is correct
	dir := cfg.ClientPath
	if !common.Exists(dir) {
		// try to create if not existing
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("Cannot create ClientPath directory: %s", dir)
		}
	}
	if !common.Exists(dir) {
		return nil, fmt.Errorf("ClientPath directory is not accessible: %s", dir)
	}
	if file, err = ioutil.TempFile(dir, ".xds_pathmap_check"); err != nil {
		return nil, fmt.Errorf(errMsg, 1, err)
	}
	// Write a specific message that will be check by server during folder add
	msg := "Pathmap checked message written by xds-agent ID: " + p.Config.AgentUID + "\n"
	if n, err := file.WriteString(msg); n != len(msg) || err != nil {
		return nil, fmt.Errorf(errMsg, 2, err)
	}
	defer func() {
		if file != nil {
			os.Remove(file.Name())
			file.Close()
		}
	}()

	// Convert to Xds folder
	fld := p.server.ProjectToFolder(cfg)
	fld.DataPathMap.CheckFile = file.Name()
	fld.DataPathMap.CheckContent = msg

	// Send request to create folder on XDS server side
	err = p.server.FolderAdd(fld, p.folder)
	if err != nil {
		return nil, fmt.Errorf("Folders mapping verification failure.\n%v", err)
	}

	// 2nd part of sanity checker
	// check specific message added by XDS Server during folder add processing
	content, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return nil, fmt.Errorf(errMsg, 3, err)
	}
	if !strings.Contains(string(content),
		"Pathmap checked message written by xds-server ID") {
		return nil, fmt.Errorf(errMsg, 4, "file content differ")
	}

	return p.GetProject(), nil
}

// Delete a project
func (p *PathMap) Delete() error {
	return p.server.FolderDelete(p.folder.ID)
}

// GetProject Get public part of project config
func (p *PathMap) GetProject() *apiv1.ProjectConfig {
	prj := p.server.FolderToProject(*p.folder)
	prj.ServerID = p.server.ID
	return &prj
}

// UpdateProject Set project config
func (p *PathMap) UpdateProject(prj apiv1.ProjectConfig) (*apiv1.ProjectConfig, error) {
	p.folder = p.server.ProjectToFolder(prj)
	np := p.GetProject()
	if err := p.events.Emit(apiv1.EVTProjectChange, np); err != nil {
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
