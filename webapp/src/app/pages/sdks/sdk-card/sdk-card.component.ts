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

import { Component, Input, Pipe, PipeTransform } from '@angular/core';
import { SdkService, ISdk, StatusType } from '../../../@core-xds/services/sdk.service';
import { AlertService } from '../../../@core-xds/services/alert.service';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ConfirmModalComponent, EType } from '../../confirm/confirm-modal/confirm-modal.component';

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
    private modalService: NgbModal,
  ) {
  }

  canRemove(sdk: ISdk) {
    return sdk.status === StatusType.INSTALLED;
  }

  remove(sdk: ISdk) {
    const modal = this.modalService.open(ConfirmModalComponent, {
      size: 'lg',
      backdrop: 'static',
      container: 'nb-layout',
    });
    modal.componentInstance.title = 'Confirm SDK deletion';
    modal.componentInstance.type = EType.YesNo;
    modal.componentInstance.question = `
    Do you <b>permanently remove '` + sdk.name + `'</b> SDK ?
    <br><br>
    <i><small>(SDK ID: ` + sdk.id + ` )</small></i>`;

    modal.result
      .then(res => {
        if (res === 'yes') {
          this.sdkSvr.remove(sdk).subscribe(
            r => { },
            err => this.alert.error(err),
          );
        }
      });
  }
}

