package swagger

import (
	"fmt"
	"sort"
)

type Synthesizer struct {
	root *Property
	rnd  *Rnd

	useEnumValues     bool
	duplicateElements map[string]int
}

type SynthesizerOption struct {
	UseEnumValues     bool
	DuplicateElements []SynthDuplicateElement
}

type SynthDuplicateElement struct {
	Cnt  int
	Addr PropertyAddr
}

func NewSynthesizer(root *Property, rnd *Rnd, opt *SynthesizerOption) (*Synthesizer, error) {
	if !root.IsMono() {
		return nil, fmt.Errorf("property is not monomorphisized")
	}
	if opt == nil {
		opt = &SynthesizerOption{}
	}
	dem := map[string]int{}
	for _, de := range opt.DuplicateElements {
		dem[de.Addr.String()] = de.Cnt
	}
	return &Synthesizer{
		root:              root,
		rnd:               rnd,
		useEnumValues:     opt.UseEnumValues,
		duplicateElements: dem,
	}, nil
}

func (syn *Synthesizer) Synthesize() interface{} {
	var synProp func(parent, p *Property) interface{}
	synProp = func(parent, p *Property) interface{} {
		switch {
		case p.Element != nil:
			n := 1
			if cnt, ok := syn.duplicateElements[p.addr.String()]; ok {
				n += cnt
			}

			var elements []interface{}
			for i := 0; i < n; i++ {
				inner := synProp(p, p.Element)
				elements = append(elements, inner)
			}

			if SchemaIsArray(p.Schema) {
				return elements
			} else {
				// map
				res := map[string]interface{}{}
				for i := 0; i < n; i++ {
					key := "KEY"
					if i != 0 {
						key = fmt.Sprintf("KEY%d", i)
					}
					inner := elements[i]
					res[key] = inner
				}
				return res
			}
		case p.Children != nil:
			// empty object
			if len(p.Children) == 0 {
				return map[string]interface{}{}
			} else {
				res := map[string]interface{}{}
				keys := make([]string, 0, len(p.Children))
				for k := range p.Children {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					res[k] = synProp(p, p.Children[k])
				}
				return res
			}
		case p.Variant != nil:
			for _, v := range p.Variant {
				// There must be at most one variant
				return synProp(p, v)
			}
		default:
			if p.Schema == nil {
				return nil
			}
			if len(p.Schema.Type) != 1 {
				panic(fmt.Sprintf("%s: schema type as array is not supported", *p))
			}
			switch t := p.Schema.Type[0]; t {
			case "string":
				if parent != nil && parent.Discriminator != "" && parent.Discriminator == p.Name() {
					// discriminator property
					return parent.DiscriminatorValue
				} else {
					// regular string
					if syn.useEnumValues && len(p.Schema.Enum) != 0 {
						return p.Schema.Enum[0].(string)
					} else {
						return syn.rnd.NextString(p.Schema.Format)
					}
				}
			case "file":
				return syn.rnd.NextString(p.Schema.Format)
			case "integer":
				return syn.rnd.NextInteger(p.Schema.Format)
			case "number":
				return syn.rnd.NextNumber(p.Schema.Format)
			case "boolean":
				return false
			case "object", "", "array":
				// Returns nothing as this implies there is a circular ref hit
			default:
				panic(fmt.Sprintf("%s: unknown schema type %s", *p, t))
			}
		}
		panic("unreachable")
	}

	return synProp(nil, syn.root)
}
