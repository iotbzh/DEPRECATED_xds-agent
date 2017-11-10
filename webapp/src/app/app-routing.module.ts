import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { HomeComponent } from './home/home.component';
import { ConfigComponent } from './config/config.component';
import { DevelComponent } from './devel/devel.component';

const routes: Routes = [
    { path: 'config', component: ConfigComponent, data: { title: 'Config' } },
    { path: 'home', component: HomeComponent, data: { title: 'Home' } },
    { path: 'devel', component: DevelComponent, data: { title: 'Build & Deploy' } },
    { path: '**', component: HomeComponent }
];

@NgModule({
    imports: [RouterModule.forRoot(routes)],
    exports: [RouterModule]
})
export class AppRoutingModule { }
