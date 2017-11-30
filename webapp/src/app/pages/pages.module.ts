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

import { NgModule } from '@angular/core';
import { ToasterModule } from 'angular2-toaster';

import { PagesComponent } from './pages.component';
import { AboutModule } from './about/about.module';
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
    AboutModule,
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
