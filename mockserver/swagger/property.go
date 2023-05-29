package swagger

import (
	"fmt"
	"net/http"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger/refutil"
)

type Expander struct {
	specpath string
	swagger  spec.Swagger
	root     Property
}

type PropertyName struct {
	Name    string
	Variant string
}

type Property struct {
	Schema *spec.Schema

	// Either child properties or array items
	Children map[PropertyName]Property

	// The resolved refs (normalized) along the way to this property, which is used to avoid cyclic reference.
	visitedRefs map[string]bool
}

func NewExpander(specpath string, ref *jsonreference.Ref) (*Expander, error) {
	doc, err := loads.Spec(specpath)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %v", specpath, err)
	}
	swg := doc.Spec()
	ptr := ref.GetPointer()
	opRaw, _, err := ptr.Get(swg)
	if err != nil {
		return nil, fmt.Errorf("referencing json pointer %s: %v", ptr, err)
	}
	op, ok := opRaw.(*spec.Operation)
	if !ok {
		return nil, fmt.Errorf("the json pointer %s points to a non-operation object %T", ptr, opRaw)
	}

	// Retrieve the 200 response schema
	if op.Responses == nil {
		return nil, fmt.Errorf("operation refed by %s has no responses defined", ptr)
	}
	// We only care about 200 for now, probably we should extend to support the others (e.g. when 200 is not defined).
	resp, ok := op.Responses.StatusCodeResponses[http.StatusOK]
	if !ok {
		return nil, fmt.Errorf("operation refed by %s has no 200 responses object defined", ptr)
	}
	// In case the response is a ref itself, follow it for one time.
	var visited map[string]bool
	if resp.Ref.String() != "" {
		var presp *spec.Response
		presp, visited, err = refutil.RResolveResponse(specpath, resp.Ref, nil)
		if err != nil {
			return nil, fmt.Errorf("recursively resolve response ref %s: %v", resp.Ref.String(), err)
		}
		resp = *presp
	}

	return &Expander{
		specpath: specpath,
		swagger:  *swg,
		root: Property{
			Schema:      resp.Schema,
			Children:    map[PropertyName]Property{},
			visitedRefs: visited,
		},
	}, nil
}
