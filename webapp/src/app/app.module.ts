import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HttpModule } from "@angular/http";
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
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
import { Routing, AppRoutingProviders } from './app.routing';
import { AppComponent } from "./app.component";
import { AlertComponent } from './alert/alert.component';
import { ConfigComponent } from "./config/config.component";
import { DlXdsAgentComponent, CapitalizePipe } from "./config/downloadXdsAgent.component";
import { ProjectCardComponent } from "./projects/projectCard.component";
import { ProjectReadableTypePipe } from "./projects/projectCard.component";
import { ProjectsListAccordionComponent } from "./projects/projectsListAccordion.component";
import { ProjectAddModalComponent} from "./projects/projectAddModal.component";
import { SdkCardComponent } from "./sdks/sdkCard.component";
import { SdksListAccordionComponent } from "./sdks/sdksListAccordion.component";
import { SdkSelectDropdownComponent } from "./sdks/sdkSelectDropdown.component";
import { SdkAddModalComponent} from "./sdks/sdkAddModal.component";

import { HomeComponent } from "./home/home.component";
import { DevelComponent } from "./devel/devel.component";
import { BuildComponent } from "./devel/build/build.component";
import { XDSAgentService } from "./services/xdsagent.service";
import { ConfigService } from "./services/config.service";
import { ProjectService } from "./services/project.service";
import { AlertService } from './services/alert.service';
import { UtilsService } from './services/utils.service';
import { SdkService } from "./services/sdk.service";



@NgModule({
    imports: [
        BrowserModule,
        HttpModule,
        FormsModule,
        ReactiveFormsModule,
        Routing,
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
        AlertComponent,
        HomeComponent,
        BuildComponent,
        DevelComponent,
        ConfigComponent,
        DlXdsAgentComponent,
        CapitalizePipe,
        ProjectCardComponent,
        ProjectReadableTypePipe,
        ProjectsListAccordionComponent,
        ProjectAddModalComponent,
        SdkCardComponent,
        SdksListAccordionComponent,
        SdkSelectDropdownComponent,
        SdkAddModalComponent,
    ],
    providers: [
        AppRoutingProviders,
        {
            provide: Window,
            useValue: window
        },
        XDSAgentService,
        ConfigService,
        ProjectService,
        AlertService,
        UtilsService,
        SdkService,
    ],
    bootstrap: [AppComponent]
})
export class AppModule {
}
