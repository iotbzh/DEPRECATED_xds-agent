import { NgModule } from '@angular/core';
import { ThemeModule } from '../../@theme/theme.module';

import { SdksComponent } from './sdks.component';
import { SdkCardComponent } from './sdk-card/sdk-card.component';
// import { SdkAddModalComponent } from './sdk-add-modal/sdk-add-modal.component';


@NgModule({
  imports: [
    ThemeModule,
  ],
  declarations: [
    SdksComponent,
    SdkCardComponent,
    // SdkAddModalComponent,
  ],
  entryComponents: [
    // SdkAddModalComponent,
  ],
})
export class SdksModule { }
