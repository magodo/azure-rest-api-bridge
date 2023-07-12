package swagger

import (
	"github.com/go-openapi/loads"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger/refutil"
)

// VariantMap maps the x-ms-discriminator-value to the model name in "/definitions".
// The format of the map is: {"parentModelName": {"childVariantValue": "childModelName"}}
// Note that the variant map is the plain translation of the swagger inheritance strucutre, it doesn't take
// cascaded variants into consideration. So always ensure use `Get()` to get the complete variant set of a model.
type VariantMap map[string]map[string]string

func (m VariantMap) Get(modelName string) (map[string]string, bool) {
	if _, ok := m[modelName]; !ok {
		return nil, false
	}
	wl := []string{}
	out := map[string]string{}
	for vValue, vName := range m[modelName] {
		out[vValue] = vName
		wl = append(wl, vName)
	}
	for {
		if len(wl) == 0 {
			break
		}
		oldWl := make([]string, len(wl))
		copy(oldWl, wl)
		wl = []string{}
		for _, modelName := range oldWl {
			mm, ok := m[modelName]
			if !ok {
				continue
			}
			for vValue, vName := range mm {
				out[vValue] = vName
				wl = append(wl, vName)
			}
		}
	}
	return out, true
}

func NewVariantMap(path string) (VariantMap, error) {
	doc, err := loads.Spec(path)
	if err != nil {
		return nil, err
	}
	definitions := doc.Spec().Definitions
	m := VariantMap{}
	for modelName, def := range definitions {
		if def.Discriminator != "" {
			m[modelName] = map[string]string{}
		}
	}

	toContinue := true
	for toContinue {
		toContinue = false
		for modelName, def := range definitions {
			if _, ok := m[modelName]; ok {
				continue
			}
			for _, allOf := range def.AllOf {
				if allOf.Ref.String() == "" {
					continue
				}
				parent := refutil.Last(allOf.Ref.Ref)
				if _, ok := m[parent]; ok {
					m[modelName] = map[string]string{}
					toContinue = true
				}
			}
		}
	}

	for modelName, def := range definitions {
		vname := modelName
		if v, ok := def.Extensions["x-ms-discriminator-value"]; ok {
			vname = v.(string)
		}

		for _, allOf := range def.AllOf {
			if allOf.Ref.String() == "" {
				continue
			}
			parent := refutil.Last(allOf.Ref.Ref)
			if mm, ok := m[parent]; ok {
				mm[vname] = modelName
			}
		}
	}
	return m, nil
}
