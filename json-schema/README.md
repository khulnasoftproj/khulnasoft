# JSON Schema

* https://raw.githubusercontent.com/khulnasoftproj/khulnasoft/main/json-schema/khulnasoft-yaml.json
* https://raw.githubusercontent.com/khulnasoftproj/khulnasoft/main/json-schema/registry.json

These JSON Schema files are generated from khulnasoft's source code powered by [invopop/jsonschema](https://github.com/invopop/jsonschema).
Don't edit these files manually.

```console
$ cmdx js # go run ./cmd/gen-jsonschema
```

If you find a CLI tool to validate configuration with JSON Schema,
[ajv-cli](https://ajv.js.org/packages/ajv-cli.html) is useful.

e.g.

```console
$ ajv --spec=draft2020 -s json-schema/khulnasoft-yaml.json -d khulnasoft.yaml
```
