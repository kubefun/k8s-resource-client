import { Component, OnInit } from '@angular/core';
import { Stats } from '../stats';

@Component({
  selector: 'app-stats',
  templateUrl: './stats.component.html',
  styleUrls: ['./stats.component.sass'],
})
export class StatsComponent implements OnInit {
  stat: Stats = {
    id: 1,
    name: 'Windstorm',
  };

  constructor() {}

  ngOnInit(): void {}
}
