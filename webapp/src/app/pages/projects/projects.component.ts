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
import { ProjectAddModalComponent } from './project-add-modal/project-add-modal.component';

import { ProjectService, IProject } from '../../@core-xds/services/project.service';

@Component({
  selector: 'xds-projects',
  styleUrls: ['./projects.component.scss'],
  templateUrl: './projects.component.html',
})
export class ProjectsComponent implements OnInit {

  projects$: Observable<IProject[]>;
  projects: IProject[];

  constructor(
    private projectSvr: ProjectService,
    private modalService: NgbModal,
  ) {
  }

  ngOnInit() {
    this.projects$ = this.projectSvr.projects$;
  }

  add() {
    const activeModal = this.modalService.open(ProjectAddModalComponent, {
      size: 'lg',
      windowClass: 'modal-xxl',
      container: 'nb-layout',
    });
  }
}
