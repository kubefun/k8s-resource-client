import { ThrowStmt } from '@angular/compiler';

export interface Watcher {
  namespace: string;
  resource: string;
  isRunning: number;
  handledEventCount: number;
  unhandledEventCount: number;
  events: number;
  queue: boolean;
  lastEvent: string;
}

export function watcherNamespace(namespace: string): string {
  if (namespace == '') {
    return 'All';
  }
  return namespace;
}
