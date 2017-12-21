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

import { Component, OnInit, Input, ViewChild, AfterViewChecked, ElementRef } from '@angular/core';
import { DomSanitizer } from '@angular/platform-browser';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';

import { AlertService } from '../../../@core-xds/services/alert.service';
import { SdkService, ISdk } from '../../../@core-xds/services/sdk.service';

@Component({
  selector: 'xds-sdk-install-modal',
  template: `
  <div tabindex="-1">
    <div class="modal-header">
      SDK installation
    </div>

    <div class="modal-body row">
      <div class="col-12 text-center">
        Installation of <b> {{ sdk?.name }} '</b> <span [innerHTML]="instStatus"></span>
      </div>
      <br>
      <br>
      <div class="col-12 text-center">
        <textarea rows="20" class="textarea-scroll" #scrollOutput [innerHtml]="installOutput"></textarea>
      </div>
      <div class="col-12 text-center">
        <button type="button" class="btn" tabindex="1"
        [ngClass]="(btnName=='Cancel')?'btn-default':'btn-primary'"
        (click)="onBtnClick()">{{ btnName }}</button>
      </div>
    </div>

    <!-- <div *ngIf="footer!=''" class="modal-footer">
      <div class="col-12 text-center">
      </div>
    </div> -->
  </div>
  `,
  styles: [`
    .btn {
      margin-top: 2em;
      min-width: 10em;
    }
    .textarea-scroll {
      font-family: monospace;
      width: 100%;
      overflow-y: scroll;
  `],
})

export class SdkInstallComponent implements OnInit {
  @Input() sdk;
  @ViewChild('scrollOutput') private scrollContainer: ElementRef;

  constructor(
    private modalRef: NgbActiveModal,
    private sanitizer: DomSanitizer,
    private alert: AlertService,
    private sdkSvr: SdkService,
  ) { }

  onInstallSub: any;
  installOutput = '';
  btnName = 'Cancel';
  instStatus = '';

  ngOnInit() {
    this.instStatus = 'in progress...';

    this.onInstallSub = this.sdkSvr.onInstall().subscribe(ev => {
      if (ev.exited) {
        this.btnName = 'OK';
        this.instStatus = '<font color="green"> Done. </font>';

        if (ev.code === 0) {
          this.alert.info('SDK ' + ev.sdk.name + ' successfully installed.');

        } else {
          if (ev.sdk.lastError !== '') {
            this.alert.error(ev.sdk.lastError);
          } else {
            this.alert.error('SDK ' + ev.sdk.name + ' installation failed. ' + ev.error);
          }
        }

      } else {
        if (ev.stdout !== '') {
          this.installOutput += ev.stdout;
        }
        if (ev.stderr !== '') {
          this.installOutput += ev.stderr;
        }
        this._scrollToBottom();
      }
    });
  }

  onBtnClick(): void {
    this.onInstallSub.unsubscribe();
    if (this.btnName === 'Cancel') {
      this.btnName = 'OK';
      this.instStatus = '<b><font color="red"> ABORTED </font></b>';
      this.sdkSvr.abortInstall(this.sdk).subscribe(r => { }, err => this.alert.error(err));
    } else {
      this.modalRef.close();
    }
  }

  private _scrollToBottom(): void {
    try {
      this.scrollContainer.nativeElement.scrollTop = this.scrollContainer.nativeElement.scrollHeight;
    } catch (err) { }
  }
}
