/**
* @license
* Copyright (C) 2017 "IoT.bzh"
* Author Sebastien Douheret <sebastien@iot.bzh>
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*   http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

import { Component, Input, ViewChild, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { FormControl, FormGroup, Validators, ValidationErrors, FormBuilder, ValidatorFn, AbstractControl } from '@angular/forms';

// Import RxJs required methods
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/filter';
import 'rxjs/add/operator/debounceTime';

import { AlertService, IAlert } from '../../../@core-xds/services/alert.service';
import { ProjectService, IProject, ProjectType, ProjectTypes } from '../../../@core-xds/services/project.service';
import { XDSConfigService } from '../../../@core-xds/services/xds-config.service';


@Component({
  selector: 'xds-project-add-modal',
  templateUrl: 'project-add-modal.component.html',
  styleUrls: ['project-add-modal.component.scss'],
})
export class ProjectAddModalComponent implements OnInit {
  // @Input('server-id') serverID: string;
  private serverID: string;

  cancelAction = false;
  userEditedLabel = false;
  projectTypes = Object.assign([], ProjectTypes);

  addProjectForm: FormGroup;
  typeCtrl: FormControl;
  pathCliCtrl: FormControl;
  pathSvrCtrl: FormControl;

  constructor(
    private alert: AlertService,
    private projectSvr: ProjectService,
    private XdsConfigSvr: XDSConfigService,
    private fb: FormBuilder,
    private activeModal: NgbActiveModal,
  ) {
    // Define types (first one is special/placeholder)
    this.projectTypes.unshift({ value: ProjectType.UNSET, display: '--Select a type--' });

    this.typeCtrl = new FormControl(this.projectTypes[0].value, this.validatorProjType);
    this.pathCliCtrl = new FormControl('', this.validatorProjPath);
    this.pathSvrCtrl = new FormControl({ value: '', disabled: true }, this.validatorProjPath);

    this.addProjectForm = fb.group({
      type: this.typeCtrl,
      pathCli: this.pathCliCtrl,
      pathSvr: this.pathSvrCtrl,
      label: ['', Validators.nullValidator],
    });
  }


  ngOnInit() {
    // Update server ID
    this.serverID = this.XdsConfigSvr.getCurServer().id;
    this.XdsConfigSvr.onCurServer().subscribe(svr => this.serverID = svr.id);

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

  closeModal() {
    this.activeModal.close();
  }

  onKeyLabel(event: any) {
    this.userEditedLabel = (this.addProjectForm.value.label !== '');
  }

  onChangeLocalProject(e) {
  }

  onSubmit() {
    if (this.cancelAction) {
      return;
    }

    const formVal = this.addProjectForm.value;

    const type = formVal['type'].value;
    this.projectSvr.add({
      serverId: this.serverID,
      label: formVal['label'],
      pathClient: formVal['pathCli'],
      pathServer: formVal['pathSvr'],
      type: formVal['type'],
      // FIXME: allow to set defaultSdkID from New Project config panel
    })
      .subscribe(prj => {
        this.alert.info('Project ' + prj.label + ' successfully created.');
        this.closeModal();

        // Reset Value for the next creation
        this.addProjectForm.reset();
        const selectedType = this.projectTypes[0].value;
        this.addProjectForm.patchValue({ type: selectedType });

      },
      err => {
        this.alert.error(err, 60);
        this.closeModal();
      });
  }

  private validatorProjType(g: FormGroup): ValidationErrors | null {
    return (g.value !== ProjectType.UNSET) ? null : { validatorProjType: { valid: false } };
  }

  private validatorProjPath(g: FormGroup): ValidationErrors | null {
    return (g.disabled || g.value !== '') ? null : { validatorProjPath: { valid: false } };
  }

}
