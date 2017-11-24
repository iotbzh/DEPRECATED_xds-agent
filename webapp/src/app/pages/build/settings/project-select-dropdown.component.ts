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
    this.projectSvr.setCurrentById(this.curPrj.id);
  }
}


