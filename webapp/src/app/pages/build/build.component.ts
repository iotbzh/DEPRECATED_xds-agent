import { Component, ViewEncapsulation, AfterViewChecked, ElementRef, ViewChild, OnInit, Input } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';

import 'rxjs/add/operator/scan';
import 'rxjs/add/operator/startWith';

import { BuildSettingsModalComponent } from './build-settings-modal/build-settings-modal.component';

import { XDSAgentService, ICmdOutput } from '../../@core-xds/services/xdsagent.service';
import { ProjectService, IProject } from '../../@core-xds/services/project.service';
import { AlertService, IAlert } from '../../@core-xds/services/alert.service';
import { SdkService } from '../../@core-xds/services/sdk.service';

@Component({
  selector: 'xds-panel-build',
  templateUrl: './build.component.html',
  styleUrls: ['./build.component.scss'],
  encapsulation: ViewEncapsulation.None,
})

export class BuildComponent implements OnInit, AfterViewChecked {
  @ViewChild('scrollOutput') private scrollContainer: ElementRef;

  // FIXME workaround of https://github.com/angular/angular-cli/issues/2034
  // should be removed with angular 5
  //  @Input() curProject: IProject;
  @Input() curProject = <IProject>null;

  public buildIsCollapsed = false;
  public cmdOutput: string;
  public cmdInfo: string;
  public curPrj: IProject;

  private startTime: Map<string, number> = new Map<string, number>();

  constructor(
    private prjSvr: ProjectService,
    private xdsSvr: XDSAgentService,
    private alertSvr: AlertService,
    private sdkSvr: SdkService,
    private modalService: NgbModal,
  ) {
    this.cmdOutput = '';
    this.cmdInfo = '';       // TODO: to be remove (only for debug)
  }

  ngOnInit() {
    // Retreive current project
    this.prjSvr.curProject$.subscribe(p => this.curPrj = p);

    // Command output data tunneling
    this.xdsSvr.CmdOutput$.subscribe(data => {
      this.cmdOutput += data.stdout;
      this.cmdOutput += data.stderr;
    });

    // Command exit
    this.xdsSvr.CmdExit$.subscribe(exit => {
      if (this.startTime.has(exit.cmdID)) {
        this.cmdInfo = 'Last command duration: ' + this._computeTime(this.startTime.get(exit.cmdID));
        this.startTime.delete(exit.cmdID);
      }

      if (exit && exit.code !== 0) {
        this.cmdOutput += '--- Command exited with code ' + exit.code + ' ---\n\n';
      }
    });

    this._scrollToBottom();
  }

  ngAfterViewChecked() {
    this._scrollToBottom();
  }

  resetOutput() {
    this.cmdOutput = '';
  }

  isSetupValid(): boolean {
    return (typeof this.curPrj !== 'undefined');
  }

  settingsShow() {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('Please select first a valid project.', true);
    }

    const activeModal = this.modalService.open(BuildSettingsModalComponent, { size: 'lg', container: 'nb-layout' });
    activeModal.componentInstance.modalHeader = 'Large Modal';
  }

  clean() {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('Please select first a valid project.', true);
    }
    this._exec(
      this.curPrj.uiSettings.cmdClean,
      this.curPrj.uiSettings.subpath,
      [],
      this.curPrj.uiSettings.envVars.join(' '));
  }

  preBuild() {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('Please select first a valid project.', true);
    }
    this._exec(
      this.curPrj.uiSettings.cmdPrebuild,
      this.curPrj.uiSettings.subpath,
      [],
      this.curPrj.uiSettings.envVars.join(' '));
  }

  build() {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('Please select first a valid project.', true);
    }
    this._exec(
      this.curPrj.uiSettings.cmdBuild,
      this.curPrj.uiSettings.subpath,
      [],
      this.curPrj.uiSettings.envVars.join(' '),
    );
  }

  populate() {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('Please select first a valid project.', true);
    }
    this._exec(
      this.curPrj.uiSettings.cmdPopulate,
      this.curPrj.uiSettings.subpath,
      [], // args
      this.curPrj.uiSettings.envVars.join(' '),
    );
  }

  execCmd() {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('Please select first a valid project.', true);
    }
    this._exec(
      this.curPrj.uiSettings.cmdArgs.join(' '),
      this.curPrj.uiSettings.subpath,
      [],
      this.curPrj.uiSettings.envVars.join(' '),
    );
  }

  private _exec(cmd: string, dir: string, args: string[], env: string) {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('No active project', true);
    }
    const prjID = this.curPrj.id;

    this.cmdOutput += this._outputHeader();

    const sdkid = this.sdkSvr.getCurrentId();

    // Detect key=value in env string to build array of string
    const envArr = [];
    env.split(';').forEach(v => envArr.push(v.trim()));

    const t0 = performance.now();
    this.cmdInfo = 'Start build of ' + prjID + ' at ' + t0;

    this.xdsSvr.exec(prjID, dir, cmd, sdkid, args, envArr)
      .subscribe(res => {
        this.startTime.set(String(res.cmdID), t0);
      },
      err => {
        this.cmdInfo = 'Last command duration: ' + this._computeTime(t0);
        this.alertSvr.error('ERROR: ' + err);
      });
  }

  private _scrollToBottom(): void {
    try {
      this.scrollContainer.nativeElement.scrollTop = this.scrollContainer.nativeElement.scrollHeight;
    } catch (err) { }
  }

  private _computeTime(t0: number, t1?: number): string {
    const enlap = Math.round((t1 || performance.now()) - t0);
    if (enlap < 1000.0) {
      return enlap.toFixed(2) + ' ms';
    } else {
      return (enlap / 1000.0).toFixed(3) + ' seconds';
    }
  }

  private _outputHeader(): string {
    return '--- ' + new Date().toString() + ' ---\n';
  }

  private _outputFooter(): string {
    return '\n';
  }
}
