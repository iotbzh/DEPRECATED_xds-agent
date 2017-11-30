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

import { ModuleWithProviders, NgModule, Optional, SkipSelf } from '@angular/core';
import { CommonModule } from '@angular/common';

import { NbAuthModule, NbDummyAuthProvider } from '@nebular/auth';
import { CookieModule } from 'ngx-cookie';

import { throwIfAlreadyLoaded } from './module-import-guard';
import { XdsServicesModule } from './services/@core-xds-services.module';
import { AnalyticsService } from '../@core/utils/analytics.service';
import { StateService } from '../@core/data/state.service';

const NB_COREXDS_PROVIDERS = [
  ...XdsServicesModule.forRoot().providers,
  ...NbAuthModule.forRoot({
    providers: {
      email: {
        service: NbDummyAuthProvider,
        config: {
          delay: 3000,
          login: {
            rememberMe: true,
          },
        },
      },
    },
  }).providers,
  AnalyticsService,
  StateService,
];

@NgModule({
  imports: [
    CommonModule,
    CookieModule.forRoot(),
  ],
  exports: [
    NbAuthModule,
  ],
  declarations: [],
})
export class CoreXdsModule {
  constructor( @Optional() @SkipSelf() parentModule: CoreXdsModule) {
    throwIfAlreadyLoaded(parentModule, 'CoreXdsModule');
  }

  static forRoot(): ModuleWithProviders {
    return <ModuleWithProviders>{
      ngModule: CoreXdsModule,
      providers: [
        ...NB_COREXDS_PROVIDERS,
      ],
    };
  }
}
