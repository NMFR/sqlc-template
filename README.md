# sqlc-template

sqlc-template is a [sqlc](https://github.com/sqlc-dev/sqlc) plugin that generates code from SQL into any language by using a user provided [Golang template](https://pkg.go.dev/text/template).
This plugin was inspired by the [sqlc-gen-from-template](https://github.com/fdietze/sqlc-gen-from-template) plugin.

## Usage

To use the plugin the following `options` must be defined in the [sqlc](https://github.com/sqlc-dev/sqlc) configuration file (usually named `sqlc.yaml`):

-   `filename`: The generated code output file name.
-   `template`: The [Golang template](https://pkg.go.dev/text/template).

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
                      text: |{{ .Text | trim | nindent 4 }}
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
      text: |
          SELECT id, name, bio FROM authors
          WHERE id = $1 LIMIT 1
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
      text: |
          SELECT id, name, bio FROM authors
          ORDER BY name
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
      text: |
          INSERT INTO authors (
            name, bio
          ) VALUES (
            $1, $2
          )
          RETURNING id, name, bio
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
      text: |
          UPDATE authors
            set name = $2,
            bio = $3
          WHERE id = $1
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
      text: |
          DELETE FROM authors
          WHERE id = $1
      params:
          - name: id
            type: bigserial
      columns: []
```

## Template

The template uses the [Golang template](https://pkg.go.dev/text/template) "language".

The data object available at the root of the template (`{{ . }}`) is the sqlc [`GenerateRequest`](internal/protos/plugin/codegen.pb.go#L967) object that provides access to the SQL schema, queries and some sqlc configuration fields.

All of the [sprig](https://masterminds.github.io/sprig/) functions are available to be called from within the template with the exception of:

-   osBase
-   osDir
-   osClean
-   osExt
-   env
-   expandenv
-   kindOf
-   kindIs
-   typeOf
-   typeIs
-   typeIsLike
-   deepEqual
-   getHostByName
