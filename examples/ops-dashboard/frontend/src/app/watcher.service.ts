import { Injectable, OnInit } from '@angular/core';
import { Observable, of } from 'rxjs';

import { Watcher } from './watcher';
import { WebsocketService } from './websocket.service';

// const WATCHERS: Watcher[] = [
//   {
//     resource: 'v1.Pod',
//     isRunning: 2,
//     queue: true,
//     events: 101,
//     eventsPerSecond: 12,
//     lastEvent: '',
//     namespace: '',
//   },
//   {
//     resource: 'apps.v1.Deployment',
//     isRunning: 1,
//     queue: true,
//     events: 101,
//     eventsPerSecond: 12,
//     lastEvent: '',
//     namespace: '',
//   },
// ];

@Injectable({
  providedIn: 'root',
})
export class WatcherService {
  constructor(private websocket: WebsocketService) {}

  getWatchers(): Observable<Watcher[]> {
    return this.websocket.dataUpdates$();
  }

  // getWatchers(): Observable<Watcher[]> {
  //   const watchers = of(WATCHERS);
  //   return watchers;
  // }
}
