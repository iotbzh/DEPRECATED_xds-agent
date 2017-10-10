import { Injectable } from '@angular/core';
import { CookieService } from 'ngx-cookie';
import { Observable } from 'rxjs/Observable';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

import { AlertService, IAlert } from "../services/alert.service";
import { UtilsService } from "../services/utils.service";

export interface IConfig {
    projectsRootDir: string;
}

@Injectable()
export class ConfigService {

    public Conf$: Observable<IConfig>;

    private confSubject: BehaviorSubject<IConfig>;
    private confStore: IConfig;

    constructor(private _window: Window,
        private cookie: CookieService,
        private alert: AlertService,
        private utils: UtilsService,
    ) {
        this.load();
        this.confSubject = <BehaviorSubject<IConfig>>new BehaviorSubject(this.confStore);
        this.Conf$ = this.confSubject.asObservable();
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

    set projectsRootDir(p: string) {
        this.confStore.projectsRootDir = p;
        this.save();
    }
}
