module github.com/honeycombio/hpsf/tests

go 1.25

require (
	github.com/honeycombio/hpsf v0.6.1
	github.com/honeycombio/hpsf/pkg/hpsftypes v0.0.0-20250729165849-4881e28cb2c5
	github.com/honeycombio/opentelemetry-collector-configs/honeycombextension v0.0.0-20250821215019-48f07307dc74
	github.com/honeycombio/opentelemetry-collector-configs/usageprocessor v0.1.0
	github.com/honeycombio/refinery v1.21.1-0.20250604165426-312ddc7c2c94
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awss3exporter v0.138.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/filterprocessor v0.138.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/groupbyattrsprocessor v0.138.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor v0.138.0
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/collector/exporter v1.44.0
	go.opentelemetry.io/collector/extension v1.44.0
	go.opentelemetry.io/collector/receiver v1.44.0
)

replace github.com/honeycombio/opentelemetry-collector-configs/usageprocessor => github.com/honeycombio/opentelemetry-collector-configs/usageprocessor v0.0.0-20250529172854-29e92f8bd7cb

require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.39.3 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.2 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.14 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.18.18 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.10 // indirect
	github.com/aws/aws-sdk-go-v2/feature/s3/manager v1.19.14 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.10 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.88.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.29.7 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.38.8 // indirect
	github.com/aws/smithy-go v1.23.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/creasty/defaults v1.8.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20250106013310-edb8663e5e33 // indirect
	github.com/ebitengine/purego v0.9.0 // indirect
	github.com/expr-lang/expr v1.17.6 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/foxboron/go-tpm-keyfiles v0.0.0-20250903184740-5d135037bd4d // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/go-tpm v0.9.6 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/itchyny/timefmt-go v0.1.7 // indirect
	github.com/jessevdk/go-flags v1.6.1 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/klauspost/compress v1.18.1 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/lightstep/go-expohisto v1.0.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20251013123823-9fd1530e3ec3 // indirect
	github.com/mostynb/go-grpc-compression v1.2.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/open-telemetry/opamp-go v0.22.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/extension/opampcustommessages v0.138.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/batchperresourceattr v0.138.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/prometheus/client_golang v1.23.2 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.67.1 // indirect
	github.com/prometheus/otlptranslator v1.0.0 // indirect
	github.com/prometheus/procfs v0.18.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/shirou/gopsutil/v4 v4.25.9 // indirect
	github.com/spf13/cobra v1.10.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/tilinna/clock v1.1.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/xxh3 v1.0.2 // indirect
	go.opentelemetry.io/collector v0.138.0 // indirect
	go.opentelemetry.io/collector/client v1.44.0 // indirect
	go.opentelemetry.io/collector/component/componentstatus v0.138.0 // indirect
	go.opentelemetry.io/collector/component/componenttest v0.138.0 // indirect
	go.opentelemetry.io/collector/config/configauth v1.44.0 // indirect
	go.opentelemetry.io/collector/config/configcompression v1.44.0 // indirect
	go.opentelemetry.io/collector/config/configgrpc v0.138.0 // indirect
	go.opentelemetry.io/collector/config/confighttp v0.138.0 // indirect
	go.opentelemetry.io/collector/config/configmiddleware v1.44.0 // indirect
	go.opentelemetry.io/collector/config/confignet v1.44.0 // indirect
	go.opentelemetry.io/collector/config/configopaque v1.44.0 // indirect
	go.opentelemetry.io/collector/config/configoptional v1.44.0 // indirect
	go.opentelemetry.io/collector/config/configretry v1.44.0 // indirect
	go.opentelemetry.io/collector/config/configtelemetry v0.138.0 // indirect
	go.opentelemetry.io/collector/config/configtls v1.44.0 // indirect
	go.opentelemetry.io/collector/confmap/xconfmap v0.138.0 // indirect
	go.opentelemetry.io/collector/connector v0.138.0 // indirect
	go.opentelemetry.io/collector/connector/connectortest v0.138.0 // indirect
	go.opentelemetry.io/collector/connector/xconnector v0.138.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror v0.138.0 // indirect
	go.opentelemetry.io/collector/consumer/consumererror/xconsumererror v0.138.0 // indirect
	go.opentelemetry.io/collector/consumer/consumertest v0.138.0 // indirect
	go.opentelemetry.io/collector/consumer/xconsumer v0.138.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper v0.138.0 // indirect
	go.opentelemetry.io/collector/exporter/exporterhelper/xexporterhelper v0.138.0 // indirect
	go.opentelemetry.io/collector/exporter/exportertest v0.138.0 // indirect
	go.opentelemetry.io/collector/exporter/xexporter v0.138.0 // indirect
	go.opentelemetry.io/collector/extension/extensionauth v1.44.0 // indirect
	go.opentelemetry.io/collector/extension/extensioncapabilities v0.138.0 // indirect
	go.opentelemetry.io/collector/extension/extensionmiddleware v0.138.0 // indirect
	go.opentelemetry.io/collector/extension/extensiontest v0.138.0 // indirect
	go.opentelemetry.io/collector/extension/xextension v0.138.0 // indirect
	go.opentelemetry.io/collector/internal/fanoutconsumer v0.138.0 // indirect
	go.opentelemetry.io/collector/internal/sharedcomponent v0.138.0 // indirect
	go.opentelemetry.io/collector/pdata/testdata v0.138.0 // indirect
	go.opentelemetry.io/collector/pdata/xpdata v0.138.0 // indirect
	go.opentelemetry.io/collector/pipeline/xpipeline v0.138.0 // indirect
	go.opentelemetry.io/collector/processor/processorhelper/xprocessorhelper v0.138.0 // indirect
	go.opentelemetry.io/collector/processor/processortest v0.138.0 // indirect
	go.opentelemetry.io/collector/processor/xprocessor v0.138.0 // indirect
	go.opentelemetry.io/collector/receiver/receiverhelper v0.138.0 // indirect
	go.opentelemetry.io/collector/receiver/receivertest v0.138.0 // indirect
	go.opentelemetry.io/collector/receiver/xreceiver v0.138.0 // indirect
	go.opentelemetry.io/collector/service v0.138.0 // indirect
	go.opentelemetry.io/collector/service/hostcapabilities v0.138.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.63.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.63.0 // indirect
	go.opentelemetry.io/contrib/otelconf v0.18.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/prometheus v0.60.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutlog v0.14.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdoutmetric v1.38.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.14.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.8.0 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/mod v0.29.0 // indirect
	gonum.org/v1/gonum v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251022142026-3a174f9686a8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/alecthomas/participle/v2 v2.1.4 // indirect
	github.com/antchfx/xmlquery v1.5.0 // indirect
	github.com/antchfx/xpath v1.3.5 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/elastic/go-grok v0.3.1 // indirect
	github.com/elastic/lunes v0.1.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20241129210726-2c02b8208cf8 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/iancoleman/strcase v0.3.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/knadh/koanf/maps v0.1.2 // indirect
	github.com/knadh/koanf/providers/confmap v1.0.0 // indirect
	github.com/knadh/koanf/v2 v2.3.0 // indirect
	github.com/magefile/mage v1.15.0 // indirect
	github.com/mitchellh/copystructure v1.2.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal v0.138.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/filter v0.138.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/pdatautil v0.138.0 // indirect
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl v0.138.0
	github.com/open-telemetry/opentelemetry-collector-contrib/pkg/pdatautil v0.138.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/ua-parser/uap-go v0.0.0-20250917011043-9c86a9b0f8f0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/component v1.44.0
	go.opentelemetry.io/collector/confmap v1.44.0
	go.opentelemetry.io/collector/consumer v1.44.0 // indirect
	go.opentelemetry.io/collector/exporter/debugexporter v0.138.0
	go.opentelemetry.io/collector/exporter/otlpexporter v0.138.0
	go.opentelemetry.io/collector/exporter/otlphttpexporter v0.138.0
	go.opentelemetry.io/collector/featuregate v1.44.0 // indirect
	go.opentelemetry.io/collector/internal/telemetry v0.138.0 // indirect
	go.opentelemetry.io/collector/otelcol v0.138.0
	go.opentelemetry.io/collector/pdata v1.44.0 // indirect
	go.opentelemetry.io/collector/pdata/pprofile v0.138.0 // indirect
	go.opentelemetry.io/collector/pipeline v1.44.0
	go.opentelemetry.io/collector/processor v1.44.0
	go.opentelemetry.io/collector/processor/processorhelper v0.138.0 // indirect
	go.opentelemetry.io/collector/receiver/otlpreceiver v0.138.0
	go.opentelemetry.io/contrib/bridges/otelzap v0.13.0 // indirect
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/log v0.14.0 // indirect
	go.opentelemetry.io/otel/metric v1.38.0 // indirect
	go.opentelemetry.io/otel/sdk v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0
	golang.org/x/exp v0.0.0-20251017212417-90e834f514db // indirect
	golang.org/x/net v0.46.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
	google.golang.org/grpc v1.76.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/honeycombio/hpsf => ../

replace github.com/honeycombio/hpsf/pkg/hpsftypes => ../pkg/hpsftypes
