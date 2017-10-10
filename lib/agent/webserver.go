package agent

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
)

// WebServer .
type WebServer struct {
	*Context
	router    *gin.Engine
	api       *APIService
	sIOServer *socketio.Server
	webApp    *gin.RouterGroup
	stop      chan struct{} // signals intentional stop
}

const indexFilename = "index.html"

// NewWebServer creates an instance of WebServer
func NewWebServer(ctx *Context) *WebServer {

	// Setup logging for gin router
	if ctx.Log.Level == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Redirect gin logs into another logger (LogVerboseOut may be stderr or a file)
	gin.DefaultWriter = ctx.Config.LogVerboseOut
	gin.DefaultErrorWriter = ctx.Config.LogVerboseOut
	log.SetOutput(ctx.Config.LogVerboseOut)

	// Creates gin router
	r := gin.New()

	svr := &WebServer{
		Context:   ctx,
		router:    r,
		api:       nil,
		sIOServer: nil,
		webApp:    nil,
		stop:      make(chan struct{}),
	}

	return svr
}

// Serve starts a new instance of the Web Server
func (s *WebServer) Serve() error {
	var err error

	// Setup middlewares
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(s.middlewareCORS())
	s.router.Use(s.middlewareXDSDetails())
	s.router.Use(s.middlewareCSRF())

	// Create REST API
	s.api = NewAPIV1(s.Context)

	// Create connections to XDS Servers
	// XXX - not sure there is no side effect to do it in background !
	go func() {
		for _, svrCfg := range s.Config.FileConf.ServersConf {
			if svr, err := s.api.AddXdsServer(svrCfg); err != nil {
				// Just log error, don't consider as critical
				s.Log.Infof("Cannot connect to XDS Server url=%s: %v", svr.BaseURL, err.Error())
			}
		}
	}()

	// Websocket routes
	s.sIOServer, err = socketio.NewServer(nil)
	if err != nil {
		s.Log.Fatalln(err)
	}

	s.router.GET("/socket.io/", s.socketHandler)
	s.router.POST("/socket.io/", s.socketHandler)
	/* TODO: do we want to support ws://...  ?
	s.router.Handle("WS", "/socket.io/", s.socketHandler)
	s.router.Handle("WSS", "/socket.io/", s.socketHandler)
	*/

	// Web Application (serve on / )
	idxFile := path.Join(s.Config.FileConf.WebAppDir, indexFilename)
	if _, err := os.Stat(idxFile); err != nil {
		s.Log.Fatalln("Web app directory not found, check/use webAppDir setting in config file: ", idxFile)
	}
	s.Log.Infof("Serve WEB app dir: %s", s.Config.FileConf.WebAppDir)
	s.router.Use(static.Serve("/", static.LocalFile(s.Config.FileConf.WebAppDir, true)))
	s.webApp = s.router.Group("/", s.serveIndexFile)
	{
		s.webApp.GET("/")
	}

	// Serve in the background
	serveError := make(chan error, 1)
	go func() {
		fmt.Printf("Web Server running on localhost:%s ...\n", s.Config.FileConf.HTTPPort)
		serveError <- http.ListenAndServe(":"+s.Config.FileConf.HTTPPort, s.router)
	}()

	fmt.Printf("XDS agent running...\n")

	// Wait for stop, restart or error signals
	select {
	case <-s.stop:
		// Shutting down permanently
		s.sessions.Stop()
		s.Log.Infoln("shutting down (stop)")
	case err = <-serveError:
		// Error due to listen/serve failure
		s.Log.Errorln(err)
	}

	return nil
}

// Stop web server
func (s *WebServer) Stop() {
	s.api.Stop()
	close(s.stop)
}

// serveIndexFile provides initial file (eg. index.html) of webapp
func (s *WebServer) serveIndexFile(c *gin.Context) {
	c.HTML(200, indexFilename, gin.H{})
}

// Add details in Header
func (s *WebServer) middlewareXDSDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("XDS-Agent-Version", s.Config.Version)
		c.Header("XDS-API-Version", s.Config.APIVersion)
		c.Next()
	}
}

func (s *WebServer) isValidAPIKey(key string) bool {
	return (s.Config.FileConf.XDSAPIKey != "" && key == s.Config.FileConf.XDSAPIKey)
}

func (s *WebServer) middlewareCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XXX - not used for now
		c.Next()
		return
		/*
			// Allow requests carrying a valid API key
			if s.isValidAPIKey(c.Request.Header.Get("X-API-Key")) {
				// Set the access-control-allow-origin header for CORS requests
				// since a valid API key has been provided
				c.Header("Access-Control-Allow-Origin", "*")
				c.Next()
				return
			}

			// Allow io.socket request
			if strings.HasPrefix(c.Request.URL.Path, "/socket.io") {
				c.Next()
				return
			}

			// FIXME Add really CSRF support

			// Allow requests for anything not under the protected path prefix,
			// and set a CSRF cookie if there isn't already a valid one.
			//if !strings.HasPrefix(c.Request.URL.Path, prefix) {
			//	cookie, err := c.Cookie("CSRF-Token-" + unique)
			//	if err != nil || !validCsrfToken(cookie.Value) {
			//		s.Log.Debugln("new CSRF cookie in response to request for", c.Request.URL)
			//		c.SetCookie("CSRF-Token-"+unique, newCsrfToken(), 600, "/", "", false, false)
			//	}
			//	c.Next()
			//	return
			//}

			// Verify the CSRF token
			//token := c.Request.Header.Get("X-CSRF-Token-" + unique)
			//if !validCsrfToken(token) {
			//	c.AbortWithError(403, "CSRF Error")
			//	return
			//}

			//c.Next()

			c.AbortWithError(403, fmt.Errorf("Not valid API key"))
		*/
	}
}

// CORS middleware
func (s *WebServer) middlewareCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-API-Key")
			c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE")
			c.Header("Access-Control-Max-Age", cookieMaxAge)
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// socketHandler is the handler for the "main" websocket connection
func (s *WebServer) socketHandler(c *gin.Context) {

	// Retrieve user session
	sess := s.sessions.Get(c)
	if sess == nil {
		c.JSON(500, gin.H{"error": "Cannot retrieve session"})
		return
	}

	s.sIOServer.On("connection", func(so socketio.Socket) {
		s.Log.Debugf("WS Connected (SID=%v)", so.Id())
		s.sessions.UpdateIOSocket(sess.ID, &so)

		so.On("disconnection", func() {
			s.Log.Debugf("WS disconnected (SID=%v)", so.Id())
			s.sessions.UpdateIOSocket(sess.ID, nil)
		})
	})

	s.sIOServer.On("error", func(so socketio.Socket, err error) {
		s.Log.Errorf("WS SID=%v Error : %v", so.Id(), err.Error())
	})

	s.sIOServer.ServeHTTP(c.Writer, c.Request)
}
