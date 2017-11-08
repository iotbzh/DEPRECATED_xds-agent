package agent

import (
	"github.com/iotbzh/xds-agent/lib/apiv1"
	st "github.com/iotbzh/xds-agent/lib/syncthing"
)

// IPROJECT interface implementation for syncthing projects

// STProject .
type STProject struct {
	*Context
	server   *XdsServer
	folder   *XdsFolderConfig
	eventIDs []int
}

// NewProjectST Create a new instance of STProject
func NewProjectST(ctx *Context, svr *XdsServer) *STProject {
	p := STProject{
		Context: ctx,
		server:  svr,
		folder:  &XdsFolderConfig{},
	}
	return &p
}

// Add a new project
func (p *STProject) Add(cfg apiv1.ProjectConfig) (*apiv1.ProjectConfig, error) {
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
		svrPrj.Status = apiv1.StatusErrorConfig
		return nil, err
	}
	if svrPrj.ID != locPrj.ID {
		p.Log.Errorf("Project ID in XDSServer and local ST differ: %s != %s", svrPrj.ID, locPrj.ID)
	}

	// Use Update function to setup remains fields
	return p.UpdateProject(*svrPrj)
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
func (p *STProject) GetProject() *apiv1.ProjectConfig {
	prj := p.server.FolderToProject(*p.folder)
	prj.ServerID = p.server.ID
	return &prj
}

// UpdateProject Update project config
func (p *STProject) UpdateProject(prj apiv1.ProjectConfig) (*apiv1.ProjectConfig, error) {
	// Update folder
	p.folder = p.server.ProjectToFolder(prj)
	svrPrj := p.GetProject()

	// Register events to update folder status
	// Register to XDS Server events
	p.server.EventOn("event:FolderStateChanged", "", p._cbServerFolderChanged)
	if err := p.server.EventRegister("FolderStateChanged", svrPrj.ID); err != nil {
		p.Log.Warningf("XDS Server EventRegister failed: %v", err)
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
	evt := data.(XdsEventFolderChange)

	// Only process event that concerns this project/folder ID
	if p.folder.ID != evt.Folder.ID {
		return nil
	}

	if evt.Folder.IsInSync != p.folder.DataCloudSync.STSvrIsInSync ||
		evt.Folder.Status != p.folder.DataCloudSync.STSvrStatus {

		p.folder.DataCloudSync.STSvrIsInSync = evt.Folder.IsInSync
		p.folder.DataCloudSync.STSvrStatus = evt.Folder.Status

		if err := p.events.Emit(apiv1.EVTProjectChange, p.server.FolderToProject(*p.folder)); err != nil {
			p.Log.Warningf("Cannot notify project change: %v", err)
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
			sts = apiv1.StatusSyncing
		case "idle":
			sts = apiv1.StatusEnable
		}
		inSync = (to == "idle")

	case st.EventFolderPaused:
		if sts == apiv1.StatusEnable {
			sts = apiv1.StatusPause
		}
		inSync = false
	}

	if prevSync != inSync || prevStatus != sts {

		p.folder.DataCloudSync.STLocIsInSync = inSync
		p.folder.DataCloudSync.STLocStatus = sts

		if err := p.events.Emit(apiv1.EVTProjectChange, p.server.FolderToProject(*p.folder)); err != nil {
			p.Log.Warningf("Cannot notify project change: %v", err)
		}
	}
}
