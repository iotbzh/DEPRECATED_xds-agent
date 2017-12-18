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
import { ProjectService, IProject, ProjectType, ProjectTypeEnum } from '../../../@core-xds/services/project.service';
import { AlertService } from '../../../@core-xds/services/alert.service';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ConfirmModalComponent, EType } from '../../confirm/confirm-modal/confirm-modal.component';

@Component({
  selector: 'xds-project-card',
  styleUrls: ['./project-card.component.scss'],
  templateUrl: './project-card.component.html',
})
export class ProjectCardComponent {

  // FIXME workaround of https://github.com/angular/angular-cli/issues/2034
  // should be removed with angular 5
  // @Input() project: IProject;
  @Input() project = <IProject>null;

  constructor(
    private alert: AlertService,
    private projectSvr: ProjectService,
    private modalService: NgbModal,
  ) {
  }

  delete(prj: IProject) {

    const modal = this.modalService.open(ConfirmModalComponent, {
      size: 'lg',
      backdrop: 'static',
      container: 'nb-layout',
    });
    modal.componentInstance.title = 'Confirm SDK deletion';
    modal.componentInstance.type = EType.YesNo;
    modal.componentInstance.question = `
      Do you <b>permanently delete '` + prj.label + `'</b> project ?
      <br><br>
      <i><small>(Project ID: ` + prj.id + ` )</small></i>`;

    modal.result
      .then(res => {
        if (res === 'yes') {
          this.projectSvr.delete(prj).subscribe(
            r => { },
            err => this.alert.error('ERROR delete: ' + err),
          );
        }
      });

  }

  sync(prj: IProject) {
    this.projectSvr.sync(prj).subscribe(
      res => { },
      err => this.alert.error('ERROR: ' + err),
    );
  }
}

// Make Project type human readable
@Pipe({
  name: 'readableType',
})

export class ProjectReadableTypePipe implements PipeTransform {
  transform(type: ProjectTypeEnum): string {
    switch (type) {
      case ProjectType.NATIVE_PATHMAP: return 'Native (path mapping)';
      case ProjectType.SYNCTHING: return 'Cloud (Syncthing)';
      default: return String(type);
    }
  }
}
