# Examples

This directory contains examples on how to use the Nasp library from Golang.
Refer to the `http`, `grpc` and `tcp` folders for the corresponding examples.
The folders contain simple Go programs that can be run through the Makefiles.

## Local test environment

If you'd like to test these examples on your local machine without connecting to a Kubernetes cluster in the cloud, you can use the [`deploy-kind.sh`](../test/deploy-kind.sh) script in the test directory. It will create a [kind](https://github.com/kubernetes-sigs/kind) cluster (or use an existing one), and install Istio and Heimdall along with an `echo` service for testing. To follow a complete example on how to configure the test environment and how to run the examples, follow the quick start guide in the [readme](../README.md).
