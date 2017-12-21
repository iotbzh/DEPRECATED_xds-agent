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
import { ThemeModule } from '../../@theme/theme.module';

import { SdksComponent } from './sdks.component';
import { SdkCardComponent } from './sdk-card/sdk-card.component';
import { SdkManagementComponent } from './sdk-management/sdk-management.component';
import { SdkInstallComponent } from './sdk-management/sdk-install.component';
import { Ng2SmartTableModule } from 'ng2-smart-table';

@NgModule({
  imports: [
    ThemeModule,
    Ng2SmartTableModule,
  ],
  declarations: [
    SdksComponent,
    SdkCardComponent,
    SdkManagementComponent,
    SdkInstallComponent,
  ],
  entryComponents: [
    SdkInstallComponent,
  ],
})
export class SdksModule { }
