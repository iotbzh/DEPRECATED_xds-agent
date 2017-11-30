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
