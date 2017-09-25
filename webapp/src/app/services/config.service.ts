import { Injectable, OnInit } from '@angular/core';
import { Http, Headers, RequestOptionsArgs, Response } from '@angular/http';
import { Location } from '@angular/common';
import { CookieService } from 'ngx-cookie';
import { Observable } from 'rxjs/Observable';
import { Subscriber } from 'rxjs/Subscriber';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';
import 'rxjs/add/operator/mergeMap';


import { XDSAgentService, IXDSProjectConfig } from "../services/xdsagent.service";
import { AlertService, IAlert } from "../services/alert.service";
import { UtilsService } from "../services/utils.service";

export interface IConfig {
    projectsRootDir: string;
    //SEB projects: IProject[];
}

@Injectable()
export class ConfigService {

    public Conf$: Observable<IConfig>;

    private confSubject: BehaviorSubject<IConfig>;
    private confStore: IConfig;
    // SEB cleanup private AgentConnectObs = null;
    // SEB cleanup private stConnectObs = null;

    constructor(private _window: Window,
        private cookie: CookieService,
        private xdsAgentSvr: XDSAgentService,
        private alert: AlertService,
        private utils: UtilsService,
    ) {
        this.load();
        this.confSubject = <BehaviorSubject<IConfig>>new BehaviorSubject(this.confStore);
        this.Conf$ = this.confSubject.asObservable();

        // force to load projects
        this.loadProjects();
    }

    // Load config
    load() {
        // Try to retrieve previous config from cookie
        let cookConf = this.cookie.getObject("xds-config");
        if (cookConf != null) {
            this.confStore = <IConfig>cookConf;
        } else {
            // Set default config
            this.confStore = {
                projectsRootDir: "",
                //projects: []
            };
        }
    }

    // Save config into cookie
    save() {
        // Notify subscribers
        this.confSubject.next(Object.assign({}, this.confStore));

        // Don't save projects in cookies (too big!)
        let cfg = Object.assign({}, this.confStore);
        this.cookie.putObject("xds-config", cfg);
    }

    loadProjects() {
        /* SEB
        // Setup connection with local XDS agent
        if (this.AgentConnectObs) {
            try {
                this.AgentConnectObs.unsubscribe();
            } catch (err) { }
            this.AgentConnectObs = null;
        }

        let cfg = this.confStore.xdsAgent;
        this.AgentConnectObs = this.xdsAgentSvr.connect(cfg.retry, cfg.URL)
            .subscribe((sts) => {
                //console.log("Agent sts", sts);
                // FIXME: load projects from local XDS Agent and
                //  not directly from local syncthing
                this._loadProjectFromLocalST();

            }, error => {
                if (error.indexOf("XDS local Agent not responding") !== -1) {
                    let url_OS_Linux = "https://en.opensuse.org/LinuxAutomotive#Installation_AGL_XDS";
                    let url_OS_Other = "https://github.com/iotbzh/xds-agent#how-to-install-on-other-platform";
                    let msg = `<span><strong>` + error + `<br></strong>
                    You may need to install and execute XDS-Agent: <br>
                        On Linux machine <a href="` + url_OS_Linux + `" target="_blank"><span
                            class="fa fa-external-link"></span></a>
                        <br>
                        On Windows machine <a href="` + url_OS_Other + `" target="_blank"><span
                            class="fa fa-external-link"></span></a>
                        <br>
                        On MacOS machine <a href="` + url_OS_Other + `" target="_blank"><span
                            class="fa fa-external-link"></span></a>
                    `;
                    this.alert.error(msg);
                } else {
                    this.alert.error(error);
                }
            });
        */
    }

    /* SEB
    private _loadProjectFromLocalST() {
        // Remove previous subscriber if existing
        if (this.stConnectObs) {
            try {
                this.stConnectObs.unsubscribe();
            } catch (err) { }
            this.stConnectObs = null;
        }

        // FIXME: move this code and all logic about syncthing inside XDS Agent
        // Setup connection with local SyncThing
        let retry = this.confStore.localSThg.retry;
        let url = this.confStore.localSThg.URL;
        this.stConnectObs = this.stSvr.connect(retry, url).subscribe((sts) => {
            this.confStore.localSThg.ID = sts.ID;
            this.confStore.localSThg.tilde = sts.tilde;
            if (this.confStore.projectsRootDir === "") {
                this.confStore.projectsRootDir = sts.tilde;
            }

            // Rebuild projects definition from local and remote syncthing
            this.confStore.projects = [];

            this.xdsServerSvr.getProjects().subscribe(remotePrj => {
                this.stSvr.getProjects().subscribe(localPrj => {
                    remotePrj.forEach(rPrj => {
                        let lPrj = localPrj.filter(item => item.id === rPrj.id);
                        if (lPrj.length > 0 || rPrj.type === ProjectType.NATIVE_PATHMAP) {
                            this._addProject(rPrj, true);
                        }
                    });
                    this.confSubject.next(Object.assign({}, this.confStore));
                }), error => this.alert.error('Could not load initial state of local projects.');
            }), error => this.alert.error('Could not load initial state of remote projects.');

        }, error => {
            if (error.indexOf("Syncthing local daemon not responding") !== -1) {
                let msg = "<span><strong>" + error + "<br></strong>";
                msg += "Please check that local XDS-Agent is running.<br>";
                msg += "</span>";
                this.alert.error(msg);
            } else {
                this.alert.error(error);
            }
        });
    }

    set syncToolURL(url: string) {
        this.confStore.localSThg.URL = url;
        this.save();
    }
    */

    set projectsRootDir(p: string) {
        /* SEB
        if (p.charAt(0) === '~') {
            p = this.confStore.localSThg.tilde + p.substring(1);
        }
        */
        this.confStore.projectsRootDir = p;
        this.save();
    }
}
