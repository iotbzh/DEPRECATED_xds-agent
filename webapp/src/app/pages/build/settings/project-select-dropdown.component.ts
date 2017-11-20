import { Component, OnInit, Input } from '@angular/core';

import { IProject, ProjectService } from '../../../@core-xds/services/project.service';

@Component({
  selector: 'xds-project-select-dropdown',
  template: `
      <div class="form-group">
      <label>Project</label>
      <select class="form-control">
        <option *ngFor="let prj of projects" (click)="select(prj)">{{prj.label}}</option>
      </select>
    </div>
    `,
})
export class ProjectSelectDropdownComponent implements OnInit {

  projects: IProject[];
  curPrj: IProject;

  constructor(private prjSvr: ProjectService) { }

  ngOnInit() {
    this.curPrj = this.prjSvr.getCurrent();
    this.prjSvr.Projects$.subscribe((s) => {
      if (s) {
        this.projects = s;
        if (this.curPrj === null || s.indexOf(this.curPrj) === -1) {
          this.prjSvr.setCurrent(this.curPrj = s.length ? s[0] : null);
        }
      }
    });
  }

  select(s) {
    this.prjSvr.setCurrent(this.curPrj = s);
  }
}


