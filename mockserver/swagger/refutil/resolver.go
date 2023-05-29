package refutil

import (
	"fmt"

	"github.com/go-openapi/spec"
)

// Recurisvely resolve response ref
func RResolveResponse(specpath string, ref spec.Ref, visitedRefs map[string]bool) (*spec.Response, map[string]bool, error) {
	visited := map[string]bool{}
	for k, v := range visitedRefs {
		visited[k] = v
	}

	var (
		err  error
		resp *spec.Response
	)
	for {
		ref, err = NormalizeFileRef(ref, specpath)
		if err != nil {
			return nil, nil, err
		}

		// Return the response if already visited
		if _, ok := visited[ref.String()]; ok {
			return resp, visited, nil
		}

		visited[ref.String()] = true

		resp, err = spec.ResolveResponse(nil, ref)
		if err != nil {
			return nil, nil, fmt.Errorf("resolving %s: %v", ref.String(), err)
		}

		specpath = ref.GetURL().Path
		ref = resp.Ref

		// Not a reference any more, just return
		if resp.Ref.String() == "" {
			return resp, visited, nil
		}
	}
}

func RResolve(specpath string, ref spec.Ref, visitedRefs map[string]bool) (*spec.Schema, map[string]bool, error) {
	visited := map[string]bool{}
	for k, v := range visitedRefs {
		visited[k] = v
	}

	var (
		err    error
		schema *spec.Schema
	)
	for {
		ref, err = NormalizeFileRef(ref, specpath)
		if err != nil {
			return nil, nil, err
		}

		// Return the response if already visited
		if _, ok := visited[ref.String()]; ok {
			return schema, visited, nil
		}

		visited[ref.String()] = true

		schema, err = spec.ResolveRefWithBase(nil, &ref, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("resolving %s: %v", ref.String(), err)
		}

		specpath = ref.GetURL().Path
		ref = schema.Ref

		// Not a reference any more, just return
		if schema.Ref.String() == "" {
			return schema, visited, nil
		}
	}
}
