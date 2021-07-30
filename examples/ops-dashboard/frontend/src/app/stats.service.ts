import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';

import { Observable, throwError } from 'rxjs';
import { catchError, retry } from 'rxjs/operators';
import { Stats } from './stats';

@Injectable({
  providedIn: 'root',
})
export class StatsService {
  constructor(private http: HttpClient) {}

  statsUrl = 'http://127.0.0.1:1234/stats';

  getStats() {
    return this.http.get<Stats>(this.statsUrl);
  }
}
