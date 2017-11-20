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
