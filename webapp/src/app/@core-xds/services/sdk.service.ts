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

import { Injectable, SecurityContext, isDevMode } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

import { XDSAgentService } from '../services/xdsagent.service';

import 'rxjs/add/observable/throw';

/* FIXME: syntax only compatible with TS>2.4.0
export enum StatusType {
  DISABLE = 'Disable',
  NOT_INSTALLED = 'Not Installed',
  INSTALLING = 'Installing',
  UNINSTALLING = 'Un-installing',
  INSTALLED = 'Installed'
}
*/
export type StatusTypeEnum = 'Disable' | 'Not Installed' | 'Installing' | 'Un-installing' | 'Installed';
export const StatusType = {
  DISABLE: 'Disable',
  NOT_INSTALLED: 'Not Installed',
  INSTALLING: 'Installing',
  UNINSTALLING: 'Un-installing',
  INSTALLED: 'Installed',
};

export interface ISdk {
  id: string;
  name: string;
  description: string;
  profile: string;
  version: string;
  arch: string;
  path: string;
  url: string;
  status: string;
  date: string;
  size: string;
  md5sum: string;
  setupFile: string;
  lastError: string;
}

export interface ISdkManagementMsg {
  cmdID: string;
  timestamp: string;
  sdk: ISdk;
  stdout: string;
  stderr: string;
  progress: number;
  exited: boolean;
  code: number;
  error: string;
}

@Injectable()
export class SdkService {
  public Sdks$: Observable<ISdk[]>;
  public curSdk$: Observable<ISdk>;

  private _sdksList = [];
  private sdksSubject = <BehaviorSubject<ISdk[]>>new BehaviorSubject(this._sdksList);
  private current: ISdk;
  private curSdkSubject = <BehaviorSubject<ISdk>>new BehaviorSubject(this.current);
  private curServerID;

  constructor(private xdsSvr: XDSAgentService) {
    this.current = null;
    this.Sdks$ = this.sdksSubject.asObservable();
    this.curSdk$ = this.curSdkSubject.asObservable();

    this.xdsSvr.XdsConfig$.subscribe(cfg => {
      if (!cfg || cfg.servers.length < 1) {
        return;
      }
      // FIXME support multiple server
      // cfg.servers.forEach(svr => {
      this.curServerID = cfg.servers[0].id;
      this.xdsSvr.getSdks(this.curServerID).subscribe((sdks) => {
        this._sdksList = [];
        sdks.forEach(s => {
          this._addSdk(s, true);
        });

        // TODO: get previous val from xds-config service / cookie
        if (this._sdksList.length > 0) {
          this.current = this._sdksList[0];
          this.curSdkSubject.next(this.current);
        }

        this.sdksSubject.next(this._sdksList);
      });
    });

    // Add listener on sdk creation, deletion and change events
    this.xdsSvr.onSdkInstall().subscribe(evMgt => {
      this._addSdk(evMgt.sdk);
    });
    this.xdsSvr.onSdkRemove().subscribe(evMgt => {
      if (evMgt.sdk.status !== StatusType.NOT_INSTALLED) {
        /* tslint:disable:no-console */
        console.log('Error: received event:sdk-remove with invalid status: evMgt=', evMgt);
        return;
      }
      this._delSdk(evMgt.sdk);
    });

  }

  public setCurrent(s: ISdk) {
    this.current = s;
  }

  public getCurrent(): ISdk {
    return this.current;
  }

  public getCurrentId(): string {
    if (this.current && this.current.id) {
      return this.current.id;
    }
    return '';
  }

  public install(sdk: ISdk): Observable<ISdk> {
    return this.xdsSvr.installSdk(this.curServerID, sdk.id);
  }

  public onInstall(): Observable<ISdkManagementMsg> {
    return this.xdsSvr.onSdkInstall();
  }

  public abortInstall(sdk: ISdk): Observable<ISdk> {
    return this.xdsSvr.abortInstall(this.curServerID, sdk.id);
  }

  public remove(sdk: ISdk): Observable<ISdk> {
    return this.xdsSvr.removeSdk(this.curServerID, sdk.id);
  }

  /** Private **/

  private _addSdk(sdk: ISdk, noNext?: boolean): ISdk {

    // check if sdk already exists
    const idx = this._sdksList.findIndex(s => s.id === sdk.id);
    if (idx >= 0) {
      this._sdksList[idx] = sdk;
    } else {
      // add new sdk
      this._sdksList.push(sdk);
    }

    // sort sdk array
    this._sdksList.sort((a, b) => {
      if (a.name < b.name) {
        return -1;
      }
      if (a.name > b.name) {
        return 1;
      }
      return 0;
    });

    if (!noNext) {
      this.sdksSubject.next(this._sdksList);
    }

    return sdk;
  }

  private _delSdk(sdk: ISdk) {
    const idx = this._sdksList.findIndex(item => item.id === sdk.id);
    if (idx === -1) {
      if (isDevMode) {
        /* tslint:disable:no-console */
        console.log('Warning: Try to delete sdk unknown id: sdk=', sdk);
      }
      return;
    }
    const delId = this._sdksList[idx].id;
    this._sdksList.splice(idx, 1);
    if (delId === this.current.id) {
      this.setCurrent(this._sdksList[0]);
    }
    this.sdksSubject.next(this._sdksList);
  }

}
