/**
* @license
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

import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { XDSConfigService } from '../../../@core-xds/services/xds-config.service';
import { IXDServerCfg } from '../../../@core-xds/services/xdsagent.service';
import { AlertService, IAlert } from '../../../@core-xds/services/alert.service';
import { NotificationsComponent } from '../../notifications/notifications.component';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';

@Component({
  selector: 'xds-config-xds',
  styleUrls: ['./config-xds.component.scss'],
  templateUrl: './config-xds.component.html',
})
export class ConfigXdsComponent {

  // TODO: cleanup agentStatus$: Observable<IAgentStatus>;
  applying = false;
  xdsServerUrl = '';
  server: IXDServerCfg = { id: '', url: 'http://localhost:8000', connRetry: 10, connected: false };

  configFormChanged = false;

  constructor(
    private XdsConfigSvr: XDSConfigService,
    private alert: AlertService,
  ) {
    // FIXME support multiple servers
    this._updateServerCfg(this.XdsConfigSvr.getCurServer());
    this.XdsConfigSvr.onCurServer().subscribe(svr => this._updateServerCfg(svr));
  }

  private _updateServerCfg(svr: IXDServerCfg) {
    if (!svr || svr.url === '') {
      return;
    }
    this.xdsServerUrl = svr.url;
    this.server = Object.assign({}, svr);
  }

  isApplyBtnEnable(): boolean {
    return this.xdsServerUrl !== '' && (!this.server.connected || this.configFormChanged);
  }

  onSubmit() {
    if (!this.configFormChanged && this.server.connected) {
      return;
    }
    this.configFormChanged = false;
    this.applying = true;
    this.server.url = this.xdsServerUrl;
    this.XdsConfigSvr.setCurServer(this.server)
      .subscribe(cfg => {
        this.alert.info('XDS Server successfully connected (' + cfg.url + ')');
        this.server = Object.assign({}, cfg);
        this.applying = false;
      },
      err => {
        this.applying = false;
        this.alert.error(err);
      });
  }

}

