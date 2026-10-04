# dd [![test](https://github.com/Code-Hex/dd/actions/workflows/test.yml/badge.svg)](https://github.com/Code-Hex/dd/actions/workflows/test.yml) [![Go Reference](https://pkg.go.dev/badge/github.com/Code-Hex/dd.svg)](https://pkg.go.dev/github.com/Code-Hex/dd)

`dd` dumps Go data structures as valid syntax in Go.

- ✅ Simple API
- ✅ Support Go 1.16 ~ (available generics!)
- ✅ Customizable dump format each types
  - Available some options in [`df`](https://github.com/Code-Hex/dd/blob/main/df/df.go) package
- ✅ Support pretty print
  - You can use any color theme you like.

There are several libraries similar to this exist. I like them all, each one leans toward debugging purposes mainly.

- [github.com/davecgh/go-spew/spew](https://github.com/davecgh/go-spew)
- [github.com/k0kubun/pp/v3](https://github.com/k0kubun/pp)

In some cases, we want to use these data structures as test data. None of them output valid Go syntax, so I had to manually modify them.

`dd` solves this problem. Output as valid syntax, we did get also more prettry and readable form.

## Synopsis

### Generator purpose

Add this import line to the file you're working in:

```go
import "github.com/Code-Hex/dd"
```

and just call `Dump` function.

```go
data := map[string]int{
  "b": 2,
  "a": 1,
  "c": 3,
}
fmt.Println(dd.Dump(data))
// map[string]int{
//   "a": 1,
//   "b": 2,
//   "c": 3,
// }

// There are also some options
fmt.Println(dd.Dump(data, dd.WithIndent(4)))
// map[string]int{
//     "a": 1,
//     "b": 2,
//     "c": 3,
// }
```

### Streaming large values

Use `DumpTo` to write directly to an `io.Writer`. `Dump` wraps the same
implementation with a `strings.Builder`.

```go
out := bufio.NewWriter(os.Stdout)
if err := dd.DumpTo(out, data, dd.WithOmitEmptyFields()); err != nil {
    return err
}
return out.Flush()
```

`DumpTo` adds no trailing newline and returns the first write error, including
`io.ErrShortWrite`. The output may be partial on failure. It does not close or
flush the caller's writer. Use a buffered writer for files or network connections
to combine small writes.

Structs and lists stream without building strings for each subtree. Cycle
tracking retains only the active traversal path. Maps still sort their keys and
use a tabwriter buffer to preserve column alignment; custom formatters also use
a tabwriter buffer. Individual quoted strings require temporary storage.
Streaming therefore avoids retaining the complete output for ordinary structs
and lists, but does not promise constant memory for every value.

Very deep values still use the Go call stack, and indentation makes the output
of a linked chain grow quadratically with depth. Choose input sizes appropriate
to available resources. Regression tests cover one million list elements and a
chain of depth 2048.

Performance measurements on Go 1.26.5 (same machine, median of five 100 ms runs):

| Workload | Previous Dump | Streaming-based Dump | Allocated bytes reduction |
| --- | ---: | ---: | ---: |
| 10,000 records | 52.5 ms | 11.1 ms | 56% |
| Parsed Go AST, 100 functions | 77.4 ms | 8.85 ms | 92% |
| Linked chain, depth 256 | 183.6 ms | 0.42 ms | 99.8% |

These are workload-specific allocation totals per call, not peak memory or
performance guarantees. Reproduce with
`go test -run '^$' -bench 'BenchmarkDump(To)?$' -benchmem -benchtime=100ms -count=5`.

### Debugging purpose

Add this import line to the file you're working in:

```go
import "github.com/Code-Hex/dd/p"
```

and just call `p.P`

```go
func main() {
  srv := &http.Server{
    Addr:    ":8080",
    Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
  }
  fmt.Println("--- monokai")
  p.P(srv)
}
```

<img width="530" alt="color.png" src="https://user-images.githubusercontent.com/6500104/159877754-976b2d48-7b58-493f-8ff7-589b782d690a.png">

You can read [examples/pretty/main.go](https://github.com/Code-Hex/dd/blob/main/examples/pretty/main.go). If you want to adopt a color theme of your own choice, the following links will help you: [pkg.go.dev/github.com/alecthomas/chroma/styles](https://pkg.go.dev/github.com/alecthomas/chroma/styles).

## Customize the format

`WithDumpFunc` option helps you if you want to customize the format for each type. This option works as code using Generics for 1.18 and above, otherwise it uses reflect.

Several wrapper options using this option are provided in the [`df`](https://github.com/Code-Hex/dd/blob/main/df/df.go) package.

```go
import "github.com/Code-Hex/dd/df"
```

and call `Dump` function with options within the package.

```go
// json.RawMessage(`{"message":"Hello, World"}`)
fmt.Println(
  dd.Dump(
    json.RawMessage(`{"message":"Hello, World"}`),
    df.WithJSONRawMessage(),
  ),
)

// func() []byte {
//   // 00000000  48 65 6c 6c 6f 2c 20 57  6f 72 6c 64              |Hello, World|
//
//   return []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64}
// }()
fmt.Println(
  dd.Dump([]byte("Hello, World"), df.WithRichBytes()),
)
```

## License

MIT License

Copyright (c) 2022 codehex

## Omit zero-value fields

Use `dd.WithOmitEmptyFields()` to omit zero-value struct fields recursively:

```go
type Config struct {
    Name string
    Retries int
}
fmt.Println(dd.Dump(Config{Name: "worker"}, dd.WithOmitEmptyFields()))
// main.Config{
//   Name: "worker",
// }
```

This uses `reflect.Value.IsZero`: nil slices and maps are omitted, but non-nil
empty slices and maps are retained. Non-nil pointers and interfaces containing
typed nil values are also retained. It composes with `WithExportedOnly`,
`WithIndent`, and custom dump functions. Fields are filtered before custom dump
functions run; custom functions control their own output.

## Go language compatibility

The CI matrix covers Go 1.16 through Go 1.27. Newer syntax tests have version
build constraints so older Go versions can still run the main test suite.
The module's `go` directive remains at 1.18; the pre-generics API is selected by
build tags on Go 1.16 and 1.17.

`dd` inspects runtime values, rather than parsing the source that created them.
The following language changes were reviewed against the official release notes:

| Go | Language changes relevant to this review | Effect on dumping |
| --- | --- | --- |
| [1.18](https://go.dev/doc/go1.18#language) | Type parameters, type sets, `any`, `comparable` | Concrete generic structs and functions are tested. Constraints are not runtime values. |
| [1.19](https://go.dev/doc/go1.19#language) | Method type-parameter scope correction; memory-model clarification | No new value syntax to emit. |
| [1.20](https://go.dev/doc/go1.20#language) | Slice-to-array conversions; comparable constraint relaxation | Resulting arrays and instantiated types use existing handlers. |
| [1.21](https://go.dev/doc/go1.21#language) | `min`, `max`, `clear`; improved type inference | Built-ins produce ordinary values. Tests no longer depend on private standard-library layouts (issue #21). |
| [1.22](https://go.dev/doc/go1.22#language) | Integer range; per-iteration loop variables | These change value creation, not dump syntax. |
| [1.23](https://go.dev/doc/go1.23#language) | Range over iterator functions | `iter.Seq` and `iter.Seq2` are tested as function values, without executing them. |
| [1.24](https://go.dev/doc/go1.24#language) | Generic type aliases | Aliases dump as their underlying runtime type; covered by tests. |
| [1.25](https://go.dev/doc/go1.25#language) | No program-affecting language changes | Existing handlers apply. |
| [1.26](https://go.dev/doc/go1.26#language) | `new(expression)`; self-referential generic constraints | Expression-created pointers use the existing pointer format; constraints do not introduce runtime kinds. |
| [1.27](https://go.dev/doc/go1.27#language) | Generic methods, field selectors in struct literal keys, broader function type inference | Instantiated method values, promoted-field literals, and inferred function assignments are tested. Struct output continues to use explicit nested fields. |

Named function dumps preserve variadic arguments and all result types. Generated
function and composite-value expressions in the compatibility tests are checked
with `go/types`, in addition to the existing syntax checks.

Dumping does not reconstruct original source expressions, alias spelling, or
function bodies. Non-nil primitive pointers keep their existing address-based
representation, including those created with `new(expression)`; that output is
for inspecting the current process, not portable serialized data. Types with
unexported fields may require `WithExportedOnly` or a custom dump function before
using the output in a different package.

Generic type names come from reflection. For type arguments from packages with
multi-segment import paths, reflection can include that path in the name (for
example, `Box[net/http.Cookie]`), which is not a valid Go type expression. Use
`WithDumpFunc` to supply the qualified type name and output for these values.
