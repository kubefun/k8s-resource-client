import { Injectable, Inject } from '@angular/core';
import { webSocket, WebSocketSubject } from 'rxjs/webSocket';
export const WS_ENDPOINT = 'ws://127.0.0.1:1234/ws';

@Injectable({
  providedIn: 'root',
})
export class WebsocketService {
  private socket$!: WebSocketSubject<any>;

  constructor() {}

  public connect(): WebSocketSubject<any> {
    if (!this.socket$ || this.socket$.closed) {
      this.socket$ = webSocket(WS_ENDPOINT);
    }
    return this.socket$;
  }

  public dataUpdates$() {
    return this.connect().asObservable();
  }

  closeConnection() {
    this.connect().complete();
  }

  sendMessage(msg: any) {
    this.socket$.next(msg);
  }
}
