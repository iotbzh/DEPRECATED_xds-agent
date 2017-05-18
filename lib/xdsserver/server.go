package xdsserver

import (
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"
	"github.com/googollee/go-socket.io"
	"github.com/iotbzh/xds-agent/lib/apiv1"
	"github.com/iotbzh/xds-agent/lib/session"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
)

// ServerService .
type ServerService struct {
	router    *gin.Engine
	api       *apiv1.APIService
	sIOServer *socketio.Server
	webApp    *gin.RouterGroup
	cfg       *xdsconfig.Config
	sessions  *session.Sessions
	log       *logrus.Logger
	stop      chan struct{} // signals intentional stop
}

const indexFilename = "index.html"
const cookieMaxAge = "3600"

// NewServer creates an instance of ServerService
func NewServer(conf *xdsconfig.Config, log *logrus.Logger) *ServerService {

	// Setup logging for gin router
	if log.Level == logrus.DebugLevel {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// TODO
	//  - try to bind gin DefaultWriter & DefaultErrorWriter to logrus logger
	//  - try to fix pb about isTerminal=false when out is in VSC Debug Console
	//gin.DefaultWriter = ??
	//gin.DefaultErrorWriter = ??

	// Creates gin router
	r := gin.New()

	svr := &ServerService{
		router:    r,
		api:       nil,
		sIOServer: nil,
		webApp:    nil,
		cfg:       conf,
		log:       log,
		sessions:  nil,
		stop:      make(chan struct{}),
	}

	return svr
}

// Serve starts a new instance of the Web Server
func (s *ServerService) Serve() error {
	var err error

	// Setup middlewares
	s.router.Use(gin.Logger())
	s.router.Use(gin.Recovery())
	s.router.Use(s.middlewareXDSDetails())
	s.router.Use(s.middlewareCORS())

	// Sessions manager
	s.sessions = session.NewClientSessions(s.router, s.log, cookieMaxAge)

	// Create REST API
	s.api = apiv1.New(s.sessions, s.cfg, s.log, s.router)

	// Websocket routes
	s.sIOServer, err = socketio.NewServer(nil)
	if err != nil {
		s.log.Fatalln(err)
	}

	s.router.GET("/socket.io/", s.socketHandler)
	s.router.POST("/socket.io/", s.socketHandler)
	/* TODO: do we want to support ws://...  ?
	s.router.Handle("WS", "/socket.io/", s.socketHandler)
	s.router.Handle("WSS", "/socket.io/", s.socketHandler)
	*/

	// Serve in the background
	serveError := make(chan error, 1)
	go func() {
		fmt.Printf("Web Server running on localhost:%s ...\n", s.cfg.HTTPPort)
		serveError <- http.ListenAndServe(":"+s.cfg.HTTPPort, s.router)
	}()

	fmt.Printf("XDS agent running...\n")

	// Wait for stop, restart or error signals
	select {
	case <-s.stop:
		// Shutting down permanently
		s.sessions.Stop()
		s.log.Infoln("shutting down (stop)")
	case err = <-serveError:
		// Error due to listen/serve failure
		s.log.Errorln(err)
	}

	return nil
}

// Stop web server
func (s *ServerService) Stop() {
	close(s.stop)
}

// Add details in Header
func (s *ServerService) middlewareXDSDetails() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("XDS-Agent-Version", s.cfg.Version)
		c.Header("XDS-API-Version", s.cfg.APIVersion)
		c.Next()
	}
}

// CORS middleware
func (s *ServerService) middlewareCORS() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", "Content-Type")
			c.Header("Access-Control-Allow-Methods", "POST, DELETE, GET, PUT")
			c.Header("Content-Type", "application/json")
			c.Header("Access-Control-Max-Age", cookieMaxAge)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// socketHandler is the handler for the "main" websocket connection
func (s *ServerService) socketHandler(c *gin.Context) {

	// Retrieve user session
	sess := s.sessions.Get(c)
	if sess == nil {
		c.JSON(500, gin.H{"error": "Cannot retrieve session"})
		return
	}

	s.sIOServer.On("connection", func(so socketio.Socket) {
		s.log.Debugf("WS Connected (SID=%v)", so.Id())
		s.sessions.UpdateIOSocket(sess.ID, &so)

		so.On("disconnection", func() {
			s.log.Debugf("WS disconnected (SID=%v)", so.Id())
			s.sessions.UpdateIOSocket(sess.ID, nil)
		})
	})

	s.sIOServer.On("error", func(so socketio.Socket, err error) {
		s.log.Errorf("WS SID=%v Error : %v", so.Id(), err.Error())
	})

	s.sIOServer.ServeHTTP(c.Writer, c.Request)
}
