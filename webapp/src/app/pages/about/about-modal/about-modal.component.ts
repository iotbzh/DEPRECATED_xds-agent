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
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';

import { XDSAgentService, IXDSVersions, IXDSVer } from '../../../@core-xds/services/xdsagent.service';


@Component({
  selector: 'xds-about-modal',
  template: `
    <div class="modal-header">
      <span>About <b>X</b>(cross) Development System</span>
      <button class="close" aria-label="Close" (click)="closeModal()">
        <span aria-hidden="true">&times;</span>
      </button>
    </div>

    <div class="modal-body row">
      <div class="col-12">
        <label class="col-sm-4">Developed by IoT.bzh</label>
        <span class="col-sm-8"><a href="http://iot.bzh/en/author" target="_blank">http://iot.bzh</a></span>
      </div>
      <div class="col-12">
        <label class="col-sm-4">Powered by</label>
        <span class="col-sm-8"><a href="https://github.com/akveo/ngx-admin" target="_blank">akveo/ngx-admin</a></span>
      </div>

      <br><br>

      <div class="col-12">
          <label class="col-sm-4">XDS Agent ID</label>
          <span class="col-sm-8">{{agent?.id}}</span>
      </div>
      <div class="col-12">
        <label class="col-sm-4">XDS Agent Version</label>
        <span class="col-sm-8">{{agent?.version}}</span>
      </div>
      <div class="col-12">
        <label class="col-sm-4">XDS Agent Sub-Version</label>
        <span class="col-sm-8">{{agent?.gitTag}}</span>
      </div>

      <div class="col-12">
        <label class="col-sm-4">XDS Server ID</label>
        <span class="col-sm-8">{{server?.id}}</span>
      </div>
      <div class="col-12">
        <label class="col-sm-4">XDS Server Version</label>
        <span class="col-sm-8">{{server?.version}}</span>
      </div>
      <div class="col-12">
        <label class="col-sm-4">XDS Server Sub-Version</label>
        <span class="col-sm-8">{{server?.gitTag}}</span>
      </div>

    </div>
  `,
})

export class AboutModalComponent implements OnInit {

  agent: IXDSVer;
  server: IXDSVer;

  constructor(
    private activeModal: NgbActiveModal,
    private xdsSvr: XDSAgentService,
  ) { }

  ngOnInit() {
    this.xdsSvr.getVersion().subscribe(v => {
      this.agent = v.client;
      if (v && v.servers.length > 0 && !v.servers[0].version.startsWith('Cannot retrieve')) {
        this.server = v.servers[0];
      }
    });
  }

  closeModal() {
    this.activeModal.close();
  }
}
