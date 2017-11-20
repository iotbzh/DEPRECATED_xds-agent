import { Component } from '@angular/core';

import { MENU_ITEMS } from './pages-menu';

@Component({
  selector: 'ngx-pages',
  template: `
    <ngx-notifications></ngx-notifications>
    <ngx-xds-layout>
      <nb-menu [items]="menu"></nb-menu>
      <router-outlet></router-outlet>
    </ngx-xds-layout>
  `,
})
export class PagesComponent {

  menu = MENU_ITEMS;
}
