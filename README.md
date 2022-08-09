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
