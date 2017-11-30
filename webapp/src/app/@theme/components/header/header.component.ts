import { Component, Input, OnInit } from '@angular/core';

import { NbMenuService, NbSidebarService } from '@nebular/theme';
// XDS_MODS
import { UserService } from '../../../@core-xds/services/users.service';
import { AnalyticsService } from '../../../@core/utils/analytics.service';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { AboutModalComponent } from '../../../pages/about/about-modal/about-modal.component';

@Component({
  selector: 'ngx-header',
  styleUrls: ['./header.component.scss'],
  templateUrl: './header.component.html',
})
export class HeaderComponent implements OnInit {


  @Input() position = 'normal';

  user: any;

  userMenu = [{ title: 'Profile' }, { title: 'Log out' }];

  // XDS_MODS - FIXME: better to define own XDS component instead of reuse nb-user
  helpName = '?';
  helpMenu = [
    {
      title: 'Online XDS documentation',
      target: '_blank',
      url: 'http://docs.automotivelinux.org/docs/devguides/en/dev/#xcross-development-system-user\'s-guide',
    },
    { title: 'About' },
  ];

  constructor(private sidebarService: NbSidebarService,
    private menuService: NbMenuService,
    private userService: UserService,
    private analyticsService: AnalyticsService,
    private modalService: NgbModal,
  ) {
  }

  ngOnInit() {
    // XDS_MODS
    this.userService.getUsers()
      .subscribe((users: any) => this.user = users.anonymous);
  }

  toggleSidebar(): boolean {
    this.sidebarService.toggle(true, 'menu-sidebar');
    return false;
  }

  toggleSettings(): boolean {
    this.sidebarService.toggle(false, 'settings-sidebar');
    return false;
  }

  goToHome() {
    this.menuService.navigateHome();
  }

  startSearch() {
    this.analyticsService.trackEvent('startSearch');
  }

  // XDS_MODS
  helpClick($event: any) {
    if ($event.title === 'About') {
        // FIXME SEB - move code in XDS part
        const activeModal = this.modalService.open(AboutModalComponent, { size: 'lg', container: 'nb-layout' });
    }

  }
}
