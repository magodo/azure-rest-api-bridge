package swagger

import (
	"fmt"
	"strings"

	"github.com/go-openapi/spec"
)

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

type Property struct {
	Schema *spec.Schema
	addr   PropertyAddr

	// The resolved refs (normalized) along the way to this property, which is used to avoid cyclic reference.
	visitedRefs map[string]bool

	// The ref (normalized) that points to the concrete schema of this property.
	// E.g. prop1's schema is "schema1", which refs "schema2", which refs "schema3".
	// Then prop1's ref is (normalized) "schema3"
	ref spec.Ref

	// Discriminator indicates the property name of the parent base schema's discriminator.
	// This only applies to property that is a variant schema.
	Discriminator string

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
