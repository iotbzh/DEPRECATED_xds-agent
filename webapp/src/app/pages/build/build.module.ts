import { NgModule } from '@angular/core';
import { ThemeModule } from '../../@theme/theme.module';

import { BuildComponent } from './build.component';
import { BuildSettingsModalComponent } from './build-settings-modal/build-settings-modal.component';
import { ProjectSelectDropdownComponent } from './settings/project-select-dropdown.component';
import { SdkSelectDropdownComponent } from './settings/sdk-select-dropdown.component';

@NgModule({
  imports: [
    ThemeModule,
  ],
  declarations: [
    BuildComponent,
    BuildSettingsModalComponent,
    ProjectSelectDropdownComponent,
    SdkSelectDropdownComponent,
  ],
  entryComponents: [
    BuildSettingsModalComponent,
  ],
})
export class BuildModule { }
