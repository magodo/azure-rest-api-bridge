package refutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/require"
)

func TestRResolve(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)

	specpathA := filepath.Join(pwd, "testdata", "a.json")
	specpathB := filepath.Join(pwd, "testdata", "b", "b.json")

	cases := []struct {
		name       string
		ref        string
		visited    map[string]bool
		outDesc    string
		outVisited map[string]bool
	}{
		{
			name:    "#/definitions/ConcreteModel",
			ref:     "#/definitions/ConcreteModel",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
		},
		{
			name: "#/definitions/ConcreteModel (visited)",
			ref:  "#/definitions/ConcreteModel",
			visited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outDesc: "",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
		},
		{
			name:    "#/definitions/Model1",
			ref:     "#/definitions/Model1",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/Model1":        true,
				specpathA + "#/definitions/ConcreteModel": true,
			},
		},
		{
			name: "#/definitions/Model1 (visited)",
			ref:  "#/definitions/Model1",
			visited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outDesc: "Model1",
			outVisited: map[string]bool{
				specpathA + "#/definitions/Model1":        true,
				specpathA + "#/definitions/ConcreteModel": true,
			},
		},
		{
			name:    "#/definitions/Model2",
			ref:     "#/definitions/Model2",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/Model1":        true,
				specpathA + "#/definitions/Model2":        true,
				specpathA + "#/definitions/ConcreteModel": true,
			},
		},
		{
			name:    "#/definitions/Circle1",
			ref:     "#/definitions/Circle1",
			visited: nil,
			outDesc: "Circle2",
			outVisited: map[string]bool{
				specpathA + "#/definitions/Circle1": true,
				specpathA + "#/definitions/Circle2": true,
			},
		},
		{
			name:    "#/definitions/Circle2",
			ref:     "#/definitions/Circle2",
			visited: nil,
			outDesc: "Circle1",
			outVisited: map[string]bool{
				specpathA + "#/definitions/Circle1": true,
				specpathA + "#/definitions/Circle2": true,
			},
		},
		{
			name:    "#/definitions/FromB",
			ref:     "#/definitions/FromB",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
				specpathA + "#/definitions/FromB":         true,
				specpathB + "#/definitions/FromA":         true,
			},
		},
		{
			name:    specpathA + "#/definitions/FromB",
			ref:     specpathA + "#/definitions/FromB",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
				specpathA + "#/definitions/FromB":         true,
				specpathB + "#/definitions/FromA":         true,
			},
		},
		{
			name:    "b/b.json#/definitions/FromA",
			ref:     "b/b.json#/definitions/FromA",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
				specpathB + "#/definitions/FromA":         true,
			},
		},
		{
			name:    specpathB + "#/definitions/FromA",
			ref:     specpathB + "#/definitions/FromA",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
				specpathB + "#/definitions/FromA":         true,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			schema, visited, err := RResolve(specpathA, spec.MustCreateRef(tt.ref), tt.visited)
			require.NoError(t, err)
			require.Equal(t, tt.outVisited, visited)
			if schema == nil {
				require.Equal(t, tt.outDesc, "")
			} else {
				require.Equal(t, tt.outDesc, schema.Description)
			}
		})
	}
}

func TestRResolveResponse(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)

	specpathA := filepath.Join(pwd, "testdata", "a.json")
	specpathB := filepath.Join(pwd, "testdata", "b", "b.json")

	cases := []struct {
		name       string
		ref        string
		visited    map[string]bool
		outDesc    string
		outVisited map[string]bool
	}{
		{
			name:    "#/paths/p1/get/responses/200",
			ref:     "#/paths/p1/get/responses/200",
			visited: nil,
			outDesc: "Concrete",
			outVisited: map[string]bool{
				specpathA + "#/paths/p1/get/responses/200": true,
				specpathA + "#/responses/Concrete":         true,
				specpathA + "#/responses/FromB":            true,
				specpathB + "#/responses/FromA":            true,
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			schema, visited, err := RResolveResponse(specpathA, spec.MustCreateRef(tt.ref), tt.visited)
			require.NoError(t, err)
			require.Equal(t, tt.outVisited, visited)
			if schema == nil {
				require.Equal(t, tt.outDesc, "")
			} else {
				require.Equal(t, tt.outDesc, schema.Description)
			}
		})
	}
}
