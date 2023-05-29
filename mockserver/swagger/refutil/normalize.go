package refutil

import (
	"fmt"
	"path/filepath"

	"github.com/go-openapi/spec"
)

func NormalizeFileRef(ref spec.Ref, base string) (spec.Ref, error) {
	if ref.String() == "" {
		return spec.Ref{}, fmt.Errorf("empty ref")
	}
	if ref.GetPointer().IsEmpty() {
		return spec.Ref{}, fmt.Errorf("empty ref pointer")
	}

	if ref.GetURL() == nil || ref.GetURL().Path == "" {
		absPath, err := filepath.Abs(base)
		if err != nil {
			return spec.Ref{}, err
		}
		return spec.NewRef(absPath + "#" + ref.GetPointer().String())
	}

	p := ref.GetURL().Path
	if filepath.IsAbs(p) {
		return ref, nil
	}

	absPath, err := filepath.Abs(filepath.Join(filepath.Dir(base), p))
	if err != nil {
		return spec.Ref{}, err
	}
	return spec.NewRef(absPath + "#" + ref.GetPointer().String())
}
