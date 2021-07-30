import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';

import { Observable, throwError } from 'rxjs';
import { catchError, retry } from 'rxjs/operators';
import { Resource } from './resource';

@Injectable({
  providedIn: 'root',
})
export class ResourceService {
  constructor(private http: HttpClient) {}

  resourceUrl = 'http://127.0.0.1:1234/resources';

  getResources(scope: string) {
    return this.http.get<Resource[]>(this.resourceUrl + '?scope=' + scope);
  }
}
