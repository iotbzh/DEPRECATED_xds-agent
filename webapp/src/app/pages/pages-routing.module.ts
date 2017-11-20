import { RouterModule, Routes } from '@angular/router';
import { NgModule } from '@angular/core';

import { PagesComponent } from './pages.component';
import { DashboardComponent } from './dashboard/dashboard.component';
import { ProjectsComponent } from './projects/projects.component';
import { SdksComponent } from './sdks/sdks.component';
import { BuildComponent } from './build/build.component';

const routes: Routes = [{
  path: '',
  component: PagesComponent,
  children: [{
    path: 'dashboard',
    component: DashboardComponent,
  }, {
    path: 'projects',
    component: ProjectsComponent,
  }, {
    path: 'sdks',
    component: SdksComponent,
  }, {
    path: 'build',
    component: BuildComponent,
  }, {
    path: 'config',
    loadChildren: './config/config.module#ConfigModule',
  },
  {
    path: '',
    redirectTo: 'dashboard',
    pathMatch: 'full',
  }],
}];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class PagesRoutingModule {
}
