import { Component, OnInit } from '@angular/core';
import { Watcher, watcherNamespace } from '../watcher';
import { WatcherService } from '../watcher.service';

@Component({
  selector: 'app-watchers',
  templateUrl: './watchers.component.html',
  styleUrls: ['./watchers.component.sass'],
})
export class WatchersComponent implements OnInit {
  constructor(private watcherService: WatcherService) {}

  watcherNamespace = watcherNamespace;

  selected: Watcher = {} as Watcher;
  watchers: Watcher[] = [];

  ngOnInit(): void {
    this.getWatchers();
  }

  onStart(): void {
    alert('start!');
  }

  onStop(): void {
    alert('stop');
  }

  getWatchers(): void {
    this.watcherService
      .getWatchers()
      .subscribe((watchers) => (this.watchers = watchers));
  }
}
