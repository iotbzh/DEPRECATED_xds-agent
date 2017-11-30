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

import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
// import { SdkAddModalComponent } from './sdk-add-modal/sdk-add-modal.component';

import { SdkService, ISdk } from '../../@core-xds/services/sdk.service';

@Component({
  selector: 'xds-sdks',
  styleUrls: ['./sdks.component.scss'],
  templateUrl: './sdks.component.html',
})
export class SdksComponent implements OnInit {

  sdks$: Observable<ISdk[]>;
  sdks: ISdk[];

  constructor(
    private sdkSvr: SdkService,
    private modalService: NgbModal,
  ) {
  }

  ngOnInit() {
    this.sdks$ = this.sdkSvr.Sdks$;
  }

  add() {
    /* SEB TODO
    const activeModal = this.modalService.open(SdkAddModalComponent, { size: 'lg', container: 'nb-layout' });
    activeModal.componentInstance.modalHeader = 'Large Modal';
    */
  }
}
