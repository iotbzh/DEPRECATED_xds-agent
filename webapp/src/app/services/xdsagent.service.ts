import { Injectable } from '@angular/core';
import { Http, Headers, RequestOptionsArgs, Response } from '@angular/http';
import { Location } from '@angular/common';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';
import * as io from 'socket.io-client';

import { AlertService } from './alert.service';
import { ISdk } from './sdk.service';
import { ProjectType} from "./project.service";

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';
import 'rxjs/add/observable/throw';
import 'rxjs/add/operator/mergeMap';
import 'rxjs/add/observable/of';
import 'rxjs/add/operator/retryWhen';


export interface IXDSConfigProject {
    id: string;
    path: string;
    clientSyncThingID: string;
    type: string;
    label?: string;
    defaultSdkID?: string;
}

interface IXDSBuilderConfig {
    ip: string;
    port: string;
    syncThingID: string;
}

export interface IXDSProjectConfig {
    id: string;
    serverId: string;
    label: string;
    clientPath: string;
    serverPath?: string;
    type: ProjectType;
    status?: string;
    isInSync?: boolean;
    defaultSdkID: string;
}

export interface IXDSVer {
    id: string;
    version: string;
    apiVersion: string;
    gitTag: string;
}

export interface IXDSVersions {
    client: IXDSVer;
    servers: IXDSVer[];
}

export interface IXDServerCfg {
    id: string;
    url: string;
    apiUrl: string;
    partialUrl: string;
    connRetry: number;
    connected: boolean;
}

export interface IXDSConfig {
    servers: IXDServerCfg[];
}

export interface ISdkMessage {
    wsID: string;
    msgType: string;
    data: any;
}

export interface ICmdOutput {
    cmdID: string;
    timestamp: string;
    stdout: string;
    stderr: string;
}

export interface ICmdExit {
    cmdID: string;
    timestamp: string;
    code: number;
    error: string;
}

export interface IAgentStatus {
    WS_connected: boolean;
}


@Injectable()
export class XDSAgentService {

    public XdsConfig$: Observable<IXDSConfig>;
    public Status$: Observable<IAgentStatus>;
    public ProjectState$ = <Subject<IXDSProjectConfig>>new Subject();
    public CmdOutput$ = <Subject<ICmdOutput>>new Subject();
    public CmdExit$ = <Subject<ICmdExit>>new Subject();

    private baseUrl: string;
    private wsUrl: string;
    private _config = <IXDSConfig>{ servers: [] };
    private _status = { WS_connected: false };

    private configSubject = <BehaviorSubject<IXDSConfig>>new BehaviorSubject(this._config);
    private statusSubject = <BehaviorSubject<IAgentStatus>>new BehaviorSubject(this._status);

    private socket: SocketIOClient.Socket;

    constructor(private http: Http, private _window: Window, private alert: AlertService) {

        this.XdsConfig$ = this.configSubject.asObservable();
        this.Status$ = this.statusSubject.asObservable();

        this.baseUrl = this._window.location.origin + '/api/v1';

        let re = this._window.location.origin.match(/http[s]?:\/\/([^\/]*)[\/]?/);
        if (re === null || re.length < 2) {
            console.error('ERROR: cannot determine Websocket url');
        } else {
            this.wsUrl = 'ws://' + re[1];
            this._handleIoSocket();
            this._RegisterEvents();
        }
    }

    private _WSState(sts: boolean) {
        this._status.WS_connected = sts;
        this.statusSubject.next(Object.assign({}, this._status));

        // Update XDS config including XDS Server list when connected
        if (sts) {
            this.getConfig().subscribe(c => {
                this._config = c;
                this.configSubject.next(
                    Object.assign({ servers: [] }, this._config)
                );
            });
        }
    }

    private _handleIoSocket() {
        this.socket = io(this.wsUrl, { transports: ['websocket'] });

        this.socket.on('connect_error', (res) => {
            this._WSState(false);
            console.error('XDS Agent WebSocket Connection error !');
        });

        this.socket.on('connect', (res) => {
            this._WSState(true);
        });

        this.socket.on('disconnection', (res) => {
            this._WSState(false);
            this.alert.error('WS disconnection: ' + res);
        });

        this.socket.on('error', (err) => {
            console.error('WS error:', err);
        });

        this.socket.on('make:output', data => {
            this.CmdOutput$.next(Object.assign({}, <ICmdOutput>data));
        });

        this.socket.on('make:exit', data => {
            this.CmdExit$.next(Object.assign({}, <ICmdExit>data));
        });

        this.socket.on('exec:output', data => {
            this.CmdOutput$.next(Object.assign({}, <ICmdOutput>data));
        });

        this.socket.on('exec:exit', data => {
            this.CmdExit$.next(Object.assign({}, <ICmdExit>data));
        });

        // Events
        // (project-add and project-delete events are managed by project.service)
        this.socket.on('event:server-config', ev => {
            if (ev && ev.data) {
                let cfg: IXDServerCfg = ev.data;
                let idx = this._config.servers.findIndex(el => el.id === cfg.id);
                if (idx >= 0) {
                    this._config.servers[idx] = Object.assign({}, cfg);
                }
                this.configSubject.next(Object.assign({}, this._config));
            }
        });

        this.socket.on('event:project-state-change', ev => {
            if (ev && ev.data) {
                this.ProjectState$.next(Object.assign({}, ev.data));
            }
        });

    }

    /**
    ** Events
    ***/
    addEventListener(ev: string, fn: Function): SocketIOClient.Emitter {
        return this.socket.addEventListener(ev, fn);
    }

    /**
    ** Misc / Version
    ***/
    getVersion(): Observable<IXDSVersions> {
        return this._get('/version');
    }

    /***
    ** Config
    ***/
    getConfig(): Observable<IXDSConfig> {
        return this._get('/config');
    }

    setConfig(cfg: IXDSConfig): Observable<IXDSConfig> {
        return this._post('/config', cfg);
    }

    setServerRetry(serverID: string, r: number) {
        let svr = this._getServer(serverID);
        if (!svr) {
            return Observable.of([]);
        }

        svr.connRetry = r;
        this.setConfig(this._config).subscribe(
            newCfg => {
                this._config = newCfg;
                this.configSubject.next(Object.assign({}, this._config));
            },
            err => {
                this.alert.error(err);
            }
        );
    }

    setServerUrl(serverID: string, url: string) {
        let svr = this._getServer(serverID);
        if (!svr) {
            return Observable.of([]);
        }
        svr.url = url;
        this.setConfig(this._config).subscribe(
            newCfg => {
                this._config = newCfg;
                this.configSubject.next(Object.assign({}, this._config));
            },
            err => {
                this.alert.error(err);
            }
        );
    }

    /***
    ** SDKs
    ***/
    getSdks(serverID: string): Observable<ISdk[]> {
        let svr = this._getServer(serverID);
        if (!svr || !svr.connected) {
            return Observable.of([]);
        }

        return this._get(svr.partialUrl + '/sdks');
    }

    /***
    ** Projects
    ***/
    getProjects(): Observable<IXDSProjectConfig[]> {
        return this._get('/projects');
    }

    addProject(cfg: IXDSProjectConfig): Observable<IXDSProjectConfig> {
        return this._post('/projects', cfg);
    }

    deleteProject(id: string): Observable<IXDSProjectConfig> {
        return this._delete('/projects/' + id);
    }

    syncProject(id: string): Observable<string> {
        return this._post('/projects/sync/' + id, {});
    }

    /***
    ** Exec
    ***/
    exec(prjID: string, dir: string, cmd: string, sdkid?: string, args?: string[], env?: string[]): Observable<any> {
        return this._post('/exec',
            {
                id: prjID,
                rpath: dir,
                cmd: cmd,
                sdkID: sdkid || "",
                args: args || [],
                env: env || [],
            });
    }

    make(prjID: string, dir: string, sdkid?: string, args?: string[], env?: string[]): Observable<any> {
        // SEB TODO add serverID
        return this._post('/make',
            {
                id: prjID,
                rpath: dir,
                sdkID: sdkid,
                args: args || [],
                env: env || [],
            });
    }


    /**
    ** Private functions
    ***/

    private _RegisterEvents() {
        // Register to all existing events
        this._post('/events/register', { "name": "event:all" })
            .subscribe(
            res => { },
            error => {
                this.alert.error("ERROR while registering to all events: ", error);
            }
            );
    }

    private _getServer(ID: string): IXDServerCfg {
        let svr = this._config.servers.filter(item => item.id === ID);
        if (svr.length < 1) {
            return null;
        }
        return svr[0];
    }

    private _attachAuthHeaders(options?: any) {
        options = options || {};
        let headers = options.headers || new Headers();
        // headers.append('Authorization', 'Basic ' + btoa('username:password'));
        headers.append('Accept', 'application/json');
        headers.append('Content-Type', 'application/json');
        // headers.append('Access-Control-Allow-Origin', '*');

        options.headers = headers;
        return options;
    }

    private _get(url: string): Observable<any> {
        return this.http.get(this.baseUrl + url, this._attachAuthHeaders())
            .map((res: Response) => res.json())
            .catch(this._decodeError);
    }
    private _post(url: string, body: any): Observable<any> {
        return this.http.post(this.baseUrl + url, JSON.stringify(body), this._attachAuthHeaders())
            .map((res: Response) => res.json())
            .catch((error) => {
                return this._decodeError(error);
            });
    }
    private _delete(url: string): Observable<any> {
        return this.http.delete(this.baseUrl + url, this._attachAuthHeaders())
            .map((res: Response) => res.json())
            .catch(this._decodeError);
    }

    private _decodeError(err: any) {
        let e: string;
        if (err instanceof Response) {
            const body = err.json() || 'Agent error';
            e = body.error || JSON.stringify(body);
            if (!e || e === "" || e === '{"isTrusted":true}') {
                e = `${err.status} - ${err.statusText || 'Unknown error'}`;
            }
        } else if (typeof err === "object") {
            if (err.statusText) {
                e = err.statusText;
            } else if (err.error) {
                e = String(err.error);
            } else {
                e = JSON.stringify(err);
            }
        } else {
            e = err.message ? err.message : err.toString();
        }
        return Observable.throw(e);
    }
}
