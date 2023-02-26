package labels

type IstioTLSMode string

const (
	IstioTLSModeIstio    IstioTLSMode = "istio"
	IstioTLSModeDisabled IstioTLSMode = "disabled"
)

type IstioProxyType string

const (
	IstioProxyTypeSidecar IstioProxyType = "sidecar"
	IstioProxyTypeGateway IstioProxyType = "gateway"
)

const (
	DefaultClusterDNSDomain = "cluster.local"
)

const (
	IstioRevisionLabel                 = "istio.io/rev"
	IstioNetworkingGatewayPortLabel    = "networking.istio.io/gatewayPort"
	IstioSecurityTlsModeLabel          = "security.istio.io/tlsMode"
	IstioServiceCanonicalNameLabel     = "service.istio.io/canonical-name"
	IstioServiceCanonicalRevisionLabel = "service.istio.io/canonical-revision"
	IstioSidecarInjectLabel            = "sidecar.istio.io/inject"
	IstioTopologyClusterLabel          = "topology.istio.io/cluster"
	IstioTopologyNetworkLabel          = "topology.istio.io/network"
	IstioTopologySubzoneLabel          = "topology.istio.io/subzone"
)

const (
	NASPMonitoringLabel     = "nasp.k8s.cisco.com/monitoring"
	NASPMonitoringPortLabel = "nasp.k8s.cisco.com/monitoring-port"
	NASPMonitoringPathLabel = "nasp.k8s.cisco.com/monitoring-path"
	NASPWorkloadgroupLabel  = "nasp.k8s.cisco.com/workloadgroup"
	NASPWorkloadUID         = "nasp.k8s.cisco.com/uid"
)

const (
	KubernetesAppNameLabel    = "app.kubernetes.io/name"
	KubernetesAppVersionLabel = "app.kubernetes.io/version"
	AppNameLabel              = "app"
	AppVersionLabel           = "version"
)

const serviceRevisionDefault = "latest"

func IstioCanonicalServiceName(labels map[string]string, workloadName string) string {
	return getLabelValueWithDefault(labels, []string{
		IstioServiceCanonicalNameLabel,
		KubernetesAppNameLabel,
		AppNameLabel,
	}, workloadName)
}

func IstioCanonicalServiceRevision(labels map[string]string) string {
	return getLabelValueWithDefault(labels, []string{
		IstioServiceCanonicalRevisionLabel,
		KubernetesAppVersionLabel,
		AppVersionLabel,
	}, serviceRevisionDefault)
}

func getLabelValueWithDefault(labels map[string]string, precedence []string, def string) string {
	for _, k := range precedence {
		if v, ok := labels[k]; ok {
			return v
		}
	}

	return def
}
