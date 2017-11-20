import { NgModule } from '@angular/core';
import { ThemeModule } from '../../@theme/theme.module';

import { BuildComponent } from './build.component';
import { ProjectSelectDropdownComponent } from './settings/project-select-dropdown.component';
import { SdkSelectDropdownComponent } from './settings/sdk-select-dropdown.component';

@NgModule({
  imports: [
    ThemeModule,
  ],
  declarations: [
    BuildComponent,
    ProjectSelectDropdownComponent,
    SdkSelectDropdownComponent,
  ],
  entryComponents: [
  ],
})
export class BuildModule { }
