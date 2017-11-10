import { Component, ViewEncapsulation, AfterViewChecked, ElementRef, ViewChild, OnInit, Input } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { FormControl, FormGroup, Validators, FormBuilder } from '@angular/forms';
import { CookieService } from 'ngx-cookie';

import 'rxjs/add/operator/scan';
import 'rxjs/add/operator/startWith';

import { XDSAgentService, ICmdOutput } from '../../services/xdsagent.service';
import { ProjectService, IProject } from '../../services/project.service';
import { AlertService, IAlert } from '../../services/alert.service';
import { SdkService } from '../../services/sdk.service';

@Component({
    selector: 'xds-panel-build',
    templateUrl: './build.component.html',
    styleUrls: ['./build.component.css'],
encapsulation: ViewEncapsulation.None
})

export class BuildComponent implements OnInit, AfterViewChecked {
    @ViewChild('scrollOutput') private scrollContainer: ElementRef;

    @Input() curProject: IProject;

    public buildForm: FormGroup;
    public subpathCtrl = new FormControl('', Validators.required);
    public debugEnable = false;
    public buildIsCollapsed = false;
    public cmdOutput: string;
    public cmdInfo: string;

    private startTime: Map<string, number> = new Map<string, number>();

    constructor(
        private xdsSvr: XDSAgentService,
        private fb: FormBuilder,
        private alertSvr: AlertService,
        private sdkSvr: SdkService,
        private cookie: CookieService,
    ) {
        this.cmdOutput = '';
        this.cmdInfo = '';      // TODO: to be remove (only for debug)
        this.buildForm = fb.group({
            subpath: this.subpathCtrl,
            cmdClean: ['', Validators.nullValidator],
            cmdPrebuild: ['', Validators.nullValidator],
            cmdBuild: ['', Validators.nullValidator],
            cmdPopulate: ['', Validators.nullValidator],
            cmdArgs: ['', Validators.nullValidator],
            envVars: ['', Validators.nullValidator],
        });
    }

    ngOnInit() {
        // Set default settings
        // TODO save & restore values from cookies
        this.buildForm.patchValue({
            subpath: '',
            cmdClean: 'rm -rf build',
            cmdPrebuild: 'mkdir -p build && cd build && cmake ..',
            cmdBuild: 'cd build && make',
            cmdPopulate: 'cd build && make remote-target-populate',
            cmdArgs: '',
            envVars: '',
        });

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

        // only use for debug
        this.debugEnable = (this.cookie.get('debug_build') === '1');
    }

    ngAfterViewChecked() {
        this._scrollToBottom();
    }

    reset() {
        this.cmdOutput = '';
    }

    clean() {
        this._exec(
            this.buildForm.value.cmdClean,
            this.buildForm.value.subpath,
            [],
            this.buildForm.value.envVars);
    }

    preBuild() {
        this._exec(
            this.buildForm.value.cmdPrebuild,
            this.buildForm.value.subpath,
            [],
            this.buildForm.value.envVars);
    }

    build() {
        this._exec(
            this.buildForm.value.cmdBuild,
            this.buildForm.value.subpath,
            [],
            this.buildForm.value.envVars
        );
    }

    populate() {
        this._exec(
            this.buildForm.value.cmdPopulate,
            this.buildForm.value.subpath,
            [], // args
            this.buildForm.value.envVars
        );
    }

    execCmd() {
        this._exec(
            this.buildForm.value.cmdArgs,
            this.buildForm.value.subpath,
            [],
            this.buildForm.value.envVars
        );
    }

    private _exec(cmd: string, dir: string, args: string[], env: string) {
        if (!this.curProject) {
            this.alertSvr.warning('No active project', true);
        }

        const prjID = this.curProject.id;

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
