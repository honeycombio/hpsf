# hpsf library changelog

## 0.3.0 2025-04-07

### Features

- feat: use http exporter in default (#73)
- feat: set otelgrpc exporter to alpha (#72)
- feat: set otel receiver to alpha (#71)

## 0.3.0 2025-04-03

### Features

- feat: Add S3 exporter component (#65)
- feat: Validate properties with the property rules (#60)
- feat: Add logos to receivers and exporters, tweak some text (#56)

### Fixes

- fix: add output config to processor (#59)

### Maintenance

- maint: Apply component property validations (#66)
- maint: only check for test files for valid component config types (#64)
- chore: add tests to validate all components (#63)
- chore: update test path (#62)

## 0.2.0 2025-03-28

### Features

- feat: Update the default template to be a complete embedded template (#52)
- feat: add log deduplication processor (#50)
- feat: add alpha status (#51)
- feat: add InstallTemplates func (#49)
- feat: add batch processor automatically (#48)
- feat: add headers/insecure support for TraceConverter (#47)
- feat: add metrics/logs to nop components, all signals to debug exporter (#45)

### Fixes

- fix: Translator generates valid refinery rules when no sampler component is provided (#54)
- fix: trace converter must support one of http or grpc (#46)

## 0.1.0 2025-03-20

This is the first release of this library.
