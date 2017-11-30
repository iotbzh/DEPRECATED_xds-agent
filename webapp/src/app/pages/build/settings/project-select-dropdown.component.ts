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
import { Observable } from 'rxjs/Observable';

import { IProject, ProjectService } from '../../../@core-xds/services/project.service';

@Component({
  selector: 'xds-project-select-dropdown',
  template: `
      <div class="form-group">
      <label>Project</label>
      <select class="form-control" [(ngModel)]="curPrj" (click)="select()">
        <option  *ngFor="let prj of projects$ | async" [ngValue]="prj">{{ prj.label }}</option>
      </select>
    </div>
    `,
})
export class ProjectSelectDropdownComponent implements OnInit {

  projects$: Observable<IProject[]>;
  curPrj: IProject;

  constructor(private projectSvr: ProjectService) { }

  ngOnInit() {
    this.curPrj = this.projectSvr.getCurrent();
    this.projects$ = this.projectSvr.projects$;
    this.projectSvr.curProject$.subscribe(p => this.curPrj = p);
  }

  select() {
    if (this.curPrj) {
      this.projectSvr.setCurrentById(this.curPrj.id);
    }
  }
}


