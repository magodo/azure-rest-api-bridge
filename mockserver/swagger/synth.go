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

func NewSynthesizer(root *Property, rnd *Rnd, opt *SynthesizerOption) Synthesizer {
	if opt == nil {
		opt = &SynthesizerOption{}
	}
	dem := map[string]int{}
	for _, de := range opt.DuplicateElements {
		dem[de.Addr.String()] = de.Cnt
	}
	return Synthesizer{
		root:              root,
		rnd:               rnd,
		useEnumValues:     opt.UseEnumValues,
		duplicateElements: dem,
	}
}

func (syn *Synthesizer) Synthesize() []interface{} {
	var synProp func(parent, p *Property) []interface{}
	synProp = func(parent, p *Property) []interface{} {
		var result []interface{}
		switch {
		case p.Element != nil:
			n := 1
			if cnt, ok := syn.duplicateElements[p.addr.String()]; ok {
				n += cnt
			}

			var innerMatrix [][]interface{}
			for i := 0; i < n; i++ {
				inners := synProp(p, p.Element)
				innerMatrix = append(innerMatrix, inners)
			}

			if SchemaIsArray(p.Schema) {
				for i := 0; i < len(innerMatrix[0]); i++ {
					var res []interface{}
					for j := 0; j < n; j++ {
						inner := innerMatrix[j][i]
						res = append(res, inner)
					}
					result = append(result, res)
				}
			} else {
				// map
				for i := 0; i < len(innerMatrix[0]); i++ {
					res := map[string]interface{}{}
					for j := 0; j < n; j++ {
						key := "KEY"
						if j != 0 {
							key = fmt.Sprintf("KEY%d", j)
						}
						inner := innerMatrix[j][i]
						res[key] = inner
					}
					result = append(result, res)
				}
			}
		case p.Children != nil:
			m := map[string][]interface{}{}
			// empty object
			if len(p.Children) == 0 {
				result = append(result, map[string]interface{}{})
			} else {
				keys := make([]string, 0, len(p.Children))
				for k := range p.Children {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				for _, k := range keys {
					m[k] = synProp(p, p.Children[k])
				}
				for _, v := range CatesianProductMap(m) {
					result = append(result, v)
				}
			}
		case p.Variant != nil:
			keys := make([]string, 0, len(p.Variant))
			for k := range p.Variant {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				result = append(result, synProp(p, p.Variant[k])...)
			}
		default:
			if p.Schema == nil {
				return result
			}
			if len(p.Schema.Type) != 1 {
				panic(fmt.Sprintf("%s: schema type as array is not supported", *p))
			}
			switch t := p.Schema.Type[0]; t {
			case "string":
				if parent != nil && parent.Discriminator != "" && parent.Discriminator == p.Name() {
					// discriminator property
					result = []interface{}{parent.DiscriminatorValue}
				} else {
					// regular string
					if syn.useEnumValues && len(p.Schema.Enum) != 0 {
						result = []interface{}{p.Schema.Enum[0].(string)}
					} else {
						result = []interface{}{syn.rnd.NextString(p.Schema.Format)}
					}
				}
			case "file":
				result = []interface{}{syn.rnd.NextString(p.Schema.Format)}
			case "integer":
				result = []interface{}{syn.rnd.NextInteger(p.Schema.Format)}
			case "number":
				result = []interface{}{syn.rnd.NextNumber(p.Schema.Format)}
			case "boolean":
				result = []interface{}{true}
			case "object", "", "array":
				// Returns nothing as this implies there is a circular ref hit
			default:
				panic(fmt.Sprintf("%s: unknown schema type %s", *p, t))
			}
		}
		return result
	}

	return synProp(nil, syn.root)
}

func CatesianProduct[T any](params ...[]T) [][]T {
	if params == nil {
		return nil
	}
	result := [][]T{}
	for _, param := range params {
		if len(param) != 0 {
			newresult := [][]T{}
			for _, v := range param {
				if len(result) == 0 {
					res := []T{v}
					newresult = append(newresult, res)
				} else {
					for _, res := range result {
						nres := make([]T, len(res))
						copy(nres, res)
						nres = append(nres, v)
						newresult = append(newresult, nres)
					}
				}
			}
			result = newresult
		}
	}
	return result
}

func CatesianProductMap[T any](params map[string][]T) []map[string]T {
	if params == nil {
		return nil
	}
	result := []map[string]T{}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		param := params[k]
		if len(param) != 0 {
			newresult := []map[string]T{}
			for _, v := range param {
				if len(result) == 0 {
					res := map[string]T{k: v}
					newresult = append(newresult, res)
				} else {
					for _, res := range result {
						nres := map[string]T{}
						for kk, vv := range res {
							nres[kk] = vv
						}
						nres[k] = v
						newresult = append(newresult, nres)
					}
				}
			}
			result = newresult
		}
	}
	return result
}
