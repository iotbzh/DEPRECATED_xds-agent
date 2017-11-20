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
