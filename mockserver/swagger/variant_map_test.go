package swagger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVariantMapNew(t *testing.T) {
	pwd, _ := os.Getwd()
	spec := filepath.Join(pwd, "testdata", "variant_map.json")
	m, err := NewVariantMap(spec)
	require.NoError(t, err)
	require.Equal(t, VariantMap{
		"Base": map[string]string{
			"Var1": "Var1",
		},
		"Var1": map[string]string{
			"Var2": "Var2",
		},
		"Var2": map[string]string{},
	}, m)
}

func TestVariantMapGet(t *testing.T) {
	pwd, _ := os.Getwd()
	spec := filepath.Join(pwd, "testdata", "variant_map.json")
	m, err := NewVariantMap(spec)
	require.NoError(t, err)
	mBase, ok := m.Get("Base")
	require.Equal(t, true, ok)
	require.Equal(t, map[string]string{
		"Var1": "Var1",
		"Var2": "Var2",
	},
		mBase)
	mVar1, ok := m.Get("Var1")
	require.Equal(t, true, ok)
	require.Equal(t, map[string]string{
		"Var2": "Var2",
	},
		mVar1)
	mVar2, ok := m.Get("Var2")
	require.Equal(t, true, ok)
	require.Equal(t, map[string]string{},
		mVar2)
	mNoVar, ok := m.Get("NoVar")
	require.Equal(t, false, ok)
	require.Equal(t, map[string]string(nil),
		mNoVar)
}
