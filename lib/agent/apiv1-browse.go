package agent

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type directory struct {
	Name     string `json:"name"`
	Fullpath string `json:"fullpath"`
}

type apiDirectory struct {
	Dir []directory `json:"dir"`
}

// browseFS used to browse local file system
func (s *APIService) browseFS(c *gin.Context) {

	response := apiDirectory{
		Dir: []directory{
			directory{Name: "TODO SEB"},
		},
	}

	c.JSON(http.StatusOK, response)
}
