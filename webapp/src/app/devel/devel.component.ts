import { Component, OnInit, ViewEncapsulation } from '@angular/core';
import { Observable } from 'rxjs/Observable';

import { ProjectService, IProject } from '../services/project.service';

@Component({
    selector: 'xds-devel',
    templateUrl: './devel.component.html',
    styleUrls: ['./devel.component.css'],
  encapsulation: ViewEncapsulation.None
})

export class DevelComponent implements OnInit {

    curPrj: IProject;
    Prjs$: Observable<IProject[]>;

    constructor(private projectSvr: ProjectService) {
    }

    ngOnInit() {
        this.Prjs$ = this.projectSvr.Projects$;
        this.Prjs$.subscribe((prjs) => {
            // Select project if no one is selected or no project exists
            if (this.curPrj && 'id' in this.curPrj) {
                this.curPrj = prjs.find(p => p.id === this.curPrj.id) || prjs[0];
            } else if (this.curPrj == null) {
                this.curPrj = prjs[0];
            } else {
                this.curPrj = null;
            }
        });
    }
}
