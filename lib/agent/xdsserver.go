package agent

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
	common "github.com/iotbzh/xds-common/golib"
	uuid "github.com/satori/go.uuid"
	sio_client "github.com/sebd71/go-socket.io-client"
)

// XdsServer .
type XdsServer struct {
	*Context
	ID           string
	BaseURL      string
	APIURL       string
	PartialURL   string
	ConnRetry    int
	Connected    bool
	Disabled     bool
	ServerConfig *XdsServerConfig

	// Events management
	CBOnError      func(error)
	CBOnDisconnect func(error)
	sockEvents     map[string][]*caller
	sockEventsLock *sync.Mutex

	// Private fields
	client    *common.HTTPClient
	ioSock    *sio_client.Client
	logOut    io.Writer
	apiRouter *gin.RouterGroup
}

// XdsServerConfig Data return by GET /config
type XdsServerConfig struct {
	ID               string           `json:"id"`
	Version          string           `json:"version"`
	APIVersion       string           `json:"apiVersion"`
	VersionGitTag    string           `json:"gitTag"`
	SupportedSharing map[string]bool  `json:"supportedSharing"`
	Builder          XdsBuilderConfig `json:"builder"`
}

// XdsBuilderConfig represents the builder container configuration
type XdsBuilderConfig struct {
	IP          string `json:"ip"`
	Port        string `json:"port"`
	SyncThingID string `json:"syncThingID"`
}

// XdsFolderType XdsServer folder type
type XdsFolderType string

const (
	XdsTypePathMap   = "PathMap"
	XdsTypeCloudSync = "CloudSync"
	XdsTypeCifsSmb   = "CIFS"
)

// XdsFolderConfig XdsServer folder config
type XdsFolderConfig struct {
	ID         string        `json:"id"`
	Label      string        `json:"label"`
	ClientPath string        `json:"path"`
	Type       XdsFolderType `json:"type"`
	Status     string        `json:"status"`
	IsInSync   bool          `json:"isInSync"`
	DefaultSdk string        `json:"defaultSdk"`
	// Specific data depending on which Type is used
	DataPathMap   XdsPathMapConfig   `json:"dataPathMap,omitempty"`
	DataCloudSync XdsCloudSyncConfig `json:"dataCloudSync,omitempty"`
}

// XdsPathMapConfig Path mapping specific data
type XdsPathMapConfig struct {
	ServerPath   string `json:"serverPath"`
	CheckFile    string `json:"checkFile"`
	CheckContent string `json:"checkContent"`
}

// XdsCloudSyncConfig CloudSync (AKA Syncthing) specific data
type XdsCloudSyncConfig struct {
	SyncThingID   string `json:"syncThingID"`
	STSvrStatus   string `json:"-"`
	STSvrIsInSync bool   `json:"-"`
	STLocStatus   string `json:"-"`
	STLocIsInSync bool   `json:"-"`
}

// XdsEventRegisterArgs arguments used to register to XDS server events
type XdsEventRegisterArgs struct {
	Name      string `json:"name"`
	ProjectID string `json:"filterProjectID"`
}

// XdsEventFolderChange Folder change event structure
type XdsEventFolderChange struct {
	Time   string          `json:"time"`
	Type   string          `json:"type"`
	Folder XdsFolderConfig `json:"folder"`
}

// caller Used to chain event listeners
type caller struct {
	id        uuid.UUID
	EventName string
	Func      func(interface{})
}

const _IDTempoPrefix = "tempo-"

// NewXdsServer creates an instance of XdsServer
func NewXdsServer(ctx *Context, conf xdsconfig.XDSServerConf) *XdsServer {
	return &XdsServer{
		Context:    ctx,
		ID:         _IDTempoPrefix + uuid.NewV1().String(),
		BaseURL:    conf.URL,
		APIURL:     conf.APIBaseURL + conf.APIPartialURL,
		PartialURL: conf.APIPartialURL,
		ConnRetry:  conf.ConnRetry,
		Connected:  false,
		Disabled:   false,

		sockEvents:     make(map[string][]*caller),
		sockEventsLock: &sync.Mutex{},
		logOut:         ctx.Log.Out,
	}
}

// Close Free and close XDS Server connection
func (xs *XdsServer) Close() error {
	xs.Connected = false
	xs.Disabled = true
	xs.ioSock = nil
	xs._NotifyState()
	return nil
}

// Connect Establish HTTP connection with XDS Server
func (xs *XdsServer) Connect() error {
	var err error
	var retry int

	xs.Disabled = false
	xs.Connected = false

	err = nil
	for retry = xs.ConnRetry; retry > 0; retry-- {
		if err = xs._CreateConnectHTTP(); err == nil {
			break
		}
		if retry == xs.ConnRetry {
			// Notify only on the first conn error
			// doing that avoid 2 notifs (conn false; conn true) on startup
			xs._NotifyState()
		}
		xs.Log.Infof("Establishing connection to XDS Server (retry %d/%d)", retry, xs.ConnRetry)
		time.Sleep(time.Second)
	}
	if retry == 0 {
		// FIXME: re-use _reconnect to wait longer in background
		return fmt.Errorf("Connection to XDS Server failure")
	}
	if err != nil {
		return err
	}

	// Check HTTP connection and establish WS connection
	err = xs._connect(false)

	return err
}

// IsTempoID returns true when server as a temporary id
func (xs *XdsServer) IsTempoID() bool {
	return strings.HasPrefix(xs.ID, _IDTempoPrefix)
}

// SetLoggerOutput Set logger ou
func (xs *XdsServer) SetLoggerOutput(out io.Writer) {
	xs.logOut = out
}

// SendCommand Send a command to XDS Server
func (xs *XdsServer) SendCommand(cmd string, body []byte) (*http.Response, error) {
	url := cmd
	if !strings.HasPrefix("/", cmd) {
		url = "/" + cmd
	}
	return xs.client.HTTPPostWithRes(url, string(body))
}

// GetVersion Send Get request to retrieve XDS Server version
func (xs *XdsServer) GetVersion(res interface{}) error {
	return xs._HTTPGet("/version", &res)
}

// GetFolders Send GET request to get current folder configuration
func (xs *XdsServer) GetFolders(folders *[]XdsFolderConfig) error {
	return xs._HTTPGet("/folders", folders)
}

// FolderAdd Send POST request to add a folder
func (xs *XdsServer) FolderAdd(fld *XdsFolderConfig, res interface{}) error {
	response, err := xs._HTTPPost("/folder", fld)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("FolderAdd error status=%s", response.Status)
	}
	// Result is a XdsFolderConfig that is equivalent to ProjectConfig
	err = json.Unmarshal(xs.client.ResponseToBArray(response), res)

	return err
}

// FolderDelete Send DELETE request to delete a folder
func (xs *XdsServer) FolderDelete(id string) error {
	return xs.client.HTTPDelete("/folder/" + id)
}

// FolderSync Send POST request to force synchronization of a folder
func (xs *XdsServer) FolderSync(id string) error {
	return xs.client.HTTPPost("/folder/sync/"+id, "")
}

// SetAPIRouterGroup .
func (xs *XdsServer) SetAPIRouterGroup(r *gin.RouterGroup) {
	xs.apiRouter = r
}

// PassthroughGet Used to declare a route that sends directly a GET request to XDS Server
func (xs *XdsServer) PassthroughGet(url string) {
	if xs.apiRouter == nil {
		xs.Log.Errorf("apiRouter not set !")
		return
	}

	xs.apiRouter.GET(url, func(c *gin.Context) {
		var data interface{}
		if err := xs._HTTPGet(url, &data); err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				xs.Connected = false
				xs._NotifyState()
			}
			common.APIError(c, err.Error())
			return
		}

		c.JSON(http.StatusOK, data)
	})
}

// PassthroughPost Used to declare a route that sends directly a POST request to XDS Server
func (xs *XdsServer) PassthroughPost(url string) {
	if xs.apiRouter == nil {
		xs.Log.Errorf("apiRouter not set !")
		return
	}

	xs.apiRouter.POST(url, func(c *gin.Context) {
		bodyReq := []byte{}
		n, err := c.Request.Body.Read(bodyReq)
		if err != nil {
			common.APIError(c, err.Error())
			return
		}

		response, err := xs._HTTPPost(url, bodyReq[:n])
		if err != nil {
			common.APIError(c, err.Error())
			return
		}
		bodyRes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			common.APIError(c, "Cannot read response body")
			return
		}
		c.JSON(http.StatusOK, string(bodyRes))
	})
}

// EventRegister Post a request to register to an XdsServer event
func (xs *XdsServer) EventRegister(evName string, id string) error {
	var err error
	_, err = xs._HTTPPost("/events/register", XdsEventRegisterArgs{
		Name:      evName,
		ProjectID: id,
	})
	return err
}

// EventOn Register a callback on events reception
func (xs *XdsServer) EventOn(evName string, f func(interface{})) (uuid.UUID, error) {
	if xs.ioSock == nil {
		return uuid.Nil, fmt.Errorf("Io.Socket not initialized")
	}

	xs.sockEventsLock.Lock()
	defer xs.sockEventsLock.Unlock()

	if _, exist := xs.sockEvents[evName]; !exist {
		// Register listener only the first time
		evn := evName

		// FIXME: use generic type: data interface{} instead of data XdsEventFolderChange
		var err error
		if evName == "event:FolderStateChanged" {
			err = xs.ioSock.On(evn, func(data XdsEventFolderChange) error {
				xs.sockEventsLock.Lock()
				defer xs.sockEventsLock.Unlock()
				for _, c := range xs.sockEvents[evn] {
					c.Func(data)
				}
				return nil
			})
		} else {
			err = xs.ioSock.On(evn, f)
		}
		if err != nil {
			return uuid.Nil, err
		}
	}

	c := &caller{
		id:        uuid.NewV1(),
		EventName: evName,
		Func:      f,
	}

	xs.sockEvents[evName] = append(xs.sockEvents[evName], c)
	return c.id, nil
}

// EventOff Un-register a (or all) callbacks associated to an event
func (xs *XdsServer) EventOff(evName string, id uuid.UUID) error {
	xs.sockEventsLock.Lock()
	defer xs.sockEventsLock.Unlock()
	if _, exist := xs.sockEvents[evName]; exist {
		if id == uuid.Nil {
			// Un-register all
			xs.sockEvents[evName] = []*caller{}
		} else {
			// Un-register only the specified callback
			for i, ff := range xs.sockEvents[evName] {
				if uuid.Equal(ff.id, id) {
					xs.sockEvents[evName] = append(xs.sockEvents[evName][:i], xs.sockEvents[evName][i+1:]...)
					break
				}
			}
		}
	}
	return nil
}

// ProjectToFolder Convert Project structure to Folder structure
func (xs *XdsServer) ProjectToFolder(pPrj ProjectConfig) *XdsFolderConfig {
	stID := ""
	if pPrj.Type == XdsTypeCloudSync {
		stID, _ = xs.SThg.IDGet()
	}
	fPrj := XdsFolderConfig{
		ID:         pPrj.ID,
		Label:      pPrj.Label,
		ClientPath: pPrj.ClientPath,
		Type:       XdsFolderType(pPrj.Type),
		Status:     pPrj.Status,
		IsInSync:   pPrj.IsInSync,
		DefaultSdk: pPrj.DefaultSdk,
		DataPathMap: XdsPathMapConfig{
			ServerPath: pPrj.ServerPath,
		},
		DataCloudSync: XdsCloudSyncConfig{
			SyncThingID:   stID,
			STLocIsInSync: pPrj.IsInSync,
			STLocStatus:   pPrj.Status,
			STSvrIsInSync: pPrj.IsInSync,
			STSvrStatus:   pPrj.Status,
		},
	}

	return &fPrj
}

// FolderToProject Convert Folder structure to Project structure
func (xs *XdsServer) FolderToProject(fPrj XdsFolderConfig) ProjectConfig {
	inSync := fPrj.IsInSync
	sts := fPrj.Status

	if fPrj.Type == XdsTypeCloudSync {
		inSync = fPrj.DataCloudSync.STSvrIsInSync && fPrj.DataCloudSync.STLocIsInSync

		sts = fPrj.DataCloudSync.STSvrStatus
		switch fPrj.DataCloudSync.STLocStatus {
		case StatusErrorConfig, StatusDisable, StatusPause:
			sts = fPrj.DataCloudSync.STLocStatus
			break
		case StatusSyncing:
			if sts != StatusErrorConfig && sts != StatusDisable && sts != StatusPause {
				sts = StatusSyncing
			}
			break
		case StatusEnable:
			// keep STSvrStatus
			break
		}
	}

	pPrj := ProjectConfig{
		ID:         fPrj.ID,
		ServerID:   xs.ID,
		Label:      fPrj.Label,
		ClientPath: fPrj.ClientPath,
		ServerPath: fPrj.DataPathMap.ServerPath,
		Type:       ProjectType(fPrj.Type),
		Status:     sts,
		IsInSync:   inSync,
		DefaultSdk: fPrj.DefaultSdk,
	}
	return pPrj
}

/***
** Private functions
***/

// Create HTTP client
func (xs *XdsServer) _CreateConnectHTTP() error {
	var err error
	xs.client, err = common.HTTPNewClient(xs.BaseURL,
		common.HTTPClientConfig{
			URLPrefix:           "/api/v1",
			HeaderClientKeyName: "Xds-Sid",
			CsrfDisable:         true,
			LogOut:              xs.logOut,
			LogPrefix:           "XDSSERVER: ",
			LogLevel:            common.HTTPLogLevelWarning,
		})

	xs.client.SetLogLevel(xs.Log.Level.String())

	if err != nil {
		msg := ": " + err.Error()
		if strings.Contains(err.Error(), "connection refused") {
			msg = fmt.Sprintf("(url: %s)", xs.BaseURL)
		}
		return fmt.Errorf("ERROR: cannot connect to XDS Server %s", msg)
	}
	if xs.client == nil {
		return fmt.Errorf("ERROR: cannot connect to XDS Server (null client)")
	}

	return nil
}

// _HTTPGet .
func (xs *XdsServer) _HTTPGet(url string, data interface{}) error {
	var dd []byte
	if err := xs.client.HTTPGet(url, &dd); err != nil {
		return err
	}
	return json.Unmarshal(dd, &data)
}

// _HTTPPost .
func (xs *XdsServer) _HTTPPost(url string, data interface{}) (*http.Response, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return xs.client.HTTPPostWithRes(url, string(body))
}

//  Re-established connection
func (xs *XdsServer) _reconnect() error {
	err := xs._connect(true)
	if err == nil {
		// Reload projects list for this server
		err = xs.projects.Init(xs)
	}
	return err
}

//  Established HTTP and WS connection and retrieve XDSServer config
func (xs *XdsServer) _connect(reConn bool) error {

	xdsCfg := XdsServerConfig{}
	if err := xs._HTTPGet("/config", &xdsCfg); err != nil {
		xs.Connected = false
		if !reConn {
			xs._NotifyState()
		}
		return err
	}

	if reConn && xs.ID != xdsCfg.ID {
		xs.Log.Warningf("Reconnected to server but ID differs: old=%s, new=%s", xs.ID, xdsCfg.ID)
	}

	// Update local XDS config
	xs.ID = xdsCfg.ID
	xs.ServerConfig = &xdsCfg

	// Establish WS connection and register listen
	if err := xs._SocketConnect(); err != nil {
		xs.Connected = false
		xs._NotifyState()
		return err
	}

	xs.Connected = true
	xs._NotifyState()
	return nil
}

// Create WebSocket (io.socket) connection
func (xs *XdsServer) _SocketConnect() error {

	xs.Log.Infof("Connecting IO.socket for server %s (url %s)", xs.ID, xs.BaseURL)

	opts := &sio_client.Options{
		Transport: "websocket",
		Header:    make(map[string][]string),
	}
	opts.Header["XDS-SID"] = []string{xs.client.GetClientID()}

	iosk, err := sio_client.NewClient(xs.BaseURL, opts)
	if err != nil {
		return fmt.Errorf("IO.socket connection error for server %s: %v", xs.ID, err)
	}
	xs.ioSock = iosk

	// Register some listeners

	iosk.On("error", func(err error) {
		xs.Log.Infof("IO.socket Error server %s; err: %v", xs.ID, err)
		if xs.CBOnError != nil {
			xs.CBOnError(err)
		}
	})

	iosk.On("disconnection", func(err error) {
		xs.Log.Infof("IO.socket disconnection server %s", xs.ID)
		if xs.CBOnDisconnect != nil {
			xs.CBOnDisconnect(err)
		}
		xs.Connected = false
		xs._NotifyState()

		// Try to reconnect during 15min (or at least while not disabled)
		go func() {
			count := 0
			waitTime := 1
			for !xs.Disabled && !xs.Connected {
				count++
				if count%60 == 0 {
					waitTime *= 5
				}
				if waitTime > 15*60 {
					xs.Log.Infof("Stop reconnection to server url=%s id=%s !", xs.BaseURL, xs.ID)
					return
				}
				time.Sleep(time.Second * time.Duration(waitTime))
				xs.Log.Infof("Try to reconnect to server %s (%d)", xs.BaseURL, count)

				xs._reconnect()
			}
		}()
	})

	// XXX - There is no connection event generated so, just consider that
	// we are connected when NewClient return successfully
	/* iosk.On("connection", func() { ... }) */
	xs.Log.Infof("IO.socket connected server url=%s id=%s", xs.BaseURL, xs.ID)

	return nil
}

// Send event to notify changes
func (xs *XdsServer) _NotifyState() {

	evSts := ServerCfg{
		ID:         xs.ID,
		URL:        xs.BaseURL,
		APIURL:     xs.APIURL,
		PartialURL: xs.PartialURL,
		ConnRetry:  xs.ConnRetry,
		Connected:  xs.Connected,
	}
	if err := xs.events.Emit(EVTServerConfig, evSts); err != nil {
		xs.Log.Warningf("Cannot notify XdsServer state change: %v", err)
	}
}
