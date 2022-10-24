## Nasp

**THIS REPO IS CURRENTLY IN PREVIEW. THE API'S ARE NOT FINAL AND ARE SUBJECT TO CHANGE WITHOUT NOTICE.**

Nasp is an **open-source, lightweight library** to expand service mesh capabilities to non-cloud environments by getting rid of the complexity of operating dedicated network proxies. It is not meant to be a complete service mesh replacement, but rather an extension. It integrates well with an Istio control plane, so applications using Nasp can be handled as standard Istio workloads.

Nasp offers the most common functionality of a sidecar proxy, so it eases the developer burden for traffic management, observability, and security. Its range of capabilities includes:

- Identity, and network traffic security using mutual TLS
- Automatic traffic management features, like HTTP, or gRPC load balancing
- Transparent observability of network traffic, especially standard Istio metrics
- Dynamic configuration through xDS support

To learn more about why we created Nasp, and where could it help, read our introduction [blog](https://techblog.cisco.com/blog/nasp-intro).

## Issues and contributions

We use GitHub to track issues and accept contributions. If you'd like to raise an issue or open a pull request with changes, refer to our [contribution guide](./CONTRIBUTING.md).


## Getting started

The easiest way to get started with Nasp is to import the library from Go code, set up a Nasp HTTP `transport`, and use that `transport` to send HTTP requests to an external server. Here's how to do it.

First, import the library:

```go
import (
  "github.com/cisco-open/nasp/pkg/istio"
  "istio_ca "github.com/cisco-open/nasp/pkg/ca/istio"
)
```

Then, create and start an `IstioIntegrationHandler`:

```go
istioHandlerConfig := &istio.IstioIntegrationHandlerConfig {
    MetricsAddress: ":15090",
    UseTLS:         true,
        IstioCAConfigGetter: func(e *environment.IstioEnvironment) (istio_ca.IstioCAClientConfig, error) {
        return istio_ca.GetIstioCAClientConfig(clusterID, istioRevision)
    },
}

iih, err := istio.NewIstioIntegrationHandler(istioHandlerConfig, klog.TODO())
if err != nil {
    panic(err)
}

iih.Run(ctx)
```

And finally, send an HTTP request through the Nasp transport layer:

```go
transport, err := iih.GetHTTPTransport(http.DefaultTransport)
if err != nil {
    panic(err)
}

httpClient := &http.Client{
    Transport: transport,
}

request, err := http.NewRequest("GET", url, nil)
if err != nil {
    return err
}

response, err := httpClient.Do(request)
if err != nil {
    return err
}
```  

## Examples

To see and test a fully working example of the above code snippet, check out the [`examples`](./examples) directory that contains examples of setting up and operating HTTP, gRPC, or TCP connections from Go code.

## Support for other languages

The core code of Nasp is written in Go, that's why the primary examples are also written in Go.
But we know that it's important that Nasp could be imported from other languages as well, so we're using C bindings generated from the core codebase to have multi-language support. Then only a thin layer should be written for specific platforms to be included in the application code.

### Mobile libraries

These C bindings also serve as the base of how Nasp can be used from iOS or Android with the help of the `go-mobile` package. To learn more about how Nasp could be used from mobile platforms, check out the [`mobile`](./experimental/mobile) directory.
