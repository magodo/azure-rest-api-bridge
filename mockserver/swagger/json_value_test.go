package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

func TestJSONValueValueMap(t *testing.T) {
	cases := []struct {
		name   string
		input  []JSONValue
		expect map[string]*JSONValuePos
	}{
		{
			name: "single object",
			input: []JSONValue{
				JSONObject{
					value: map[string]JSONValue{
						"p1": JSONPrimitive[float64]{
							value: 0.5,
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("p1"),
								Addr: ParseAddr("p1"),
							},
						},
						"p2": JSONPrimitive[string]{
							value: "abc",
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("p2"),
								Addr: ParseAddr("p2"),
							},
						},
						"p3": JSONPrimitive[bool]{
							value: true,
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("p3"),
								Addr: ParseAddr("p3"),
							},
						},
					},
				},
			},
			expect: map[string]*JSONValuePos{
				"0.5": {
					Ref:  jsonreference.MustCreateRef("p1"),
					Addr: ParseAddr("p1"),
				},
				"abc": {
					Ref:  jsonreference.MustCreateRef("p2"),
					Addr: ParseAddr("p2"),
				},
				"TRUE": {
					Ref:  jsonreference.MustCreateRef("p3"),
					Addr: ParseAddr("p3"),
				},
			},
		},
		{
			name: "single array",
			input: []JSONValue{
				JSONArray{
					value: []JSONValue{
						JSONPrimitive[bool]{
							value: true,
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("*"),
								Addr: ParseAddr("*"),
							},
						},
					},
				},
			},
			expect: map[string]*JSONValuePos{
				"TRUE": {
					Ref:  jsonreference.MustCreateRef("*"),
					Addr: ParseAddr("*"),
				},
			},
		},
		{
			name: "mixed with duplicated values",
			input: []JSONValue{
				JSONObject{
					value: map[string]JSONValue{
						"p1": JSONPrimitive[float64]{
							value: 0.5,
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("p1"),
								Addr: ParseAddr("p1"),
							},
						},
						"p2": JSONPrimitive[string]{
							value: "abc",
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("p2"),
								Addr: ParseAddr("p2"),
							},
						},
						"p3": JSONPrimitive[bool]{
							value: true,
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("p3"),
								Addr: ParseAddr("p3"),
							},
						},
					},
				},
				JSONArray{
					value: []JSONValue{
						JSONPrimitive[bool]{
							value: true,
							pos: &JSONValuePos{
								Ref:  jsonreference.MustCreateRef("*"),
								Addr: ParseAddr("*"),
							},
						},
					},
				},
			},
			expect: map[string]*JSONValuePos{
				"0.5": {
					Ref:  jsonreference.MustCreateRef("p1"),
					Addr: ParseAddr("p1"),
				},
				"abc": {
					Ref:  jsonreference.MustCreateRef("p2"),
					Addr: ParseAddr("p2"),
				},
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
								pos: &JSONValuePos{
									Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/array/items"),
									Addr: ParseAddr("array.*"),
								},
							},
						},
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/array"),
							Addr: ParseAddr("array"),
						},
					},
					"boolean": JSONPrimitive[bool]{
						value: true,
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/boolean"),
							Addr: ParseAddr("boolean"),
						},
					},
					"emptyObject": JSONObject{
						value: map[string]JSONValue{
							"OBJKEY": JSONPrimitive[string]{
								value: "OBJVAL",
								pos:   nil,
							},
						},
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/emptyObject"),
							Addr: ParseAddr("emptyObject"),
						},
					},
					"integer": JSONPrimitive[float64]{
						value: 1,
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/integer"),
							Addr: ParseAddr("integer"),
						},
					},
					"map": JSONObject{
						value: map[string]JSONValue{
							"KEY": JSONPrimitive[string]{
								value: "c",
								pos: &JSONValuePos{
									Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/map/additionalProperties"),
									Addr: ParseAddr("map.*"),
								},
							},
						},
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/map"),
							Addr: ParseAddr("map"),
						},
					},
					"number": JSONPrimitive[float64]{
						value: 1.5,
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/number"),
							Addr: ParseAddr("number"),
						},
					},
					"object": JSONObject{
						value: map[string]JSONValue{
							"p1": JSONPrimitive[string]{
								value: "d",
								pos: &JSONValuePos{
									Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/object/properties/p1"),
									Addr: ParseAddr("object.p1"),
								},
							},
							"obj": JSONObject{
								value: map[string]JSONValue{
									"pp1": JSONPrimitive[float64]{
										value: 2,
										pos: &JSONValuePos{
											Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/object/properties/obj/properties/pp1"),
											Addr: ParseAddr("object.obj.pp1"),
										},
									},
								},
								pos: &JSONValuePos{
									Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/object/properties/obj"),
									Addr: ParseAddr("object.obj"),
								},
							},
						},
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/object"),
							Addr: ParseAddr("object"),
						},
					},
					"string": JSONPrimitive[string]{
						value: "e",
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object/properties/string"),
							Addr: ParseAddr("string"),
						},
					},
				},
				pos: &JSONValuePos{
					Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/object"),
					Addr: ParseAddr(""),
				},
			},
		},
		{
			ref: specpathSyn + "#/definitions/base",
			input: `
{
  "type": "var1",
  "prop1": "foo"
}`,
			expect: JSONObject{
				value: map[string]JSONValue{
					"type": JSONPrimitive[string]{
						value: "var1",
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/base/properties/type"),
							Addr: ParseAddr("{var1}.type"),
						},
					},
					"prop1": JSONPrimitive[string]{
						value: "foo",
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/var1/properties/prop1"),
							Addr: ParseAddr("{var1}.prop1"),
						},
					},
				},
				pos: &JSONValuePos{
					Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/var1"),
					Addr: ParseAddr("{var1}"),
				},
			},
		},
		{
			ref: specpathSyn + "#/definitions/use_base",
			input: `
{
  "prop": {
	  "type": "var1",
	  "prop1": "foo"
	}
}`,
			expect: JSONObject{
				value: map[string]JSONValue{
					"prop": JSONObject{
						value: map[string]JSONValue{
							"type": JSONPrimitive[string]{
								value: "var1",
								pos: &JSONValuePos{
									Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/base/properties/type"),
									Addr: ParseAddr("prop.{var1}.type"),
								},
							},
							"prop1": JSONPrimitive[string]{
								value: "foo",
								pos: &JSONValuePos{
									Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/var1/properties/prop1"),
									Addr: ParseAddr("prop.{var1}.prop1"),
								},
							},
						},
						pos: &JSONValuePos{
							Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/var1"),
							Addr: ParseAddr("prop.{var1}"),
						},
					},
				},
				pos: &JSONValuePos{
					Ref:  jsonreference.MustCreateRef(specpathSyn + "#/definitions/use_base"),
					Addr: ParseAddr(""),
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.ref, func(t *testing.T) {
			ref := spec.MustCreateRef(tt.ref)
			exp, err := NewExpander(ref, nil)
			require.NoError(t, err)
			require.NoError(t, exp.Expand())
			v, err := UnmarshalJSONToJSONValue([]byte(tt.input), exp.root)
			require.NoError(t, err)
			require.Equal(t, tt.expect, v)
		})
	}
}

func TestUnmarshalJSONValuePos(t *testing.T) {
	var pos JSONValuePos
	input := []byte(`{
  "root_model": {
	"path_ref": "p1#/paths/path1",
	"operation": "get",
	"version": "2021-05-01"
  },
  "ref": "p1#/foo/bar",
  "addr": "a.b",
  "link_local": "p1/p2:1:2",
  "link_github": "https://github.com/blah"
}`)
	require.NoError(t, json.Unmarshal(input, &pos))
	fmt.Println(pos.RootModel)
	b, err := json.Marshal(pos)
	require.NoError(t, err)
	require.JSONEq(t, string(input), string(b))
}
