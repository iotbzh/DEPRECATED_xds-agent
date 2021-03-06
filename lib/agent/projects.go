/*
 * Copyright (C) 2017 "IoT.bzh"
 * Author Sebastien Douheret <sebastien@iot.bzh>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package agent

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/franciscocpg/reflectme"
	"github.com/iotbzh/xds-agent/lib/syncthing"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	common "github.com/iotbzh/xds-common/golib"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
	"github.com/syncthing/syncthing/lib/sync"
)

// Projects Represent a an XDS Projects
type Projects struct {
	*Context
	SThg     *st.SyncThing
	projects map[string]*IPROJECT
}

// Mutex to make add/delete atomic
var pjMutex = sync.NewMutex()

// NewProjects Create a new instance of Project Model
func NewProjects(ctx *Context, st *st.SyncThing) *Projects {
	return &Projects{
		Context:  ctx,
		SThg:     st,
		projects: make(map[string]*IPROJECT),
	}
}

// Init Load Projects configuration
func (p *Projects) Init(server *XdsServer) error {

	svrList := make(map[string]*XdsServer)
	// If server not set, load for all servers
	if server == nil {
		svrList = p.xdsServers
	} else {
		svrList[server.ID] = server
	}
	errMsg := ""
	for _, svr := range svrList {
		if svr.Disabled {
			continue
		}
		xFlds := []xsapiv1.FolderConfig{}
		if err := svr.GetFolders(&xFlds); err != nil {
			errMsg += fmt.Sprintf("Cannot retrieve folders config of XDS server ID %s : %v \n", svr.ID, err.Error())
			continue
		}
		p.Log.Debugf("Connected to XDS Server %s, %d projects detected", svr.ID, len(xFlds))
		for _, prj := range xFlds {
			newP := svr.FolderToProject(prj)
			if _, err := p.createUpdate(newP, false, true); err != nil {
				// Don't consider that as an error (allow support config without CloudSync support)
				if p.Context.SThg == nil && strings.Contains(err.Error(), "Server doesn't support project type CloudSync") {
					continue
				}

				errMsg += "Error while creating project id " + prj.ID + ": " + err.Error() + "\n "
				continue
			}
		}
	}

	p.Log.Infof("Number of loaded Projects: %d", len(p.projects))

	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}
	return nil
}

// ResolveID Complete a Project ID (helper for user that can use partial ID value)
func (p *Projects) ResolveID(id string) (string, error) {
	if id == "" {
		return "", nil
	}

	match := []string{}
	for iid := range p.projects {
		if strings.HasPrefix(iid, id) {
			match = append(match, iid)
		}
	}

	if len(match) == 1 {
		return match[0], nil
	} else if len(match) == 0 {
		return id, fmt.Errorf("Unknown id")
	}
	return id, fmt.Errorf("Multiple IDs found with provided prefix: " + id)
}

// Get returns the folder config or nil if not existing
func (p *Projects) Get(id string) *IPROJECT {
	if id == "" {
		return nil
	}
	fc, exist := p.projects[id]
	if !exist {
		return nil
	}
	return fc
}

// GetProjectArr returns the config of all folders as an array
func (p *Projects) GetProjectArr() []xaapiv1.ProjectConfig {
	pjMutex.Lock()
	defer pjMutex.Unlock()

	return p.GetProjectArrUnsafe()
}

// GetProjectArrUnsafe Same as GetProjectArr without mutex protection
func (p *Projects) GetProjectArrUnsafe() []xaapiv1.ProjectConfig {
	conf := []xaapiv1.ProjectConfig{}
	for _, v := range p.projects {
		prj := (*v).GetProject()
		conf = append(conf, *prj)
	}
	return conf
}

// Add adds a new folder
func (p *Projects) Add(newP xaapiv1.ProjectConfig, fromSid, requestURL string) (*xaapiv1.ProjectConfig, error) {
	prj, err := p.createUpdate(newP, true, false)
	if err != nil {
		return prj, err
	}

	// Create xds-project.conf file
	prjConfFile := filepath.Join(prj.ClientPath, "xds-project.conf")
	if !common.Exists(prjConfFile) {
		fd, err := os.OpenFile(prjConfFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
		if err != nil {
			return prj, fmt.Errorf("Cannot create xds-project.conf: %v", err.Error())
		}
		fd.WriteString("# XDS project settings\n")
		fd.WriteString("export XDS_AGENT_URL=" + requestURL + "\n")
		fd.WriteString("export XDS_PROJECT_ID=" + prj.ID + "\n")
		if prj.DefaultSdk != "" {
			fd.WriteString("export XDS_SDK_ID=" + prj.DefaultSdk + "\n")
		} else {
			fd.WriteString("#export XDS_SDK_ID=???\n")
		}
		fd.Close()
	}

	// Notify client with event
	if err := p.events.Emit(xaapiv1.EVTProjectAdd, *prj, fromSid); err != nil {
		p.Log.Warningf("Cannot notify project deletion: %v", err)
	}

	return prj, err
}

// CreateUpdate creates or update a folder
func (p *Projects) createUpdate(newF xaapiv1.ProjectConfig, create bool, initial bool) (*xaapiv1.ProjectConfig, error) {
	var err error

	pjMutex.Lock()
	defer pjMutex.Unlock()

	// Sanity check
	if _, exist := p.projects[newF.ID]; create && exist {
		return nil, fmt.Errorf("ID already exists")
	}
	if newF.ClientPath == "" {
		return nil, fmt.Errorf("ClientPath must be set")
	}
	if newF.ServerID == "" {
		return nil, fmt.Errorf("Server ID must be set")
	}
	var svr *XdsServer
	var exist bool
	if svr, exist = p.xdsServers[newF.ServerID]; !exist {
		return nil, fmt.Errorf("Unknown Server ID %s", newF.ServerID)
	}

	// Check type supported
	b, exist := svr.ServerConfig.SupportedSharing[string(newF.Type)]
	if !exist || !b {
		return nil, fmt.Errorf("Server doesn't support project type %s", newF.Type)
	}

	// Create a new folder object
	var fld IPROJECT
	switch newF.Type {
	// SYNCTHING
	case xaapiv1.TypeCloudSync:
		if p.SThg != nil {
			fld = NewProjectST(p.Context, svr)
		} else {
			return nil, fmt.Errorf("Cloud Sync project not supported")
		}

	// PATH MAP
	case xaapiv1.TypePathMap:
		fld = NewProjectPathMap(p.Context, svr)
	default:
		return nil, fmt.Errorf("Unsupported folder type")
	}

	var newPrj *xaapiv1.ProjectConfig
	if create {
		// Add project on server
		if newPrj, err = fld.Add(newF); err != nil {
			newF.Status = xaapiv1.StatusErrorConfig
			log.Printf("ERROR Adding project: %v\n", err)
			return newPrj, err
		}
	} else {
		// Just update project config
		if newPrj, err = fld.Setup(newF); err != nil {
			newF.Status = xaapiv1.StatusErrorConfig
			log.Printf("ERROR Updating project: %v\n", err)
			return newPrj, err
		}
	}

	// Sanity check
	if newPrj.ID == "" {
		log.Printf("ERROR project ID empty: %v", newF)
		return newPrj, fmt.Errorf("Project ID empty")
	}

	// Add to folders list
	p.projects[newPrj.ID] = &fld

	// Force sync to get an initial sync status
	// (need to defer to be sure that WS events will arrive after HTTP creation reply)
	go func() {
		time.Sleep(time.Millisecond * 500)
		fld.Sync()
	}()

	return newPrj, nil
}

// Delete deletes a specific folder
func (p *Projects) Delete(id, fromSid string) (xaapiv1.ProjectConfig, error) {
	var err error

	pjMutex.Lock()
	defer pjMutex.Unlock()

	fld := xaapiv1.ProjectConfig{}
	fc, exist := p.projects[id]
	if !exist {
		return fld, fmt.Errorf("Unknown id")
	}

	prj := (*fc).GetProject()

	if err = (*fc).Delete(); err != nil {
		return *prj, err
	}

	delete(p.projects, id)

	// Notify client with event
	if err := p.events.Emit(xaapiv1.EVTProjectDelete, *prj, fromSid); err != nil {
		p.Log.Warningf("Cannot notify project deletion: %v", err)
	}

	return *prj, err
}

// ForceSync Force the synchronization of a folder
func (p *Projects) ForceSync(id string) error {
	fc := p.Get(id)
	if fc == nil {
		return fmt.Errorf("Unknown id")
	}
	return (*fc).Sync()
}

// IsProjectInSync Returns true when folder is in sync
func (p *Projects) IsProjectInSync(id string) (bool, error) {
	fc := p.Get(id)
	if fc == nil {
		return false, fmt.Errorf("Unknown id")
	}
	return (*fc).IsInSync()
}

// Update Update some field of a project
func (p *Projects) Update(id string, prj xaapiv1.ProjectConfig, fromSid string) (*xaapiv1.ProjectConfig, error) {

	pjMutex.Lock()
	defer pjMutex.Unlock()

	fc, exist := p.projects[id]
	if !exist {
		return nil, fmt.Errorf("Unknown id")
	}

	// Copy current in a new object to change nothing in case of an error rises
	newFld := xaapiv1.ProjectConfig{}
	reflectme.Copy((*fc).GetProject(), &newFld)

	// Only update some fields
	dirty := false
	for _, fieldName := range xaapiv1.ProjectConfigUpdatableFields {
		valNew, err := reflectme.GetField(prj, fieldName)
		if err == nil {
			valCur, err := reflectme.GetField(newFld, fieldName)
			if err == nil && valNew != valCur {
				err = reflectme.SetField(&newFld, fieldName, valNew)
				if err != nil {
					return nil, err
				}
				dirty = true
			}
		}
	}

	if !dirty {
		return &newFld, nil
	}

	upPrj, err := (*fc).Update(newFld)
	if err != nil {
		return nil, err
	}

	// Notify client with event
	if err := p.events.Emit(xaapiv1.EVTProjectChange, *upPrj, fromSid); err != nil {
		p.Log.Warningf("Cannot notify project change: %v", err)
	}
	return upPrj, err
}
