# Developing components

To create a new component, you create a new file in the `pkg/data/components` directory. The file name should be the `kind` of the component, with a `.yaml` extension.

You can look at the [README](./README.md) for information on each of the fields.

## Naming conventions


## UI first development

Each component is built into 2 parts. The `metadata` and `properties` sections, then the `template` section.

The Metadata and properties section is required, but the template section is optional. This ability allows you to see the visualisation of the component without having ti generate any output.

## Testing

There are 4 approaches to testing your components.

### UI testing in a local Honeycomb instance

If you have access to honeycomb, you can overwrite the references to `honeycombio/hpsf` in `go.mod` to point to your local copy of the repository.

```go
replace github.com/honeycombio/hpsf =>  ../hpsf
```

This will allow you to see the visualisation of all the components.

This step is optional, but really useful to understand how the component's UX will work when interacting with other components.

### Golden Master tests

Inside `translator/testdata`, every component requires 2 tests. The first is to ensure that the defaults are rendered correctly, the second is to ensure that all the properties available are overriden correctly.

From each these templates, the test will generate an output for each of the supported config types into the `testdata/<config_type>` directory with the same name as the file.

These tests ensure that the output generated from the components doesn't change without the engineers realise. These tests are run during CI and will fail if you have changed the generated output and not updated the information in the `testdata` directory.

#### Generating the testdata output

The `_all.yaml` and `_defaults.yaml` must be created manually, but the output files can be regenerated automatically by the test.

To regenerate the output files:

edit `pkg/translator/translator_test.go` and set the `overwrite` variable to `true`. then run the test:

```shell
go test ./pkg/translator -v -run "TestGenerateConfigForAllComponents" -count=1
```

 This will regenerate the testdata files with the new output.

### End to end tests

The end to end tests are located in `tests/hpsftests`. They are designed to test full templates of single or multiple components, and ensure that the generated configs are valid and have the desired settings.

The tests are located in `tests/hpsftests` and have 2 parts.

* `<test_name>.yaml` which is the HPSF template to test
* `<test_name>_test.go` which is the test file that runs the test

For more information on these tests, look at the [README](../tests/hpsftests/README.md) in the test directory.

### Smoke tests

These are tests which use the hpsf templates from `pkg/data/templates` to generate configs using the cli in `cmd/hpsf` then run the config in the collector and refinery to ensure that they start up correctly.

These tests are run as part of CI, and will fail the build if they do not work.
