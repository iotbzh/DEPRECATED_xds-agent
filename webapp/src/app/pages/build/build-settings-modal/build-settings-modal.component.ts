import { Component, Input, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { FormControl, FormGroup, Validators, ValidationErrors, FormBuilder, ValidatorFn, AbstractControl } from '@angular/forms';

import { AlertService } from '../../../@core-xds/services/alert.service';
import { ProjectService, IProject } from '../../../@core-xds/services/project.service';


@Component({
  selector: 'xds-build-settings-modal',
  templateUrl: 'build-settings-modal.component.html',
})

export class BuildSettingsModalComponent implements OnInit {
  // @Input('server-id') serverID: string;
  private serverID: string;

  closeAction = false;
  userEditedLabel = false;

  settingsProjectForm: FormGroup;
  subpathCtrl = new FormControl('', Validators.nullValidator);

  private curPrj: IProject;

  constructor(
    private alert: AlertService,
    private projectSvr: ProjectService,
    private fb: FormBuilder,
    private activeModal: NgbActiveModal,
  ) {
    this.settingsProjectForm = fb.group({
      subpath: this.subpathCtrl,
      cmdClean: ['', Validators.required],
      cmdPrebuild: ['', Validators.nullValidator],
      cmdBuild: ['', Validators.required],
      cmdPopulate: ['', Validators.nullValidator],
      cmdArgs: ['', Validators.nullValidator],
      envVars: ['', Validators.nullValidator],
    });
  }

  ngOnInit() {
    this.curPrj = this.projectSvr.getCurrent();
    this.settingsProjectForm.patchValue(this.curPrj.uiSettings);
  }

  closeModal() {
    this.activeModal.close();
  }

  resetDefault() {
    this.settingsProjectForm.patchValue(this.projectSvr.getDefaultSettings());
  }

  onSubmit() {
    if (!this.closeAction) {
      return;
    }

    this.curPrj.uiSettings = this.settingsProjectForm.value;
    this.projectSvr.setSettings(this.curPrj)
    .subscribe(prj => {
      this.alert.info('Settings of project "' + prj.label + '" successfully updated.');
      this.closeModal();

      // Reset Value for the next creation
      this.settingsProjectForm.reset();
    },
    err => {
      this.alert.error(err, 60);
      this.closeModal();
    });
  }

}
