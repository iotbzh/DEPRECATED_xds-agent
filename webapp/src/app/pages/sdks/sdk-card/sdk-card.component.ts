import { Component, Input, Pipe, PipeTransform } from '@angular/core';
import { SdkService, ISdk } from '../../../@core-xds/services/sdk.service';
import { AlertService } from '../../../@core-xds/services/alert.service';


@Component({
  selector: 'xds-sdk-card',
  styleUrls: ['./sdk-card.component.scss'],
  templateUrl: './sdk-card.component.html',
})
export class SdkCardComponent {

  // FIXME workaround of https://github.com/angular/angular-cli/issues/2034
  // should be removed with angular 5
  // @Input() sdk: ISdk;
  @Input() sdk = <ISdk>null;

  constructor(
    private alert: AlertService,
    private sdkSvr: SdkService,
  ) {
  }

  labelGet(sdk: ISdk) {
    return sdk.profile + '-' + sdk.arch + '-' + sdk.version;
  }

  delete(sdk: ISdk) {
    this.sdkSvr.delete(sdk).subscribe(
      res => { },
      err => this.alert.error('ERROR delete: ' + err),
    );
  }
}

