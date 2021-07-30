import { Component, OnInit } from '@angular/core';
import { Stats } from '../stats';
import { StatsService } from '../stats.service';

@Component({
  selector: 'app-stats',
  templateUrl: './stats.component.html',
  styleUrls: ['./stats.component.sass'],
})
export class StatsComponent implements OnInit {
  stats: Stats = {} as Stats;

  constructor(private statsService: StatsService) {
    this.showStats();
  }

  showStats() {
    this.statsService.getStats().subscribe(
      (data: Stats) =>
        (this.stats = {
          total: data.total,
          stopped: data.stopped,
          running: data.running,
        })
    );
  }

  ngOnInit(): void {}
}
