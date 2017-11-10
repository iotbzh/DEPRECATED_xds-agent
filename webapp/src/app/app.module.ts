import { NgModule } from '@angular/core';
import { HttpClientModule, HttpClient } from '@angular/common/http';
import { BrowserModule } from '@angular/platform-browser';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { TranslateModule, TranslateLoader } from '@ngx-translate/core';
import { TranslateHttpLoader } from '@ngx-translate/http-loader';
import { FileUploadModule } from 'ng2-file-upload';
import { LocationStrategy, HashLocationStrategy } from '@angular/common';
import { CookieModule } from 'ngx-cookie';

// Import bootstrap
import { AlertModule } from 'ngx-bootstrap/alert';
import { ModalModule } from 'ngx-bootstrap/modal';
import { AccordionModule } from 'ngx-bootstrap/accordion';
import { CarouselModule } from 'ngx-bootstrap/carousel';
import { PopoverModule } from 'ngx-bootstrap/popover';
import { CollapseModule } from 'ngx-bootstrap/collapse';
import { BsDropdownModule } from 'ngx-bootstrap/dropdown';

// Import the application components and services.
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { AlertComponent } from './alert/alert.component';
import { HomeComponent } from './home/home.component';
import { ConfigComponent } from './config/config.component';
import { DwnlAgentComponent } from './config/downloadXdsAgent.component';
import { DevelComponent } from './devel/devel.component';
import { BuildComponent } from './devel/build/build.component';
import { ProjectCardComponent } from './projects/projectCard.component';
import { ProjectReadableTypePipe } from './projects/projectCard.component';
import { ProjectsListAccordionComponent } from './projects/projectsListAccordion.component';
import { ProjectAddModalComponent } from './projects/projectAddModal.component';
import { SdkCardComponent } from './sdks/sdkCard.component';
import { SdksListAccordionComponent } from './sdks/sdksListAccordion.component';
import { SdkSelectDropdownComponent } from './sdks/sdkSelectDropdown.component';
import { SdkAddModalComponent } from './sdks/sdkAddModal.component';

import { AlertService } from './services/alert.service';
import { ConfigService } from './services/config.service';
import { ProjectService } from './services/project.service';
import { SdkService } from './services/sdk.service';
import { UtilsService } from './services/utils.service';
import { XDSAgentService } from './services/xdsagent.service';

import { SafePipe } from './common/safe.pipe';

export function createTranslateLoader(http: HttpClient) {
    return new TranslateHttpLoader(http, './assets/i18n/', '.json');
}

@NgModule({
    imports: [
        BrowserModule,
        FormsModule,
        ReactiveFormsModule,
        HttpClientModule,
        AppRoutingModule,
        FileUploadModule,
        TranslateModule.forRoot({
            loader: {
                provide: TranslateLoader,
                useFactory: (createTranslateLoader),
                deps: [HttpClient]
            }
        }),
        CookieModule.forRoot(),
        AlertModule.forRoot(),
        ModalModule.forRoot(),
        AccordionModule.forRoot(),
        CarouselModule.forRoot(),
        PopoverModule.forRoot(),
        CollapseModule.forRoot(),
        BsDropdownModule.forRoot(),
    ],
    declarations: [
        AppComponent,
        HomeComponent,
        AlertComponent,
        ConfigComponent,
        DwnlAgentComponent,
        DevelComponent,
        BuildComponent,
        ProjectCardComponent,
        ProjectReadableTypePipe,
        ProjectsListAccordionComponent,
        ProjectAddModalComponent,
        SdkCardComponent,
        SdksListAccordionComponent,
        SdkSelectDropdownComponent,
        SdkAddModalComponent,
        SafePipe
    ],
    providers: [
        {
            provide: LocationStrategy, useClass: HashLocationStrategy,
        },
        AlertService,
        ConfigService,
        ProjectService,
        SdkService,
        UtilsService,
        XDSAgentService
    ],
    bootstrap: [AppComponent]
})
export class AppModule {
    constructor() { }
}
