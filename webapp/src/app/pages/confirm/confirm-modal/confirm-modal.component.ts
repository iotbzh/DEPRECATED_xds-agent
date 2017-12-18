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

import { Component, OnInit, Input } from '@angular/core';
import { DomSanitizer } from '@angular/platform-browser';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';

export enum EType {
  YesNo = 1,
  OKCancel,
  OK,
}

@Component({
  selector: 'xds-confirm-modal',
  template: `
  <div tabindex="-1">
    <div class="modal-header">
      {{ title }}
    </div>

    <div class="modal-body row">
      <div class="col-12 text-center">
        <div [innerHtml]="question"></div>
      </div>
      <div class="col-12 text-center" style="margin-top: 2em;">
        <button *ngIf="textBtn[0] != ''" type="button" class="btn btn-primary" tabindex="2"
          (click)="onClick(textBtn[0])">{{textBtn[0]}}</button>
        <button *ngIf="textBtn[1] != ''" type="button" class="btn btn-default" tabindex="1"
          (click)="onClick(textBtn[1])">{{textBtn[1]}}</button>
        <button *ngIf="textBtn[2] != ''" type="button" class="btn btn-default" tabindex="3"
          (click)="onClick(textBtn[2])">{{textBtn[2]}}</button>
      </div>
    </div>

    <div *ngIf="footer!=''" class="modal-footer">
      <div class="col-12 text-center">
        <div [innerHtml]="footer"></div>
      </div>
    </div>
  </div>
  `,
})

export class ConfirmModalComponent implements OnInit {
  @Input() title;
  @Input() footer = '';
  @Input() type;
  @Input() question;

  bodyQuestion = '';
  textBtn: Array<string> = ['', '', ''];

  constructor(
    private modalRef: NgbActiveModal,
    private sanitizer: DomSanitizer,
  ) { }

  ngOnInit() {
    switch (this.type) {
      case EType.OK:
        this.textBtn = [ 'OK', '', '' ];
        break;

      case EType.OKCancel:
        this.textBtn = [ 'OK', 'Cancel', '' ];
      break;

      default:
      case EType.YesNo:
        this.textBtn = [ 'Yes', 'No', '' ];
        break;
    }
  }

  onClick(txt: string): void {
    this.modalRef.close(txt.toLowerCase());
  }
}
