package dgsee

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetFilter(t *testing.T) {
	var testData = []struct {
		in    string
		out   string
		panic bool
	}{
		{
			in:  "1",
			out: "NOT eq(val(CNT), 1)",
		},
		{
			in:  "2",
			out: "NOT eq(val(CNT), 2)",
		},
		{
			in:  "0..2",
			out: "lt(val(CNT), 0) OR gt(val(CNT), 2)",
		},
		{
			in:  "1..3",
			out: "lt(val(CNT), 1) OR gt(val(CNT), 3)",
		},
		{
			in:  "1..n",
			out: "lt(val(CNT), 1)",
		},
		{
			in:  "n",
			out: "",
		},
		{
			in:  "0..n",
			out: "",
		},
		{
			in:  "-",
			out: "",
		},
		{
			in:  "42",
			out: "NOT eq(val(CNT), 42)",
		},
	}

	for i, tt := range testData {
		if tt.panic {
			require.Panics(t, func() { getFilter(tt.in) }, "Should have panicked")
		} else {
			require.Exactly(t, tt.out, getFilter(tt.in), "Test i=%d", i)
		}
	}
}
