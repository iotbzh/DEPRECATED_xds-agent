import { Component, OnInit, ViewChild, ViewEncapsulation } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { CollapseModule } from 'ngx-bootstrap/collapse';

import { ConfigService, IConfig } from '../services/config.service';
import { ProjectService, IProject } from '../services/project.service';
import { XDSAgentService, IAgentStatus, IXDSConfig } from '../services/xdsagent.service';
import { AlertService } from '../services/alert.service';
import { ProjectAddModalComponent } from '../projects/projectAddModal.component';
import { SdkService, ISdk } from '../services/sdk.service';
import { SdkAddModalComponent } from '../sdks/sdkAddModal.component';

@Component({
  selector: 'app-config',
  templateUrl: './config.component.html',
  styleUrls: ['./config.component.css'],
  encapsulation: ViewEncapsulation.None
})

// Inspired from https://embed.plnkr.co/jgDTXknPzAaqcg9XA9zq/
// and from http://plnkr.co/edit/vCdjZM?p=preview

export class ConfigComponent implements OnInit {
    @ViewChild('childProjectModal') childProjectModal: ProjectAddModalComponent;
    @ViewChild('childSdkModal') childSdkModal: SdkAddModalComponent;

    config$: Observable<IConfig>;
    projects$: Observable<IProject[]>;
    sdks$: Observable<ISdk[]>;
    agentStatus$: Observable<IAgentStatus>;

    curProj: number;
    curServer: number;
    curServerID: string;
    userEditedLabel = false;

    gConfigIsCollapsed = true;
    sdksIsCollapsed = true;
    projectsIsCollapsed = false;

    // TODO replace by reactive FormControl + add validation
    xdsServerConnected = false;
    xdsServerUrl: string;
    xdsServerRetry: string;
    projectsRootDir: string;    // FIXME: should be remove when projectAddModal will always return full path
    showApplyBtn = {    // Used to show/hide Apply buttons
        'retry': false,
        'rootDir': false,
    };

    constructor(
        private configSvr: ConfigService,
        private projectSvr: ProjectService,
        private xdsAgentSvr: XDSAgentService,
        private sdkSvr: SdkService,
        private alert: AlertService,
    ) {
    }

    ngOnInit() {
        this.config$ = this.configSvr.Conf$;
        this.projects$ = this.projectSvr.Projects$;
        this.sdks$ = this.sdkSvr.Sdks$;
        this.agentStatus$ = this.xdsAgentSvr.Status$;

        // FIXME support multiple servers
        this.curServer = 0;

        // Bind xdsServerUrl to baseURL
        this.xdsAgentSvr.XdsConfig$.subscribe(cfg => {
            if (!cfg || cfg.servers.length < 1) {
                return;
            }
            const svr = cfg.servers[this.curServer];
            this.curServerID = svr.id;
            this.xdsServerConnected = svr.connected;
            this.xdsServerUrl = svr.url;
            this.xdsServerRetry = String(svr.connRetry);
            this.projectsRootDir = ''; // SEB FIXME: add in go config? cfg.projectsRootDir;
        });
    }

    submitGlobConf(field: string) {
        switch (field) {
            case 'retry':
                const re = new RegExp('^[0-9]+$');
                const rr = parseInt(this.xdsServerRetry, 10);
                if (re.test(this.xdsServerRetry) && rr >= 0) {
                    this.xdsAgentSvr.setServerRetry(this.curServerID, rr);
                } else {
                    this.alert.warning('Not a valid number', true);
                }
                break;
            case 'rootDir':
                this.configSvr.projectsRootDir = this.projectsRootDir;
                break;
            default:
                return;
        }
        this.showApplyBtn[field] = false;
    }

    xdsAgentRestartConn() {
        const url = this.xdsServerUrl;
        this.xdsAgentSvr.setServerUrl(this.curServerID, url);
    }

}
