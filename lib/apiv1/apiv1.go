package apiv1

import (
	"github.com/Sirupsen/logrus"
	"github.com/gin-gonic/gin"

	"github.com/iotbzh/xds-agent/lib/session"
	"github.com/iotbzh/xds-agent/lib/xdsconfig"
)

// APIService .
type APIService struct {
	router    *gin.Engine
	apiRouter *gin.RouterGroup
	sessions  *session.Sessions
	cfg       *xdsconfig.Config
	log       *logrus.Logger
}

// New creates a new instance of API service
func New(sess *session.Sessions, conf *xdsconfig.Config, log *logrus.Logger, r *gin.Engine) *APIService {
	s := &APIService{
		router:    r,
		sessions:  sess,
		apiRouter: r.Group("/api/v1"),
		cfg:       conf,
		log:       log,
	}

	s.apiRouter.GET("/version", s.getVersion)

	s.apiRouter.GET("/config", s.getConfig)
	s.apiRouter.POST("/config", s.setConfig)

	return s
}
