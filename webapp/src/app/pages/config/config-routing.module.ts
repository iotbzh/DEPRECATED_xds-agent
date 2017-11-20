import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';

import { ConfigComponent } from './config.component';
import { ConfigGlobalComponent } from './config-global/config-global.component';
import { ConfigXdsComponent } from './config-xds/config-xds.component';

const routes: Routes = [{
  path: '',
  component: ConfigComponent,
  children: [
    {
      path: 'global',
      component: ConfigGlobalComponent,
    }, {
      path: 'xds',
      component: ConfigXdsComponent,
    },
  ],
}];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule],
})
export class ConfigRoutingModule { }

export const routedConfig = [
  ConfigComponent,
  ConfigGlobalComponent,
  ConfigXdsComponent,
];
