package ctrl

import (
	"fmt"

	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
)

type Synthesizer struct {
	root *swagger.Property
	m    map[string]interface{}
}

func NewSynthesizer(root *swagger.Property) Synthesizer {
	return Synthesizer{
		root: root,
		m:    map[string]interface{}{},
	}
}

func (syn *Synthesizer) Synthesize() []interface{} {
	var synProp func(parent, p *swagger.Property) []interface{}
	synProp = func(parent, p *swagger.Property) []interface{} {
		var result []interface{}
		switch {
		case p.Element != nil:
			inners := synProp(p, p.Element)
			for _, inner := range inners {
				var res interface{}
				if swagger.SchemaIsArray(p.Schema) {
					res = []interface{}{inner}
				} else {
					// map
					res = map[string]interface{}{"key": inner}
				}
				result = append(result, res)
			}
		case p.Children != nil:
			m := map[string][]interface{}{}

			// empty object
			if len(p.Children) == 0 {
				result = append(result, map[string]interface{}{"empty": "empty"})
			} else {
				for k, v := range p.Children {
					m[k] = synProp(p, v)
				}
				for _, v := range CatesianProductMap(m) {
					result = append(result, v)
				}
			}
		case p.Variant != nil:
			for _, pp := range p.Variant {
				result = append(result, synProp(p, pp)...)
			}
		default:
			// discriminator
			if parent != nil {
				if v := parent.VariantName(); v != "" {
					return []interface{}{v}
				}
			}
			if len(p.Schema.Type) != 1 {
				panic(fmt.Sprintf("%s: schema type as array is not supported", *p))
			}
			switch t := p.Schema.Type[0]; t {
			case "string":
				result = []interface{}{"test string"}
			case "integer":
				result = []interface{}{0}
			case "number":
				result = []interface{}{1.2}
			case "boolean":
				result = []interface{}{true}
			default:
				panic(fmt.Sprintf("%s: unknwon schema type %s", *p, t))
			}
		}
		return result
	}

	return synProp(nil, syn.root)
}

func CatesianProduct(params ...[]interface{}) [][]interface{} {
	if params == nil {
		return nil
	}
	result := [][]interface{}{}
	for _, param := range params {
		if len(param) != 0 {
			newresult := [][]interface{}{}
			for _, v := range param {
				if len(result) == 0 {
					res := []interface{}{v}
					newresult = append(newresult, res)
				} else {
					for _, res := range result {
						nres := make([]interface{}, len(res))
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

func CatesianProductMap(params map[string][]interface{}) []map[string]interface{} {
	if params == nil {
		return nil
	}
	result := []map[string]interface{}{}
	for k, param := range params {
		if len(param) != 0 {
			newresult := []map[string]interface{}{}
			for _, v := range param {
				if len(result) == 0 {
					res := map[string]interface{}{k: v}
					newresult = append(newresult, res)
				} else {
					for _, res := range result {
						nres := map[string]interface{}{}
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
