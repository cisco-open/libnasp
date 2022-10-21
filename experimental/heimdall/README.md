# Heimdall

A small microservice to allow external services to obtain an Istio configuration based on some external authentication mechanism, like database or OAuth, etc...

Heimdall is meant to run on Kubernetes.

### Currently supported auth methods for clients
- Kubernetes ConfigMap database (ClientID/ClientSecret)
- ...
- TODO

## Installation

Apply the manifests, configure them beforehand:

```bash
kubectl apply -f experimental/heimdall/deploy
```

Take note of the exposed Heimdall service, this can be used in mobile applications for example, to get get an Istio entry confgiuration:

```bash
kubectl describe services -n heimdall heimdall-gw | grep Ingress | awk '{print $3}'
```

## Populate the client database

The client database holds the list of clients (indexed by ClientID) and their attributes/configuration for the mesh.A

A sample entry:

```yaml
  16362813-F46B-41AC-B191-A390DB1F6BDF: |
    {
    "ClientSecret": "16362813-F46B-41AC-B191-A390DB1F6BDF",
    "ClientOS": "ios",
    "WorkloadName": "ios-mobile-app",
    "PodNamespace": "external",
    "Network": "network2",
    "MeshID": "mesh1"
    }
```
