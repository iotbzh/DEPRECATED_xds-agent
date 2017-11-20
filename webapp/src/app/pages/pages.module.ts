import { NgModule } from '@angular/core';
import { ToasterModule } from 'angular2-toaster';

import { PagesComponent } from './pages.component';
import { DashboardModule } from './dashboard/dashboard.module';
import { BuildModule } from './build/build.module';
import { ProjectsModule } from './projects/projects.module';
import { SdksModule } from './sdks/sdks.module';
import { PagesRoutingModule } from './pages-routing.module';
import { NotificationsComponent } from './notifications/notifications.component';
import { ThemeModule } from '../@theme/theme.module';

const PAGES_COMPONENTS = [
  PagesComponent,
  NotificationsComponent,
];

@NgModule({
  imports: [
    PagesRoutingModule,
    ThemeModule,
    BuildModule,
    DashboardModule,
    ProjectsModule,
    SdksModule,
    ToasterModule,
  ],
  declarations: [
    ...PAGES_COMPONENTS,
  ],
})
export class PagesModule {
}
