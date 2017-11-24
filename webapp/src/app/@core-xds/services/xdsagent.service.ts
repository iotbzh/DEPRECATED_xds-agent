import { Injectable, Inject, isDevMode } from '@angular/core';
import { HttpClient, HttpHeaders, HttpErrorResponse } from '@angular/common/http';
import { DOCUMENT } from '@angular/common';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';
import * as io from 'socket.io-client';

import { AlertService } from './alert.service';
import { ISdk } from './sdk.service';
import { ProjectType, ProjectTypeEnum } from './project.service';

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
  type: ProjectTypeEnum;
  status?: string;
  isInSync?: boolean;
  defaultSdkID: string;
  clientData?: string;
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
  apiUrl?: string;
  partialUrl?: string;
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

export interface IServerStatus {
  id: string;
  connected: boolean;
}

export interface IAgentStatus {
  connected: boolean;
  servers: IServerStatus[];
}


@Injectable()
export class XDSAgentService {

  public XdsConfig$: Observable<IXDSConfig>;
  public Status$: Observable<IAgentStatus>;
  public CmdOutput$ = <Subject<ICmdOutput>>new Subject();
  public CmdExit$ = <Subject<ICmdExit>>new Subject();

  protected projectAdd$ = new Subject<IXDSProjectConfig>();
  protected projectDel$ = new Subject<IXDSProjectConfig>();
  protected projectChange$ = new Subject<IXDSProjectConfig>();

  private baseUrl: string;
  private wsUrl: string;
  private httpSessionID: string;
  private _config = <IXDSConfig>{ servers: [] };
  private _status = { connected: false, servers: [] };

  private configSubject = <BehaviorSubject<IXDSConfig>>new BehaviorSubject(this._config);
  private statusSubject = <BehaviorSubject<IAgentStatus>>new BehaviorSubject(this._status);

  private socket: SocketIOClient.Socket;

  constructor( @Inject(DOCUMENT) private document: Document,
    private http: HttpClient, private alert: AlertService) {

    this.XdsConfig$ = this.configSubject.asObservable();
    this.Status$ = this.statusSubject.asObservable();

    const originUrl = this.document.location.origin;
    this.baseUrl = originUrl + '/api/v1';

    // Retrieve Session ID / token
    this.http.get(this.baseUrl + '/version', { observe: 'response' })
      .subscribe(
      resp => {
        this.httpSessionID = resp.headers.get('xds-agent-sid');

        const re = originUrl.match(/http[s]?:\/\/([^\/]*)[\/]?/);
        if (re === null || re.length < 2) {
          console.error('ERROR: cannot determine Websocket url');
        } else {
          this.wsUrl = 'ws://' + re[1];
          this._handleIoSocket();
          this._RegisterEvents();
        }
      },
      err => {
        /* tslint:disable:no-console */
        console.error('ERROR while retrieving session id:', err);
      });
  }

  private _NotifyXdsAgentState(sts: boolean) {
    this._status.connected = sts;
    this.statusSubject.next(Object.assign({}, this._status));

    // Update XDS config including XDS Server list when connected
    if (sts) {
      this.getConfig().subscribe(c => {
        this._config = c;
        this._NotifyXdsServerState();
        this.configSubject.next(Object.assign({ servers: [] }, this._config));
      });
    }
  }

  private _NotifyXdsServerState() {
    this._status.servers = this._config.servers.map(svr => {
      return { id: svr.id, connected: svr.connected };
    });
    this.statusSubject.next(Object.assign({}, this._status));
  }

  private _handleIoSocket() {
    this.socket = io(this.wsUrl, { transports: ['websocket'] });

    this.socket.on('connect_error', (res) => {
      this._NotifyXdsAgentState(false);
      console.error('XDS Agent WebSocket Connection error !');
    });

    this.socket.on('connect', (res) => {
      this._NotifyXdsAgentState(true);
    });

    this.socket.on('disconnection', (res) => {
      this._NotifyXdsAgentState(false);
      this.alert.error('WS disconnection: ' + res);
    });

    this.socket.on('error', (err) => {
      console.error('WS error:', err);
    });

    // XDS Events decoding

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

    this.socket.on('event:server-config', ev => {
      if (ev && ev.data) {
        const cfg: IXDServerCfg = ev.data;
        const idx = this._config.servers.findIndex(el => el.id === cfg.id);
        if (idx >= 0) {
          this._config.servers[idx] = Object.assign({}, cfg);
          this._NotifyXdsServerState();
        }
        this.configSubject.next(Object.assign({}, this._config));
      }
    });

    this.socket.on('event:project-add', (ev) => {
      if (ev && ev.data && ev.data.id) {
        this.projectAdd$.next(Object.assign({}, ev.data));
        if (ev.sessionID !== this.httpSessionID && ev.data.label) {
          this.alert.info('Project "' + ev.data.label + '" has been added by another tool.');
        }
      } else if (isDevMode) {
        /* tslint:disable:no-console */
        console.log('Warning: received event:project-add with unknown data: ev=', ev);
      }
    });

    this.socket.on('event:project-delete', (ev) => {
      if (ev && ev.data && ev.data.id) {
        this.projectDel$.next(Object.assign({}, ev.data));
        if (ev.sessionID !== this.httpSessionID && ev.data.label) {
          this.alert.info('Project "' + ev.data.label + '" has been deleted by another tool.');
        }
      } else if (isDevMode) {
        console.log('Warning: received event:project-delete with unknown data: ev=', ev);
      }
    });

    this.socket.on('event:project-state-change', ev => {
      if (ev && ev.data) {
        this.projectChange$.next(Object.assign({}, ev.data));
      } else if (isDevMode) {
        console.log('Warning: received event:project-state-change with unknown data: ev=', ev);
      }
    });

  }

  /**
  ** Events registration
  ***/
  onProjectAdd(): Observable<IXDSProjectConfig> {
    return this.projectAdd$.asObservable();
  }

  onProjectDelete(): Observable<IXDSProjectConfig> {
    return this.projectDel$.asObservable();
  }

  onProjectChange(): Observable<IXDSProjectConfig> {
    return this.projectChange$.asObservable();
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

  setServerRetry(serverID: string, retry: number): Observable<IXDSConfig> {
    const svr = this._getServer(serverID);
    if (!svr) {
      return Observable.throw('Unknown server ID');
    }
    if (retry < 0 || Number.isNaN(retry) || retry == null) {
      return Observable.throw('Not a valid number');
    }
    svr.connRetry = retry;
    return this._setConfig();
  }

  setServerUrl(serverID: string, url: string, retry: number): Observable<IXDSConfig> {
    const svr = this._getServer(serverID);
    if (!svr) {
      return Observable.throw('Unknown server ID');
    }
    svr.connected = false;
    svr.url = url;
    if (!Number.isNaN(retry) && retry > 0) {
      svr.connRetry = retry;
    }
    this._NotifyXdsServerState();
    return this._setConfig();
  }

  private _setConfig(): Observable<IXDSConfig> {
    return this.setConfig(this._config)
      .map(newCfg => {
        this._config = newCfg;
        this.configSubject.next(Object.assign({}, this._config));
        return this._config;
      });
  }

  /***
  ** SDKs
  ***/
  getSdks(serverID: string): Observable<ISdk[]> {
    const svr = this._getServer(serverID);
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

  updateProject(cfg: IXDSProjectConfig): Observable<IXDSProjectConfig> {
    return this._put('/projects/' + cfg.id, cfg);
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
        sdkID: sdkid || '',
        args: args || [],
        env: env || [],
      });
  }

  /**
  ** Private functions
  ***/

  private _RegisterEvents() {
    // Register to all existing events
    this._post('/events/register', { 'name': 'event:all' })
      .subscribe(
      res => { },
      error => {
        this.alert.error('ERROR while registering to all events: ' + error);
      },
    );
  }

  private _getServer(ID: string): IXDServerCfg {
    const svr = this._config.servers.filter(item => item.id === ID);
    if (svr.length < 1) {
      return null;
    }
    return svr[0];
  }

  private _attachAuthHeaders(options?: any) {
    options = options || {};
    const headers = options.headers || new HttpHeaders();
    // headers.append('Authorization', 'Basic ' + btoa('username:password'));
    headers.append('Accept', 'application/json');
    headers.append('Content-Type', 'application/json');
    // headers.append('Access-Control-Allow-Origin', '*');

    options.headers = headers;
    return options;
  }

  private _get(url: string): Observable<any> {
    return this.http.get(this.baseUrl + url, this._attachAuthHeaders())
      .catch(this._decodeError);
  }
  private _post(url: string, body: any): Observable<any> {
    return this.http.post(this.baseUrl + url, JSON.stringify(body), this._attachAuthHeaders())
      .catch((error) => {
        return this._decodeError(error);
      });
  }
  private _put(url: string, body: any): Observable<any> {
    return this.http.put(this.baseUrl + url, JSON.stringify(body), this._attachAuthHeaders())
      .catch((error) => {
        return this._decodeError(error);
      });
  }
  private _delete(url: string): Observable<any> {
    return this.http.delete(this.baseUrl + url, this._attachAuthHeaders())
      .catch(this._decodeError);
  }

  private _decodeError(err: any) {
    let e: string;
    if (err instanceof HttpErrorResponse) {
      e = (err.error && err.error.error) ? err.error.error : err.message || 'Unknown error';
    } else if (typeof err === 'object') {
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
    /* tslint:disable:no-console */
    if (isDevMode) {
      console.log('xdsagent.service - ERROR: ', e);
    }
    return Observable.throw(e);
  }
}
