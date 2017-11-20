import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
//import { SdkAddModalComponent } from './sdk-add-modal/sdk-add-modal.component';

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
