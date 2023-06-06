package swagger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

func TestJSONValueValueMap(t *testing.T) {
	cases := []struct {
		name   string
		input  []JSONValue
		expect map[string]string
	}{
		{
			name: "single object",
			input: []JSONValue{
				JSONObject{
					value: map[string]JSONValue{
						"p1": JSONPrimitive[float64]{
							value: 0.5,
							addr:  "p1",
						},
						"p2": JSONPrimitive[string]{
							value: "abc",
							addr:  "p2",
						},
						"p3": JSONPrimitive[bool]{
							value: true,
							addr:  "p3",
						},
					},
				},
			},
			expect: map[string]string{
				"0.5":  "p1",
				"abc":  "p2",
				"TRUE": "p3",
			},
		},
		{
			name: "single array",
			input: []JSONValue{
				JSONArray{
					value: []JSONValue{
						JSONPrimitive[bool]{
							value: true,
							addr:  "*",
						},
					},
				},
			},
			expect: map[string]string{
				"TRUE": "*",
			},
		},
		{
			name: "mixed with duplicated values",
			input: []JSONValue{
				JSONArray{
					value: []JSONValue{
						JSONPrimitive[bool]{
							value: true,
							addr:  "*",
						},
					},
				},
				JSONObject{
					value: map[string]JSONValue{
						"p1": JSONPrimitive[float64]{
							value: 0.5,
							addr:  "p1",
						},
						"p2": JSONPrimitive[string]{
							value: "abc",
							addr:  "p2",
						},
						"p3": JSONPrimitive[bool]{
							value: true,
							addr:  "p3",
						},
					},
				},
			},
			expect: map[string]string{
				"0.5": "p1",
				"abc": "p2",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			m, err := JSONValueValueMap(tt.input...)
			require.NoError(t, err)
			require.Equal(t, tt.expect, m)
		})
	}
}

func TestUnmarshalJSONToJSONValue(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	specpathSyn := filepath.Join(pwd, "testdata", "syn.json")

	cases := []struct {
		ref    string
		input  string
		expect JSONValue
	}{
		{
			ref: specpathSyn + "#/definitions/object",
			input: `
{
  "array": [
    "b"
  ],
  "boolean": true,
  "emptyObject": {
  	"OBJKEY": "OBJVAL"
  },
  "integer": 1,
  "map": {
    "KEY": "c"
  },
  "number": 1.5,
  "object": {
  	"p1": "d",
	"obj": {
		"pp1": 2
	}
  },
  "string": "e"
}`,
			expect: JSONObject{
				value: map[string]JSONValue{
					"array": JSONArray{
						value: []JSONValue{
							JSONPrimitive[string]{
								value: "b",
								addr:  "array.*",
							},
						},
						addr: "array",
					},
					"boolean": JSONPrimitive[bool]{
						value: true,
						addr:  "boolean",
					},
					"emptyObject": JSONObject{
						value: map[string]JSONValue{
							"OBJKEY": JSONPrimitive[string]{
								value: "OBJVAL",
							},
						},
						addr: "emptyObject",
					},
					"integer": JSONPrimitive[float64]{
						value: 1,
						addr:  "integer",
					},
					"map": JSONObject{
						value: map[string]JSONValue{
							"KEY": JSONPrimitive[string]{
								value: "c",
								addr:  "map.*",
							},
						},
						addr: "map",
					},
					"number": JSONPrimitive[float64]{
						value: 1.5,
						addr:  "number",
					},
					"object": JSONObject{
						value: map[string]JSONValue{
							"p1": JSONPrimitive[string]{
								value: "d",
								addr:  "object.p1",
							},
							"obj": JSONObject{
								value: map[string]JSONValue{
									"pp1": JSONPrimitive[float64]{
										value: 2,
										addr:  "object.obj.pp1",
									},
								},
								addr: "object.obj",
							},
						},
						addr: "object",
					},
					"string": JSONPrimitive[string]{
						value: "e",
						addr:  "string",
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.ref, func(t *testing.T) {
			ref := spec.MustCreateRef(tt.ref)
			exp, err := NewExpander(ref)
			require.NoError(t, err)
			require.NoError(t, exp.Expand())
			v, err := UnmarshalJSONToJSONValue([]byte(tt.input), exp.root)
			require.NoError(t, err)
			require.Equal(t, tt.expect, v)
		})
	}
}
