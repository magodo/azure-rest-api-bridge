package swagger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

func TestExpand(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)

	specpathA := filepath.Join(pwd, "testdata", "exp_a.json")
	specpathB := filepath.Join(pwd, "testdata", "exp_b.json")

	cases := []struct {
		name       string
		ref        string
		otherspecs []string
		opt        *ExpanderOption
		verify     func(*testing.T, *Property, ...*spec.Swagger)
	}{
		{
			name: "object",
			ref:  specpathA + "#/definitions/object",
			verify: func(t *testing.T, root *Property, swgs ...*spec.Swagger) {
				swg := swgs[0]
				expect := &Property{
					Schema: ptr(swg.Definitions["object"]),
					addr:   RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/object": true,
					},
					ref: spec.MustCreateRef(specpathA + "#/definitions/object"),
					Children: map[string]*Property{
						"number": {
							Schema: ptr(swg.Definitions["object"].Properties["number"]),
							addr:   ParseAddr("number"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/number"),
						},
						"integer": {
							Schema: ptr(swg.Definitions["object"].Properties["integer"]),
							addr:   ParseAddr("integer"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/integer"),
						},
						"string": {
							Schema: ptr(swg.Definitions["object"].Properties["string"]),
							addr:   ParseAddr("string"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/string"),
						},
						"boolean": {
							Schema: ptr(swg.Definitions["object"].Properties["boolean"]),
							addr:   ParseAddr("boolean"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/boolean"),
						},
						"object": {
							Schema: ptr(swg.Definitions["object"].Properties["object"]),
							addr:   ParseAddr("object"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/object"),
							Children: map[string]*Property{
								"p1": {
									Schema: ptr(swg.Definitions["object"].Properties["object"].Properties["p1"]),
									addr:   ParseAddr("object.p1"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/object": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/object/properties/p1"),
								},
								"obj": {
									Schema: ptr(swg.Definitions["object"].Properties["object"].Properties["obj"]),
									addr:   ParseAddr("object.obj"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/object": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/object/properties/obj"),
									Children: map[string]*Property{
										"pp1": {
											Schema: ptr(swg.Definitions["object"].Properties["object"].Properties["obj"].Properties["pp1"]),
											addr:   ParseAddr("object.obj.pp1"),
											visitedRefs: map[string]bool{
												specpathA + "#/definitions/object": true,
											},
											ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/object/properties/obj/properties/pp1"),
										},
									},
								},
							},
						},
						"emptyObject": {
							Schema: ptr(swg.Definitions["object"].Properties["emptyObject"]),
							addr:   ParseAddr("emptyObject"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref:      spec.MustCreateRef(specpathA + "#/definitions/object/properties/emptyObject"),
							Children: map[string]*Property{},
						},
						"array": {
							Schema: ptr(swg.Definitions["object"].Properties["array"]),
							addr:   ParseAddr("array"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/array"),
							Element: &Property{
								Schema: swg.Definitions["object"].Properties["array"].Items.Schema,
								addr:   ParseAddr("array.*"),
								visitedRefs: map[string]bool{
									specpathA + "#/definitions/object": true,
								},
								ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/array/items"),
							},
						},
						"map": {
							Schema: ptr(swg.Definitions["object"].Properties["map"]),
							addr:   ParseAddr("map"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/map"),
							Element: &Property{
								Schema: swg.Definitions["object"].Properties["map"].AdditionalProperties.Schema,
								addr:   ParseAddr("map.*"),
								visitedRefs: map[string]bool{
									specpathA + "#/definitions/object": true,
								},
								ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/map/additionalProperties"),
							},
						},
						"map2": {
							Schema: ptr(swg.Definitions["object"].Properties["map2"]),
							addr:   ParseAddr("map2"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/object": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/map2"),
							Element: &Property{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: spec.StringOrArray{"string"},
									},
								},
								addr: ParseAddr("map2.*"),
								visitedRefs: map[string]bool{
									specpathA + "#/definitions/object": true,
								},
								ref: spec.MustCreateRef(specpathA + "#/definitions/object/properties/map2/additionalProperties"),
							},
						},
					},
				}
				require.Equal(t, expect, root)
			},
		},
		{
			name: "Pet",
			ref:  specpathA + "#/definitions/Pet",
			verify: func(t *testing.T, root *Property, swgs ...*spec.Swagger) {
				swg := swgs[0]
				expect := &Property{
					Schema: ptr(swg.Definitions["Pet"]),
					addr:   RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/Pet": true,
					},
					ref: spec.MustCreateRef(specpathA + "#/definitions/Pet"),
					Variant: map[string]*Property{
						"Dog": {
							Schema:             ptr(swg.Definitions["Dog"]),
							Discriminator:      "type",
							DiscriminatorValue: "Dog",
							addr:               ParseAddr("{Dog}"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/Dog": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/Dog"),
							Children: map[string]*Property{
								"type": {
									Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
									addr:   ParseAddr("{Dog}.type"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Dog": true,
										specpathA + "#/definitions/Pet": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/type"),
								},
								"nickname": {
									Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
									addr:   ParseAddr("{Dog}.nickname"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Dog": true,
										specpathA + "#/definitions/Pet": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/nickname"),
								},
								"cat_friends": {
									Schema: ptr(swg.Definitions["Dog"].Properties["cat_friends"]),
									addr:   ParseAddr("{Dog}.cat_friends"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Dog": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/Dog/properties/cat_friends"),
									Element: &Property{
										Schema:             ptr(swg.Definitions["Cat"]),
										Discriminator:      "type",
										DiscriminatorValue: "Cat",
										addr:               ParseAddr("{Dog}.cat_friends.*"),
										visitedRefs: map[string]bool{
											specpathA + "#/definitions/Cat": true,
											specpathA + "#/definitions/Dog": true,
										},
										ref: spec.MustCreateRef(specpathA + "#/definitions/Cat"),
										Children: map[string]*Property{
											"type": {
												Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
												addr:   ParseAddr("{Dog}.cat_friends.*.type"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
												ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/type"),
											},
											"nickname": {
												Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
												addr:   ParseAddr("{Dog}.cat_friends.*.nickname"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
												ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/nickname"),
											},
											"dog_friends": {
												Schema: ptr(swg.Definitions["Cat"].Properties["dog_friends"]),
												addr:   ParseAddr("{Dog}.cat_friends.*.dog_friends"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Cat": true,
												},
												ref: spec.MustCreateRef(specpathA + "#/definitions/Cat/properties/dog_friends"),
											},
										},
									},
								},
							},
						},
						"Cat": {
							Schema:             ptr(swg.Definitions["Cat"]),
							Discriminator:      "type",
							DiscriminatorValue: "Cat",
							addr:               ParseAddr("{Cat}"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/Cat": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/Cat"),
							Children: map[string]*Property{
								"type": {
									Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
									addr:   ParseAddr("{Cat}.type"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Cat": true,
										specpathA + "#/definitions/Pet": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/type"),
								},
								"nickname": {
									Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
									addr:   ParseAddr("{Cat}.nickname"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Cat": true,
										specpathA + "#/definitions/Pet": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/nickname"),
								},
								"dog_friends": {
									Schema: ptr(swg.Definitions["Cat"].Properties["dog_friends"]),
									addr:   ParseAddr("{Cat}.dog_friends"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Cat": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/Cat/properties/dog_friends"),
									Element: &Property{
										Schema:             ptr(swg.Definitions["Dog"]),
										Discriminator:      "type",
										DiscriminatorValue: "Dog",
										addr:               ParseAddr("{Cat}.dog_friends.*"),
										visitedRefs: map[string]bool{
											specpathA + "#/definitions/Cat": true,
											specpathA + "#/definitions/Dog": true,
										},
										ref: spec.MustCreateRef(specpathA + "#/definitions/Dog"),
										Children: map[string]*Property{
											"type": {
												Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
												addr:   ParseAddr("{Cat}.dog_friends.*.type"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
												ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/type"),
											},
											"nickname": {
												Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
												addr:   ParseAddr("{Cat}.dog_friends.*.nickname"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
												ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/nickname"),
											},
											"cat_friends": {
												Schema: ptr(swg.Definitions["Dog"].Properties["cat_friends"]),
												addr:   ParseAddr("{Cat}.dog_friends.*.cat_friends"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
												},
												ref: spec.MustCreateRef(specpathA + "#/definitions/Dog/properties/cat_friends"),
											},
										},
									},
								},
							},
						},
					},
				}
				require.Equal(t, expect, root)
			},
		},
		{
			name: "Dog",
			ref:  specpathA + "#/definitions/Dog",
			verify: func(t *testing.T, root *Property, swgs ...*spec.Swagger) {
				swg := swgs[0]
				expect := &Property{
					Schema:             ptr(swg.Definitions["Dog"]),
					Discriminator:      "type",
					DiscriminatorValue: "Dog",
					addr:               RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/Dog": true,
					},
					ref: spec.MustCreateRef(specpathA + "#/definitions/Dog"),
					Children: map[string]*Property{
						"type": {
							Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
							addr:   ParseAddr("type"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/Dog": true,
								specpathA + "#/definitions/Pet": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/type"),
						},
						"nickname": {
							Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
							addr:   ParseAddr("nickname"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/Dog": true,
								specpathA + "#/definitions/Pet": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/nickname"),
						},
						"cat_friends": {
							Schema: ptr(swg.Definitions["Dog"].Properties["cat_friends"]),
							addr:   ParseAddr("cat_friends"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/Dog": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/Dog/properties/cat_friends"),
							Element: &Property{
								Schema:             ptr(swg.Definitions["Cat"]),
								Discriminator:      "type",
								DiscriminatorValue: "Cat",
								addr:               ParseAddr("cat_friends.*"),
								visitedRefs: map[string]bool{
									specpathA + "#/definitions/Cat": true,
									specpathA + "#/definitions/Dog": true,
								},
								ref: spec.MustCreateRef(specpathA + "#/definitions/Cat"),
								Children: map[string]*Property{
									"type": {
										Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
										addr:   ParseAddr("cat_friends.*.type"),
										visitedRefs: map[string]bool{
											specpathA + "#/definitions/Cat": true,
											specpathA + "#/definitions/Dog": true,
											specpathA + "#/definitions/Pet": true,
										},
										ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/type"),
									},
									"nickname": {
										Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
										addr:   ParseAddr("cat_friends.*.nickname"),
										visitedRefs: map[string]bool{
											specpathA + "#/definitions/Cat": true,
											specpathA + "#/definitions/Dog": true,
											specpathA + "#/definitions/Pet": true,
										},
										ref: spec.MustCreateRef(specpathA + "#/definitions/Pet/properties/nickname"),
									},
									"dog_friends": {
										Schema: ptr(swg.Definitions["Cat"].Properties["dog_friends"]),
										addr:   ParseAddr("cat_friends.*.dog_friends"),
										visitedRefs: map[string]bool{
											specpathA + "#/definitions/Dog": true,
											specpathA + "#/definitions/Cat": true,
										},
										ref: spec.MustCreateRef(specpathA + "#/definitions/Cat/properties/dog_friends"),
									},
								},
							},
						},
					},
				}
				require.Equal(t, expect, root)
			},
		},
		{
			name: "MsPet",
			ref:  specpathA + "#/definitions/MsPet",
			verify: func(t *testing.T, root *Property, swgs ...*spec.Swagger) {
				swg := swgs[0]
				expect := &Property{
					Schema: ptr(swg.Definitions["MsPet"]),
					addr:   RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/MsPet": true,
					},
					ref: spec.MustCreateRef(specpathA + "#/definitions/MsPet"),
					Variant: map[string]*Property{
						"CuteDog": {
							Schema:             ptr(swg.Definitions["MsDog"]),
							Discriminator:      "type",
							DiscriminatorValue: "CuteDog",
							addr:               ParseAddr("{CuteDog}"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/MsDog": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/MsDog"),
							Children: map[string]*Property{
								"type": {
									Schema: ptr(swg.Definitions["MsPet"].Properties["type"]),
									addr:   ParseAddr("{CuteDog}.type"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/MsPet": true,
										specpathA + "#/definitions/MsDog": true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/MsPet/properties/type"),
								},
							},
						},
					},
				}
				require.Equal(t, expect, root)
			},
		},
		{
			name: "ConflictBase",
			ref:  specpathA + "#/definitions/ConflictBase",
			verify: func(t *testing.T, root *Property, swgs ...*spec.Swagger) {
				swg := swgs[0]
				expect := &Property{
					Schema: ptr(swg.Definitions["ConflictBase"]),
					addr:   RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/ConflictBase": true,
					},
					ref: spec.MustCreateRef(specpathA + "#/definitions/ConflictBase"),
					Variant: map[string]*Property{
						"ConflictVar": {
							Schema:             ptr(swg.Definitions["RealConflictVar"]),
							Discriminator:      "type",
							DiscriminatorValue: "ConflictVar",
							addr:               ParseAddr("{ConflictVar}"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/RealConflictVar": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/RealConflictVar"),
							Children: map[string]*Property{
								"type": {
									Schema: ptr(swg.Definitions["ConflictBase"].Properties["type"]),
									addr:   ParseAddr("{ConflictVar}.type"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/RealConflictVar": true,
										specpathA + "#/definitions/ConflictBase":    true,
									},
									ref: spec.MustCreateRef(specpathA + "#/definitions/ConflictBase/properties/type"),
								},
							},
						},
					},
				}
				require.Equal(t, expect, root)
			},
		},
		{
			name:       "UseExtBase",
			ref:        specpathA + "#/definitions/UseExtBase",
			otherspecs: []string{specpathB},
			verify: func(t *testing.T, root *Property, swgs ...*spec.Swagger) {
				swgA, swgB := swgs[0], swgs[1]
				expect := &Property{
					Schema: ptr(swgA.Definitions["UseExtBase"]),
					addr:   RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/UseExtBase": true,
					},
					ref: spec.MustCreateRef(specpathA + "#/definitions/UseExtBase"),
					Children: map[string]*Property{
						"foo": {
							Schema: ptr(swgB.Definitions["BBase"]),
							addr:   ParseAddr("foo"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/UseExtBase": true,
								specpathB + "#/definitions/BBase":      true,
							},
							ref: spec.MustCreateRef(specpathB + "#/definitions/BBase"),
							Variant: map[string]*Property{
								"BVar": {
									Schema:             ptr(swgB.Definitions["BarVar"]),
									Discriminator:      "type",
									DiscriminatorValue: "BVar",
									addr:               ParseAddr("foo.{BVar}"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/UseExtBase": true,
										specpathB + "#/definitions/BarVar":     true,
									},
									ref: spec.MustCreateRef(specpathB + "#/definitions/BarVar"),
									Children: map[string]*Property{
										"type": {
											Schema: ptr(swgB.Definitions["BBase"].Properties["type"]),
											addr:   ParseAddr("foo.{BVar}.type"),
											visitedRefs: map[string]bool{
												specpathA + "#/definitions/UseExtBase": true,
												specpathB + "#/definitions/BBase":      true,
												specpathB + "#/definitions/BarVar":     true,
											},
											ref: spec.MustCreateRef(specpathB + "#/definitions/BBase/properties/type"),
										},
									},
								},
							},
						},
					},
				}
				require.Equal(t, expect, root)
			},
		},
		{
			name: "empty object regard as string",
			ref:  specpathA + "#/definitions/empty",
			opt:  &ExpanderOption{EmptyObjAsStr: true},
			verify: func(t *testing.T, root *Property, swgs ...*spec.Swagger) {
				swg := swgs[0]
				expect := &Property{
					Schema: ptr(swg.Definitions["empty"]),
					addr:   RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/empty": true,
					},
					ref: spec.MustCreateRef(specpathA + "#/definitions/empty"),
					Children: map[string]*Property{
						"emptyObject": {
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type: []string{"string"},
								},
							},
							addr: ParseAddr("emptyObject"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/empty": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/empty/properties/emptyObject"),
						},
						"emptyKey": {
							Schema: ptr(swg.Definitions["empty"].Properties["emptyKey"]),
							addr:   ParseAddr("emptyKey"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/empty": true,
							},
							ref: spec.MustCreateRef(specpathA + "#/definitions/empty/properties/emptyKey"),
							Element: &Property{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type: []string{"string"},
									},
								},
								addr: ParseAddr("emptyKey.*"),
								visitedRefs: map[string]bool{
									specpathA + "#/definitions/empty": true,
								},
								ref: spec.MustCreateRef(specpathA + "#/definitions/empty/properties/emptyKey/additionalProperties"),
							},
						},
					},
				}
				require.Equal(t, expect, root)
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ref := spec.MustCreateRef(tt.ref)
			exp, err := NewExpander(ref, tt.opt)
			require.NoError(t, err)
			require.NoError(t, exp.Expand())
			doc, err := loads.Spec(ref.GetURL().Path)
			require.NoError(t, err)

			specs := []*spec.Swagger{doc.Spec()}
			for _, spec := range tt.otherspecs {
				doc, err := loads.Spec(spec)
				require.NoError(t, err)
				specs = append(specs, doc.Spec())
			}
			tt.verify(t, exp.root, specs...)
		})
	}
}

func TestExpandWithCache(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	specpathA := filepath.Join(pwd, "testdata", "exp_a.json")
	ref := spec.MustCreateRef(specpathA + "#/definitions/object")

	// First run with no cache
	exp1, err := NewExpander(ref, nil)
	require.NoError(t, err)
	require.NoError(t, exp1.Expand())

	// Then run with cache
	cache := NewExpanderCache()
	exp2, err := NewExpander(ref, &ExpanderOption{Cache: cache})
	require.NoError(t, err)
	require.NoError(t, exp2.Expand())

	// Two runs should be the same
	require.Equal(t, exp1.root, exp2.root)

	// Run again with the same cache
	exp3, err := NewExpander(ref, &ExpanderOption{Cache: cache})
	require.NoError(t, err)
	require.NoError(t, exp3.Expand())

	// Two runs should be the same
	require.Equal(t, exp1.root, exp3.root)
}

func ptr[T any](input T) *T {
	return &input
}
