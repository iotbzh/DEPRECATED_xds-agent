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
  server: IXDServerCfg;

  configFormChanged = false;

  constructor(
    private XdsConfigSvr: XDSConfigService,
    private alert: AlertService,
  ) {
    // FIXME support multiple servers
    this.XdsConfigSvr.onCurServer().subscribe(svr => {
      this.xdsServerUrl = svr.url;
      this.server = Object.assign({}, svr);
    });
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

