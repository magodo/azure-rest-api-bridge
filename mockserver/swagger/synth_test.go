package swagger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

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
						Addr: MustParseAddr("array"),
					},
					{
						Cnt:  2,
						Addr: MustParseAddr("map"),
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
			"prop2": "b"
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
				"prop2": "b"
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
					"prop2": "b"
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
						Addr: MustParseAddr(""),
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
					"prop1": "c"
				}
			}
		]
						`,
				`
		[
			{
				"prop": {
					"type": "var2",
					"prop2": "b"
				}
			},
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
		"p22": "b"
	}
}
				`,
				`
{
	"type": "L1Var2",
	"p12": "b"
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
			propInstances := Monomorphization(exp.Root())
			require.Len(t, propInstances, len(tt.expect))
			for i, v := range propInstances {
				propInstance := v
				syn, err := NewSynthesizer(&propInstance, ptr(NewRnd(nil)), tt.opt)
				require.NoError(t, err)
				res, ok := syn.Synthesize()
				require.True(t, ok)
				b, err := json.Marshal(res)
				require.NoError(t, err)
				require.JSONEq(t, tt.expect[i], string(b))
			}
		})
	}
}
