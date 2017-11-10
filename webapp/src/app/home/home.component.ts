import { Component, OnInit } from '@angular/core';

export interface ISlide {
    img?: string;
    imgAlt?: string;
    hText?: string;
    hHtml?: string;
    text?: string;
    html?: string;
    btn?: string;
    btnHref?: string;
}

@Component({
    selector: 'xds-home',
    templateUrl: 'home.component.html',
    styleUrls: ['home.component.css']
})

export class HomeComponent {

    public carInterval = 4000;
    public activeSlideIndex = 0;

    // FIXME SEB - Add more slides and info
    public slides: ISlide[] = [
        {
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: 'iot graphx image',
            hText: 'Welcome to XDS Dashboard !',
            text: 'X(cross) Development System allows developers to easily cross-compile applications.',
        },
        {
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: 'iot graphx image',
            hText: 'Create, Build, Deploy, Enjoy !',
        },
        {
            img: 'assets/images/iot-graphx.jpg',
            imgAlt: 'iot graphx image',
            hHtml: '<p>To Start: click on <i class=\'fa fa-cog\' style=\'color:#9d9d9d;\'></i> icon and add new folder</p>',
        }
    ];

    constructor() { }
}
