import { NgModule } from '@angular/core';
import { ThemeModule } from '../../@theme/theme.module';

import { AboutModalComponent } from './about-modal/about-modal.component';

@NgModule({
  imports: [
    ThemeModule,
  ],
  declarations: [
    AboutModalComponent,
  ],
  entryComponents: [
    AboutModalComponent,
  ],
})
export class AboutModule { }
