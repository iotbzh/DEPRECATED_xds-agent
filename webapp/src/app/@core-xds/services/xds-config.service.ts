import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

import { AlertService, IAlert } from '../services/alert.service';
import { XDSAgentService, IAgentStatus, IXDServerCfg } from '../../@core-xds/services/xdsagent.service';

import 'rxjs/add/operator/publish';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/catch';


@Injectable()
export class XDSConfigService {

  // Conf$: Observable<IXdsConfig>;
  xdsServers: IXDServerCfg[];

  // private confSubject: BehaviorSubject<IXdsConfig>;
  // private confStore: IXdsConfig;

  private _curServer: IXDServerCfg = { id: '', url: '', connRetry: 0, connected: false };
  private curServer$ = new Subject<IXDServerCfg>();

  constructor(
    private alert: AlertService,
    private xdsAgentSvr: XDSAgentService,
  ) {
    /*
    this.confSubject = <BehaviorSubject<IXdsConfig>>new BehaviorSubject(this.confStore);
    this.Conf$ = this.confSubject.asObservable();
    */

    // Update servers list
    this.xdsAgentSvr.XdsConfig$.subscribe(cfg => {
      if (!cfg || cfg.servers.length < 1) {
        return;
      }
      this.xdsServers = cfg.servers;
      this._updateCurServer();
    });
  }

  onCurServer(): Observable<IXDServerCfg> {
    return this.curServer$.publish().refCount();
  }

  getCurServer(): IXDServerCfg {
    return this._curServer;
  }

  setCurServer(svr: IXDServerCfg): Observable<IXDServerCfg> {
    const curSvr = this._getCurServer();

    if (!curSvr.connected || curSvr.url !== svr.url) {
      return this.xdsAgentSvr.setServerUrl(curSvr.id, svr.url, svr.connRetry)
        .map(cfg => this._updateCurServer())
        .catch(err => {
          this._curServer.connected = false;
          this.curServer$.next(this._curServer);
          return Observable.throw(err);
        });
    } else {
      if (curSvr.connRetry !== svr.connRetry) {
        return this.xdsAgentSvr.setServerRetry(curSvr.id, svr.connRetry)
          .map(cfg => this._updateCurServer())
          .catch(err => {
            this.curServer$.next(this._curServer);
            return Observable.throw(err);
          });
      }
    }
    return Observable.of(curSvr);
  }

  private _updateCurServer(): IXDServerCfg {
    this._curServer = this._getCurServer();
    this.curServer$.next(this._curServer);
    return this._curServer;
  }

  private _getCurServer(url?: string): IXDServerCfg {
    if (!this.xdsServers) {
      return this._curServer;
    }

    // Init the 1st time
    let svrUrl = url || this._curServer.url;
    if (this._curServer.url === '' && this.xdsServers.length > 0) {
      svrUrl = this.xdsServers[0].url;
    }

    const svr = this.xdsServers.filter(s => s.url === svrUrl);
    return svr[0];
  }

}
