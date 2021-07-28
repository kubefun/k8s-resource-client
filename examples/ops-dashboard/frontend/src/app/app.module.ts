import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';

import { AppComponent } from './app.component';
import { ClarityModule } from '@clr/angular';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { StatsComponent } from './stats/stats.component';
import { WatchersComponent } from './watchers/watchers.component';
import { HttpClientModule } from '@angular/common/http';

import '@cds/core/icon/register.js';
import { ClarityIcons, playIcon, stopIcon } from '@cds/core/icon';

ClarityIcons.addIcons(playIcon, stopIcon);

@NgModule({
  declarations: [AppComponent, StatsComponent, WatchersComponent],
  imports: [
    BrowserModule,
    ClarityModule,
    BrowserAnimationsModule,
    HttpClientModule,
  ],
  providers: [],
  bootstrap: [AppComponent],
})
export class AppModule {}
