import { Component, Input, ViewChild, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { ModalDirective } from 'ngx-bootstrap/modal';
import { FormControl, FormGroup, Validators, FormBuilder, ValidatorFn, AbstractControl } from '@angular/forms';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/filter';
import 'rxjs/add/operator/debounceTime';

import { AlertService, IAlert } from '../services/alert.service';
import { ProjectService, IProject, ProjectType, ProjectTypes } from '../services/project.service';


@Component({
    selector: 'xds-project-add-modal',
    templateUrl: 'projectAddModal.component.html',
    styleUrls: ['projectAddModal.component.css']
})
export class ProjectAddModalComponent implements OnInit {
    @ViewChild('childProjectModal') public childProjectModal: ModalDirective;
    @Input() title?: string;
    @Input('server-id') serverID: string;

    cancelAction = false;
    userEditedLabel = false;
    projectTypes = ProjectTypes;

    addProjectForm: FormGroup;
    typeCtrl: FormControl;
    pathCliCtrl: FormControl;
    pathSvrCtrl: FormControl;

    constructor(
        private alert: AlertService,
        private projectSvr: ProjectService,
        private fb: FormBuilder
    ) {
        // Define types (first one is special/placeholder)
        this.projectTypes.unshift({ value: ProjectType.UNSET, display: '--Select a type--' });

        this.typeCtrl = new FormControl(this.projectTypes[0].value, Validators.pattern('[A-Za-z]+'));
        this.pathCliCtrl = new FormControl('', Validators.required);
        this.pathSvrCtrl = new FormControl({ value: '', disabled: true }, [Validators.required, Validators.minLength(1)]);

        this.addProjectForm = fb.group({
            type: this.typeCtrl,
            pathCli: this.pathCliCtrl,
            pathSvr: this.pathSvrCtrl,
            label: ['', Validators.nullValidator],
        });
    }

    ngOnInit() {
        // Auto create label name
        this.pathCliCtrl.valueChanges
            .debounceTime(100)
            .filter(n => n)
            .map(n => {
                const last = n.split('/');
                let nm = n;
                if (last.length > 0) {
                    nm = last.pop();
                    if (nm === '' && last.length > 0) {
                        nm = last.pop();
                    }
                }
                return 'Project_' + nm;
            })
            .subscribe(value => {
                if (value && !this.userEditedLabel) {
                    this.addProjectForm.patchValue({ label: value });
                }
            });

        // Handle disabling of Server path
        this.typeCtrl.valueChanges
            .debounceTime(500)
            .subscribe(valType => {
                const dis = (valType === String(ProjectType.SYNCTHING));
                this.pathSvrCtrl.reset({ value: '', disabled: dis });
            });
    }

    show() {
        this.cancelAction = false;
        this.userEditedLabel = false;
        this.childProjectModal.show();
    }

    hide() {
        this.childProjectModal.hide();
    }

    onKeyLabel(event: any) {
        this.userEditedLabel = (this.addProjectForm.value.label !== '');
    }

    /* FIXME: change input to file type
     <td><input type='file' id='select-local-path' webkitdirectory
     formControlName='pathCli' placeholder='myProject' (change)='onChangeLocalProject($event)'></td>

    onChangeLocalProject(e) {
        if e.target.files.length < 1 {
            console.log('NO files');
        }
        let dir = e.target.files[0].webkitRelativePath;
        console.log('files: ' + dir);
        let u = URL.createObjectURL(e.target.files[0]);
    }
    */
    onChangeLocalProject(e) {
    }

    onSubmit() {
        if (this.cancelAction) {
            return;
        }

        const formVal = this.addProjectForm.value;

        const type = formVal['type'].value;
        this.projectSvr.Add({
            serverId: this.serverID,
            label: formVal['label'],
            pathClient: formVal['pathCli'],
            pathServer: formVal['pathSvr'],
            type: formVal['type'],
            // FIXME: allow to set defaultSdkID from New Project config panel
        })
            .subscribe(prj => {
                this.alert.info('Project ' + prj.label + ' successfully created.');
                this.hide();

                // Reset Value for the next creation
                this.addProjectForm.reset();
                const selectedType = this.projectTypes[0].value;
                this.addProjectForm.patchValue({ type: selectedType });

            },
            err => {
                this.alert.error(err, 60);
                this.hide();
            });
    }

}
