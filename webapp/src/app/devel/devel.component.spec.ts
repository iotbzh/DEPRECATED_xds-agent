import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { DevelComponent } from './devel.component';

describe('DevelComponent', () => {
  let component: DevelComponent;
  let fixture: ComponentFixture<DevelComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ DevelComponent ]
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(DevelComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
