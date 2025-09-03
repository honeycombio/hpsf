# hpsf library changelog

## 0.16.0 2025-08-27

### Features
- feat: add input/output tags (#212) | [Kent Quirk](https://github.com/kentquirk)

### Fixes
- fix: don't use property names where component names belong (#211) | [Kent Quirk](https://github.com/kentquirk)

## 0.15.0 2025-08-18

New components, new organization, many improvements to validation, better processes.

### Features
- feat: Add FieldExistsCondition (#200) | [Kent Quirk](https://github.com/kentquirk)
- feat: Add FieldStartsWith and FieldContains conditions (#194) | [Kent Quirk](https://github.com/kentquirk)
- feat: add library_version field, add FromYAML func (#195) | [Alex Boten](https://github.com/codeboten)
- feat: add tool to help with bulk-editing components (#204) | [Kent Quirk](https://github.com/kentquirk)
- feat: add type-specific comparison and scope to conditions (#197) | [Kent Quirk](https://github.com/kentquirk)
- feat: add validation for component version (#184) | [Alex Boten](https://github.com/codeboten)
- feat: Implement condition for checking existence of root span (#196) | [Kent Quirk](https://github.com/kentquirk)
- feat: Regexp Component (#198) | [Kent Quirk](https://github.com/kentquirk)
- feat: Reorganize category tags for updated UI. (#208) | [Kent Quirk](https://github.com/kentquirk)
- feat: split up hpsftypes into separate package (#190) | [Alex Boten](https://github.com/codeboten)
- feat: Update SamplingSequencer to support 12 rules (#186) | [Kent Quirk](https://github.com/kentquirk)

### Fixes
- fix: Better sampler validation (conforms to new sampler design). (#182) | [Kent Quirk](https://github.com/kentquirk)
- fix: RootSpanCondition property type (#201) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: add smoke tests to CI (#185) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: run all the files through the reformatter (#205) | [Kent Quirk](https://github.com/kentquirk)
- maint: update smoke tests to test components (#183) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- refactor: remove EnsureYAML public func (#207) | [Alex Boten](https://github.com/codeboten)

### maintenance
- chore: accidentally bumped up go version in previous commit (#191) | [Alex Boten](https://github.com/codeboten)
- chore: add make lint target (#188) | [Alex Boten](https://github.com/codeboten)
- chore: add omitempty to certain fields (#199) | [Alex Boten](https://github.com/codeboten)
- chore: clean up following linter suggestions (#187) | [Alex Boten](https://github.com/codeboten)
- chore: minor refactor (#189) | [Alex Boten](https://github.com/codeboten)
- chore: remove deprecated consts (#192) | [Alex Boten](https://github.com/codeboten)
- chore: remove unnecessary call to func (#206) | [Alex Boten](https://github.com/codeboten)
- chore: remove unused function LoadEmbeddedDefaultTemplate (#193) | [Alex Boten](https://github.com/codeboten)

## 0.14.0 2025-07-21

### Fixes

- fix: add default signal to AttributeJSONParsingProcessor (#180) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: refinery rules bugs (#176) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: handle paths with just a complex sampler (#175) | [Kent Quirk](https://github.com/KentQuirk)
- fix: nil pointer failure for sampler indexing (#174) | [Kent Quirk](https://github.com/KentQuirk)
- fix: Set the name field properly with downstream samplers (#172) | [Kent Quirk](https://github.com/KentQuirk)

### Maintenance

- maint: rename component files (#173) | [Kent Quirk](https://github.com/KentQuirk)

## 0.11.1 2025-07-18 (hotfix)

### Fixes

- maint: update the default template to use honeycomb exporter (#160) | [Tyler Helmuth](https://github.com/TylerHelmuth)
    - Plus test data fixes and Go version consistency

## 0.13.0 2025-07-16

### Feature

- feat: Use the sampler name as the rule name when generating refinery configs (#170) | [Kent Quirk](https://github.com/KentQuirk)

## 0.12.0 2025-07-16

### Feature

- feat: rewrite the way refinery samplers work (#159) | [Kent Quirk](https://github.com/KentQuirk)

### Maintenance

- maint: bump sampling components to alpha (#168) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: Add a couple of Refinery scenario tests (#167) | [Kent Quirk](https://github.com/KentQuirk)
- maint: rename processors components to match naming guidelines (#166) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: rename receivers components to match naming guidelines (#165) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: rename exporters components to match naming guidelines (#164) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: rename startsamplers/droppers components to match naming guidelines (#163) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: rename condition components to match naming guidelines (#162) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: rename sampler components to match naming guidelines (#161) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: update the default template to use honeycomb exporter (#160) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- maint: update env vars to use HTP_ prefix (#158) | [Tyler Helmuth](https://github.com/TylerHelmuth)

## 0.11.0 2025-06-25

### Fixes

- fix: automatically set otlp http exporter scheme based on insecure (#153) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: Fix rule rendering bug related to indexing (#154) | [Kent Quirk](https://github.com/KentQuirk)

## 0.10.0 2025-06-24

### Features

- feat: More sophisticated validations for hpsf (#148) | [Kent Quirk](https://github.com/kentquirk)

### Fixes

- fix: minor updates to component names & properties (#151) | [Jessica Parsons](https://github.com/verythorough)

## 0.9.0 2025-06-18

### Features

- feat: Add syntax-level validations to HPSF (#143) | [Kent Quirk](https://github.com/kentquirk)

### Maintenance

- maint: Update existing component ports with a consistent order (#144) | [Mike Goldsmith](https://github.com/MikeGoldsmith)

## 0.8.0 2025-06-16

### Features

- feat: Add Symbolicator component (#109) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat: Add APIPort parameter to HoneycombExporter (#139) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat: Add Redaction Processor component (#129) | [Mike Goldsmith](https://github.com/MikeGoldsmith)

- fix: don't ever generate "Field", always use "Fields" (#141) | [Kent Quirk](https://github.com/kentquirk)
- fix(validation): Fixed the validation for properties that aren't supplied (#136) | [MartinDotNet](https://github.com/martinjt)
- fix: Remove nonempty validators for redact properties (#140) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- test: Added test for the various multiple pipeline scenarios (#134) | [MartinDotNet](https://github.com/martinjt)
- fix: Update StartSampling UseTLS label (#138) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- fix: Update redaction component tests logs pipeline names (#137) | [Mike Goldsmith](https://github.com/MikeGoldsmith)

## 0.7.0 2025-06-13

### Work

- fix: resolve and render pipelines individually.  (#132) | [Kent Quirk](https://github.com/kentquirk)
- test: Update Tests to support named pipelines (#133) | [MartinDotNet](https://github.com/martinjt)
- fix: Add subtype to HoneycombExporter Mode property (#131) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat: Add subtype labels for bool properties (#130) | [Mike Goldsmith](https://github.com/MikeGoldsmith)


## 0.6.1 2025-06-10

- maint: remove extra test file and update version file. | [Kent Quirk](https://github.com/kentquirk)

## 0.6.0 2025-06-10

### Features

- feat: Renamed all Port names and types (#126) | [MartinDotNet](https://github.com/martinjt)
- feat: move dev component to alpha (#125) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat(components): Rename TraceConverter to StartSampling (#123) | [MartinDotNet](https://github.com/martinjt)
- feat(components): Rename KeepSlowTraces to SampleByTraceDuration (#122) | [MartinDotNet](https://github.com/martinjt)
- feat(component): Rename SampleErrors to SampleByHTTPStatus (#121) | [MartinDotNet](https://github.com/martinjt)
- feat(components): rename S3Exporter to SendToS3Archive (#120) | [MartinDotNet](https://github.com/martinjt)
- feat: Update HoneycombExporter to also create collector exporter (#108) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- test: Added make target for regenerating the translator config tests (#119) | [MartinDotNet](https://github.com/martinjt)
- feat: add headers to default otlp exporter (#85) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- feat: add s3 template (#116) | [Jamie Danielson](https://github.com/JamieDanielson)
- feat: Add smoke tests for collector configs using HPSF templates (#88) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat: Add smoke tests for refinery components (#83) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat: add subtype to components (#82) | [Kent Quirk](https://github.com/kentquirk)
- feat: change the way rules files are generated (#90) | [Kent Quirk](https://github.com/kentquirk)
- feat: Extend bracket notation everywhere; work with multiple components (#107) | [Kent Quirk](https://github.com/kentquirk)
- feat: new rules components (#100) | [Kent Quirk](https://github.com/kentquirk)
- feat: set SendKeyMode when api key present (#80) | [Alex Boten](https://github.com/codeboten)
- feat: Simplify using default values in components (#94) | [Kent Quirk](https://github.com/kentquirk)
- feat: switch s3exporter default marshaler to proto (#105) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- feat: Update logs dedup processor port names (#118) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat: Update Refinery smoke tests (#84) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- feat: update s3 exporter with batch config (#95) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- feat: use batch config in exporters (#97) | [Alex Boten](https://github.com/codeboten)
- feat(component): Added FilterLogsBySeverity processor (#113) | [MartinDotNet](https://github.com/martinjt)
- feat(components): Added JSON parsing components (#110) | [MartinDotNet](https://github.com/martinjt)
- feat(hpsf): Added an upper function to the go templates (#111) | [MartinDotNet](https://github.com/martinjt)

### Fixes

- fix: batch setting should be nested under batch section (#98) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: honecyomb exporter default values (#93) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: make it easier to run smoke tests for templates (#115) | [Jamie Danielson](https://github.com/JamieDanielson)
- fix: make it possible to have multiple templates per kind (#112) | [Kent Quirk](https://github.com/kentquirk)
- fix: Replace duplicate HoneycombExporter in test data (#78) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- fix: s3exporter compression render location (#104) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: traceconverter batch settings (#99) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: TraceConverter default values (#92) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- fix: Update component positions in templates due to bigger nodes in new design (#91) | [Candice Pang](https://github.com/dustxd)
- fix: Update EMA throughput component to produce valid configs (#86) | [Mike Goldsmith](https://github.com/MikeGoldsmith)
- fix: update emathroughput template to use updated fields (#114) | [Jamie Danielson](https://github.com/JamieDanielson)
- fix: Update Honeycomb exporter test data to use valid API key (#87) | [Mike Goldsmith](https://github.com/MikeGoldsmith)

### Maintenance
- maint: update status for nop components (#101) | [Alex Boten](https://github.com/codeboten)
- maint: Use numbers wherever possible in sampler (#96) | [Kent Quirk](https://github.com/kentquirk)
- maint: add github workflows (#81) | [Alex Boten](https://github.com/codeboten)
- maint: README for components; fixing names of things (#79) | [Kent Quirk](https://github.com/kentquirk)
- test: Added end to end framework for config rendering (#106) | [MartinDotNet](https://github.com/martinjt)


## 0.5.0 2025-04-09

### Features

- feat: add logo to components (#76)

### Maintenance

- maint: Update template layouts to be valid (#75)

## 0.4.0 2025-04-07

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
