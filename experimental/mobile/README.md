# NASP on mobile

NASP is written mostly in Go - and some parts in other languages, but compiled to web assembly. Thanks to this, we can leverage [Go mobile](https://pkg.go.dev/golang.org/x/mobile) to compile the library to the most well-known mobile platforms to languages like Java/Kotlin (Android) and Objective-C/Swift (iOS) with [gobind](https://pkg.go.dev/golang.org/x/mobile/cmd/gobind). Not to mention that Java and Swift bindings allow us to exercise these bindings in server-side applications as well.

If you are interested in how the bindings work in the background there is a [great blogpost](https://medium.com/@matryer/tutorial-calling-go-code-from-swift-on-ios-and-vice-versa-with-gomobile-7925620c17a4) describing the internals.

This small library exposes only basic HTTP request capabilities and is not yet complete, there is more to come.

## Compile

It's worth having a look at the official Go mobile wiki to have an initial understanding of how it works: https://github.com/golang/go/wiki/Mobile 

### Prepare gomobile first

First, we need to setup `gomobile` and `gobind` from the Go mobile project:

```bash
go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest
```

### Compile the library to a specific mobile platform library

### iOS/XCFramework

Xcode installation is necessary.

```bash
cd experimental/mobile
make ios
```

### Android/AAR

Android SDK/NDK installation is necessary.

```bash
cd experimental/mobile
make android
```

### Example applications

There are two basic mobile applications in the `experimental/mobile/examples` directory for each aforementioned platform.
It is a basic application, which can connect to an existing Istio cluster via getting the initial Istio configuration through [Heimdall](../heimdall).

You need to create the Kubernetes ServiceAccounts for the mobile application after installing Heimdall, like:

```bash
kubectl create serviceaccount -n external ios-mobile
kubectl create serviceaccount -n external android-mobile
```

#### Mobile mesh metrics with Prometheus (optional)

In order to ship the mesh stats from the mobile application to Prometheus we use Pushgateway currently, which is required to be available on the target cluster (and inside the same mesh) that the mobile application talks to.

Install the Pushgateway with Helm:

```bash
# Prepare the pushgateway namespace with istio enabled
kubectl create namespace prometheus-pushgateway
kubectl label namespace prometheus-pushgateway istio.io/rev=icp-v115x.istio-system

# Install pushgateway into it
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm upgrade --install push-gw -n prometheus-pushgateway prometheus-community/prometheus-pushgateway

# Create a ServiceMonitor to collect data from the Pushgateway (expects Prometheus operator)
kubectl apply -f experimental/mobile/deploy/servicemonitor.yaml
```
