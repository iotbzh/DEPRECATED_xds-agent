import { Injectable, SecurityContext, isDevMode } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { BehaviorSubject } from 'rxjs/BehaviorSubject';

import { XDSAgentService, IXDSProjectConfig } from '../services/xdsagent.service';

/* FIXME: syntax only compatible with TS>2.4.0
export enum ProjectType {
    UNSET = '',
    NATIVE_PATHMAP = 'PathMap',
    SYNCTHING = 'CloudSync'
}
*/
export type ProjectTypeEnum = '' | 'PathMap' | 'CloudSync';
export const ProjectType = {
  UNSET: '',
  NATIVE_PATHMAP: 'PathMap',
  SYNCTHING: 'CloudSync',
};

export const ProjectTypes = [
  { value: ProjectType.NATIVE_PATHMAP, display: 'Path mapping' },
  { value: ProjectType.SYNCTHING, display: 'Cloud Sync' },
];

export const ProjectStatus = {
  ErrorConfig: 'ErrorConfig',
  Disable: 'Disable',
  Enable: 'Enable',
  Pause: 'Pause',
  Syncing: 'Syncing',
};

export interface IUISettings {
  subpath: string;
  cmdClean: string;
  cmdPrebuild: string;
  cmdBuild: string;
  cmdPopulate: string;
  cmdArgs: string[];
  envVars: string[];
}
export interface IProject {
  id?: string;
  serverId: string;
  label: string;
  pathClient: string;
  pathServer?: string;
  type: ProjectTypeEnum;
  status?: string;
  isInSync?: boolean;
  isUsable?: boolean;
  serverPrjDef?: IXDSProjectConfig;
  isExpanded?: boolean;
  visible?: boolean;
  defaultSdkID?: string;
  uiSettings?: IUISettings;
}

const defaultUISettings: IUISettings = {
  subpath: '',
  cmdClean: 'rm -rf build && echo Done',
  cmdPrebuild: 'mkdir -p build && cd build && cmake ..',
  cmdBuild: 'cd build && make',
  cmdPopulate: 'cd build && make remote-target-populate',
  cmdArgs: [],
  envVars: [],
};

@Injectable()
export class ProjectService {
  projects$: Observable<IProject[]>;
  curProject$: Observable<IProject>;

  private _prjsList: IProject[] = [];
  private prjsSubject = <BehaviorSubject<IProject[]>>new BehaviorSubject(this._prjsList);
  private _current: IProject;
  private curPrjSubject = <BehaviorSubject<IProject>>new BehaviorSubject(this._current);

  constructor(private xdsSvr: XDSAgentService) {
    this._current = null;
    this.projects$ = this.prjsSubject.asObservable();
    this.curProject$ = this.curPrjSubject.asObservable();

    // Load initial projects list
    this.xdsSvr.getProjects().subscribe((projects) => {
      this._prjsList = [];
      projects.forEach(p => {
        this._addProject(p, true);
      });

      // TODO: get previous val from xds-config service / cookie
      if (this._prjsList.length > 0) {
        this._current = this._prjsList[0];
        this.curPrjSubject.next(this._current);
      }

      this.prjsSubject.next(this._prjsList);
    });

    // Add listener on projects creation, deletion and change events
    this.xdsSvr.onProjectAdd().subscribe(prj => this._addProject(prj));
    this.xdsSvr.onProjectDelete().subscribe(prj => this._delProject(prj));
    this.xdsSvr.onProjectChange().subscribe(prj => this._updateProject(prj));
  }

  setCurrent(p: IProject): IProject | undefined {
    if (!p) {
      this._current = null;
      return undefined;
    }
    return this.setCurrentById(p.id);
  }

  setCurrentById(id: string): IProject | undefined {
    const p = this._prjsList.find(item => item.id === id);
    if (p) {
      this._current = p;
      this.curPrjSubject.next(this._current);
    }
    return this._current;
  }

  getCurrent(): IProject {
    return this._current;
  }

  add(prj: IProject): Observable<IProject> {
    // Send config to XDS server
    return this.xdsSvr.addProject(this._convToIXdsProject(prj))
      .map(xp => this._convToIProject(xp));
  }

  delete(prj: IProject): Observable<IProject> {
    const idx = this._getProjectIdx(prj.id);
    const delPrj = prj;
    if (idx === -1) {
      throw new Error('Invalid project id (id=' + prj.id + ')');
    }
    return this.xdsSvr.deleteProject(prj.id)
      .map(res => delPrj);
  }

  sync(prj: IProject): Observable<string> {
    const idx = this._getProjectIdx(prj.id);
    if (idx === -1) {
      throw new Error('Invalid project id (id=' + prj.id + ')');
    }
    return this.xdsSvr.syncProject(prj.id);
  }

  setSettings(prj: IProject): Observable<IProject> {
    return this.xdsSvr.updateProject(this._convToIXdsProject(prj))
      .map(xp => this._convToIProject(xp));
  }

  getDefaultSettings(): IUISettings {
    return defaultUISettings;
  }

  /***  Private functions  ***/

  private _isUsableProject(p) {
    return p && p.isInSync &&
      (p.status === ProjectStatus.Enable) &&
      (p.status !== ProjectStatus.Syncing);
  }

  private _getProjectIdx(id: string): number {
    return this._prjsList.findIndex((item) => item.id === id);
  }


  private _convToIXdsProject(prj: IProject): IXDSProjectConfig {
    const xPrj: IXDSProjectConfig = {
      id: prj.id || '',
      serverId: prj.serverId,
      label: prj.label || '',
      clientPath: prj.pathClient.trim(),
      serverPath: prj.pathServer,
      type: prj.type,
      defaultSdkID: prj.defaultSdkID,
      clientData: JSON.stringify(prj.uiSettings || defaultUISettings),
    };
    return xPrj;
  }

  private _convToIProject(rPrj: IXDSProjectConfig): IProject {
    let settings = defaultUISettings;
    if (rPrj.clientData && rPrj.clientData !== '') {
      settings = JSON.parse(rPrj.clientData);
    }

    // Convert XDSFolderConfig to IProject
    const pp: IProject = {
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
      uiSettings: settings,
    };
    return pp;
  }

  private _addProject(prj: IXDSProjectConfig, noNext?: boolean): IProject {

    // Convert XDSFolderConfig to IProject
    const pp = this._convToIProject(prj);

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
      this.prjsSubject.next(this._prjsList);
    }

    return pp;
  }

  private _delProject(prj: IXDSProjectConfig) {
    const idx = this._prjsList.findIndex(item => item.id === prj.id);
    if (idx === -1) {
      if (isDevMode) {
        /* tslint:disable:no-console */
        console.log('Warning: Try to delete project unknown id: prj=', prj);
      }
      return;
    }
    const delId = this._prjsList[idx].id;
    this._prjsList.splice(idx, 1);
    if (this._prjsList[idx].id === this._current.id) {
      this.setCurrent(this._prjsList[0]);
    }
    this.prjsSubject.next(this._prjsList);
  }

  private _updateProject(prj: IXDSProjectConfig) {
    const i = this._getProjectIdx(prj.id);
    if (i >= 0) {
      // XXX for now, only isInSync and status may change
      this._prjsList[i].isInSync = prj.isInSync;
      this._prjsList[i].status = prj.status;
      this._prjsList[i].isUsable = this._isUsableProject(prj);
      this.prjsSubject.next(this._prjsList);
    }
  }

}
