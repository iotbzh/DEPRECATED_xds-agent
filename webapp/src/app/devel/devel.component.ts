import { Component } from '@angular/core';

import { Observable } from 'rxjs';

import { ProjectService, IProject } from "../services/project.service";

@Component({
    selector: 'devel',
    moduleId: module.id,
    templateUrl: './devel.component.html',
    styleUrls: ['./devel.component.css'],
})

export class DevelComponent {

    curPrj: IProject;
    Prjs$: Observable<IProject[]>;

    constructor(private projectSvr: ProjectService) {
    }

    ngOnInit() {
        this.Prjs$ = this.projectSvr.Projects$;
        this.Prjs$.subscribe((prjs) => {
            // Select project if no one is selected or no project exists
            if (this.curPrj && "id" in this.curPrj) {
                this.curPrj = prjs.find(p => p.id === this.curPrj.id) || prjs[0];
            } else if (this.curPrj == null) {
                this.curPrj = prjs[0];
            } else {
                this.curPrj = null;
            }
        });
    }
}
