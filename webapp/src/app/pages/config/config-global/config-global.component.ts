import { Component, OnInit } from '@angular/core';

import { ConfigService, IConfig } from '../../../@core-xds/services/config.service';

@Component({
  selector: 'xds-config-global',
  styleUrls: ['./config-global.component.scss'],
  templateUrl: './config-global.component.html',
})
export class ConfigGlobalComponent implements OnInit {

  public configFormChanged = false;

  constructor(
    private configSvr: ConfigService,
  ) {
  }

  ngOnInit() {
  }

  onSubmit() {
  }
}

