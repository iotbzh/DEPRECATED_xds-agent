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

import { NgModule, ModuleWithProviders } from '@angular/core';
import { CommonModule } from '@angular/common';

import { AlertService } from './alert.service';
import { ConfigService } from './config.service';
import { ProjectService } from './project.service';
import { SdkService } from './sdk.service';
import { UserService } from './users.service';
import { XDSConfigService } from './xds-config.service';
import { XDSAgentService } from './xdsagent.service';

const SERVICES = [
  AlertService,
  ConfigService,
  ProjectService,
  SdkService,
  UserService,
  XDSConfigService,
  XDSAgentService,
];

@NgModule({
  imports: [
    CommonModule,
  ],
  providers: [
    ...SERVICES,
  ],
})
export class XdsServicesModule {
  static forRoot(): ModuleWithProviders {
    return <ModuleWithProviders>{
      ngModule: XdsServicesModule,
      providers: [
        ...SERVICES,
      ],
    };
  }
}
