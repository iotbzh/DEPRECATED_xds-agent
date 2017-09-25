import { Injectable, SecurityContext } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

import { XDSAgentService, IXDSProjectConfig } from "../services/xdsagent.service";

export enum ProjectType {
    UNSET = "",
    NATIVE_PATHMAP = "PathMap",
    SYNCTHING = "CloudSync"
}

export var ProjectTypes = [
    { value: ProjectType.NATIVE_PATHMAP, display: "Path mapping" },
    { value: ProjectType.SYNCTHING, display: "Cloud Sync" }
];

export var ProjectStatus = {
    ErrorConfig: "ErrorConfig",
    Disable: "Disable",
    Enable: "Enable",
    Pause: "Pause",
    Syncing: "Syncing"
};

export interface IProject {
    id?: string;
    serverId: string;
    label: string;
    pathClient: string;
    pathServer?: string;
    type: ProjectType;
    status?: string;
    isInSync?: boolean;
    isUsable?: boolean;
    serverPrjDef?: IXDSProjectConfig;
    isExpanded?: boolean;
    visible?: boolean;
    defaultSdkID?: string;
}

@Injectable()
export class ProjectService {
    public Projects$: Observable<IProject[]>;

    private _prjsList: IProject[] = [];
    private current: IProject;
    private prjsSubject = <BehaviorSubject<IProject[]>>new BehaviorSubject(this._prjsList);

    constructor(private xdsSvr: XDSAgentService) {
        this.current = null;
        this.Projects$ = this.prjsSubject.asObservable();

        this.xdsSvr.getProjects().subscribe((projects) => {
            this._prjsList = [];
            projects.forEach(p => {
                this._addProject(p, true);
            });
            this.prjsSubject.next(Object.assign([], this._prjsList));
        });

        // Update Project data
        this.xdsSvr.ProjectState$.subscribe(prj => {
            let i = this._getProjectIdx(prj.id);
            if (i >= 0) {
                // XXX for now, only isInSync and status may change
                this._prjsList[i].isInSync = prj.isInSync;
                this._prjsList[i].status = prj.status;
                this._prjsList[i].isUsable = this._isUsableProject(prj);
                this.prjsSubject.next(Object.assign([], this._prjsList));
            }
        });

        // Add listener on create and delete project events
        this.xdsSvr.addEventListener('event:project-add', (ev) => {
            if (ev && ev.data && ev.data.id) {
                this._addProject(ev.data);
            } else {
                console.log("Warning: received events with unknown data: ev=", ev);
            }
        });
        this.xdsSvr.addEventListener('event:project-delete', (ev) => {
            if (ev && ev.data && ev.data.id) {
                let idx = this._prjsList.findIndex(item => item.id === ev.data.id);
                if (idx === -1) {
                    console.log("Warning: received events on unknown project id: ev=", ev);
                    return;
                }
                this._prjsList.splice(idx, 1);
                this.prjsSubject.next(Object.assign([], this._prjsList));
            } else {
                console.log("Warning: received events with unknown data: ev=", ev);
            }
        });

    }

    public setCurrent(s: IProject) {
        this.current = s;
    }

    public getCurrent(): IProject {
        return this.current;
    }

    public getCurrentId(): string {
        if (this.current && this.current.id) {
            return this.current.id;
        }
        return "";
    }

    Add(prj: IProject): Observable<IProject> {
        let xdsPrj: IXDSProjectConfig = {
            id: "",
            serverId: prj.serverId,
            label: prj.label || "",
            clientPath: prj.pathClient.trim(),
            serverPath: prj.pathServer,
            type: prj.type,
            defaultSdkID: prj.defaultSdkID,
        };
        // Send config to XDS server
        return this.xdsSvr.addProject(xdsPrj)
            .map(xdsPrj => this._convToIProject(xdsPrj));
    }

    Delete(prj: IProject): Observable<IProject> {
        let idx = this._getProjectIdx(prj.id);
        let delPrj = prj;
        if (idx === -1) {
            throw new Error("Invalid project id (id=" + prj.id + ")");
        }
        return this.xdsSvr.deleteProject(prj.id)
            .map(res => { return delPrj; });
    }

    Sync(prj: IProject): Observable<string> {
        let idx = this._getProjectIdx(prj.id);
        if (idx === -1) {
            throw new Error("Invalid project id (id=" + prj.id + ")");
        }
        return this.xdsSvr.syncProject(prj.id);
    }

    private _isUsableProject(p) {
        return p && p.isInSync &&
            (p.status === ProjectStatus.Enable) &&
            (p.status !== ProjectStatus.Syncing);
    }

    private _getProjectIdx(id: string): number {
        return this._prjsList.findIndex((item) => item.id === id);
    }

    private _convToIProject(rPrj: IXDSProjectConfig): IProject {
        // Convert XDSFolderConfig to IProject
        let pp: IProject = {
            id: rPrj.id,
            serverId: rPrj.serverId,
            label: rPrj.label,
            pathClient: rPrj.clientPath,
            pathServer: rPrj.serverPath,
            type: rPrj.type,
            status: rPrj.status,
            isInSync: rPrj.isInSync,
            isUsable: this._isUsableProject(rPrj),
            defaultSdkID: rPrj.defaultSdkID,
            serverPrjDef: Object.assign({}, rPrj),  // do a copy
        };
        return pp;
    }

    private _addProject(rPrj: IXDSProjectConfig, noNext?: boolean): IProject {

        // Convert XDSFolderConfig to IProject
        let pp = this._convToIProject(rPrj);

        // add new project
        this._prjsList.push(pp);

        // sort project array
        this._prjsList.sort((a, b) => {
            if (a.label < b.label) {
                return -1;
            }
            if (a.label > b.label) {
                return 1;
            }
            return 0;
        });

        if (!noNext) {
            this.prjsSubject.next(Object.assign([], this._prjsList));
        }

        return pp;
    }
}
