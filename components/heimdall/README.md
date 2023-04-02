# Heimdall

![Heimdall](images/heimdall-logo.png)

Heimdall is a lightweight microservice that is designed to provide seamless Istio environment configuration for microservices-based applications that use Nasp, our new library that integrates with Istio without requiring an envoy sidecar. Heimdall is intended to run on Kubernetes and acts as a component that handles the seamless integration of Nasp with Istio. For authentication of external services, Heimdall relies on the [WorkloadGroup](https://istio.io/latest/docs/reference/config/networking/workload-group/) resource offered by Istio.

With Heimdall, developers can easily configure the Istio environment for services running inside and outside the Kubernetes cluster. The microservice eliminates the need for developers to manually configure Istio settings for applications that use Nasp, by providing automated configuration management that seamlessly integrates with Nasp.

Nasp itself simplifies the overall application architecture, by removing the need for an envoy sidecar and providing a more efficient, scalable, and secure microservices-based architecture that is easier to build, test, and deploy, while Heimdall acts as a mediator between Istio and applications that use Nasp, ensuring seamless integration between the two.

Heimdall is compatible with Istio installations using both `istioctl` and `istio operator from Cisco`.

## Installation

Install the Helm chart of Heimdall:

> The default configuration values are based on the assumption that Istio is installed and managed using the Istio operator.
>
> Use `--set istio.withIstioOperator=false` otherwise.

```bash
helm install -n heimdall --create-namespace heimdall ./components/heimdall/deploy/charts/heimdall
```

### Expose heimdall

By default, Heimdall is not exposed externally, and you must explicitly enable it using the --set expose.enabled=true flag.

Please take note of the external address of the exposed Heimdall service, as it is required for establishing a connection:

```bash
kubectl describe services -n heimdall heimdall-gw | grep Ingress | awk '{print $3}'
```

## Internal applications with Nasp

Heimdall leverages a k8s mutating webhook to provide environment configuration to internal applications. To enable Nasp integration for an application, set the `nasp.k8s.cisco.com/inject=true` label to the Pod.

## License

```text
Copyright 2022-2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
