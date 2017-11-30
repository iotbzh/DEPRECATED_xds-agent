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
