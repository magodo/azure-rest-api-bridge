package swagger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

func TestExpand(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)

	specpathA := filepath.Join(pwd, "testdata", "a.json")

	cases := []struct {
		name     string
		ref      string
		specpath string
		verify   func(*testing.T, *Property, *spec.Swagger)
	}{
		{
			name:     specpathA,
			ref:      "#/paths/~1p1/get",
			specpath: specpathA,
			verify: func(t *testing.T, root *Property, swg *spec.Swagger) {
				expect := &Property{
					Schema: ptr(swg.Definitions["Pet"]),
					addr:   RootAddr,
					visitedRefs: map[string]bool{
						specpathA + "#/definitions/Pet": true,
					},
					ownRef: ptr(spec.MustCreateRef(specpathA + "#/definitions/Pet")),
					Variant: map[string]*Property{
						"Dog": {
							Schema: ptr(swg.Definitions["Dog"]),
							addr:   ParseAddr("{Dog}"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/Dog": true,
							},
							ownRef: ptr(spec.MustCreateRef(specpathA + "#/definitions/Dog")),
							Children: map[string]*Property{
								"type": {
									Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
									addr:   ParseAddr("{Dog}.type"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Dog": true,
										specpathA + "#/definitions/Pet": true,
									},
								},
								"nickname": {
									Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
									addr:   ParseAddr("{Dog}.nickname"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Dog": true,
										specpathA + "#/definitions/Pet": true,
									},
								},
								"cat_friends": {
									Schema: ptr(swg.Definitions["Dog"].Properties["cat_friends"]),
									addr:   ParseAddr("{Dog}.cat_friends"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Dog": true,
									},
									Element: &Property{
										Schema: ptr(swg.Definitions["Cat"]),
										addr:   ParseAddr("{Dog}.cat_friends.*"),
										visitedRefs: map[string]bool{
											specpathA + "#/definitions/Cat": true,
											specpathA + "#/definitions/Dog": true,
										},
										ownRef: ptr(spec.MustCreateRef(specpathA + "#/definitions/Cat")),
										Children: map[string]*Property{
											"type": {
												Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
												addr:   ParseAddr("{Dog}.cat_friends.*.type"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
											},
											"nickname": {
												Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
												addr:   ParseAddr("{Dog}.cat_friends.*.nickname"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
											},
											"dog_friends": {
												Schema: ptr(swg.Definitions["Cat"].Properties["dog_friends"]),
												addr:   ParseAddr("{Dog}.cat_friends.*.dog_friends"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Cat": true,
												},
											},
										},
									},
								},
							},
						},
						"Cat": {
							Schema: ptr(swg.Definitions["Cat"]),
							addr:   ParseAddr("{Cat}"),
							visitedRefs: map[string]bool{
								specpathA + "#/definitions/Cat": true,
							},
							ownRef: ptr(spec.MustCreateRef(specpathA + "#/definitions/Cat")),
							Children: map[string]*Property{
								"type": {
									Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
									addr:   ParseAddr("{Cat}.type"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Cat": true,
										specpathA + "#/definitions/Pet": true,
									},
								},
								"nickname": {
									Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
									addr:   ParseAddr("{Cat}.nickname"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Cat": true,
										specpathA + "#/definitions/Pet": true,
									},
								},
								"dog_friends": {
									Schema: ptr(swg.Definitions["Cat"].Properties["dog_friends"]),
									addr:   ParseAddr("{Cat}.dog_friends"),
									visitedRefs: map[string]bool{
										specpathA + "#/definitions/Cat": true,
									},
									Element: &Property{
										Schema: ptr(swg.Definitions["Dog"]),
										addr:   ParseAddr("{Cat}.dog_friends.*"),
										visitedRefs: map[string]bool{
											specpathA + "#/definitions/Cat": true,
											specpathA + "#/definitions/Dog": true,
										},
										ownRef: ptr(spec.MustCreateRef(specpathA + "#/definitions/Dog")),
										Children: map[string]*Property{
											"type": {
												Schema: ptr(swg.Definitions["Pet"].Properties["type"]),
												addr:   ParseAddr("{Cat}.dog_friends.*.type"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
											},
											"nickname": {
												Schema: ptr(swg.Definitions["Pet"].Properties["nickname"]),
												addr:   ParseAddr("{Cat}.dog_friends.*.nickname"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
													specpathA + "#/definitions/Pet": true,
												},
											},
											"cat_friends": {
												Schema: ptr(swg.Definitions["Dog"].Properties["cat_friends"]),
												addr:   ParseAddr("{Cat}.dog_friends.*.cat_friends"),
												visitedRefs: map[string]bool{
													specpathA + "#/definitions/Cat": true,
													specpathA + "#/definitions/Dog": true,
												},
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
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ref := jsonreference.MustCreateRef(tt.ref)
			exp, err := NewExpander(tt.specpath, &ref)
			require.NoError(t, err)
			require.NoError(t, exp.Expand())
			tt.verify(t, exp.root, &exp.swagger)
		})
	}
}

func ptr[T any](input T) *T {
	return &input
}
