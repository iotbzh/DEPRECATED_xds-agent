import { Component } from '@angular/core';

@Component({
    selector: 'xds-dwnl-agent',
    template: `
        <template #popTemplate>
            <h3>Install xds-agent:</h3>
            <ul>
                <li>On Linux machine <a href="{{url_OS_Linux}}" target="_blank">
                <span class="fa fa-external-link"></span></a></li>

                <li>On Windows machine <a href="{{url_OS_Other}}" target="_blank"><span class="fa fa-external-link"></span></a></li>

                <li>On MacOS machine <a href="{{url_OS_Other}}" target="_blank"><span class="fa fa-external-link"></span></a></li>
            </ul>
            <button type="button" class="btn btn-sm" (click)="pop.hide()"> Cancel </button>
        </template>
        <button type="button" class="btn btn-link fa fa-download fa-size-x2"
            [popover]="popTemplate"
            #pop="bs-popover"
            placement="left">
        </button>
        `,
    styles: [`
        .fa-size-x2 {
            font-size: 20px;
        }
    `]
})

export class DwnlAgentComponent {

    public url_OS_Linux = 'https://en.opensuse.org/LinuxAutomotive#Installation_AGL_XDS';
    public url_OS_Other = 'https://github.com/iotbzh/xds-agent#how-to-install-on-other-platform';
}
