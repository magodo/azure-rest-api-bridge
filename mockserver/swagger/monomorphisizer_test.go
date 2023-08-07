package swagger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMonomorphization(t *testing.T) {
	cases := []struct {
		name   string
		input  Property
		expect []Property
	}{
		{
			name: "Basic",
			input: Property{
				addr: RootAddr,
				Children: map[string]*Property{
					"p1": {
						addr: ParseAddr("p1"),
					},
				},
			},
			expect: []Property{
				{
					addr: RootAddr,
					Children: map[string]*Property{
						"p1": {
							addr: ParseAddr("p1"),
						},
					},
				},
			},
		},
		{
			name: "Polymorphic",
			input: Property{
				addr: RootAddr,
				Variant: map[string]*Property{
					"V1": {
						addr: RootAddr,
					},
					"V2": {
						addr: RootAddr,
					},
					"V3": {
						addr: RootAddr,
					},
				},
			},
			expect: []Property{
				{
					addr: RootAddr,
					Variant: map[string]*Property{
						"V1": {
							addr: RootAddr,
						},
					},
				},
				{
					addr: RootAddr,
					Variant: map[string]*Property{
						"V2": {
							addr: RootAddr,
						},
					},
				},
				{
					addr: RootAddr,
					Variant: map[string]*Property{
						"V3": {
							addr: RootAddr,
						},
					},
				},
			},
		},
		{
			name: "Object key is polymorphic",
			input: Property{
				addr: RootAddr,
				Children: map[string]*Property{
					"p1": {
						addr: ParseAddr("p1"),
						Variant: map[string]*Property{
							"V1": {
								addr: ParseAddr("p1"),
							},
							"V2": {
								addr: ParseAddr("p1"),
							},
							"V3": {
								addr: ParseAddr("p1"),
							},
						},
					},
				},
			},
			expect: []Property{
				{
					addr: RootAddr,
					Children: map[string]*Property{
						"p1": {
							addr: ParseAddr("p1"),
							Variant: map[string]*Property{
								"V1": {
									addr: ParseAddr("p1"),
								},
							},
						},
					},
				},
				{
					addr: RootAddr,
					Children: map[string]*Property{
						"p1": {
							addr: ParseAddr("p1"),
							Variant: map[string]*Property{
								"V2": {
									addr: ParseAddr("p1"),
								},
							},
						},
					},
				},
				{
					addr: RootAddr,
					Children: map[string]*Property{
						"p1": {
							addr: ParseAddr("p1"),
							Variant: map[string]*Property{
								"V3": {
									addr: ParseAddr("p1"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "Element is polymorphic",
			input: Property{
				addr: RootAddr,
				Element: &Property{
					Variant: map[string]*Property{
						"V1": {
							addr: RootAddr,
						},
						"V2": {
							addr: RootAddr,
						},
						"V3": {
							addr: RootAddr,
						},
					},
				},
			},
			expect: []Property{
				{
					addr: RootAddr,
					Element: &Property{
						Variant: map[string]*Property{
							"V1": {
								addr: RootAddr,
							},
						},
					},
				},
				{
					addr: RootAddr,
					Element: &Property{
						Variant: map[string]*Property{
							"V2": {
								addr: RootAddr,
							},
						},
					},
				},
				{
					addr: RootAddr,
					Element: &Property{
						Variant: map[string]*Property{
							"V3": {
								addr: RootAddr,
							},
						},
					},
				},
			},
		},
		{
			name: "Mixed",
			input: Property{
				addr: RootAddr,
				Children: map[string]*Property{
					"p1": {
						addr: ParseAddr("p1"),
						Element: &Property{
							addr: ParseAddr("p1.*"),
							Variant: map[string]*Property{
								"V1": {
									addr: ParseAddr("p1.*"),
									Children: map[string]*Property{
										"pp1": {
											addr: ParseAddr("p1.*.pp1"),
										},
									},
								},
								"V2": {
									addr: ParseAddr("p1.*"),
									Children: map[string]*Property{
										"pp1": {
											addr: ParseAddr("p1.*.pp1"),
											Variant: map[string]*Property{
												"W1": {
													addr: ParseAddr("p1.*.pp1"),
												},
												"W2": {
													addr: ParseAddr("p1.*.pp1"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expect: []Property{
				{
					addr: RootAddr,
					Children: map[string]*Property{
						"p1": {
							addr: ParseAddr("p1"),
							Element: &Property{
								addr: ParseAddr("p1.*"),
								Variant: map[string]*Property{
									"V1": {
										addr: ParseAddr("p1.*"),
										Children: map[string]*Property{
											"pp1": {
												addr: ParseAddr("p1.*.pp1"),
											},
										},
									},
								},
							},
						},
					},
				},
				{
					addr: RootAddr,
					Children: map[string]*Property{
						"p1": {
							addr: ParseAddr("p1"),
							Element: &Property{
								addr: ParseAddr("p1.*"),
								Variant: map[string]*Property{
									"V2": {
										addr: ParseAddr("p1.*"),
										Children: map[string]*Property{
											"pp1": {
												addr: ParseAddr("p1.*.pp1"),
												Variant: map[string]*Property{
													"W1": {
														addr: ParseAddr("p1.*.pp1"),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					addr: RootAddr,
					Children: map[string]*Property{
						"p1": {
							addr: ParseAddr("p1"),
							Element: &Property{
								addr: ParseAddr("p1.*"),
								Variant: map[string]*Property{
									"V2": {
										addr: ParseAddr("p1.*"),
										Children: map[string]*Property{
											"pp1": {
												addr: ParseAddr("p1.*.pp1"),
												Variant: map[string]*Property{
													"W2": {
														addr: ParseAddr("p1.*.pp1"),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect, Monomorphization(&tt.input))
		})
	}
}
