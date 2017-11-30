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

  execCmd(cmdName: string) {
    if (!this.isSetupValid()) {
      return this.alertSvr.warning('Please select first a valid project.', true);
    }

    if (!this.curPrj.uiSettings) {
      return this.alertSvr.warning('Invalid setting structure', true);
    }

    let cmd = '';
    switch (cmdName) {
      case 'clean':
        cmd = this.curPrj.uiSettings.cmdClean;
        break;
      case 'prebuild':
        cmd = this.curPrj.uiSettings.cmdPrebuild;
        break;
      case 'build':
        cmd = this.curPrj.uiSettings.cmdBuild;
        break;
      case 'populate':
        cmd = this.curPrj.uiSettings.cmdPopulate;
        break;
      case 'exec':
        if (this.curPrj.uiSettings.cmdArgs instanceof Array)  {
          cmd = this.curPrj.uiSettings.cmdArgs.join(' ');
        } else {
          cmd = this.curPrj.uiSettings.cmdArgs;
        }
        break;
      default:
        return this.alertSvr.warning('Unknown command name ' + cmdName);
    }

    const prjID = this.curPrj.id;
    const dir = this.curPrj.uiSettings.subpath;
    const args: string[] = [];
    const sdkid = this.sdkSvr.getCurrentId();

    let env = '';
    if (this.curPrj.uiSettings.envVars instanceof Array) {
      env = this.curPrj.uiSettings.envVars.join(' ');
    } else {
      env = this.curPrj.uiSettings.envVars;
    }

    this.cmdOutput += this._outputHeader();

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
