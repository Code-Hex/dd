package dd_test

import (
	"context"
	"fmt"
	"go/parser"
	"math"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/Code-Hex/dd"
)

const (
	intSize = 32 << (^uint(0) >> 63) // 32 or 64

	MaxInt  = 1<<(intSize-1) - 1
	MinInt  = -1 << (intSize - 1)
	MaxUint = 1<<intSize - 1
)

func TestDumpBasic(t *testing.T) {
	cases := []struct {
		name string
		v    interface{}
		want string
	}{
		{
			name: "immediate nil",
			v:    nil,
			want: "nil",
		},
		{
			name: "string",
			v:    "Hello, World",
			want: strconv.Quote("Hello, World"),
		},
		{
			name: "true",
			v:    true,
			want: "true",
		},
		{
			name: "false",
			v:    false,
			want: "false",
		},
		{
			name: "max int",
			v:    int(MaxInt),
			want: strconv.FormatInt(MaxInt, 10),
		},
		{
			name: "min int",
			v:    int(MinInt),
			want: strconv.FormatInt(MinInt, 10),
		},
		{
			name: "max int8",
			v:    int8(math.MaxInt8),
			want: strconv.FormatInt(math.MaxInt8, 10),
		},
		{
			name: "min int8",
			v:    int8(math.MinInt8),
			want: strconv.FormatInt(math.MinInt8, 10),
		},
		{
			name: "max int16",
			v:    int16(math.MaxInt16),
			want: strconv.FormatInt(math.MaxInt16, 10),
		},
		{
			name: "min int16",
			v:    int16(math.MinInt16),
			want: strconv.FormatInt(math.MinInt16, 10),
		},
		{
			name: "max int32",
			v:    int32(math.MaxInt32),
			want: strconv.FormatInt(math.MaxInt32, 10),
		},
		{
			name: "min int32",
			v:    int32(math.MinInt32),
			want: strconv.FormatInt(math.MinInt32, 10),
		},
		{
			name: "max int64",
			v:    int64(math.MaxInt64),
			want: strconv.FormatInt(math.MaxInt64, 10),
		},
		{
			name: "min int64",
			v:    int64(math.MinInt64),
			want: strconv.FormatInt(math.MinInt64, 10),
		},
		{
			name: "max uint",
			v:    uint(MaxUint),
			want: strconv.FormatUint(MaxUint, 10),
		},
		{
			name: "max uint8",
			v:    uint8(math.MaxUint8),
			want: strconv.FormatUint(math.MaxUint8, 10),
		},
		{
			name: "max uint16",
			v:    uint16(math.MaxUint16),
			want: strconv.FormatUint(math.MaxUint16, 10),
		},
		{
			name: "max uint32",
			v:    uint32(math.MaxUint32),
			want: strconv.FormatUint(math.MaxUint32, 10),
		},
		{
			name: "max uint64",
			v:    uint64(math.MaxUint64),
			want: strconv.FormatUint(math.MaxUint64, 10),
		},
		{
			name: "max float32",
			v:    float32(math.MaxFloat32),
			want: fmt.Sprintf("%f", float32(math.MaxFloat32)),
		},
		{
			name: "max float64",
			v:    float64(math.MaxFloat64),
			want: fmt.Sprintf("%f", float64(math.MaxFloat64)),
		},
		{
			name: "max complex64",
			v:    complex64(complex(float32(math.MaxFloat32), float32(math.MaxFloat32))),
			want: fmt.Sprintf("%v",
				complex64(complex(float32(math.MaxFloat32), float32(math.MaxFloat32))),
			),
		},
		{
			name: "max complex128",
			v:    complex128(complex(float64(math.MaxFloat64), float64(math.MaxFloat64))),
			want: fmt.Sprintf("%v",
				complex128(complex(float64(math.MaxFloat64), float64(math.MaxFloat64))),
			),
		},
		{
			name: "array [0]int{}",
			v:    [0]int{},
			want: "[0]int{}",
		},
		{
			name: "array [2]int{}",
			v:    [2]int{1, 2},
			want: "[2]int{\n  1,\n  2,\n}",
			// [2]int{
			//   1, // 2 spaces indent
			//   2,
			// }
		},
		{
			name: "array [2]interface {}{}",
			v:    [2]interface{}{1, "hello"},
			want: "[2]interface {}{\n  1,\n  \"hello\",\n}",
			// [2]interface {}{
			//   1, // 2 spaces indent
			//   "hello",
			// }
		},
		{
			name: "slice []int{}",
			v:    []int{},
			want: "[]int{}",
		},
		{
			name: "slice ([]int)(nil)",
			v:    ([]int)(nil),
			want: "([]int)(nil)",
		},
		{
			name: "slice []int{1, 2}",
			v:    []int{1, 2},
			want: "[]int{\n  1,\n  2,\n}",
			// []int{
			//   1, // 2 spaces indent
			//   2,
			// }
		},
		{
			name: "slice []interface {}{}",
			v:    []interface{}{1, "hello"},
			want: "[]interface {}{\n  1,\n  \"hello\",\n}",
			// []int{
			//   1, // 2 spaces indent
			//   "hello",
			// }
		},
		{
			name: "(map[string]int)(nil)",
			v:    (map[string]int)(nil),
			want: "(map[string]int)(nil)",
		},
		{
			name: "map[string]int{}",
			v:    map[string]int{},
			want: "map[string]int{}",
		},
		{
			name: "map[string]int{}",
			v: map[string]int{
				"b": 2,
				"a": 1,
				"c": 3,
			},
			want: "map[string]int{\n  \"a\": 1,\n  \"b\": 2,\n  \"c\": 3,\n}",
			// map[string]int{
			//   "a": 1,
			//   "b": 2,
			//   "c": 3,
			// }
		},
		{
			name: "chan int",
			v:    (chan int)(nil),
			want: "(chan int)(nil)",
		},
		{
			name: "<-chan int",
			v:    (<-chan int)(nil),
			want: "(<-chan int)(nil)",
		},
		{
			name: "chan<- int",
			v:    (chan<- int)(nil),
			want: "(chan<- int)(nil)",
		},
		{
			name: "struct {}{}",
			v:    struct{}{},
			want: "struct {}{}",
		},
		{
			name: "struct {age int}{name: 10}",
			v:    struct{ age int }{age: 10},
			want: "struct { age int }{\n  age: 10,\n}",
		},
		{
			name: "empty func",
			v:    func() {},
			want: "func() {\n  // ...\n}",
		},
		{
			name: "nil empty func",
			v:    (func())(nil),
			want: "(func())(nil)",
		},
		{
			name: "func(int, int) bool { return false }",
			v:    func(int, int) bool { return false },
			want: "func(int, int) bool {\n  // ...\n  return false\n}",
		},
		{
			name: "func(int, int) (bool, error) { return false, nil }",
			v:    func(int, int) (bool, error) { return false, nil },
			want: "func(int, int) (bool, error) {\n  // ...\n  return false, nil\n}",
		},
		{
			name: "nil func(int, int) bool",
			v:    (func(int, int) bool)(nil),
			want: "(func(int, int) bool)(nil)",
		},
		{
			name: "reflect.Value{}",
			v:    reflect.Value{},
			want: "reflect.Value{\n  typ: (*reflect.rtype)(nil),\n  ptr: unsafe.Pointer(uintptr(0)),\n  flag: 0,\n}",
		},
		{
			name: "context.CancelFunc",
			v:    context.CancelFunc(func() {}),
			want: "context.CancelFunc(func() {\n  // ...\n})",
		},
		{
			name: "http.HandlerFunc",
			v:    http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
			want: "http.HandlerFunc(func(http.ResponseWriter, *http.Request) {\n  // ...\n})",
		},
		{
			name: "(context.CancelFunc)(nil)",
			v:    (context.CancelFunc)(nil),
			want: "(context.CancelFunc)(nil)",
		},
		{
			name: "(http.HandlerFunc)(nil)",
			v:    (http.HandlerFunc)(nil),
			want: "(http.HandlerFunc)(nil)",
		},
	}
	t.Run("typed", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := dd.Dump(tc.v)
				if tc.want != got {
					t.Fatalf("want %q, but got %q", tc.want, got)
				}
				if _, err := parser.ParseExpr(got); err != nil {
					t.Fatal(err)
				}
			})
		}
	})

	t.Run("wrapped with interface", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := dd.Dump(interface{}(tc.v))
				if tc.want != got {
					t.Fatalf("want %q, but got %q", tc.want, got)
				}
				if _, err := parser.ParseExpr(got); err != nil {
					t.Fatal(err)
				}
			})
		}
	})

}

func TestPointer(t *testing.T) {
	cases := []struct {
		name string
		v    interface{}
		want string
	}{
		{
			name: "nil pointer of int",
			v:    (*int)(nil),
			want: "(*int)(nil)",
		},
		{
			name: "nil pointer of pointer of int",
			v:    (**int)(nil),
			want: "(**int)(nil)",
		},
		{
			name: "pointer of int",
			v:    new(int),
			want: "(*int)(unsafe.Pointer(uintptr(",
		},
		{
			name: "pointer of string",
			v:    new(string),
			want: "(*string)(unsafe.Pointer(uintptr(",
		},
		{
			name: "pointer of bool",
			v:    new(bool),
			want: "(*bool)(unsafe.Pointer(uintptr(",
		},
		{
			name: "pointer of uint8",
			v:    new(uint8),
			want: "(*uint8)(unsafe.Pointer(uintptr(",
		},
		{
			name: "pointer of struct",
			v:    &struct{ age int }{age: 10},
			want: "&struct { age int }{\n  age: 10,\n}",
		},
		{
			name: "pointer of pointer of struct",
			v: func() interface{} {
				a := &struct{ age int }{age: 10}
				return &a
			}(),
			want: "(**struct { age int })(unsafe.Pointer(uintptr(",
		},
		{
			name: "pointer of slice",
			v:    &[]int{1, 2},
			want: "&[]int{\n  1,\n  2,\n}",
		},
		{
			name: "pointer of array",
			v:    &[2]int{1, 2},
			want: "&[2]int{\n  1,\n  2,\n}",
		},
		{
			name: "unsafe.Pointer",
			v:    (unsafe.Pointer(&[2]int{1, 2})),
			want: "unsafe.Pointer(uintptr(",
		},
		{
			name: "chan int",
			v:    make(chan int),
			want: "(chan int)(unsafe.Pointer(uintptr(",
		},
	}
	t.Run("typed", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := dd.Dump(tc.v)
				if !strings.Contains(got, tc.want) {
					t.Fatalf("want %q, but got %q", tc.want, got)
				}
				if _, err := parser.ParseExpr(got); err != nil {
					t.Fatal(err)
				}
			})
		}
	})

	t.Run("wrapped with interface", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := dd.Dump(interface{}(tc.v))
				if !strings.Contains(got, tc.want) {
					t.Fatalf("want %q, but got %q", tc.want, got)
				}
				if _, err := parser.ParseExpr(got); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
}

func TestWithIndent(t *testing.T) {
	want := "[]int{\n    1,\n    2,\n}"
	got := dd.Dump([]int{1, 2}, dd.WithIndent(4))
	if want != got {
		t.Fatalf("want %q, but got %q", want, got)
	}
	if _, err := parser.ParseExpr(got); err != nil {
		t.Fatal(err)
	}
}

func TestWithExportedOnly(t *testing.T) {
	// contains exported and unexported
	fh := &multipart.FileHeader{
		Filename: "file1",
		Header:   make(textproto.MIMEHeader),
		Size:     10,
	}
	got := dd.Dump(fh, dd.WithExportedOnly())
	want := "&multipart.FileHeader{\n  Filename: \"file1\",\n  Header: textproto.MIMEHeader{},\n  Size: 10,\n}"
	if want != got {
		t.Fatalf("want %q, but got %q", want, got)
	}
	if _, err := parser.ParseExpr(got); err != nil {
		t.Fatal(err)
	}
}

func TestWithUintFormat(t *testing.T) {
	cases := []struct {
		name       string
		v          interface{}
		want       string
		dumpOption dd.OptionFunc
	}{
		{
			name:       "uint8 binary format",
			v:          uint8(0),
			want:       "0b00000000",
			dumpOption: dd.WithUintFormat(dd.BinaryUint),
		},
		{
			name:       "uint8 hex format",
			v:          uint8(0),
			want:       "0x00",
			dumpOption: dd.WithUintFormat(dd.HexUint),
		},
		{
			name:       "uint16 binary format",
			v:          uint16(0),
			want:       "0b0000000000000000",
			dumpOption: dd.WithUintFormat(dd.BinaryUint),
		},
		{
			name:       "uint16 hex format",
			v:          uint16(0),
			want:       "0x0000",
			dumpOption: dd.WithUintFormat(dd.HexUint),
		},
		{
			name:       "uint32 binary format",
			v:          uint32(0),
			want:       "0b00000000000000000000000000000000",
			dumpOption: dd.WithUintFormat(dd.BinaryUint),
		},
		{
			name:       "uint32 hex format",
			v:          uint32(0),
			want:       "0x00000000",
			dumpOption: dd.WithUintFormat(dd.HexUint),
		},
		{
			name:       "uint64 binary format",
			v:          uint64(0),
			want:       "0b0000000000000000000000000000000000000000000000000000000000000000",
			dumpOption: dd.WithUintFormat(dd.BinaryUint),
		},
		{
			name:       "uint64 hex format",
			v:          uint64(0),
			want:       "0x0000000000000000",
			dumpOption: dd.WithUintFormat(dd.HexUint),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := dd.Dump(tc.v, tc.dumpOption)
			if tc.want != got {
				t.Fatalf("want %q, but got %q", tc.want, got)
			}
			if _, err := parser.ParseExpr(got); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCircularRefs(t *testing.T) {
	cases := []struct {
		name string
		v    interface{}
		want string
	}{
		{
			name: "struct",
			v: func() interface{} {
				type A struct {
					a *A
				}
				a := &A{}
				a.a = a
				return a
			}(),
			want: "a: (*dd_test.A)(unsafe.Pointer(uintptr(",
		},
		{
			name: "map",
			v: func() interface{} {
				a := map[struct{}]interface{}{}
				a[struct{}{}] = a
				return a
			}(),
			want: "struct {}{}: (map[struct {}]interface {})(unsafe.Pointer(uintptr(",
		},
		{
			name: "slice",
			v: func() interface{} {
				a := make([]interface{}, 1)
				a[0] = a
				return a
			}(),
			want: "([]interface {})(unsafe.Pointer(uintptr(",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := dd.Dump(tc.v)
			if !strings.Contains(got, tc.want) {
				t.Log(got)
				t.Fatalf("want %q, but got %q", tc.want, got)
			}
			if _, err := parser.ParseExpr(got); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWithListBreakLineSize(t *testing.T) {
	cases := []struct {
		name       string
		v          interface{}
		want       string
		dumpOption dd.OptionFunc
	}{
		{
			name:       "slice break line at 1",
			v:          []int{1, 2, 3, 4},
			want:       "[]int{\n    1,\n    2,\n    3,\n    4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 1),
		},
		{
			name:       "slice break line at 2",
			v:          []int{1, 2, 3, 4},
			want:       "[]int{\n    1, 2,\n    3, 4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 2),
		},
		{
			name:       "slice break line at 3",
			v:          []int{1, 2, 3, 4},
			want:       "[]int{\n    1, 2, 3,\n    4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 3),
		},
		{
			name:       "slice break line at 4",
			v:          []int{1, 2, 3, 4},
			want:       "[]int{\n    1, 2, 3, 4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 4),
		},
		{
			name:       "slice break line at 5",
			v:          []int{1, 2, 3, 4},
			want:       "[]int{\n    1, 2, 3, 4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 5),
		},
		{
			name:       "array break line at 1",
			v:          [4]int{1, 2, 3, 4},
			want:       "[4]int{\n    1,\n    2,\n    3,\n    4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 1),
		},
		{
			name:       "array break line at 2",
			v:          [4]int{1, 2, 3, 4},
			want:       "[4]int{\n    1, 2,\n    3, 4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 2),
		},
		{
			name:       "array break line at 3",
			v:          [4]int{1, 2, 3, 4},
			want:       "[4]int{\n    1, 2, 3,\n    4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 3),
		},
		{
			name:       "array break line at 4",
			v:          [4]int{1, 2, 3, 4},
			want:       "[4]int{\n    1, 2, 3, 4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 4),
		},
		{
			name:       "array break line at 5",
			v:          [4]int{1, 2, 3, 4},
			want:       "[4]int{\n    1, 2, 3, 4,\n}",
			dumpOption: dd.WithListBreakLineSize(int(0), 5),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := dd.Dump(tc.v, tc.dumpOption, dd.WithIndent(4))
			if tc.want != got {
				t.Fatalf("want %q, but got %q", tc.want, got)
			}
			if _, err := parser.ParseExpr(got); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUnexportedField(t *testing.T) {
	type stringKey struct{}
	ctx := context.WithValue(
		context.Background(),
		stringKey{},
		"value",
	)
	got := dd.Dump(ctx, dd.WithDumpFunc(func(s string, w dd.Writer) {
		w.Write(s)
	}))
	got = addressReplaceRegexp.ReplaceAllString(got, "0x0")
	want := "&context.valueCtx{\n  Context: (*context.emptyCtx)(unsafe.Pointer(0x0)),\n  key: dd_test.stringKey{},\n  val: value,\n}"
	if want != got {
		t.Fatalf("want %q, but got %q", want, got)
	}
	if _, err := parser.ParseExpr(got); err != nil {
		t.Fatal(err)
	}
}
