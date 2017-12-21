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

import { Component, ViewEncapsulation, OnInit, isDevMode } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { LocalDataSource } from 'ng2-smart-table';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ConfirmModalComponent, EType } from '../../confirm/confirm-modal/confirm-modal.component';
import { SdkInstallComponent } from './sdk-install.component';

import { AlertService } from '../../../@core-xds/services/alert.service';
import { SdkService, ISdk } from '../../../@core-xds/services/sdk.service';
import { ISdkMessage } from '../../../@core-xds/services/xdsagent.service';

interface ISdkMgt extends ISdk {
  link: string;
  selected: boolean;
}

/*
 * FIXME / TODO:
 *  - support install of multi SDKs  (see settings.selectMode: 'multi')
 *  - add Uninstall button (use delete)
 *  - add (mouseover) to display description, date, size, ...
 */

@Component({
  selector: 'xds-sdk-management',
  templateUrl: 'sdk-management.component.html',
  styleUrls: ['sdk-management.component.scss'],
  encapsulation: ViewEncapsulation.None,
})

export class SdkManagementComponent implements OnInit {

  sdks$: Observable<ISdk[]>;
  sdks: ISdkMgt[];
  source: LocalDataSource = new LocalDataSource();

  settings = {
    mode: 'external',
    actions: {
      add: false,
      edit: false,
      delete: false,  // TODO, add delete == uninstall
      custom: [
        { name: 'install', title: '<i class="nb-plus"></i>' },
      ],
    },
    delete: {
      deleteButtonContent: '<i class="nb-trash"></i>',
      confirmDelete: true,
    },
    columns: {
      name: { title: 'Name', editable: false },
      profile: { title: 'Profile', editable: false, filter: {} },
      arch: { title: 'Architecture', editable: false, filter: {} },
      version: { title: 'Version', editable: false },
      // TODO: add status when delete supported:
      // status: { title: 'Status', editable: false },
      link: { title: 'Link', editable: false, type: 'html', filter: false, width: '2%' },
    },
  };

  constructor(
    private alert: AlertService,
    private sdkSvr: SdkService,
    private modalService: NgbModal,
  ) { }

  ngOnInit() {
    this.sdkSvr.Sdks$.subscribe(sdks => {
      const profMap = {};
      const archMap = {};
      this.sdks = [];
      sdks.forEach(s => {
        // only display not installed SDK
        if (s.status !== 'Not Installed') {
          return;
        }
        profMap[s.profile] = s.profile;
        archMap[s.arch] = s.arch;

        const sm = <ISdkMgt>s;
        sm.selected = false;
        if (s.url !== '') {
          sm.link = '<a href="' + s.url.substr(0, s.url.lastIndexOf('/')) + '" target="_blank" class="fa fa-external-link"></a>';
        }
        this.sdks.push(sm);

      });

      // Add text box filter for Profile and Arch columns
      const profList = []; Object.keys(profMap).forEach(a => profList.push({ value: a, title: a }));
      this.settings.columns.profile.filter = {
        type: 'list',
        config: { selectText: 'Select...', list: profList },
      };

      const archList = []; Object.keys(archMap).forEach(a => archList.push({ value: a, title: a }));
      this.settings.columns.arch.filter = {
        type: 'list',
        config: { selectText: 'Select...', list: archList },
      };

      // update sources
      this.source.load(this.sdks);

    });
  }

  onCustom(event): void {
    if (event.action === 'install') {
      const sdk = <ISdkMgt>event.data;
      const modal = this.modalService.open(ConfirmModalComponent, {
        size: 'lg',
        backdrop: 'static',
        container: 'nb-layout',
      });
      modal.componentInstance.title = 'Confirm SDK installation';
      modal.componentInstance.type = EType.YesNo;
      modal.componentInstance.question = `
      Please confirm installation of <b>` + sdk.name + `'</b> SDK ?<br>
      <br>
      <i><small>(size: ` + sdk.size + `, date: ` + sdk.date + `)</small></i>`;

      modal.result.then(res => {
        if (res === 'yes') {
          // Request installation
          this.sdkSvr.install(sdk).subscribe(r => { }, err => this.alert.error(err));

          const modalInstall = this.modalService.open(SdkInstallComponent, {
            size: 'lg',
            backdrop: 'static',
            container: 'nb-layout',
          });
          modalInstall.componentInstance.sdk = sdk;
        }
      });


    } else if (event.action === 'uninstall') {
      // TODO

    } else {
      /* tslint:disable:no-console */
      if (isDevMode) {
        console.error('onCustom: unknown event action: ', event);
      }
    }
  }

}
