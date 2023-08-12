package ctrl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONValueMap(t *testing.T) {
	cases := []struct {
		name   string
		input  map[string]interface{}
		expect map[string]string
	}{
		{
			name: "simple object",
			input: map[string]interface{}{
				"str":    "foo",
				"number": 1,
				"bool":   true,
				"array": []interface{}{
					1,
				},
				"object": map[string]interface{}{
					"p1": "bar",
				},
				"null": nil,
			},
			expect: map[string]string{
				"bar":  "/object/p1",
				"foo":  "/str",
				"TRUE": "/bool",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			out := jsonValueMap(tt.input)
			require.Equal(t, tt.expect, out)
		})
	}
}
