module github.com/asraa/test-sigstore-root

go 1.16

require (
	github.com/peterbourgon/ff/v3 v3.1.2
	github.com/pkg/errors v0.9.1
	github.com/sigstore/cosign v1.7.1
	github.com/sigstore/sigstore v1.2.1-0.20220401110139-0e610e39782f
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.10.1
	github.com/tent/canonical-json-go v0.0.0-20130607151641-96e4ba3a7613
	github.com/theupdateframework/go-tuf v0.1.0
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/theupdateframework/go-tuf => github.com/mnm678/go-tuf v0.0.0-20220304223310-cbf366a5642a
