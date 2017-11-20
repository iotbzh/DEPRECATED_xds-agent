import { NgModule } from '@angular/core';

import { ThemeModule } from '../../@theme/theme.module';
import { ConfigRoutingModule, routedConfig } from './config-routing.module';

@NgModule({
  imports: [
    ThemeModule,
    ConfigRoutingModule,
  ],
  declarations: [
    ...routedConfig,
  ]
})
export class ConfigModule { }
