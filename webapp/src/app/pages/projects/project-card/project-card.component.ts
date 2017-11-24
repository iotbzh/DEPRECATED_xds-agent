import { Component, Input, Pipe, PipeTransform } from '@angular/core';
import { ProjectService, IProject, ProjectType, ProjectTypeEnum } from '../../../@core-xds/services/project.service';
import { AlertService } from '../../../@core-xds/services/alert.service';


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
  ) {
  }

  delete(prj: IProject) {
    this.projectSvr.delete(prj).subscribe(
      res => { },
      err => this.alert.error('ERROR delete: ' + err),
    );
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
