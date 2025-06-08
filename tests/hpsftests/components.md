| Kind | Summary | Notes | Prioritization notes |
|------|---------|-------|---------------------|
| HoneycombExporter | | | |
| SendToArchive | | S3 Exporter configured for Enhance | |
| OTelHTTPExporter | | | |
| OTelGRPCExporter | | | |
| OtelDebugExporter | | | |
| ColumnDeletion | | For hiding sensitive info, or reducing noise | |
| RedactFields | | Bindplane's implementation: https://bindplane.com/docs/resources/processors/mask-sensitive-data<br><br>Semantic gotcha: What we name this probably matters. Customers tend to search for "redact" even if that's not exactly what they want to do. "Mask" implies the ability to unmask. Other options: obfuscate | |
| RenameAttributes | | Change field keys based on a pattern | |
| NormalizeFields | | Patterned changes to multiple field values (find/replace) | |
| ParseUserAgent | | | Parses a user agent string and breaks it out into attributes., takes a field to use, and uses the transformprocessor function to parse it out. |
| ChangeServiceName | | Group by processor for a span attribute and a transform processor to copy it to service.name. It will group by the attribute, then copy the resource attribute it's now in to the service.name resource attribute | |
| FilterByTargetUrl | | Filters spans based on the `http.target` field, used for filtering healthchecks. It could be made more specific by adding in all the different health type urls (/health, /healthz, etc.) | |
| HashField | | Hash a field (different to redact), uses the hash function of the transformprocessor | Valuable, but lower priority than Redact. |
| AllowOnlyTheseFields | | Apply an AllowList of fields, this uses the keep_keys function of the transformprocessor, and takes a user input that's a string list of field names. | Valuable, but since you can technically achieve similar results with ColumnDeletion (albeit with more effort, potentially), other processors are higher priority for now. |
| FilterLogsByLibrary | | Filter processor to delete logs based on the library, uses the filter processor and the library.name attribute. It should allow the user to provide a list of attributes | Sounds useful. Needs info on how we would intepret the library and corresponding behavior |
