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
    const activeModal = this.modalService.open(ProjectAddModalComponent, { size: 'lg', container: 'nb-layout' });
    activeModal.componentInstance.modalHeader = 'Large Modal';
  }
}
