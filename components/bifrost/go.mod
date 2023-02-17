module github.com/cisco-open/nasp/components/bifrost

go 1.19

replace github.com/cisco-open/nasp => ../..

require (
	github.com/cisco-open/nasp v0.0.0-00010101000000-000000000000
	github.com/go-logr/logr v1.2.3
	github.com/spf13/cobra v1.6.0
	k8s.io/klog/v2 v2.90.0
)

require (
	emperror.dev/errors v0.8.1 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/xtaci/smux v1.5.24 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
)
