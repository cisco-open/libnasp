# Python HTTP client leveraging Nasp

This example demonstrates how a client application written in Python can connect to HTTP service running in an Istio mesh through Nasp library.

## Prerequisites

### 1. Istio

To be able to join an existing service mesh, Nasp needs to connect to an Istio control plane. To install Istio, we suggest to use the Banzai Cloud [Istio Operator](https://github.com/banzaicloud/istio-operator/). While it should just work with the official installation methods, we are using our operator for testing, and the test deployment scripts are also using this tool.
To get started with the operator, follow the deployment steps on [GitHub](https://github.com/banzaicloud/istio-operator/#build-and-deploy).

### 2. Heimdall

Heimdall is a small tool that allows external services to obtain an Istio configuration from a Kubernetes cluster based on an external authentication mechanism. When running a service outside a cluster that you'd like to put in the mesh, it has to know what network, cluster or mesh to join, or where to get certificates from. Heimdall helps getting that configuration from the cluster.
Heimdall can be found in this repo, follow its [readme](../heimdall) to get it up and running on your cluster.

### 3. Compile

Nasp library is written in Go thus it can be used from Python through CGO bindings. Run `make dynamic-lib` to build Nasp as dynamic library that Python can consume.

### 4. Install requires python dependencies 

`pip3 install -r nasp/requirements.txt`

## Connecting to HTTP service using Nasp

At high level create a new HTTP transport with Nasp than bind it to HTTP requests issued from Python code through the `NaspHTTPTransportAdapter` [request adapter](https://requests.readthedocs.io/en/latest/user/advanced/#transport-adapters) 

```python
with nasp.new_http_transport(...) as http_transport:
    with requests.Session() as s:
        s.mount("http://", nasp.NaspHTTPTransportAdapter(http_transport.id))
        with s.get("http://test.example.org") as resp:
            # process response

```

For a complete example see [main.py](./main.py)


## Test locally with Kind

If you'd like to test these examples on your local machine without connecting to a Kubernetes cluster in the cloud, you can use the [`deploy-kind.sh`](../../test/deploy-kind.sh) script in the test directory. It will create a [kind](https://github.com/kubernetes-sigs/kind) cluster (or use an existing one), and install Istio and Heimdall along with an `echo` service for testing.
