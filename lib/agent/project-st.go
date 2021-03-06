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
	"encoding/json"
	"fmt"

	st "github.com/iotbzh/xds-agent/lib/syncthing"
	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// IPROJECT interface implementation for syncthing projects

// STProject .
type STProject struct {
	*Context
	server   *XdsServer
	folder   *xsapiv1.FolderConfig
	eventIDs []int
}

// NewProjectST Create a new instance of STProject
func NewProjectST(ctx *Context, svr *XdsServer) *STProject {
	p := STProject{
		Context: ctx,
		server:  svr,
		folder:  &xsapiv1.FolderConfig{},
	}
	return &p
}

// Add a new project
func (p *STProject) Add(cfg xaapiv1.ProjectConfig) (*xaapiv1.ProjectConfig, error) {
	var err error

	// Add project/folder into XDS Server
	err = p.server.FolderAdd(p.server.ProjectToFolder(cfg), p.folder)
	if err != nil {
		return nil, err
	}
	svrPrj := p.GetProject()

	// Declare project into local Syncthing
	id, err := p.SThg.FolderChange(st.FolderChangeArg{
		ID:           svrPrj.ID,
		Label:        svrPrj.Label,
		RelativePath: cfg.ClientPath,
		SyncThingID:  p.server.ServerConfig.Builder.SyncThingID,
	})
	if err != nil {
		return nil, err
	}

	locPrj, err := p.SThg.FolderConfigGet(id)
	if err != nil {
		svrPrj.Status = xaapiv1.StatusErrorConfig
		return nil, err
	}
	if svrPrj.ID != locPrj.ID {
		p.Log.Errorf("Project ID in XDSServer and local ST differ: %s != %s", svrPrj.ID, locPrj.ID)
	}

	// Use Setup function to setup remains fields
	return p.Setup(*svrPrj)
}

// Delete a project
func (p *STProject) Delete() error {
	errSvr := p.server.FolderDelete(p.folder.ID)
	errLoc := p.SThg.FolderDelete(p.folder.ID)
	if errSvr != nil {
		return errSvr
	}
	return errLoc
}

// GetProject Get public part of project config
func (p *STProject) GetProject() *xaapiv1.ProjectConfig {
	prj := p.server.FolderToProject(*p.folder)
	prj.ServerID = p.server.ID
	return &prj
}

// Setup Setup local project config
func (p *STProject) Setup(prj xaapiv1.ProjectConfig) (*xaapiv1.ProjectConfig, error) {
	// Update folder
	p.folder = p.server.ProjectToFolder(prj)
	svrPrj := p.GetProject()

	// Register events to update folder status
	// Register to XDS Server events
	if _, err := p.server.EventOn(xsapiv1.EVTFolderStateChange, "", p._cbServerFolderChanged); err != nil {
		p.Log.Errorf("XDS Server EventOn '%s' failed: %v", xsapiv1.EVTFolderStateChange, err)
		return svrPrj, err
	}

	// Register to Local Syncthing events
	for _, evName := range []string{st.EventStateChanged, st.EventFolderPaused} {
		evID, err := p.SThg.Events.Register(evName, p._cbLocalSTEvents, svrPrj.ID, nil)
		if err != nil {
			return nil, err
		}
		p.eventIDs = append(p.eventIDs, evID)
	}

	return svrPrj, nil
}

// Update Update some field of a project
func (p *STProject) Update(prj xaapiv1.ProjectConfig) (*xaapiv1.ProjectConfig, error) {

	if p.folder.ID != prj.ID {
		return nil, fmt.Errorf("Invalid id")
	}

	err := p.server.FolderUpdate(p.server.ProjectToFolder(prj), p.folder)
	if err != nil {
		return nil, err
	}

	return p.GetProject(), nil
}

// GetServer Get the XdsServer that holds this project
func (p *STProject) GetServer() *XdsServer {
	return p.server
}

// Sync Force project files synchronization
func (p *STProject) Sync() error {
	if err := p.server.FolderSync(p.folder.ID); err != nil {
		return err
	}
	return p.SThg.FolderScan(p.folder.ID, "")
}

// IsInSync Check if project files are in-sync
func (p *STProject) IsInSync() (bool, error) {
	// Should be up-to-date by callbacks (see below)
	return p.folder.IsInSync, nil
}

/**
** Private functions
***/

// callback use to update (XDS Server) folder IsInSync status

func (p *STProject) _cbServerFolderChanged(pData interface{}, data interface{}) error {
	evt := xsapiv1.EventMsg{}
	d, err := json.Marshal(data)
	if err != nil {
		p.Log.Errorf("Cannot marshal XDS Server event folder-change err=%v", err)
		return err
	}
	if err = json.Unmarshal(d, &evt); err != nil {
		p.Log.Errorf("Cannot unmarshal XDS Server event folder-change err=%v", err)
		return err
	}

	fld, err := evt.DecodeFolderConfig()
	if err != nil {
		p.Log.Errorf("Cannot decode FolderChanged event: %v", data)
	}

	// Only process event that concerns this project/folder ID
	if p.folder.ID != fld.ID {
		return nil
	}

	if fld.IsInSync != p.folder.DataCloudSync.STSvrIsInSync ||
		fld.Status != p.folder.DataCloudSync.STSvrStatus {

		p.folder.DataCloudSync.STSvrIsInSync = fld.IsInSync
		p.folder.DataCloudSync.STSvrStatus = fld.Status

		if err := p.events.Emit(xaapiv1.EVTProjectChange, p.server.FolderToProject(*p.folder), ""); err != nil {
			p.Log.Warningf("Cannot notify project change (from server): %v", err)
		}
	}
	return nil
}

// callback use to update IsInSync status
func (p *STProject) _cbLocalSTEvents(ev st.Event, data *st.EventsCBData) {

	inSync := p.folder.DataCloudSync.STLocIsInSync
	sts := p.folder.DataCloudSync.STLocStatus
	prevSync := inSync
	prevStatus := sts

	switch ev.Type {

	case st.EventStateChanged:
		to := ev.Data["to"]
		switch to {
		case "scanning", "syncing":
			sts = xaapiv1.StatusSyncing
		case "idle":
			sts = xaapiv1.StatusEnable
		}
		inSync = (to == "idle")

	case st.EventFolderPaused:
		if sts == xaapiv1.StatusEnable {
			sts = xaapiv1.StatusPause
		}
		inSync = false
	}

	if prevSync != inSync || prevStatus != sts {

		p.folder.DataCloudSync.STLocIsInSync = inSync
		p.folder.DataCloudSync.STLocStatus = sts

		if err := p.events.Emit(xaapiv1.EVTProjectChange, p.server.FolderToProject(*p.folder), ""); err != nil {
			p.Log.Warningf("Cannot notify project change (local): %v", err)
		}
	}
}
