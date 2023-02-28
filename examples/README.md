# Examples

This directory contains examples on how to use the Nasp library from Go.
Refer to the `http`, `grpc` and `tcp` folders for the corresponding examples.
The folders contain simple Go files that could be run with `go run .`

## Prerequisites

### 1. Istio

To be able to join an existing service mesh, Nasp needs to connect to an Istio control plane. To install Istio, we suggest to use the Banzai Cloud [Istio Operator](https://github.com/banzaicloud/istio-operator/). While it should just work with the official installation methods, we are using our operator for testing, and the test deployment scripts are also using this tool.
To get started with the operator, follow the deployment steps on [GitHub](https://github.com/banzaicloud/istio-operator/#build-and-deploy).

### 2. Heimdall

Heimdall is a small tool that allows external services to obtain an Istio configuration from a Kubernetes cluster based on an external authentication mechanism. When running a service outside a cluster that you'd like to put in the mesh, it has to know what network, cluster or mesh to join, or where to get certificates from. Heimdall helps getting that configuration from the cluster.
Heimdall can be found in this repo, follow its [readme](../components/heimdall) to get it up and running on your cluster.

## Test locally with Kind

If you'd like to test these examples on your local machine without connecting to a Kubernetes cluster in the cloud, you can use the [`deploy-kind.sh`](../test/deploy-kind.sh) script in the test directory. It will create a [kind](https://github.com/kubernetes-sigs/kind) cluster (or use an existing one), and install Istio and Heimdall along with an `echo` service for testing.

## Mobile examples

To see examples of how to use Nasp from mobile platforms, check out the [`mobile`](../experimental/mobile) examples.
