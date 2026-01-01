module github.com/honeycombio/hpsf

go 1.25

require (
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33
	github.com/honeycombio/hpsf/pkg/hpsftypes v0.0.0-00010101000000-000000000000
	github.com/jessevdk/go-flags v1.6.1
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
	github.com/stretchr/testify v1.11.1
	golang.org/x/mod v0.31.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/honeycombio/hpsf/pkg/hpsftypes => ./pkg/hpsftypes
