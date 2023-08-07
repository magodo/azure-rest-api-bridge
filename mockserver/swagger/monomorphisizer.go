package swagger

import "sort"

func Monomorphization(prop *Property) []Property {
	var monomorph func(p *Property) []Property
	monomorph = func(p *Property) []Property {
		var result []Property
		switch {
		case p.Element != nil:
			elements := monomorph(p.Element)
			for _, elem := range elements {
				elem := elem
				np := *p
				np.Element = &elem
				result = append(result, np)
			}
		case p.Children != nil:
			if len(p.Children) == 0 {
				// empty object
				np := *p
				return []Property{np}
			}

			m := map[string][]Property{}
			var keys []string
			for k := range p.Children {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				m[k] = monomorph(p.Children[k])
			}
			for _, mm := range CatesianProductMap(m) {
				np := *p
				np.Children = map[string]*Property{}
				for k, child := range mm {
					child := child
					np.Children[k] = &child
				}
				result = append(result, np)
			}
		case p.Variant != nil:
			keys := make([]string, 0, len(p.Variant))
			for k := range p.Variant {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				for _, variant := range monomorph(p.Variant[k]) {
					variant := variant
					np := *p
					np.Variant = map[string]*Property{
						k: &variant,
					}
					result = append(result, np)
				}
			}
		default:
			result = []Property{*p}
		}
		return result
	}
	return monomorph(prop)
}
