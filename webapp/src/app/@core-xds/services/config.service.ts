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

import { AlertService, IAlert } from '../services/alert.service';

export interface IConfig {
    language: string;
    projectsRootDir: string;
}

@Injectable()
export class ConfigService {

    public Conf$: Observable<IConfig>;

    private confSubject: BehaviorSubject<IConfig>;
    private confStore: IConfig;

    constructor(
        private cookie: CookieService,
        private alert: AlertService,
    ) {
        this.load();
        this.confSubject = <BehaviorSubject<IConfig>>new BehaviorSubject(this.confStore);
        this.Conf$ = this.confSubject.asObservable();
    }

    // Load config
    load() {
        // Try to retrieve previous config from cookie
        const cookConf = this.cookie.getObject('xds-config');
        if (cookConf != null) {
            this.confStore = <IConfig>cookConf;
        } else {
            // Set default config
            this.confStore = {
                language: 'ENG',
                projectsRootDir: '',
                // projects: []
            };
        }
    }

    // Save config into cookie
    save() {
        // Notify subscribers
        this.confSubject.next(Object.assign({}, this.confStore));

        // Don't save projects in cookies (too big!)
        const cfg = Object.assign({}, this.confStore);
        this.cookie.putObject('xds-config', cfg);
    }

    set projectsRootDir(p: string) {
        this.confStore.projectsRootDir = p;
        this.save();
    }

}
