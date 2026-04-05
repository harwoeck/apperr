# apperr

[![Go Reference](https://pkg.go.dev/badge/github.com/harwoeck/apperr.svg)](https://pkg.go.dev/github.com/harwoeck/apperr)

**apperr** provides a unified application-error interface for Go services.
Errors are created with a status code and optional details, resolved with
localization support, and then converted into the native error type of your
transport layer — Connect RPC, gRPC, Twirp, plain HTTP, or structured
terminal output.

```bash
go get github.com/harwoeck/apperr
```

## Quick start

```go
// 1. Create an application error with optional details
err := apperr.NotFound("book not found",
    apperr.ResourceInfo("library.v1.Book", bookID, "", ""),
    apperr.Localize("BOOK_NOT_FOUND"))

// 2. Resolve it (applies localization when a provider is available)
resolved, _, resolveErr := resolve.Error(err,
    resolve.WithLocalizationProvider(i18nAdapter))
if resolveErr != nil {
    // handle error
}

// 3. Convert to your transport's native error type
//    The second argument controls whether DebugInfo is kept in the
//    response (false = strip, which is the secure default).
connectErr, _ := connecterr.Convert(resolved, false)     // Connect RPC
grpcStatus, _ := grpcerr.Convert(resolved, false)        // gRPC
httpStatus, body, _ := httperr.Convert(resolved, false)  // HTTP
twirpErr, _ := twirperr.Convert(resolved, false)         // Twirp
terminalerr.Log(logger, resolved, false)                 // structured terminal log
```

## Why three separate steps?

apperr splits error handling into **creation**, **resolution**, and
**conversion** — three independent concerns that can each be extended or
replaced without affecting the others.

Because resolution and conversion happen separately from creation, an
interceptor or middleware can sit between them to enrich errors with
contextual information, perform i18n lookups, attach request metadata, or
log structured diagnostics — all without the business logic that created
the error needing to know about any of it.

This separation is what makes the ready-made
[interceptors](#interceptors--middleware) for Connect RPC, gRPC, and Twirp possible:
they catch a plain `*apperr.AppError`, resolve it with localization, and
convert it to the framework-native error type in one place. Your business
logic never imports a transport package or concerns itself with how a
particular framework expects errors and error details to be shaped —
keeping the codebase decoupled and following the dependency-inversion
principle.

To take this further, [`stricterr`](#stricterr--aip-193-enforcement-at-compile-time)
adds compile-time guardrails that ensure the right details are always
attached. It is modelled after Google's AIP-193 — not because every
project needs to follow Google's conventions, but because the underlying
idea is sound: an error should tell the caller *what* went wrong, *which*
resource or field is affected, and *whether* a retry makes sense. That
makes `stricterr` useful as practical guidance for good error handling
regardless of whose style guide you subscribe to.

## `stricterr` — AIP-193 enforcement at compile time

The base `apperr` package is intentionally flexible: every detail is
optional. In services that follow Google's
[AIP-193](https://google.aip.dev/193) error model this flexibility can
lead to accidental omissions — a `NotFound` without `ResourceInfo`, or
an `InvalidArgument` without field violations.

**`stricterr`** solves this with a factory that makes the required
details mandatory function parameters, so missing information becomes a
compile error instead of a runtime surprise.

```bash
go get github.com/harwoeck/apperr/stricterr
```

```go
// Create a factory bound to your service's ErrorInfo domain.
var errors = stricterr.NewFactory("api.store.example.com")

// NotFound requires a ResourceInfoOption — the compiler enforces it.
return errors.NotFound("book not found", "BOOK_NOT_FOUND", nil,
    apperr.ResourceInfo("store.v1.Book", bookID, "", ""),
    apperr.Localize("BOOK_NOT_FOUND"))

// InvalidArgument requires []FieldViolationOption.
return errors.InvalidArgument("bad request", "INVALID_TITLE", nil,
    []apperr.FieldViolationOption{
        apperr.FieldViolation("title", "must not be empty"),
    },
    apperr.Localize("INVALID_TITLE"))
```

### Which codes require extra details?

| Factory method | Required parameter | AIP-193 detail |
|---|---|---|
| `NotFound`, `AlreadyExists` | `ResourceInfoOption` | ResourceInfo |
| `InvalidArgument` | `[]FieldViolationOption` | BadRequest / FieldViolation |
| `ResourceExhausted` | `[]QuotaViolationOption` | QuotaFailure / QuotaViolation |
| `FailedPrecondition` | `[]PreconditionViolationOption` | PreconditionFailure / PreconditionViolation |
| `Unavailable` | `*RetryInfoOption` (nil-able) | RetryInfo |

All other codes (`Canceled`, `Unknown`, `Internal`, etc.) only require
`reason` and `metadata` for `ErrorInfo` — no additional typed parameter.

### `LocalizeWithReason`

When the `reason` key doubles as the localization message ID, use
`LocalizeWithReason` to avoid repeating it at every call site:

```go
var errors = stricterr.NewFactory("api.store.example.com",
    stricterr.LocalizeWithReason())

// apperr.Localize(reason) is appended automatically.
return errors.NotFound("book not found", "BOOK_NOT_FOUND", nil,
    apperr.ResourceInfo("store.v1.Book", bookID, "", ""))
```

An explicit `apperr.Localize(...)` in `opts` overrides the automatic
value ([last write wins](#option-semantics)).

## Convert and Interceptors

Use the `Convert` function from your installed converter to translate a
resolved error into your framework's native error type. For Connect RPC,
gRPC, Twirp, and HTTP the packages also ship ready-made **interceptors /
middleware** that catch `*apperr.AppError` values, resolve them (with
optional localization), and convert them automatically.

### Converters

- [Connect RPC](https://pkg.go.dev/github.com/harwoeck/apperr/x/connecterr)
  - Install converter `go get github.com/harwoeck/apperr/x/connecterr`
  - ```go
    var connectErr *connect.Error
    connectErr, err := connecterr.Convert(resolved, false)
    ```
- [gRPC](https://pkg.go.dev/github.com/harwoeck/apperr/x/grpcerr)
  - Install converter `go get github.com/harwoeck/apperr/x/grpcerr`
  - ```go
    var grpcStatus *status.Status
    grpcStatus, err := grpcerr.Convert(resolved, false)
    ```
- [HTTP](https://pkg.go.dev/github.com/harwoeck/apperr/x/httperr)
  - Install converter `go get github.com/harwoeck/apperr/x/httperr`
  - ```go
    var httpStatus int
    var httpBody []byte
    httpStatus, httpBody, err := httperr.Convert(resolved, false)
    ```
- [Twirp](https://pkg.go.dev/github.com/harwoeck/apperr/x/twirperr)
  - Install converter `go get github.com/harwoeck/apperr/x/twirperr`
  - ```go
    var twirpErr twirp.Error
    twirpErr, err := twirperr.Convert(resolved, false)
    ```
- [Terminal or Console](https://pkg.go.dev/github.com/harwoeck/apperr/x/terminalerr)
  - Install converter `go get github.com/harwoeck/apperr/x/terminalerr`
  - ```go
    terminalerr.Log(logger, resolved, false)
    ```

### Interceptors / Middleware

The interceptors and middleware handle the full resolve → convert pipeline
for you. They catch any `*apperr.AppError` returned by your handler,
resolve it through the localization pipeline, and return the
framework-native error. Errors that are already in the framework's native
type are passed through unchanged; any other error is wrapped as internal.

All interceptor/middleware options are optional. Without a
`LocalizationProvider`, errors are resolved without translation. Without a
`GetLogFunc`, a noop logger is used.

#### Connect RPC

```go
interceptor := connecterr.NewInterceptor(
    connecterr.WithLocalizationProvider(i18nAdapter),
    connecterr.WithGetClientLanguagesFunc(func(ctx context.Context) []language.Tag {
        return langsFromContext(ctx)
    }),
    // connecterr.EnableDebugInfo(), // opt-in: forward DebugInfo to clients
)

_, handler := examplev1connect.NewBookServiceHandler(svc,
    connect.WithInterceptors(interceptor))
```

#### gRPC

```go
unary := grpcerr.NewUnaryServerInterceptor(
    grpcerr.WithLocalizationProvider(i18nAdapter),
    grpcerr.WithGetClientLanguagesFunc(func(ctx context.Context) []language.Tag {
        return langsFromContext(ctx)
    }),
    // grpcerr.EnableDebugInfo(), // opt-in: forward DebugInfo to clients
)
stream := grpcerr.NewStreamServerInterceptor(
    grpcerr.WithLocalizationProvider(i18nAdapter),
)

srv := grpc.NewServer(
    grpc.UnaryInterceptor(unary),
    grpc.StreamInterceptor(stream),
)
```

#### Twirp

```go
interceptor := twirperr.NewInterceptor(
    twirperr.WithLocalizationProvider(i18nAdapter),
    twirperr.WithGetClientLanguagesFunc(func(ctx context.Context) []language.Tag {
        return langsFromContext(ctx)
    }),
    // twirperr.EnableDebugInfo(), // opt-in: forward DebugInfo to clients
)

srv := examplev1.NewBookServiceServer(svc,
    twirp.WithServerInterceptors(interceptor))
```

#### HTTP

```go
mw := httperr.NewMiddleware(
    httperr.WithLocalizationProvider(i18nAdapter),
    httperr.WithGetClientLanguagesFunc(func(ctx context.Context) []language.Tag {
        return langsFromContext(ctx)
    }),
    // httperr.EnableDebugInfo(), // opt-in: forward DebugInfo to clients
)

http.Handle("/books", mw(func(w http.ResponseWriter, r *http.Request) error {
    // Return an *apperr.AppError to have it resolved and written as JSON.
    return apperr.NotFound("book not found",
        apperr.ResourceInfo("library.v1.Book", bookID, "", ""))
}))
```

---

## Further Details

### Wrapping errors

`*AppError` supports Go's standard error wrapping via `Unwrap()`. Pass
`apperr.Wrap(cause)` as an option to preserve the original error chain:

```go
row := db.QueryRowContext(ctx, "SELECT ...")
if err := row.Scan(&book); err != nil {
    return apperr.NotFound("book not found",
        apperr.Wrap(err),
        apperr.ResourceInfo("library.v1.Book", bookID, "", ""))
}
```

This allows standard `errors.Is` and `errors.As` matching through the
`*AppError`:

```go
if errors.Is(err, sql.ErrNoRows) { /* still works */ }
```

The wrapped error's message is included in `AppError.Error()` output for
logging and debugging, but is **not** exposed in the resolved error sent to
clients — internal details like `sql.ErrNoRows` never leak.

### Checking error codes

Two helpers let you inspect an `*AppError` from an arbitrary `error`
without manual type assertions:

```go
// Check whether the error chain contains a specific code:
if apperr.IsCode(err, errdetails.NotFound) {
    // handle not-found
}

// Extract the *AppError from the chain:
if ae, ok := apperr.AsAppError(err); ok {
    log.Println(ae.Code(), ae.Message())
}
```

Both use `errors.As` internally and work through wrapped error chains.

### Option semantics

Options are applied in order during resolution. Their merge behaviour
depends on the detail type:

| Behaviour | Option types |
|---|---|
| **Last write wins** — only the final option of this type takes effect | `RequestInfo`, `ResourceInfo`, `ErrorInfo`, `Localize` / `LocalizeAny`, `RetryInfo` |
| **Append** — every option adds to a list; all entries are preserved | `HelpLink`, `FieldViolation`, `PreconditionViolation`, `QuotaViolation` |
| **Aggregate** — values are merged into a single `http.Header` map via `Add` | `Headers` |

For the "last write wins" types, if you attach two options of the same
kind (e.g. two `Localize` calls), only the last one is used — with no
warning or error. This is intentional and enables patterns like
`stricterr.LocalizeWithReason()` where an automatic option can be
overridden by an explicit one.

`Headers` is the exception: it uses **aggregate** semantics. Each
`Headers(…)` call merges its entries into a shared `http.Header` map
via `Add`, so values from multiple call sites are preserved. This
allows different parts of the application to contribute response
headers independently.

---

## References

- [AIP-193 — Errors](https://google.aip.dev/193) — Google's API Improvement Proposal for structured error responses with ErrorInfo, ResourceInfo, and other standard detail types.
- [Writing helpful error messages](https://developers.google.com/tech-writing/error-messages) — Google's technical writing guide on crafting clear, actionable error messages.
