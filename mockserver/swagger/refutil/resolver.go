package refutil

import (
	"fmt"

	"github.com/go-openapi/spec"
)

// RResolveResponse recursively resolve a response (pointed by ref) by following its refs until it is a concrete response (no ref anymore), or until an already visited reference is hit.
// The 2nd to last return value is used to indicate whether the resolving ends normally (true) or due to hit a cyclic ref (false).
// Only when it is normally ended, the final response, its pointing reference and all the visited references are returned.
// Note that the visited references only include explicitly defined reference during the reference following, which doesn't include the input ref, unless explicitly mark the input is ref via the last param.
func RResolveResponse(ref spec.Ref, visitedRefs map[string]bool, inputIsRef bool) (*spec.Response, spec.Ref, map[string]bool, bool, error) {
	visited := map[string]bool{}
	for k, v := range visitedRefs {
		visited[k] = v
	}

	if !ref.HasFullFilePath {
		return nil, spec.Ref{}, nil, false, fmt.Errorf("Only normalized reference is allowed")
	}

	if _, ok := visited[ref.String()]; ok {
		return nil, spec.Ref{}, nil, false, nil
	}
	if inputIsRef {
		visited[ref.String()] = true
	}

	for {
		resp, err := spec.ResolveResponseWithBase(nil, ref, nil)
		if err != nil {
			return nil, spec.Ref{}, nil, false, fmt.Errorf("resolving %s: %v", ref.String(), err)
		}

		if resp.Ref.String() == "" {
			return resp, ref, visited, true, nil
		}

		ref, err = NormalizeFileRef(resp.Ref, ref.GetURL().Path)
		if err != nil {
			return nil, spec.Ref{}, nil, false, err
		}

		if _, ok := visited[ref.String()]; ok {
			return nil, spec.Ref{}, nil, false, nil
		}
		visited[ref.String()] = true
	}
}

// RResolve recursively resolve a schema (pointed by ref) by following its refs until it is a concrete schema (no ref anymore), or until an already visited reference is hit.
// The 2nd to last return value is used to indicate whether the resolving ends normally (true) or due to hit a cyclic ref (false).
// Only when it is normally ended, the final schema, its pointing reference and all the visited references are returned.
// Note that the visited references only include explicitly defined reference during the reference following, which doesn't include the input ref, unless explicitly mark the input is ref via the last param.
func RResolve(ref spec.Ref, visitedRefs map[string]bool, inputIsRef bool) (*spec.Schema, spec.Ref, map[string]bool, bool, error) {
	visited := map[string]bool{}
	for k, v := range visitedRefs {
		visited[k] = v
	}

	if !ref.HasFullFilePath {
		return nil, spec.Ref{}, nil, false, fmt.Errorf("Only normalized reference is allowed")
	}

	if _, ok := visited[ref.String()]; ok {
		return nil, spec.Ref{}, nil, false, nil
	}
	if inputIsRef {
		visited[ref.String()] = true
	}

	for {
		schema, err := spec.ResolveRefWithBase(nil, &ref, nil)
		if err != nil {
			return nil, spec.Ref{}, nil, false, fmt.Errorf("resolving %s: %v", ref.String(), err)
		}

		if schema.Ref.String() == "" {
			return schema, ref, visited, true, nil
		}

		ref, err = NormalizeFileRef(schema.Ref, ref.GetURL().Path)
		if err != nil {
			return nil, spec.Ref{}, nil, false, err
		}

		if _, ok := visited[ref.String()]; ok {
			return nil, spec.Ref{}, nil, false, nil
		}
		visited[ref.String()] = true
	}
}
