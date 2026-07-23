# json-schematics-v2

[![Go Reference](https://pkg.go.dev/badge/github.com/ashbeelghouri/json-schematics-v2.svg)](https://pkg.go.dev/github.com/ashbeelghouri/json-schematics-v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/ashbeelghouri/json-schematics-v2)](https://goreportcard.com/report/github.com/ashbeelghouri/json-schematics-v2)
[![CI](https://github.com/ashbeelghouri/json-schematics-v2/actions/workflows/ci.yml/badge.svg)](https://github.com/ashbeelghouri/json-schematics-v2/actions/workflows/ci.yml)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**Validate and transform arbitrary JSON in Go using a small, declarative schema — with no third-party dependencies.**

> 📖 Full documentation, live examples, and an interactive playground:
> **[jsonschematics.ashbeelghouri.com](https://jsonschematics.ashbeelghouri.com)**

`json-schematics-v2` flattens any document into dotted keys (`user.profile.name`,
`tags.0`), lets each schema field target one or more of those keys (literally,
with a `*` wildcard, or with a regular expression), and runs every matched value
through an ordered chain of **validators** and **operators**.

This is a ground-up redesign of the original `jsonschematics`. See
[MIGRATION.md](MIGRATION.md) for what changed and why.

## Features

- 🎯 **Target anything** — match flattened keys literally, with a `*` wildcard, or a full regex.
- 🧩 **Batteries included** — 41 validators, 12 operators, and 3 conditions built in.
- 🛡️ **Never panics** — wrong types return typed errors; unknown rule names are caught up front by `Check()`.
- 🔍 **Catches target typos** — `ValidateSchema()` matches every field's target against sample data, so a typo like `"target": "mane"` fails your tests instead of silently matching nothing in production.
- 🌍 **Localized errors** — per-locale messages and format templates; `ValidationError` marshals to clean JSON.
- ⚙️ **Validate _and_ transform** — a separate operator pass, plus a shared context DB for cross-field logic.
- 🌐 **HTTP request validation** — validate headers, query, and body per endpoint.
- 📦 **Zero dependencies** — standard library only, Go 1.24+.

## Install

```sh
go get github.com/ashbeelghouri/json-schematics-v2@latest
```

```go
import schematics "github.com/ashbeelghouri/json-schematics-v2"
```

## Quick start

```go
s := schematics.New()
if err := s.LoadFile("schema.json"); err != nil {
    log.Fatal(err)
}

data := map[string]any{
    "user": map[string]any{"profile": map[string]any{"name": "a", "age": 200}},
}

if err := s.Validate(data); err != nil {
    var ve *schematics.ValidationErrors
    if errors.As(err, &ve) {
        for _, msg := range ve.Strings("en", "%target: %message") {
            fmt.Println(msg)
        }
    }
}
```

`Validate` returns:

- `nil` when the data is valid,
- a `*ValidationErrors` when the data breaks the rules,
- a `*SchemaError` when the schema references an unknown rule/operator/condition.

Use `errors.As` to tell them apart.

## The schema

```json
{
  "version": "2.0",
  "separator": ".",
  "arrayIdKey": "id",
  "locale": "en",
  "db": { "minAge": 18 },
  "fields": [
    {
      "target": "user.profile.name",
      "type": "string",
      "required": true,
      "dependsOn": ["user.profile.email"],
      "addToDB": false,
      "when":     [ { "condition": "fieldPresent", "args": { "field": "user.profile.email" } } ],
      "validate": [ { "rule": "minLength", "args": { "min": 2 }, "message": "too short", "messages": { "ar": "قصير جدا" } } ],
      "operate":  [ { "op": "trim" }, { "op": "capitalize" } ]
    }
  ]
}
```

### Field keys

| Key | Meaning |
|-----|---------|
| `target` | Flattened key to match. Literal, `*` wildcard (one segment), or a regex when `targetRegex` is `true`. |
| `targetRegex` | Treat `target` as a Go regular expression. |
| `required` | Fail if the target selects no value. |
| `dependsOn` | Other targets that must be present for this field to be validated. |
| `addToDB` | Copy the matched value into the shared `db` before validation. |
| `when` | Conditions that must all hold for the field to run. |
| `validate` | Ordered validators. The **first** failing rule reports one error and stops that field. |
| `operate` | Ordered operators, applied by `Operate`. |
| `name`, `displayName`, `type`, `description`, `tags`, `meta` | Metadata; carried through untouched. |

### Targeting arrays

Flattening turns `{"tags":["a","b"]}` into `tags.0`, `tags.1`. To validate the
**array itself** (length, uniqueness), target the parent key — the engine
reconstructs the array for you:

```json
{ "target": "tags", "validate": [ { "rule": "minItems", "args": { "min": 1 } }, { "rule": "unique" } ] }
```

To validate **each element**, use a wildcard: `"target": "tags.*"`.

Validating an array of objects? Pass a `[]map[string]any` (or JSON array) to
`Validate`; set `arrayIdKey` so each error carries a `RowID`.

## Built-in validators

**Strings:** `isString`, `notEmpty`, `email`, `maxLength`, `minLength`,
`lengthBetween`, `noSpecialChars`, `hasSpecialChars`, `hasUpper`, `hasLower`,
`hasDigit`, `isURL`, `notURL`, `urlHasHost`, `urlHasQuery`, `isHTTPS`, `isUUID`,
`equals`, `inOptions`, `matchRegex`, `like`.

**Numbers:** `isNumber`, `isInteger`, `isFloat`, `max`, `min`, `between`,
`positive`, `negative`, `nonNegative`.

**Dates:** `isDate`, `beforeNow`, `afterNow`, `before`, `after`, `betweenTime`.

**Arrays:** `isArray`, `maxItems`, `minItems`, `unique`, `itemsInOptions`.

Every validator inspects its input safely — the wrong type returns an error, it
never panics.

## Built-in operators

`trim`, `capitalize`, `upper`, `lower`, `toString`, `add`, `subtract`,
`multiply`, `divide`, `round`, `default`, `arrayToObject`.

```go
out, err := s.Operate(data) // returns the transformed document
```

## Built-in conditions

`fieldPresent`, `fieldAbsent`, `fieldEquals` — used in a field's `when` list.
Set `"negate": true` on a condition to invert it.

## Custom rules, operators, conditions

Signatures are typed and context-aware. Register your own before validating:

```go
s.RegisterRule("isAsh", func(v any, args schematics.Args, ctx *schematics.Context) error {
    if v == "ash" {
        return nil
    }
    return errors.New("must be ash")
})

s.RegisterOperator("reverse", func(v any, args schematics.Args, ctx *schematics.Context) (any, error) {
    str, _ := v.(string)
    r := []rune(str)
    for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r), nil
})
```

`Args` provides panic-free typed accessors: `args.String("k")`, `args.Int("k")`,
`args.Float("k")`, `args.Bool("k")`, `args.Strings("k")`, plus `args.StringOr` /
`args.FloatOr`. `Context` exposes the shared `DB`, `Locale`, `Separator`, the
full flattened document `Flat`, the current `RowID`, a read-only `Field` view,
and `ctx.Ctx` (a `context.Context`) for cancellation.

## Configuration

```go
s := schematics.New(
    schematics.WithSeparator("."),
    schematics.WithLocale("en"),
    schematics.WithArrayIDKey("id"),
    schematics.WithDB(map[string]any{"minAge": 18}),
    schematics.WithLogger(slog.Default()),
)
```

Options take precedence over values set inside the schema file.

## Error handling

```go
var ve *schematics.ValidationErrors
if errors.As(err, &ve) {
    ve.Strings("en", "%target: %message") // []string, formatted
    ve.Messages("ar")                      // []string, localized messages only
    ve.ForTarget("user.profile.name")      // []*ValidationError
    for _, e := range ve.Errors {
        _ = e.Target // "user.profile.name"
        _ = e.Rule   // "minLength"
        _ = e.Value  // the offending value
        _ = e.RowID  // array row id, if any
        _ = e.Message("ar")
    }
}
```

Format tokens: `%message`, `%target`, `%rule` (alias `%validator`), `%value`,
`%id`. Each `*ValidationError` also marshals to a stable JSON object.

## Catching target typos before they ship

Because a schema is data, not code, a typo in a field's `target` doesn't fail
to compile — it just quietly matches nothing at runtime. `Check()` (run
automatically before every `Validate`) catches a typo in a JSON *key*, like
`"tagret"` instead of `"target"`: `encoding/json` drops the unrecognized key,
`Target` ends up `""`, and `Check` reports `"field #N has an empty target"`.

A typo in the target's *value* — `"target": "mane"` instead of `"name"` — is a
perfectly valid string, so `Check` alone has nothing to flag. `ValidateSchema`
closes that gap by matching every target against a sample of your real data:

```go
s := schematics.New()
s.LoadFile("schema.json")

sample := map[string]any{"name": "Ada", "email": "ada@example.com"}
if err := s.ValidateSchema(sample); err != nil {
    log.Fatal(err) // *SchemaError: field #0: target "mane" does not match anything in the sample data
}
```

Run it once in a test alongside a fixture (or a golden payload) so schema/data
drift fails CI instead of failing silently in production:

```go
func TestSchemaMatchesFixture(t *testing.T) {
    s := schematics.New()
    if err := s.LoadFile("schema.json"); err != nil {
        t.Fatal(err)
    }
    if err := s.ValidateSchema(loadFixture(t)); err != nil {
        t.Fatal(err) // schema drifted from the real payload shape — fix the typo
    }
}
```

If a field is legitimately optional and absent from every fixture you have,
pass its target to `ignoreTargets` so it isn't flagged:

```go
s.ValidateSchema(sample, "user.profile.middleName")
```

`ValidateSchema` runs every check `Check` does, so it also still catches
unknown rule/operator/condition names and duplicate targets.

## HTTP request validation

Validate `*http.Request` headers, query, and body against per-endpoint field
sets:

```go
api := schematics.NewAPI()
if err := api.LoadFile("api.schema.json"); err != nil {
    log.Fatal(err)
}
// api.Base().RegisterRule(...) to add custom rules used by the request schema
if err := api.ValidateRequest(r); err != nil {
    // *ValidationErrors or *SchemaError
}
```

Endpoint paths support `:name` (one segment) and a trailing `*` wildcard. See
[`examples/api.schema.json`](examples/api.schema.json).

## Examples

Runnable schema and data live in [`examples/`](examples/). The package-level
`Example` in the tests loads them end to end. Try any schema live in the
[playground](https://jsonschematics.ashbeelghouri.com/playground).

## Development

```sh
go test ./...
go test -race -cover ./...
go vet ./...
```

## License

MIT — see [LICENSE](LICENSE).
