package swagger

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAddr(t *testing.T) {
	cases := []struct {
		input  string
		expect PropertyAddr
		iserr  bool
	}{
		{
			input:  "",
			expect: PropertyAddr{},
		},
		{
			input: "*",
			expect: PropertyAddr{
				{
					Type: PropertyAddrStepTypeIndex,
				},
			},
		},
		{
			input: "a*",
			expect: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a*",
				},
			},
		},
		{
			input: "*a",
			expect: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "*a",
				},
			},
		},
		{
			input: "{Foo}",
			expect: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Variant: "Foo",
				},
			},
		},
		{
			input: `{{}`,
			expect: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Variant: "{",
				},
			},
		},
		{
			input: "a",
			expect: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
			},
		},
		{
			input: "a.b",
			expect: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
		},
		{
			input: "a\\.b",
			expect: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a.b",
				},
			},
		},
		{
			input: "a.*.b",
			expect: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
				{
					Type: PropertyAddrStepTypeIndex,
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
		},
		{
			input: "a.*{Foo}.b",
			expect: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
				{
					Type:    PropertyAddrStepTypeIndex,
					Variant: "Foo",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
		},
		{
			input: "a{Foo}.b",
			expect: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Value:   "a",
					Variant: "Foo",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
		},
		{
			input: "a{Foo.Bar}.b",
			expect: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Value:   "a",
					Variant: "Foo.Bar",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
		},
		{
			input: `a{Foo.{Bar\}}.b`,
			expect: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Value:   "a",
					Variant: "Foo.{Bar}",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
		},
		{
			input: `a\a`,
			iserr: true,
		},
		{
			input: `.`,
			iserr: true,
		},
		{
			input: `{}`,
			iserr: true,
		},
		{
			input: `a{}`,
			iserr: true,
		},
		{
			input: `a{Foo}a`,
			iserr: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			addr, err := ParseAddr(tt.input)
			if tt.iserr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expect, *addr)
		})
	}
}

func TestPropertyAddrString(t *testing.T) {
	cases := []struct {
		input  PropertyAddr
		expect string
	}{
		{
			input:  PropertyAddr{},
			expect: "",
		},
		{
			input: PropertyAddr{
				{
					Type: PropertyAddrStepTypeIndex,
				},
			},
			expect: "*",
		},
		{
			input: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a*",
				},
			},
			expect: "a*",
		},
		{
			input: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "*a",
				},
			},
			expect: "*a",
		},
		{
			input: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Variant: "Foo",
				},
			},
			expect: "{Foo}",
		},
		{
			input: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
			},
			expect: "a",
		},
		{
			input: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
			expect: "a.b",
		},
		{
			input: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a.b",
				},
			},
			expect: "a\\.b",
		},
		{
			input: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
				{
					Type: PropertyAddrStepTypeIndex,
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
			expect: "a.*.b",
		},
		{
			input: PropertyAddr{
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "a",
				},
				{
					Type:    PropertyAddrStepTypeIndex,
					Variant: "Foo",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
			expect: "a.*{Foo}.b",
		},
		{
			input: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Value:   "a",
					Variant: "Foo",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
			expect: "a{Foo}.b",
		},
		{
			input: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Value:   "a",
					Variant: "Foo.Bar",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
			expect: "a{Foo.Bar}.b",
		},
		{
			input: PropertyAddr{
				{
					Type:    PropertyAddrStepTypeProp,
					Value:   "a",
					Variant: "Foo.{Bar}",
				},
				{
					Type:  PropertyAddrStepTypeProp,
					Value: "b",
				},
			},
			expect: `a{Foo.{Bar\}}.b`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.expect, func(t *testing.T) {
			require.Equal(t, tt.expect, tt.input.String())
		})
	}
}
