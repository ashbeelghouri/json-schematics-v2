# Contributing to json-schematics-v2

Thanks for your interest in improving json-schematics-v2. This is a small,
dependency-free Go library, and contributions of all sizes are welcome — bug
reports, docs, new validators, or core improvements.

## Ground rules

- **No third-party dependencies.** The standard library only. A PR that adds a
  module to `go.mod` will be asked to remove it.
- **Never panic on input.** Validators, operators, and conditions must handle
  unexpected types by returning an error, not panicking. Use the `Args`
  accessors and the value helpers in `conv.go`.
- **Keep it Go-idiomatic.** Formatted, vetted, and tested (see below).

## Getting started

```sh
git clone https://github.com/ashbeelghouri/json-schematics-v2.git
cd json-schematics-v2
go test ./...
```

Requires Go 1.24 or newer.

## Before you open a PR

Run the same checks CI runs:

```sh
gofmt -l .          # should print nothing
go vet ./...
go test -race -cover ./...
```

Please add or update tests for any behavior you change. The suite is
table-driven; follow the style in `validators_test.go` and `schematics_test.go`.

## Adding a built-in validator / operator / condition

1. Implement the function in the matching file:
   - validators → `builtins_string.go` / `builtins_number.go` /
     `builtins_date.go` / `builtins_array.go`
   - operators → `builtins_operators.go`
   - conditions → `builtins_conditions.go`
2. Register it under a stable, camelCase name in `builtins.go`.
3. Add tests (both the passing and failing case).
4. Document the name in `README.md` under the relevant list.

Signatures:

```go
type Rule      func(value any, args Args, ctx *Context) error
type Operator  func(value any, args Args, ctx *Context) (any, error)
type Condition func(args Args, ctx *Context) bool
```

## Reporting bugs

Open an issue with:

- what you did (a minimal schema + data snippet is ideal),
- what you expected,
- what happened instead, and
- the Go version and library version/commit.

## Commits & pull requests

- Keep commits focused; write clear messages ("fix: before/after date
  comparison", "add: slug validator").
- Reference the issue you're addressing where relevant.
- One logical change per PR keeps review fast.

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE).
