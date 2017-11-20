import { NgModule } from '@angular/core';
import { ThemeModule } from '../../@theme/theme.module';

import { ProjectsComponent } from './projects.component';
import { ProjectCardComponent, ProjectReadableTypePipe } from './project-card/project-card.component';
import { ProjectAddModalComponent } from './project-add-modal/project-add-modal.component';


@NgModule({
  imports: [
    ThemeModule,
  ],
  declarations: [
    ProjectsComponent,
    ProjectCardComponent,
    ProjectAddModalComponent,
    ProjectReadableTypePipe,
  ],
  entryComponents: [
    ProjectAddModalComponent
  ],
})
export class ProjectsModule { }
