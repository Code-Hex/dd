package dd

import (
	"encoding/json"
	"go/parser"
	"math/big"
	"testing"
	"time"
)

func TestWithDumpFunc(t *testing.T) {
	cases := []struct {
		name       string
		v          any
		want       string
		dumpOption OptionFunc
	}{
		{
			name:       "time unix date",
			v:          time.Date(2022, 3, 6, 12, 0, 0, 0, time.UTC),
			want:       "func() time.Time {\n  tmp, _ := time.Parse(\"Mon Jan _2 15:04:05 MST 2006\", \"Sun Mar  6 12:00:00 UTC 2022\")\n  return tmp\n}",
			dumpOption: WithTime(time.UnixDate),
		},
		{
			name:       "big int",
			v:          big.NewInt(10),
			want:       "func() *big.Int {\n  tmp := new(big.Int)\n  tmp.SetString(\"10\")\n  return tmp\n}",
			dumpOption: WithBigInt(),
		},
		{
			name:       "big float",
			v:          big.NewFloat(12345.6789),
			want:       "func() *big.Float {\n  tmp := new(big.Float)\n  tmp.SetString(\"12345.6789\")\n  return tmp\n}",
			dumpOption: WithBigFloat(),
		},
		{
			name:       "json.RawMessage",
			v:          json.RawMessage(`{"hello":"world"}`),
			want:       "json.RawMessage(`{\"hello\":\"world\"}`)",
			dumpOption: WithJSONRawMessage(),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := Dump(tc.v, tc.dumpOption)
			if tc.want != got {
				t.Fatalf("want %q, but got %q", tc.want, got)
			}
			if _, err := parser.ParseExpr(got); err != nil {
				t.Fatal(err)
			}
		})
	}
}
