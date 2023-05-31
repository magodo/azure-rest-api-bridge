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
		inputIsRef bool
		outDesc    string
		outVisited map[string]bool
		outOwnRef  string
		outOK      bool
		err        bool
	}{
		{
			name:       "#/definitions/ConcreteModel",
			ref:        specpathA + "#/definitions/ConcreteModel",
			visited:    nil,
			outDesc:    "ConcreteModel",
			outVisited: map[string]bool{},
			outOwnRef:  specpathA + "#/definitions/ConcreteModel",
			outOK:      true,
		},
		{
			name:       "#/definitions/ConcreteModel (input is ref)",
			ref:        specpathA + "#/definitions/ConcreteModel",
			visited:    nil,
			inputIsRef: true,
			outDesc:    "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outOwnRef: specpathA + "#/definitions/ConcreteModel",
			outOK:     true,
		},
		{
			name: "#/definitions/ConcreteModel (visited)",
			ref:  specpathA + "#/definitions/ConcreteModel",
			visited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outOK: false,
		},
		{
			name:    "#/definitions/Model1",
			ref:     specpathA + "#/definitions/Model1",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outOwnRef: specpathA + "#/definitions/ConcreteModel",
			outOK:     true,
		},
		{
			name: "#/definitions/Model1 (visited)",
			ref:  specpathA + "#/definitions/Model1",
			visited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outOK: false,
		},
		{
			name:    "#/definitions/Model2",
			ref:     specpathA + "#/definitions/Model2",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/Model1":        true,
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outOwnRef: specpathA + "#/definitions/ConcreteModel",
			outOK:     true,
		},
		{
			name:    "#/definitions/Circle1",
			ref:     specpathA + "#/definitions/Circle1",
			visited: nil,
			outOK:   false,
		},
		{
			name:    "#/definitions/Circle2",
			ref:     specpathA + "#/definitions/Circle2",
			visited: nil,
			outOK:   false,
		},
		{
			name:    "#/definitions/FromB",
			ref:     specpathA + "#/definitions/FromB",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
				specpathB + "#/definitions/FromA":         true,
			},
			outOwnRef: specpathA + "#/definitions/ConcreteModel",
			outOK:     true,
		},
		{
			name:    specpathB + "#/definitions/FromA",
			ref:     specpathB + "#/definitions/FromA",
			visited: nil,
			outDesc: "ConcreteModel",
			outVisited: map[string]bool{
				specpathA + "#/definitions/ConcreteModel": true,
			},
			outOwnRef: specpathA + "#/definitions/ConcreteModel",
			outOK:     true,
		},
		{
			name: "b/b.json#/definitions/FromA",
			ref:  "b/b.json#/definitions/FromA",
			err:  true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			schema, ownRef, visited, ok, err := RResolve(spec.MustCreateRef(tt.ref), tt.visited, tt.inputIsRef)
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.outVisited, visited)
			require.Equal(t, tt.outOK, ok)
			if schema == nil {
				require.Equal(t, tt.outDesc, "")
			} else {
				require.Equal(t, tt.outDesc, schema.Description)
			}
			require.Equal(t, tt.outOwnRef, ownRef.String())
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
		outOwnRef  string
		outOK      bool
	}{
		{
			name:    "#/paths/p1/get/responses/200",
			ref:     specpathA + "#/paths/p1/get/responses/200",
			visited: nil,
			outDesc: "Concrete",
			outVisited: map[string]bool{
				specpathA + "#/responses/Concrete": true,
				specpathA + "#/responses/FromB":    true,
				specpathB + "#/responses/FromA":    true,
			},
			outOwnRef: specpathA + "#/responses/Concrete",
			outOK:     true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			schema, ownRef, visited, ok, err := RResolveResponse(spec.MustCreateRef(tt.ref), tt.visited, false)
			require.NoError(t, err)
			require.Equal(t, tt.outVisited, visited)
			require.Equal(t, tt.outOK, ok)
			if schema == nil {
				require.Equal(t, tt.outDesc, "")
			} else {
				require.Equal(t, tt.outDesc, schema.Description)
			}
			require.Equal(t, tt.outOwnRef, ownRef.String())
		})
	}
}
