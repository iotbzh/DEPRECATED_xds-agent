package agent

import (
	"fmt"
	"log"
	"time"

	"github.com/iotbzh/xds-agent/lib/syncthing"
	"github.com/syncthing/syncthing/lib/sync"
)

// Projects Represent a an XDS Projects
type Projects struct {
	*Context
	SThg     *st.SyncThing
	projects map[string]*IPROJECT
	//SEB registerCB []RegisteredCB
}

/* SEB
type RegisteredCB struct {
	cb   *EventCB
	data *EventCBData
}
*/

// Mutex to make add/delete atomic
var pjMutex = sync.NewMutex()

// NewProjects Create a new instance of Project Model
func NewProjects(ctx *Context, st *st.SyncThing) *Projects {
	return &Projects{
		Context:  ctx,
		SThg:     st,
		projects: make(map[string]*IPROJECT),
		//registerCB: []RegisteredCB{},
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
		xFlds := []FolderConfig{}
		if err := svr.HTTPGet("/folders", &xFlds); err != nil {
			errMsg += fmt.Sprintf("Cannot retrieve folders config of XDS server ID %s : %v \n", svr.ID, err.Error())
			continue
		}
		p.Log.Debugf("Server %s, %d projects detected", svr.ID[:8], len(xFlds))
		for _, prj := range xFlds {
			newP := svr.FolderToProject(prj)
			if /*nPrj*/ _, err := p.createUpdate(newP, false, true); err != nil {
				errMsg += "Error while creating project id " + prj.ID + ": " + err.Error() + "\n "
				continue
			}

			/* FIXME emit EVTProjectChange event ?
			if err := p.events.Emit(EVTProjectChange, *nPrj); err != nil {
				p.Log.Warningf("Cannot notify project change: %v", err)
			}
			*/
		}

	}

	p.Log.Infof("Number of loaded Projects: %d", len(p.projects))

	if errMsg != "" {
		return fmt.Errorf(errMsg)
	}
	return nil
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
func (p *Projects) GetProjectArr() []ProjectConfig {
	pjMutex.Lock()
	defer pjMutex.Unlock()

	return p.GetProjectArrUnsafe()
}

// GetProjectArrUnsafe Same as GetProjectArr without mutex protection
func (p *Projects) GetProjectArrUnsafe() []ProjectConfig {
	conf := []ProjectConfig{}
	for _, v := range p.projects {
		prj := (*v).GetProject()
		conf = append(conf, *prj)
	}
	return conf
}

// Add adds a new folder
func (p *Projects) Add(newF ProjectConfig) (*ProjectConfig, error) {
	prj, err := p.createUpdate(newF, true, false)
	if err != nil {
		return prj, err
	}

	// Notify client with event
	if err := p.events.Emit(EVTProjectAdd, *prj); err != nil {
		p.Log.Warningf("Cannot notify project deletion: %v", err)
	}

	return prj, err
}

// CreateUpdate creates or update a folder
func (p *Projects) createUpdate(newF ProjectConfig, create bool, initial bool) (*ProjectConfig, error) {
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
	case TypeCloudSync:
		if p.SThg != nil {
			/*SEB fld = f.SThg.NewFolderST(f.Conf)*/
			fld = NewProjectST(p.Context, svr)
		} else {
			return nil, fmt.Errorf("Cloud Sync project not supported")
		}

	// PATH MAP
	case TypePathMap:
		fld = NewProjectPathMap(p.Context, svr)
	default:
		return nil, fmt.Errorf("Unsupported folder type")
	}

	var newPrj *ProjectConfig
	if create {
		// Add project on server
		if newPrj, err = fld.Add(newF); err != nil {
			newF.Status = StatusErrorConfig
			log.Printf("ERROR Adding folder: %v\n", err)
			return newPrj, err
		}
	} else {
		// Just update project config
		newPrj = fld.SetProject(newF)
	}

	// Sanity check
	if newPrj.ID == "" {
		log.Printf("ERROR project ID empty: %v", newF)
		return newPrj, fmt.Errorf("Project ID empty")
	}

	// Add to folders list
	p.projects[newPrj.ID] = &fld

	// Force sync after creation
	// (need to defer to be sure that WS events will arrive after HTTP creation reply)
	go func() {
		time.Sleep(time.Millisecond * 500)
		fld.Sync()
	}()

	return newPrj, nil
}

// Delete deletes a specific folder
func (p *Projects) Delete(id string) (ProjectConfig, error) {
	var err error

	pjMutex.Lock()
	defer pjMutex.Unlock()

	fld := ProjectConfig{}
	fc, exist := p.projects[id]
	if !exist {
		return fld, fmt.Errorf("unknown id")
	}

	prj := (*fc).GetProject()

	if err = (*fc).Delete(); err != nil {
		return *prj, err
	}

	delete(p.projects, id)

	// Notify client with event
	if err := p.events.Emit(EVTProjectDelete, *prj); err != nil {
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
