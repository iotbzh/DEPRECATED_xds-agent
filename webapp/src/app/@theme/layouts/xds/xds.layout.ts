import { Component, OnDestroy, EventEmitter } from '@angular/core';
import {
  NbMediaBreakpoint,
  NbMediaBreakpointsService,
  NbMenuItem,
  NbMenuService,
  NbSidebarService,
  NbThemeService,
} from '@nebular/theme';

import { StateService } from '../../../@core/data/state.service';

import { Subscription } from 'rxjs/Subscription';
import { Observable } from 'rxjs/Observable';
import 'rxjs/add/operator/withLatestFrom';
import 'rxjs/add/operator/delay';
import 'rxjs/add/observable/of';
import 'rxjs/add/operator/takeUntil';

// TODO: move layouts into the framework
@Component({
  selector: 'ngx-xds-layout',
  styleUrls: ['./xds.layout.scss'],
  templateUrl: './xds.layout.html',
})

export class XdsLayoutComponent implements OnDestroy {

  subMenu: NbMenuItem[] = [];
  layout: any = {};
  sidebar: any = {};
  sidebarCompact = true;

  protected layoutState$: Subscription;
  protected sidebarState$: Subscription;
  protected menuClick$: Subscription;

  private _mouseEnterStream: EventEmitter<any> = new EventEmitter();
  private _mouseLeaveStream: EventEmitter<any> = new EventEmitter();

  constructor(protected stateService: StateService,
    protected menuService: NbMenuService,
    protected themeService: NbThemeService,
    protected bpService: NbMediaBreakpointsService,
    protected sidebarService: NbSidebarService) {
    this.layoutState$ = this.stateService.onLayoutState()
      .subscribe((layout: string) => this.layout = layout);

    this.sidebarState$ = this.stateService.onSidebarState()
      .subscribe((sidebar: string) => {
        this.sidebar = sidebar;
      });

    const isBp = this.bpService.getByName('is');
    this.menuClick$ = this.menuService.onItemSelect()
      .withLatestFrom(this.themeService.onMediaQueryChange())
      .delay(20)
      .subscribe(([item, [bpFrom, bpTo]]: [any, [NbMediaBreakpoint, NbMediaBreakpoint]]) => {

        this.sidebarCompact = false;
        if (bpTo.width <= isBp.width) {
          this.sidebarService.collapse('menu-sidebar');
        }
      });

    // Set sidebarCompact according to sidebar state changes
    this.sidebarService.onToggle().subscribe(s => s.tag === 'menu-sidebar' && (this.sidebarCompact = !this.sidebarCompact));
    this.sidebarService.onCollapse().subscribe(s => s.tag === 'menu-sidebar' && (this.sidebarCompact = true));
    this.sidebarService.onExpand().subscribe(() => this.sidebarCompact = false);
    this.menuService.onSubmenuToggle().subscribe(i => i.item && i.item.expanded && (this.sidebarCompact = false));

    // Automatically expand sidebar on mouse over
    this._mouseEnterStream.flatMap(e => {
      return Observable
        .of(e)
        .delay(200)
        .takeUntil(this._mouseLeaveStream);
    })
      .subscribe(e => (this.sidebarCompact) && this.sidebarService.toggle(true, 'menu-sidebar'));

    // Automatically collapse sidebar on mouse leave
    this._mouseLeaveStream.flatMap(e => {
      return Observable
        .of(e)
        .delay(500)
        .takeUntil(this._mouseEnterStream);
    })
      .subscribe(e => this.sidebarService.toggle(true, 'menu-sidebar'));
  }

  onMouseEnter($event) {
    this._mouseEnterStream.emit($event);
  }

  onMouseLeave($event) {
    this._mouseLeaveStream.emit($event);
  }

  toogleSidebar() {
    this.sidebarService.toggle(true, 'menu-sidebar');
  }

  ngOnDestroy() {
    this.layoutState$.unsubscribe();
    this.sidebarState$.unsubscribe();
    this.menuClick$.unsubscribe();
  }
}
