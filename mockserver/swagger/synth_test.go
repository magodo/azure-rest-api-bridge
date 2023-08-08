package swagger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

func TestCatesianProduct(t *testing.T) {
	cases := []struct {
		name   string
		params [][]interface{}
		expect [][]interface{}
	}{
		{
			name:   "nil",
			params: nil,
			expect: nil,
		},
		{
			name:   "empty",
			params: [][]interface{}{},
			expect: [][]interface{}{},
		},
		{
			name:   "[1]",
			params: [][]interface{}{{1}},
			expect: [][]interface{}{{1}},
		},
		{
			name:   "[1, 2]",
			params: [][]interface{}{{1, 2}},
			expect: [][]interface{}{{1}, {2}},
		},
		{
			name:   "[1, 2] x [3]",
			params: [][]interface{}{{1, 2}, {3}},
			expect: [][]interface{}{{1, 3}, {2, 3}},
		},
		{
			name:   "[1] x []",
			params: [][]interface{}{{1}, {}},
			expect: [][]interface{}{{1}},
		},
		{
			name:   "[1, 2] x [3, 4] x [5, 6]",
			params: [][]interface{}{{1, 2}, {3, 4}, {5, 6}},
			expect: [][]interface{}{
				{1, 3, 5},
				{2, 3, 5},
				{1, 4, 5},
				{2, 4, 5},
				{1, 3, 6},
				{2, 3, 6},
				{1, 4, 6},
				{2, 4, 6},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect, CatesianProduct(tt.params...))
		})
	}
}

func TestCatesianProductMap(t *testing.T) {
	cases := []struct {
		name   string
		params map[string][]interface{}
		expect []map[string]interface{}
	}{
		{
			name:   "nil",
			params: nil,
			expect: nil,
		},
		{
			name:   "empty",
			params: map[string][]interface{}{},
			expect: []map[string]interface{}{},
		},
		{
			name:   "{a: []}",
			params: map[string][]interface{}{"a": {}},
			expect: []map[string]interface{}{},
		},
		{
			name:   "{a: [1]}",
			params: map[string][]interface{}{"a": {1}},
			expect: []map[string]interface{}{{"a": 1}},
		},
		{
			name:   "{a: [1, 2]}",
			params: map[string][]interface{}{"a": {1, 2}},
			expect: []map[string]interface{}{{"a": 1}, {"a": 2}},
		},
		{
			name:   "{a: [1], b: []}",
			params: map[string][]interface{}{"a": {1}, "b": {}},
			expect: []map[string]interface{}{{"a": 1}},
		},
		{
			name:   "{a: [1, 2]} x {b: [3]}",
			params: map[string][]interface{}{"a": {1, 2}, "b": {3}},
			expect: []map[string]interface{}{{"a": 1, "b": 3}, {"a": 2, "b": 3}},
		},
		{
			name:   "{a: [1, 2]} x {b: [3]} x {c: [40, 50]}",
			params: map[string][]interface{}{"a": {1, 2}, "b": {3}, "c": {40, 50}},
			expect: []map[string]interface{}{
				{"a": 1, "b": 3, "c": 40},
				{"a": 2, "b": 3, "c": 40},
				{"a": 1, "b": 3, "c": 50},
				{"a": 2, "b": 3, "c": 50},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := CatesianProductMap(tt.params)
			require.Equal(t, tt.expect, result)
		})
	}
}

func TestSynthesize(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	specpathSyn := filepath.Join(pwd, "testdata", "syn.json")

	cases := []struct {
		name   string
		ref    string
		opt    *SynthesizerOption
		expect []string
	}{
		{
			name: specpathSyn + "#/definitions/object",
			ref:  specpathSyn + "#/definitions/object",
			expect: []string{
				`
		{
		  "array": [
		    "b"
		  ],
		  "boolean": true,
		  "emptyObject": {},
		  "integer": 1,
		  "map": {
		    "KEY": "c"
		  },
		  "map2": {
		    "KEY": "d"
		  },
		  "number": 1.5,
		  "object": {
		  	"p1": "e",
			"obj": {
				"pp1": 2
			}
		  },
		  "string": "f"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/object (duplicate)",
			ref:  specpathSyn + "#/definitions/object",
			opt: &SynthesizerOption{
				DuplicateElements: []SynthDuplicateElement{
					{
						Cnt:  1,
						Addr: ParseAddr("array"),
					},
					{
						Cnt:  2,
						Addr: ParseAddr("map"),
					},
				},
			},
			expect: []string{
				`
		{
		  "array": [
		    "b",
			"c"
		  ],
		  "boolean": true,
		  "emptyObject": {},
		  "integer": 1,
		  "map": {
		    "KEY": "d",
		    "KEY1": "e",
		    "KEY2": "f"
		  },
		  "map2": {
		    "KEY": "g"
		  },
		  "number": 1.5,
		  "object": {
		  	"p1": "h",
			"obj": {
				"pp1": 2
			}
		  },
		  "string": "i"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/base",
			ref:  specpathSyn + "#/definitions/base",
			expect: []string{
				`
		{
			"type": "var1",
			"prop1": "b"
		}
						`,
				`
		{
			"type": "var2",
			"prop2": "c"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/var1",
			ref:  specpathSyn + "#/definitions/var1",
			expect: []string{
				`
		{
			"type": "c",
			"prop1": "b"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/msbase",
			ref:  specpathSyn + "#/definitions/msbase",
			expect: []string{
				`
		{
			"type": "xvar1"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/msvar1",
			ref:  specpathSyn + "#/definitions/msvar1",
			expect: []string{
				`
		{
			"type": "b"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/enumobject",
			ref:  specpathSyn + "#/definitions/enumobject",
			opt:  &SynthesizerOption{UseEnumValues: true},
			expect: []string{
				`
		{
			"prop": "foo"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/use_base",
			ref:  specpathSyn + "#/definitions/use_base",
			expect: []string{
				`
		{
			"prop": {
				"type": "var1",
				"prop1": "b"
			}
		}
						`,
				`
		{
			"prop": {
				"type": "var2",
				"prop2": "c"
			}
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/array_base",
			ref:  specpathSyn + "#/definitions/array_base",
			expect: []string{
				`
		[
			{
				"prop": {
					"type": "var1",
					"prop1": "b"
				}
			}
		]
						`,
				`
		[
			{
				"prop": {
					"type": "var2",
					"prop2": "c"
				}
			}
		]
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/array_base (duplicate)",
			ref:  specpathSyn + "#/definitions/array_base",
			opt: &SynthesizerOption{
				DuplicateElements: []SynthDuplicateElement{
					{
						Cnt:  1,
						Addr: ParseAddr(""),
					},
				},
			},
			expect: []string{
				`
		[
			{
				"prop": {
					"type": "var1",
					"prop1": "b"
				}
			},
			{
				"prop": {
					"type": "var1",
					"prop1": "d"
				}
			}
		]
						`,
				`
		[
			{
				"prop": {
					"type": "var2",
					"prop2": "c"
				}
			},
			{
				"prop": {
					"type": "var2",
					"prop2": "e"
				}
			}
		]
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/conflictbase",
			ref:  specpathSyn + "#/definitions/conflictbase",
			expect: []string{
				`
		{
			"type": "conflictvar",
			"prop": "b"
		}
						`,
			},
		},
		{
			name: specpathSyn + "#/definitions/L1Base",
			ref:  specpathSyn + "#/definitions/L1Base",
			expect: []string{
				`
{
	"type": "L1Var1",
	"p11": {
		"type": "L2Var1",
		"p21": "b"
	}
}
				`,
				`
{
	"type": "L1Var1",
	"p11": {
		"type": "L2Var2",
		"p22": "c"
	}
}
				`,
				`
{
	"type": "L1Var2",
	"p12": "d"
}
				`,
			},
		},
		{
			name: specpathSyn + "#/definitions/XBase",
			ref:  specpathSyn + "#/definitions/XBase",
			expect: []string{
				`
{
  "type": "XVar1"
}
			`,
				`
{
  "type": "XVar2"
}
			`,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ref := spec.MustCreateRef(tt.ref)
			exp, err := NewExpander(ref, nil)
			require.NoError(t, err)
			require.NoError(t, exp.Expand())
			syn := NewSynthesizer(exp.Root(), ptr(NewRnd(nil)), tt.opt)
			results := syn.Synthesize()
			require.Len(t, results, len(tt.expect))
			for i, res := range results {
				b, err := json.Marshal(res)
				require.NoError(t, err)
				require.JSONEq(t, tt.expect[i], string(b))
			}
		})
	}
}
