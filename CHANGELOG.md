# hpsf library changelog

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