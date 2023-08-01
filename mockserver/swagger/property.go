package swagger

import (
	"github.com/go-openapi/spec"
)

type PropertyName struct {
	Name    string
	Variant string
}

type Property struct {
	Schema *spec.Schema

	// The API path whose response is the main model of this property.
	apiPath string

	// The property address starting from the main model.
	addr PropertyAddr

	// The resolved refs (normalized) along the way to this property, which is used to avoid cyclic reference.
	visitedRefs map[string]bool

	// The ref (normalized) that points to the concrete schema of this property.
	// E.g. prop1's schema is "schema1", which refs "schema2", which refs "schema3".
	// Then prop1's ref is (normalized) "schema3"
	ref spec.Ref

	// Discriminator indicates the property name of the parent base schema's discriminator.
	// This only applies to property that is a variant schema.
	Discriminator string

	// DiscriminatorValue indicates the discriminator value.
	// This only applies to property that is a variant schema.
	DiscriminatorValue string

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

// PropWalkFunc is invoked during the property tree walking. If it returns false, it will stop walking at that property.
type PropWalkFunc func(p *Property) bool

// Walk walks the property tree in depth first order
func (prop *Property) Walk(fn PropWalkFunc) {
	if prop == nil {
		return
	}
	if !fn(prop) {
		return
	}
	for _, p := range prop.Children {
		p.Walk(fn)
	}
	prop.Element.Walk(fn)
	for _, p := range prop.Variant {
		p.Walk(fn)
	}
}

func (prop Property) SchemaName() string {
	tks := prop.ref.GetPointer().DecodedTokens()
	if len(tks) != 2 || tks[0] != "definitions" {
		return ""
	}
	return tks[1]
}

func (prop Property) Name() string {
	if len(prop.addr) == 0 {
		return ""
	}
	lastStep := prop.addr[len(prop.addr)-1]
	if lastStep.Type != PropertyAddrStepTypeProp {
		return ""
	}
	return lastStep.Value
}

func (prop Property) String() string {
	return prop.addr.String()
}
