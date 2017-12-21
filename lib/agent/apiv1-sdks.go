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

	"github.com/iotbzh/xds-agent/lib/xaapiv1"
	"github.com/iotbzh/xds-server/lib/xsapiv1"
)

// sdksPassthroughInit Declare passthrough routes for sdks
func (s *APIService) sdksPassthroughInit(svr *XdsServer) error {
	svr.PassthroughGet("/sdks")
	svr.PassthroughGet("/sdks/:id")
	svr.PassthroughPost("/sdks")
	svr.PassthroughPost("/sdks/abortinstall")
	svr.PassthroughDelete("/sdks/:id")

	return nil
}

// sdksEventsForwardInit Register events forwarder for sdks
func (s *APIService) sdksEventsForwardInit(svr *XdsServer) error {

	if !svr.Connected {
		return fmt.Errorf("Cannot register events: XDS Server %v not connected", svr.ID)
	}

	// Forward SDK events from XDS-server to client
	if _, err := svr.EventOn(xsapiv1.EVTSDKInstall, "", s._sdkEventInstallCB); err != nil {
		s.Log.Errorf("XDS Server EventOn '%s' failed: %v", xsapiv1.EVTSDKInstall, err)
		return err
	}

	if _, err := svr.EventOn(xsapiv1.EVTSDKRemove, "", s._sdkEventRemoveCB); err != nil {
		s.Log.Errorf("XDS Server EventOn '%s' failed: %v", xsapiv1.EVTSDKRemove, err)
		return err
	}

	return nil
}

func (s *APIService) _sdkEventInstallCB(privD interface{}, data interface{}) error {
	// assume that xsapiv1.SDKManagementMsg == xaapiv1.SDKManagementMsg
	evt := xaapiv1.SDKManagementMsg{}
	evtName := xaapiv1.EVTSDKInstall
	d, err := json.Marshal(data)
	if err != nil {
		s.Log.Errorf("Cannot marshal XDS Server %s: err=%v", evtName, err)
		return err
	}
	if err = json.Unmarshal(d, &evt); err != nil {
		s.Log.Errorf("Cannot unmarshal XDS Server %s: err=%v", evtName, err)
		return err
	}

	if err := s.events.Emit(evtName, evt, ""); err != nil {
		s.Log.Warningf("Cannot notify %s (from server): %v", evtName, err)
		return err
	}
	return nil
}

func (s *APIService) _sdkEventRemoveCB(privD interface{}, data interface{}) error {
	// assume that xsapiv1.SDKManagementMsg == xaapiv1.SDKManagementMsg
	evt := xaapiv1.SDKManagementMsg{}
	evtName := xaapiv1.EVTSDKRemove
	d, err := json.Marshal(data)
	if err != nil {
		s.Log.Errorf("Cannot marshal XDS Server %s: err=%v", evtName, err)
		return err
	}
	if err = json.Unmarshal(d, &evt); err != nil {
		s.Log.Errorf("Cannot unmarshal XDS Server %s: err=%v", evtName, err)
		return err
	}

	if err := s.events.Emit(evtName, evt, ""); err != nil {
		s.Log.Warningf("Cannot notify %s (from server): %v", evtName, err)
		return err
	}
	return nil
}
