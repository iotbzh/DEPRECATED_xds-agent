import { Component, OnInit, OnDestroy } from '@angular/core';
import { TranslateService } from '@ngx-translate/core';
import { ConfigService, IConfig } from './services/config.service';

@Component({
    selector: 'app-root',
    templateUrl: 'app.component.html'
})

export class AppComponent implements OnInit, OnDestroy {
    private defaultLanguage = 'en';
    public isCollapsed = true;

    constructor(private translate: TranslateService, private configSvr: ConfigService) {
    }

    ngOnInit() {
        this.translate.addLangs(['en', 'fr']);
        this.translate.setDefaultLang(this.defaultLanguage);

        const browserLang = this.translate.getBrowserLang();
        this.translate.use(browserLang.match(/en|fr/) ? browserLang : this.defaultLanguage);

        this.configSvr.Conf$.subscribe((cfg: IConfig) => {
            let lang: string;
            switch (cfg.language) {
                case 'ENG':
                    lang = 'en';
                    break;
                case 'FRA':
                    lang = 'fr';
                    break;
                default:
                    lang = this.defaultLanguage;
            }
            this.translate.use(lang);
        });
    }

    ngOnDestroy(): void {
        // this.aglIdentityService.loginResponse.unsubscribe();
        // this.aglIdentityService.logoutResponse.unsubscribe();
    }
}
