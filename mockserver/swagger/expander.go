package swagger

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger/refutil"
)

type Expander struct {
	specpath string
	swagger  spec.Swagger
	root     *Property
}

type PropertyName struct {
	Name    string
	Variant string
}

type PropertyAddr []PropertyAddrStep

type PropertyAddrStep struct {
	Type  PropertyAddrStepType
	Value string
}

var RootAddr = PropertyAddr{}

type PropertyAddrStepType int

const (
	PropertyAddrStepTypeProp PropertyAddrStepType = iota
	PropertyAddrStepTypeIndex
	PropertyAddrStepTypeVariant
)

type Property struct {
	Schema *spec.Schema
	addr   PropertyAddr

	// The resolved refs (normalized) along the way to this property, which is used to avoid cyclic reference.
	visitedRefs map[string]bool

	// The ref (normalized) that owns the schema of this property, if any.
	// E.g. prop1's schema is "schema1", which refs "schema2", which refs "schema3".
	// So prop1's ownRef is (normalized) "schema3"
	// E.g. prop1's schema is inlined, then the ownRef is nil
	ownRef *spec.Ref

	// Children represents the child properties of an object
	// At most one of Children, Element and Variant is non nil
	Children map[string]*Property

	// Element represents the element property of an array or a map (additionalProperties of an object)
	// At most one of Children, Element and Variant is non nil
	Element *Property

	// Variant represents the current property is a polymorphic schema, which is then expanded to multiple variant schemas
	// At most one of Children, Element and Variant is non nil
	Variant map[string]*Property
}

// NewExpander create a expander for the successful response schema of an operation referenced by the input json reference.
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

	// In case the response is a ref itself, follow it
	presp, respRef, visited, ok, err := refutil.RResolveResponse(specpath, resp, nil)
	if err != nil {
		return nil, fmt.Errorf("recursively resolve response ref %s: %v", resp.Ref.String(), err)
	}
	if !ok {
		return nil, fmt.Errorf("circular ref found when resolving response ref %s", resp.Ref.String())
	}

	sspecpath := specpath
	if respRef != nil {
		sspecpath = respRef.GetURL().Path
	}

	psch, ownRef, visited, ok, err := refutil.RResolve(sspecpath, *presp.Schema, visited)
	if err != nil {
		return nil, fmt.Errorf("recursively resolve response schema: %v", err)
	}
	if !ok {
		return nil, fmt.Errorf("circular ref found when resolving response schema")
	}

	return &Expander{
		specpath: specpath,
		swagger:  *swg,
		root: &Property{
			Schema:      psch,
			ownRef:      ownRef,
			addr:        RootAddr,
			visitedRefs: visited,
		},
	}, nil
}

func (e *Expander) Expand() error {
	wl := []*Property{e.root}
	for {
		if len(wl) == 0 {
			break
		}
		nwl := []*Property{}
		for _, prop := range wl {
			if err := e.expandPropStep(prop); err != nil {
				return err
			}
			if prop.Element != nil {
				nwl = append(nwl, prop.Element)
			}
			for _, v := range prop.Children {
				nwl = append(nwl, v)
			}
			for _, v := range prop.Variant {
				nwl = append(nwl, v)
			}
		}
		wl = nwl
	}
	return nil
}

func (e *Expander) expandPropStep(prop *Property) error {
	if len(prop.Schema.Type) > 1 {
		return fmt.Errorf("%s: type of property type is an array (not supported yet)", prop.addr)
	}
	schema := prop.Schema
	t := "object"
	if len(schema.Type) == 1 {
		t = schema.Type[0]
	}
	switch t {
	case "array":
		return e.expandPropStepAsArray(prop)
	case "object":
		if schema.Discriminator == "" {
			if schemaIsMap(schema) {
				return e.expandPropAsMap(prop)
			}
			return e.expandPropAsRegularObject(prop)
		}
		return e.expandPropAsPolymorphicObject(prop)
	}
	return nil
}

func (e *Expander) expandPropStepAsArray(prop *Property) error {
	schema := prop.Schema
	if !schemaIsArray(schema) {
		return fmt.Errorf("%s: is not array", prop.addr)
	}
	addr := append(prop.addr, PropertyAddrStep{
		Type: PropertyAddrStepTypeIndex,
	})
	if schema.Items.Schema == nil {
		return fmt.Errorf("%s: items of property is not a single schema (not supported yet)", addr)
	}
	schema, ownRef, visited, ok, err := refutil.RResolve(e.specpath, *schema.Items.Schema, prop.visitedRefs)
	if err != nil {
		return fmt.Errorf("%s: recursively resolving items: %v", addr, err)
	}
	if !ok {
		return nil
	}
	prop.Element = &Property{
		Schema:      schema,
		ownRef:      ownRef,
		addr:        addr,
		visitedRefs: visited,
	}
	return nil
}

func (e *Expander) expandPropAsMap(prop *Property) error {
	schema := prop.Schema
	if !schemaIsMap(schema) {
		return fmt.Errorf("%s: is not map", prop.addr)
	}
	addr := append(PropertyAddr{}, prop.addr...)
	addr = append(addr, PropertyAddrStep{
		Type: PropertyAddrStepTypeIndex,
	})
	if schema.AdditionalProperties.Schema == nil {
		return fmt.Errorf("%s: additionalProperties is not a single schema (not supported yet)", addr)
	}
	schema, ownRef, visited, ok, err := refutil.RResolve(e.specpath, *schema.AdditionalProperties.Schema, prop.visitedRefs)
	if err != nil {
		return fmt.Errorf("%s: recursively resolving additionalProperties: %v", addr, err)
	}
	if !ok {
		return nil
	}
	prop.Element = &Property{
		Schema:      schema,
		ownRef:      ownRef,
		addr:        addr,
		visitedRefs: visited,
	}
	return nil
}

func (e *Expander) expandPropAsRegularObject(prop *Property) error {
	schema := prop.Schema

	if !schemaIsObject(schema) {
		return fmt.Errorf("%s: is not object", prop.addr)
	}

	prop.Children = map[string]*Property{}

	// Expanding the regular properties
	for k, sch := range schema.Properties {
		addr := append(PropertyAddr{}, prop.addr...)
		addr = append(addr, PropertyAddrStep{
			Type:  PropertyAddrStepTypeProp,
			Value: k,
		})
		schema, ownRef, visited, ok, err := refutil.RResolve(e.specpath, sch, prop.visitedRefs)
		if err != nil {
			return fmt.Errorf("%s: recursively resolving property %s: %v", addr, k, err)
		}
		if !ok {
			continue
		}
		prop.Children[k] = &Property{
			Schema:      schema,
			ownRef:      ownRef,
			addr:        addr,
			visitedRefs: visited,
		}
	}

	// Inheriting the allOf schemas
	for i, sch := range schema.AllOf {
		schema, ownRef, visited, ok, err := refutil.RResolve(e.specpath, sch, prop.visitedRefs)
		if err != nil {
			return fmt.Errorf("%s: recursively resolving %d-th allOf schema: %v", prop.addr, i, err)
		}
		if !ok {
			continue
		}
		tmpExp := Expander{
			specpath: e.specpath,
			swagger:  e.swagger,
			root: &Property{
				Schema:      schema,
				ownRef:      ownRef,
				addr:        prop.addr,
				visitedRefs: visited,
			},
		}
		// The base schema of a variant schema is always a regular object.
		if err := tmpExp.expandPropAsRegularObject(tmpExp.root); err != nil {
			return fmt.Errorf("%s: expanding the %d-th (temporary) allOf schema: %v", prop.addr, i, err)
		}
		for k, v := range tmpExp.root.Children {
			prop.Children[k] = v
		}
	}

	return nil
}

func (e *Expander) expandPropAsPolymorphicObject(prop *Property) error {
	schema := prop.Schema

	if !schemaIsObject(schema) {
		return fmt.Errorf("%s: is not object", prop.addr)
	}

	prop.Variant = map[string]*Property{}
	dsch, _, _, _, err := refutil.RResolve(e.specpath, schema.Properties[schema.Discriminator], prop.visitedRefs)
	if err != nil {
		return fmt.Errorf("%s: recursively resolving discriminator property's(%s) schema: %v", prop.addr, schema.Discriminator, err)
	}
	for _, name := range dsch.Enum {
		name := name.(string)
		addr := append(PropertyAddr{}, prop.addr...)
		addr = append(addr, PropertyAddrStep{
			Type:  PropertyAddrStepTypeVariant,
			Value: name,
		})
		sch, ok := e.swagger.Definitions[name]
		if !ok {
			return fmt.Errorf("%s: no variant schema named %s found", addr, name)
		}

		// Though we are not explicitly following any reference of the variant, while we are effectively doing so. Hence construct the reference to the variant schema and mark it as visited.
		// Meanwhile, we'll remove the owning ref of the base schema from visited set in order to allow the later allOf inheritance.
		vref, err := refutil.NormalizeFileRef(spec.MustCreateRef("#/definitions/"+name), e.specpath)
		if err != nil {
			return err
		}
		visited := map[string]bool{}
		for k, v := range prop.visitedRefs {
			if k == prop.ownRef.String() {
				continue
			}
			visited[k] = v
		}
		visited[vref.String()] = true

		psch, ownRef, visited, ok, err := refutil.RResolve(e.specpath, sch, visited)
		if err != nil {
			return fmt.Errorf("%s: recursively resolving variant schema (%s): %v", addr, name, err)
		}
		if !ok {
			continue
		}
		if ownRef == nil {
			ownRef = &vref
		}
		prop.Variant[name] = &Property{
			Schema:      psch,
			ownRef:      ownRef,
			addr:        addr,
			visitedRefs: visited,
		}
	}
	return nil
}

func (prop Property) isVariant() bool {
	return len(prop.addr) != 0 && prop.addr[len(prop.addr)-1].Type == PropertyAddrStepTypeVariant
}

func (addr PropertyAddr) String() string {
	var addrs []string
	for _, step := range addr {
		switch step.Type {
		case PropertyAddrStepTypeProp:
			addrs = append(addrs, step.Value)
		case PropertyAddrStepTypeIndex:
			addrs = append(addrs, "*")
		case PropertyAddrStepTypeVariant:
			addrs = append(addrs, "{"+step.Value+"}")
		default:
			panic(fmt.Sprintf("unknown step type: %d", step.Type))
		}
	}
	return strings.Join(addrs, ".")
}

func ParseAddr(input string) PropertyAddr {
	if input == "" {
		return RootAddr
	}
	var addr PropertyAddr
	for _, part := range strings.Split(input, ".") {
		var step PropertyAddrStep
		if part == "*" {
			step = PropertyAddrStep{Type: PropertyAddrStepTypeIndex}
		} else if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			step = PropertyAddrStep{Type: PropertyAddrStepTypeVariant, Value: strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")}
		} else {
			step = PropertyAddrStep{Type: PropertyAddrStepTypeProp, Value: part}
		}
		addr = append(addr, step)
	}
	return addr
}

func schemaTypeIsObject(schema *spec.Schema) bool {
	return len(schema.Type) == 0 || len(schema.Type) == 1 && schema.Type[0] == "object"
}

func schemaIsArray(schema *spec.Schema) bool {
	return len(schema.Type) == 1 && schema.Type[0] == "array"
}

func schemaIsObject(schema *spec.Schema) bool {
	return schemaTypeIsObject(schema) && !schemaIsMap(schema)
}

func schemaIsMap(schema *spec.Schema) bool {
	return schemaTypeIsObject(schema) && len(schema.Properties) == 0 && schema.AdditionalProperties != nil
}
