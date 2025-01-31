# sqlc-template

sqlc-template is a [sqlc](https://github.com/sqlc-dev/sqlc) plugin that generates code from SQL into any language by using a user provided [Golang template](https://pkg.go.dev/text/template).
This plugin was inspired by the [sqlc-gen-from-template](https://github.com/fdietze/sqlc-gen-from-template) plugin.

## Usage

To use the plugin the following `options` must be defined in the [sqlc](https://github.com/sqlc-dev/sqlc) configuration file (usually named `sqlc.yaml`):

-   `filename`: The generated code output file name.
-   `template`: The [Golang template](https://pkg.go.dev/text/template).

The template has access to the [sqlc](https://github.com/sqlc-dev/sqlc) [`GenerateRequest`](internal/protos/plugin/codegen.pb.go#L967) object as the root data object.

Usage example:

`sqlc.yaml`:

```yaml
version: "2"
plugins:
    - name: sqlc-template
      wasm:
          url: https://github.com/NMFR/sqlc-template/releases/download/v1.0.0/sqlc-template.wasm
          sha256: 62b052a392a2ee631af1e54ae07e4653ba81ae296d2df62caf02c3bb4fa70be4
sql:
    - engine: "postgresql"
      queries: "example/database/postgresql/query.sql"
      schema: "example/database/postgresql/schema.sql"
      codegen:
          - out: example/test/
            plugin: sqlc-template
            options:
                filename: queries.yaml
                template: |
                    queries:
                    {{- range .Queries }}
                    - name: {{ .Name | ToLowerCamel }}
                      cmd: {{ .Cmd }}
                      params: {{ if (eq (len .Params) 0) }}[]{{ end }}
                      {{- range .Params }}
                      - name: {{ .Column.Name | ToLowerCamel }}
                        type: {{ .Column.Type.Name -}}
                      {{ end }}
                      columns: {{ if (eq (len .Columns) 0) }}[]{{ end }}
                      {{- range .Columns }}
                      - name: {{ .Name | ToLowerCamel }}
                        type: {{ .Type.Name -}}
                      {{ end }}
                    {{ end -}}
```

Running the `sqlc generate` command will create the following file:

`queries.yaml`:

```yaml
queries:
    - name: getAuthor
      cmd: :one
      params:
          - name: id
            type: bigserial
      columns:
          - name: id
            type: bigserial
          - name: name
            type: text
          - name: bio
            type: text

    - name: listAuthors
      cmd: :many
      params: []
      columns:
          - name: id
            type: bigserial
          - name: name
            type: text
          - name: bio
            type: text

    - name: createAuthor
      cmd: :one
      params:
          - name: name
            type: text
          - name: bio
            type: text
      columns:
          - name: id
            type: bigserial
          - name: name
            type: text
          - name: bio
            type: text

    - name: updateAuthor
      cmd: :exec
      params:
          - name: id
            type: bigserial
          - name: name
            type: text
          - name: bio
            type: text
      columns: []

    - name: deleteAuthor
      cmd: :exec
      params:
          - name: id
            type: bigserial
      columns: []
```
