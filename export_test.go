package dd_test

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	dd "github.com/Code-Hex/go-data-dumper"
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
			v:    int(math.MaxInt),
			want: strconv.FormatInt(math.MaxInt, 10),
		},
		{
			name: "min int",
			v:    int(math.MinInt),
			want: strconv.FormatInt(math.MinInt, 10),
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
			v:    uint(math.MaxUint),
			want: strconv.FormatUint(math.MaxUint, 10),
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
			name: "func(int, int) bool { return false }",
			v:    func(int, int) bool { return false },
			want: "func(int, int) bool {\n  // ...\n  return false\n}",
		},
		{
			name: "func(int, int) (bool, error) { return false, nil }",
			v:    func(int, int) (bool, error) { return false, nil },
			want: "func(int, int) (bool, error) {\n  // ...\n  return false, nil\n}",
		},
	}
	t.Run("typed", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				if got := dd.Dump(tc.v); tc.want != got {
					t.Fatalf("want %q, but got %q", tc.want, got)
				}
			})
		}
	})

	t.Run("wrapped with interface", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				if got := dd.Dump(interface{}(tc.v)); tc.want != got {
					t.Fatalf("want %q, but got %q", tc.want, got)
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
	}
	t.Run("typed", func(t *testing.T) {
		for _, tc := range cases {
			tc := tc
			t.Run(tc.name, func(t *testing.T) {
				got := dd.Dump(tc.v)
				if !strings.Contains(got, tc.want) {
					t.Fatalf("want %q, but got %q", tc.want, got)
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
			})
		}
	})
}

func TestWithDumpFunc(t *testing.T) {
	cases := []struct {
		name       string
		v          interface{}
		want       string
		dumpOption dd.OptionFunc
	}{
		{
			name:       "time unix date",
			v:          time.Date(2022, 3, 6, 12, 0, 0, 0, time.UTC),
			want:       "func() time.Time {\n  tmp, _ := time.Parse(\"Mon Jan _2 15:04:05 MST 2006\", \"Sun Mar  6 12:00:00 UTC 2022\")\n  return tmp\n}",
			dumpOption: dd.WithTime(time.UnixDate),
		},
		{
			name:       "big int",
			v:          big.NewInt(10),
			want:       "func() *big.Int {\n  tmp := new(big.Int)\n  tmp.SetString(\"10\")\n  return tmp\n}",
			dumpOption: dd.WithBigInt(),
		},
		{
			name:       "big float",
			v:          big.NewFloat(12345.6789),
			want:       "func() *big.Float {\n  tmp := new(big.Float)\n  tmp.SetString(\"12345.6789\")\n  return tmp\n}",
			dumpOption: dd.WithBigFloat(),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := dd.Dump(tc.v, tc.dumpOption)
			if tc.want != got {
				t.Fatalf("want %q, but got %q", tc.want, got)
			}
		})
	}
}
