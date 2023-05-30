package refutil

import (
	"fmt"

	"github.com/go-openapi/spec"
)

// RResolveResponse recursively resolve a response's ref until it is a concrete schema (no ref anymore), or until an already visited reference is hit.
// All the visited references are recorded and returned. Meanwhile, the final response and its normalized pointing reference is returned together.
// If an already visited reference is hit, the response and the pointing reference is the one before the already visited reference, and the returned ok is false.
func RResolveResponse(specpath string, resp spec.Response, visitedRefs map[string]bool) (*spec.Response, *spec.Ref, map[string]bool, bool, error) {
	visited := map[string]bool{}
	for k, v := range visitedRefs {
		visited[k] = v
	}

	var respRef *spec.Ref
	for {
		ref := resp.Ref
		if ref.String() == "" {
			return &resp, respRef, visited, true, nil
		}

		ref, err := NormalizeFileRef(ref, specpath)
		if err != nil {
			return nil, nil, nil, false, err
		}

		// Return the response if already visited
		if _, ok := visited[ref.String()]; ok {
			return &resp, respRef, visited, false, nil
		}

		respRef = &ref
		specpath = ref.GetURL().Path

		visited[ref.String()] = true

		presp, err := spec.ResolveResponse(nil, *respRef)
		if err != nil {
			return nil, nil, nil, false, fmt.Errorf("resolving %s: %v", ref.String(), err)
		}
		resp = *presp
	}
}

// RResolve recursively resolve a schema's ref until it is a concrete schema (no ref anymore), or until an already visited reference is hit.
// All the visited references are recorded and returned. Meanwhile, the final schema and its normalized pointing reference is returned together.
// If an already visited reference is hit, the schema and the pointing reference is the one before the already visited reference, and the returned ok is false.
func RResolve(specpath string, schema spec.Schema, visitedRefs map[string]bool) (*spec.Schema, *spec.Ref, map[string]bool, bool, error) {
	visited := map[string]bool{}
	for k, v := range visitedRefs {
		visited[k] = v
	}

	var schemaRef *spec.Ref

	for {
		ref := schema.Ref
		if ref.String() == "" {
			return &schema, schemaRef, visited, true, nil
		}

		ref, err := NormalizeFileRef(ref, specpath)
		if err != nil {
			return nil, nil, nil, false, err
		}

		// Return the schema if already visited
		if _, ok := visited[ref.String()]; ok {
			return &schema, schemaRef, visited, false, nil
		}
		visited[ref.String()] = true

		schemaRef = &ref
		specpath = ref.GetURL().Path

		pschema, err := spec.ResolveRefWithBase(nil, schemaRef, nil)
		if err != nil {
			return nil, nil, nil, false, fmt.Errorf("resolving %s: %v", schemaRef.String(), err)
		}
		schema = *pschema
	}
}
