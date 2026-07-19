# Migrating from `jsonschematics` to `json-schematics-v2`

v2 is a full redesign. The core idea is unchanged — flatten JSON, match target
keys, run validators and operators — but the package layout, the Go API, and the
schema format are all new, and a number of correctness bugs are fixed.

## Why a rewrite

The original had a maze of `data/v0|v1|v2`, `api/v0|v1|v2`, and a duplicate
`structures` package; the README documented a root API that did not exist; and
several built-ins were broken. v2 is a single package with one front door.

## Bugs fixed in v2

| Original behavior | v2 |
|---|---|
| `IsURL` was registered under the same name as the UUID validator, silently overwriting it. | `isURL` and `isUUID` are distinct and both work. |
| `NoSpecialCharacters` required special characters (logic inverted). | `noSpecialChars` rejects special characters; `hasSpecialChars` requires one. |
| `IsLesserThanZero` passed a `max` arg to a min-only function and always failed. | `negative`, `positive`, `nonNegative` are correct. |
| `IsBefore`/`IsAfter` both read `maxTime` and both said "is after". | `before`/`after` read a `date` arg and report correctly. |
| Required-field failures were added to an error that was never recorded — silently dropped. | `required` failures are always reported. |
| Conditions were dead code (type assertion could never succeed). | `when` conditions work; unknown names are caught by `Check`. |
| `add_to_db` wrote into a possibly-nil map and could panic. | `addToDB` is panic-free. |
| Many `i.(string)` / `i.(float64)` casts panicked on unexpected data. | All built-ins return errors instead of panicking. |
| Stray `log.Println` fired even with logging off. | Uses `log/slog`; silent by default. |

## Package / API changes

- **One package.** `import schematics "github.com/ashbeelghouri/json-schematics-v2"`.
- **Construct with `New(...Option)`**, then `LoadFile` / `LoadBytes` / `LoadMap`.
- **`Validate(data) error`** returns `nil`, `*ValidationErrors`, or `*SchemaError`.
- **Typed signatures**, no `interface{}`:
  - Validator: `func(value any, args Args, ctx *Context) error`
  - Operator: `func(value any, args Args, ctx *Context) (any, error)` (was `*interface{}`)
  - Condition: `func(args Args, ctx *Context) bool`
- **`Args`** has panic-free typed accessors (`String`, `Int`, `Float`, `Bool`, `Strings`, …).
- **`Check()`** verifies a schema up front; unknown rule names become loud errors.

## Schema format changes

| Original | v2 |
|---|---|
| `fields` as a map keyed by target, or a `validators` array with `name` | `fields` array; each has `target`; validators live under `validate` with `rule` |
| `operators` with `name` | `operate` with `op` |
| `conditions` | `when` with `condition` (+ optional `negate`) |
| `attributes` | `args` |
| `err` / `error` + nested `l10n` | `message` + `messages` (locale → string) |
| `target_key`, `add_to_db`, `depends_on` | `target`, `addToDB`, `dependsOn` (camelCase) |
| top-level `DB` | `db` |

### Before

```json
{
  "version": "2",
  "fields": [
    { "target_key": "user.name", "required": true,
      "validators": [ { "name": "MinLengthAllowed", "attributes": { "min": 2 }, "error": "too short" } ],
      "operators":  [ { "name": "Capitalize" } ] }
  ]
}
```

### After

```json
{
  "version": "2.0",
  "fields": [
    { "target": "user.name", "required": true,
      "validate": [ { "rule": "minLength", "args": { "min": 2 }, "message": "too short" } ],
      "operate":  [ { "op": "capitalize" } ] }
  ]
}
```

## Validator name mapping (old → new)

`IsString`→`isString`, `NotEmpty`→`notEmpty`, `IsEmail`→`email`,
`MaxLengthAllowed`→`maxLength`, `MinLengthAllowed`→`minLength`,
`InBetweenLengthAllowed`→`lengthBetween`, `NoSpecialCharacters`→`noSpecialChars`,
`HaveSpecialCharacters`→`hasSpecialChars`, `LeastOneUpperCase`→`hasUpper`,
`LeastOneLowerCase`→`hasLower`, `LeastOneDigit`→`hasDigit`, `IsURL`→`isURL`,
`IsNotURL`→`notURL`, `HaveURLHostName`→`urlHasHost`,
`HaveQueryParameter`→`urlHasQuery`, `IsHttps`→`isHTTPS`, `IsValidUuid`→`isUUID`,
`LIKE`→`like`, `MatchRegex`→`matchRegex`, `IsNumber`→`isNumber`,
`MaxAllowed`→`max`, `MinAllowed`→`min`, `InBetween`→`between`,
`IsValidDate`→`isDate`, `IsLessThanNow`→`beforeNow`, `IsMoreThanNow`→`afterNow`,
`IsBefore`→`before`, `IsAfter`→`after`, `IsInBetweenTime`→`betweenTime`,
`ArrayLengthMax`→`maxItems`, `ArrayLengthMin`→`minItems`,
`StringInOptions`→`inOptions`, `StringsExistsInOptions`→`itemsInOptions`.

Operators: `Capitalize`→`capitalize`, `UpperCase`→`upper`, `LowerCase`→`lower`,
`Add`→`add` (arg `add_with`→`value`), `Subtract`→`subtract`,
`Multiply`→`multiply`, `Divide`→`divide`, `ArrayOfObjToObj`→`arrayToObject`
(arg `unique_string_key`→`key`).
