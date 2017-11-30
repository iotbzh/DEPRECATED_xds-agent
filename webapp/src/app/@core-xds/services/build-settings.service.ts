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

import { Injectable } from '@angular/core';
import { CookieService } from 'ngx-cookie';
import { Observable } from 'rxjs/Observable';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

export interface IBuildSettings {
  subpath: string;
  cmdClean: string;
  cmdPrebuild: string;
  cmdBuild: string;
  cmdPopulate: string;
  cmdArgs: string[];
  envVars: string[];
}

@Injectable()
export class BuildSettingsService {
  public settings$: Observable<IBuildSettings>;

  private settingsSubject: BehaviorSubject<IBuildSettings>;
  private settingsStore: IBuildSettings;

  constructor(
    private cookie: CookieService,
  ) {
    this._load();
  }

  // Load build settings from cookie
  private _load() {
    // Try to retrieve previous config from cookie
    const cookConf = this.cookie.getObject('xds-build-settings');
    if (cookConf != null) {
      this.settingsStore = <IBuildSettings>cookConf;
    } else {
      // Set default config
      this.settingsStore = {
        subpath: '',
        cmdClean: 'rm -rf build && echo Done',
        cmdPrebuild: 'mkdir -p build && cd build && cmake ..',
        cmdBuild: 'cd build && make',
        cmdPopulate: 'cd build && make remote-target-populate',
        cmdArgs: [],
        envVars: [],
      };
    }
  }

  // Save config into cookie
  private _save() {
    // Notify subscribers
    this.settingsSubject.next(Object.assign({}, this.settingsStore));

    const cfg = Object.assign({}, this.settingsStore);
    this.cookie.putObject('xds-build-settings', cfg);
  }

  // Get whole config values
  get(): IBuildSettings {
    return this.settingsStore;
  }

  // Get whole config values
  set(bs: IBuildSettings) {
    this.settingsStore = bs;
    this._save();
  }

  get subpath(): string {
    return this.settingsStore.subpath;
  }

  set subpath(p: string) {
    this.settingsStore.subpath = p;
    this._save();
  }

}
