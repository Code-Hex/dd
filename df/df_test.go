package df_test

import (
	"encoding/json"
	"go/parser"
	"math/big"
	"testing"
	"time"

	"github.com/Code-Hex/dd"
	"github.com/Code-Hex/dd/df"
)

func TestWithDumpFunc(t *testing.T) {
	cases := []struct {
		name       string
		v          any
		want       string
		dumpOption dd.OptionFunc
	}{
		{
			name:       "time unix date",
			v:          time.Date(2022, 3, 6, 12, 0, 0, 0, time.UTC),
			want:       "func() time.Time {\n  tmp, _ := time.Parse(\"Mon Jan _2 15:04:05 MST 2006\", \"Sun Mar  6 12:00:00 UTC 2022\")\n  return tmp\n}()",
			dumpOption: df.WithTime(time.UnixDate),
		},
		{
			name:       "big int",
			v:          big.NewInt(10),
			want:       "func() *big.Int {\n  tmp := new(big.Int)\n  tmp.SetString(\"10\")\n  return tmp\n}()",
			dumpOption: df.WithBigInt(),
		},
		{
			name:       "big float",
			v:          big.NewFloat(12345.6789),
			want:       "func() *big.Float {\n  tmp := new(big.Float)\n  tmp.SetString(\"12345.6789\")\n  return tmp\n}()",
			dumpOption: df.WithBigFloat(),
		},
		{
			name:       "json.RawMessage",
			v:          json.RawMessage(`{"hello":"world"}`),
			want:       "json.RawMessage(`{\"hello\":\"world\"}`)",
			dumpOption: df.WithJSONRawMessage(),
		},
		{
			name:       "[]byte",
			v:          []byte("Hello, World"),
			want:       "func() []byte {\n  // 00000000  48 65 6c 6c 6f 2c 20 57  6f 72 6c 64              |Hello, World|\n  // \n  return []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64}\n}()",
			dumpOption: df.WithRichBytes(),
		},
	}
	for _, tc := range cases {
		tc := tc
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
