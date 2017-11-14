import { Component, ViewEncapsulation } from '@angular/core';

@Component({
    selector: 'app-topnav',
    templateUrl: './app-topnav.component.html',
    styleUrls: ['./app-topnav.component.css'],
    encapsulation: ViewEncapsulation.None
})
export class AppTopnavComponent {
    public isCollapsed = false;

    constructor() { }
}
