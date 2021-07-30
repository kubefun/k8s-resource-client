import { Component, OnInit } from '@angular/core';
import { Resource } from '../resource';
import { ResourceService } from '../resources.service';

@Component({
  selector: 'app-resources',
  templateUrl: './resources.component.html',
  styleUrls: ['./resources.component.sass'],
})
export class ResourcesComponent implements OnInit {
  clusterResources: Resource[] = [];
  namespacedResources: Resource[] = [];

  constructor(private resourceService: ResourceService) {
    this.getClusterResources();
    this.getNamespacedResources();
  }

  getClusterResources() {
    this.resourceService
      .getResources('')
      .subscribe((data: Resource[]) => (this.clusterResources = data));
  }

  getNamespacedResources() {
    this.resourceService
      .getResources('namespace')
      .subscribe((data: Resource[]) => (this.namespacedResources = data));
  }

  ngOnInit(): void {}
}
