# k8s-resource-client

The goals of the `k8s-resource-client` library is to create an out-of-cluster client experience that enables easily querying up-to-date resources from a Kubernetes cluster while dealing with the class of concerns around API authorization. To do this, the client has two modes of operation `auto` and `explicit`. In both cases, the `auth.v1.SelfSubjectAccessReview` API will be used to validate that a minimal set of access is provided for a resource. If that minimal access is not allowed with the currently configured credentials the resource will be marked as `FailedSubjectAccessReview` and attempts to fetch object(s) for that resource endpoint will return an nil response and a typed error containing the details of the failure.

## Namespaces

Namespaces are treated as a special resource and can have their mode set to `auto` (default) or `explicit` indpendently of the mode set for other resources.

## Modes of Operation

Minimal RBAC requirements for this client are the `List` and `Watch` verbs for the resource you wish to view objects for. By default, the client will attempt to validate the minimal RBAC requirements by issuing a `SelfSubjectAccessReview` request for a resource. This behavior may be explictily skippend by the user.

### Auto (default)

In greedy mode the client will do best effort to discover Kubernetes resources. After discovering the resources a subject access review will be created for every discovered resource unless that behavior has been explicitly disabled.

### Explicit

In explicit mode the client will be provided a list of resources. An attempt to query any resources not configured when in explict mode will produce a `ResourceNotSynced` error.

## Configurable Options

- namespaces: auto, explicit
- namespace-scoped-resources: auto, explicit
- cluster-scoped-resources: auto, explicit
- refresh-subject-access-interval: default 5m
