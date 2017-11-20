import { Component, OnInit, Input } from '@angular/core';

import { ISdk, SdkService } from '../../../@core-xds/services/sdk.service';

@Component({
    selector: 'xds-sdk-select-dropdown',
    template: `
      <div class="form-group">
      <label>SDK</label>
      <select class="form-control">
        <option *ngFor="let sdk of sdks" (click)="select(sdk)">{{sdk.name}}</option>
      </select>
    </div>
    `,
})
export class SdkSelectDropdownComponent implements OnInit {

    // FIXME investigate to understand why not working with sdks as input
    // <xds-sdk-select-dropdown [sdks]="(sdks$ | async)"></xds-sdk-select-dropdown>
    // @Input() sdks: ISdk[];
    sdks: ISdk[];

    curSdk: ISdk;

    constructor(private sdkSvr: SdkService) { }

    ngOnInit() {
        this.curSdk = this.sdkSvr.getCurrent();
        this.sdkSvr.Sdks$.subscribe((s) => {
            if (s) {
                this.sdks = s;
                if (this.curSdk === null || s.indexOf(this.curSdk) === -1) {
                    this.sdkSvr.setCurrent(this.curSdk = s.length ? s[0] : null);
                }
            }
        });
    }

    select(s) {
        this.sdkSvr.setCurrent(this.curSdk = s);
    }
}


