package swagger

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger/refutil"
)

type Expander struct {
	// Operation ref
	ref  spec.Ref
	root *Property

	// This map is not initialized until the first time failed to resolve the model by the discriminator enum value.
	// The key represents each swagger spec
	variantMaps map[string]VariantMap

	// Regard empty object type (no properties&allOf&additionalProperties) as of string type
	// This is for some poorly defined Swagger that defines property as empty objects, but actually return strings (e.g. Azure data factory RP).
	emptyObjAsStr bool

	// Once specified, it will be used for expanding the property. If no hit, it will also update the cache accordingly.
	cache *expanderCache
}

type ExpanderOption struct {
	EmptyObjAsStr bool
	Cache         *expanderCache
}

// NewExpander create a expander for the schema referenced by the input json reference.
// The reference must be a normalized reference.
func NewExpander(ref spec.Ref, opt *ExpanderOption) (*Expander, error) {
	if opt == nil {
		opt = &ExpanderOption{}
	}

	psch, ownRef, visited, ok, err := refutil.RResolve(ref, nil, true)
	if err != nil {
		return nil, fmt.Errorf("recursively resolve schema %s: %v", &ref, err)
	}
	if !ok {
		return nil, fmt.Errorf("circular ref found when resolving schema: %s", &ref)
	}

	return &Expander{
		ref: ref,
		root: &Property{
			Schema:      psch,
			ref:         ownRef,
			addr:        RootAddr,
			visitedRefs: visited,
		},
		variantMaps:   map[string]VariantMap{},
		emptyObjAsStr: opt.EmptyObjAsStr,
		cache:         opt.Cache,
	}, nil
}

// NewExpanderFromOpRef create a expander for the successful response schema of an operation referenced by the input json reference.
// The reference must be a normalized reference to the operation.
func NewExpanderFromOpRef(ref spec.Ref, opt *ExpanderOption) (*Expander, error) {
	if !ref.HasFullFilePath {
		return nil, fmt.Errorf("reference %s is not normalized", &ref)
	}
	tks := ref.GetPointer().DecodedTokens()
	if len(tks) == 0 {
		return nil, fmt.Errorf("reference %s is an empty pointer", &ref)
	}
	opKind := tks[len(tks)-1]

	piref := refutil.Parent(ref)
	pi, err := spec.ResolvePathItemWithBase(nil, piref, nil)
	if err != nil {
		return nil, fmt.Errorf("resolving path item ref %s: %v", &piref, err)
	}

	var op *spec.Operation
	switch strings.ToLower(opKind) {
	case "get":
		op = pi.Get
	case "post":
		op = pi.Post
	default:
		return nil, fmt.Errorf("operation `%s` defined by path item %s is not supported", opKind, &piref)
	}

	if op.Responses == nil {
		return nil, fmt.Errorf("operation refed by %s has no responses defined", &ref)
	}
	// We only care about 200 for now, probably we should extend to support the others (e.g. when 200 is not defined).
	if _, ok := op.Responses.StatusCodeResponses[http.StatusOK]; !ok {
		return nil, fmt.Errorf("operation refed by %s has no 200 responses object defined", &ref)
	}

	// In case the response is a ref itself, follow it
	respref := refutil.Append(ref, "responses", "200")
	_, respref, _, ok, err := refutil.RResolveResponse(respref, nil, false)
	if err != nil {
		return nil, fmt.Errorf("recursively resolve response ref %s: %v", &respref, err)
	}
	if !ok {
		return nil, fmt.Errorf("circular ref found when resolving response ref %s", &respref)
	}

	return NewExpander(refutil.Append(respref, "schema"), opt)
}

func (e *Expander) Root() *Property {
	return e.root
}

func (e *Expander) Expand() error {
	wl := []*Property{e.root}

	if e.cache != nil {
		if e.cache.load(e) {
			return nil
		}
	}

	for {
		if len(wl) == 0 {
			break
		}
		nwl := []*Property{}
		for _, prop := range wl {
			log.Trace("expand", "prop", prop.addr.String(), "ref", prop.ref.String())
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

	if e.cache != nil {
		e.cache.save(e)
	}

	return nil
}

func (e *Expander) expandPropStep(prop *Property) error {
	if prop.Schema == nil {
		return nil
	}
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
		log.Trace("expand step", "type", "array", "prop", prop.addr.String(), "ref", prop.ref.String())
		return e.expandPropStepAsArray(prop)
	case "object":
		if SchemaIsMap(schema) {
			log.Trace("expand step", "type", "map", "prop", prop.addr.String(), "ref", prop.ref.String())
			return e.expandPropAsMap(prop)
		}
		log.Trace("expand step", "type", "object", "prop", prop.addr.String(), "ref", prop.ref.String())
		return e.expandPropAsObject(prop)
	}
	return nil
}

func (e *Expander) expandPropStepAsArray(prop *Property) error {
	schema := prop.Schema
	if !SchemaIsArray(schema) {
		return fmt.Errorf("%s: is not array", prop.addr)
	}
	addr := append(prop.addr, PropertyAddrStep{
		Type: PropertyAddrStepTypeIndex,
	})
	if schema.Items.Schema == nil {
		return fmt.Errorf("%s: items of property is not a single schema (not supported yet)", addr)
	}
	schema, ownRef, visited, ok, err := refutil.RResolve(refutil.Append(prop.ref, "items"), prop.visitedRefs, false)
	if err != nil {
		return fmt.Errorf("%s: recursively resolving items: %v", addr, err)
	}
	if !ok {
		return nil
	}
	prop.Element = &Property{
		Schema:      schema,
		ref:         ownRef,
		addr:        addr,
		visitedRefs: visited,
	}
	return nil
}

func (e *Expander) expandPropAsMap(prop *Property) error {
	schema := prop.Schema
	if !SchemaIsMap(schema) {
		return fmt.Errorf("%s: is not map", prop.addr)
	}
	addr := append(PropertyAddr{}, prop.addr...)
	addr = append(addr, PropertyAddrStep{
		Type: PropertyAddrStepTypeIndex,
	})

	// For definition as below, the .Schema is nil. While .Allow is always true when .AdditionalProperties != nil:
	//   "map": {
	//       "type": "object",
	//       "additionalProperties": true
	//   }
	if schema.AdditionalProperties.Schema == nil {
		prop.Element = &Property{
			Schema: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			},
			ref:         refutil.Append(prop.ref, "additionalProperties"),
			addr:        addr,
			visitedRefs: prop.visitedRefs,
		}
		return nil
	}

	schema, ownRef, visited, ok, err := refutil.RResolve(refutil.Append(prop.ref, "additionalProperties"), prop.visitedRefs, false)
	if err != nil {
		return fmt.Errorf("%s: recursively resolving additionalProperties: %v", addr, err)
	}
	if !ok {
		return nil
	}

	if SchemaIsEmptyObject(schema) && e.emptyObjAsStr {
		//schema.Type = []string{"string"}
		schema = &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
	}

	prop.Element = &Property{
		Schema:      schema,
		ref:         ownRef,
		addr:        addr,
		visitedRefs: visited,
	}
	return nil
}

func (e *Expander) expandPropAsObject(prop *Property) error {
	schema := prop.Schema
	if !SchemaIsObject(schema) {
		return fmt.Errorf("%s: is not object", prop.addr)
	}

	if SchemaIsEmptyObject(schema) && e.emptyObjAsStr {
		//schema.Type = []string{"string"}
		*schema = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
		return nil
	}

	// Figure out whether this is a regular object (including leaf variant model) or a polymorphic object
	vm, err := e.initVariantMap(prop.ref.GetURL().Path)
	if err != nil {
		return err
	}
	mm, ok := vm.Get(prop.SchemaName())
	if !ok || len(mm) == 0 {
		// A regualr object (inlcuding leaf polymorphic model)
		log.Trace("expand step", "type", "regular object", "prop", prop.addr.String(), "ref", prop.ref.String())
		return e.expandPropAsRegularObject(prop)
	} else {
		// A non-leaf polymorphic model
		log.Trace("expand step", "type", "polymorphic object", "prop", prop.addr.String(), "ref", prop.ref.String())
		return e.expandPropAsPolymorphicObject(prop, mm)
	}
}

func (e *Expander) expandPropAsRegularObject(prop *Property) error {
	schema := prop.Schema

	if !SchemaIsObject(schema) {
		return fmt.Errorf("%s: is not object", prop.addr)
	}

	if SchemaIsEmptyObject(schema) && e.emptyObjAsStr {
		//schema.Type = []string{"string"}
		*schema = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type: []string{"string"},
			},
		}
		return nil
	}

	prop.Children = map[string]*Property{}

	// Expanding the regular properties
	for k := range schema.Properties {
		addr := append(PropertyAddr{}, prop.addr...)
		addr = append(addr, PropertyAddrStep{
			Type:  PropertyAddrStepTypeProp,
			Value: k,
		})
		schema, ownRef, visited, ok, err := refutil.RResolve(refutil.Append(prop.ref, "properties", k), prop.visitedRefs, false)
		if err != nil {
			return fmt.Errorf("%s: recursively resolving property %s: %v", addr, k, err)
		}
		if !ok {
			continue
		}
		prop.Children[k] = &Property{
			Schema:      schema,
			ref:         ownRef,
			addr:        addr,
			visitedRefs: visited,
		}
	}

	// Inheriting the allOf schemas
	for i := range schema.AllOf {
		schema, ownRef, visited, ok, err := refutil.RResolve(refutil.Append(prop.ref, "allOf", strconv.Itoa(i)), prop.visitedRefs, false)
		if err != nil {
			return fmt.Errorf("%s: recursively resolving %d-th allOf schema: %v", prop.addr, i, err)
		}
		if !ok {
			continue
		}
		tmpExp := Expander{
			ref: ownRef,
			root: &Property{
				Schema:      schema,
				ref:         ownRef,
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

func (e *Expander) expandPropAsPolymorphicObject(prop *Property, dvals map[string]string) error {
	schema := prop.Schema
	if !SchemaIsObject(schema) {
		return fmt.Errorf("%s: is not object", prop.addr)
	}
	prop.Variant = map[string]*Property{}
	for vValue, vName := range dvals {
		addr := append(PropertyAddr{}, prop.addr...)
		addr = append(addr, PropertyAddrStep{
			Type:  PropertyAddrStepTypeVariant,
			Value: vValue,
		})
		visited := map[string]bool{}
		for k, v := range prop.visitedRefs {
			// Remove the owning ref of the base schema from visited set in order to allow the later allOf inheritance.
			if k == prop.ref.String() {
				continue
			}
			visited[k] = v
		}

		vref := spec.MustCreateRef(prop.ref.GetURL().Path + "#/definitions/" + vName)
		psch, ownRef, visited, ok, err := refutil.RResolve(vref, visited, true)
		if err != nil {
			return fmt.Errorf("%s: recursively resolving variant schema %q by variant value %q: %v", addr, vName, vValue, err)
		}
		if !ok {
			continue
		}
		prop.Variant[vValue] = &Property{
			Schema:             psch,
			ref:                ownRef,
			addr:               addr,
			visitedRefs:        visited,
			Discriminator:      schema.Discriminator,
			DiscriminatorValue: vValue,
		}

	}
	return nil
}

func (e *Expander) initVariantMap(path string) (VariantMap, error) {
	if m := e.variantMaps[path]; m != nil {
		return m, nil
	}
	m, err := NewVariantMap(path)
	if err != nil {
		return nil, err
	}
	e.variantMaps[path] = m
	return m, nil
}

func (e *Expander) cacheKey() string {
	key := e.root.ref.String() + "|"
	if e.emptyObjAsStr {
		key += "1"
	} else {
		key += "0"
	}
	return key
}

func schemaTypeIsObject(schema *spec.Schema) bool {
	return len(schema.Type) == 0 || len(schema.Type) == 1 && schema.Type[0] == "object"
}

func SchemaIsArray(schema *spec.Schema) bool {
	return len(schema.Type) == 1 && schema.Type[0] == "array"
}

func SchemaIsObject(schema *spec.Schema) bool {
	return schemaTypeIsObject(schema) && !SchemaIsMap(schema)
}

func SchemaIsMap(schema *spec.Schema) bool {
	return schemaTypeIsObject(schema) && len(schema.Properties) == 0 && schema.AdditionalProperties != nil
}

func SchemaIsEmptyObject(schema *spec.Schema) bool {
	return SchemaIsObject(schema) && len(schema.Properties) == 0 && len(schema.AllOf) == 0
}
