import { Component } from '@angular/core';

@Component({
  selector: 'ngx-footer',
  styleUrls: ['./footer.component.scss'],
  template: `
    <span class="created-by">Created by
    <b><a href="http://iot.bzh" target="_blank">IoT.bzh</a></b> 2017
    &nbsp;&nbsp;
    <span style="font-size: small;">(powered by <a href="https://github.com/akveo/ngx-admin" target="_blank">akveo/ngx-admin</a>)</span>
    </span>
    <!-- MODS_XDS
    <div class="socials">
      <a href="#" target="_blank" class="ion ion-social-github"></a>
      <a href="#" target="_blank" class="ion ion-social-facebook"></a>
      <a href="#" target="_blank" class="ion ion-social-twitter"></a>
      <a href="#" target="_blank" class="ion ion-social-linkedin"></a>
    </div>
    -->
  `,
})
export class FooterComponent {
}
