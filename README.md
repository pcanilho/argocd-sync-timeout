[![CodeQL](https://github.com/pcanilho/argocd-sync-timeout/actions/workflows/codeql.yml/badge.svg)](https://github.com/pcanilho/argocd-sync-timeout/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/pcanilho/argocd-sync-timeout)](https://goreportcard.com/report/github.com/pcanilho/argocd-sync-timeout)

<br>

<p align="center" width="100%">
    <img src="https://github.com/pcanilho/argocd-sync-timeout/blob/main/docs/images/banner.png?raw=true" width="220"></img>
    <br>
    <i>argocd-sync-timeout</i>
    <br>
    🔎 <a href="#features">Features</a> | ⚙️ <a href="#requirements">Requirements</a> | 🚀 <a href="#installation">Installation</a> | 📝 <a href="#configuration">Configuration</a>
    <br><br>
</p>

A service that enforces GitOps semantics on ArgoCD and addresses the issue of having ArgoCD sync operations frozen
indefinitely.

Upstream issue:

* https://github.com/argoproj/argo-cd/pull/15603

## Features

* Enables `true` GitOps semantics on a per-application basis by enforcing a timeout on sync operations.
* Allows either global or per-application configuration.
* Provides a way to specify per-destination cell configuration.

## Requirements

* ArgoCD v2.8 or later

## Installation

### Helm

* Installs the `argocd-sync-timeout` service in the `argocd` namespace as stand-alone service.

```shell
helm upgrade --install ast oci://ghcr.io/pcanilho/argocd-sync-timeout -n argocd --version 0.1.0
```

### ArgoCD

* Installs the `argocd-sync-timeout` service in the `argocd` namespace as a native ArgoCD application.

```shell
argocd app create argocd-sync-timeout -f argo/application.yaml --upsert
```

* Alternatively, you can leverage `kubectl` directly in the `argocd` namespace:

```shell
kubectl apply -f argo/application.yaml -n argocd
```

### Stand-alone

* Expose the required environment variables or inline:
    * `AST_PERIOD` - The service period to check for sync operations
    * `AST_CONFIG` - The path to the configuration file

```shell
cd source && AST_PERIOD=10s AST_CONFIG=<path_to_yaml_config> go run .
```

## Configuration

### Description

| Key                                   | Description                                                                                                        |
|---------------------------------------|--------------------------------------------------------------------------------------------------------------------|
| `timeout`                             | The global default timeout for all applications                                                                    |
| `deferSync`                           | Whether a sync operation should be deferred to run after the previous sync operation has finished                  |
| `applications`                        | [Optional] Application name configuration mapping                                                                  |
| `applications.<app>.timeout`          | [Optional] Application-specific timeout configuration mapping. Defaults the to global default `timeout`            |
| `applications.<app>.overrides`        | [Optional] Application-specific destination cell configuration mapping.                                            |
| `applications.<app>.overrides.<cell>` | [Optional] Destination cell-specific timeout configuration mapping. Defaults to the application-specific `timeout` |

* The `cell` overrides can will be used as a direct match **first** and as a prefix match **second**.

#### Cell override example:

* Cells:
    * `cluster-test-1`
    * `cluster-test-2`

```
overrides:
  cluster-test: 6m
  cluster-test-1: 1m
```

The `cluster-test-1` cell will have a timeout of `1m` and the `cluster-test-2` cell will have a timeout of `6m`.

### Example

```yaml
---
timeout: 10s
deferSync: false
applications:
  my-argo-app:
    timeout: 1s
    overrides:
      <cell_prefix_works>: 1m
      <cell_name>: 6m
```
