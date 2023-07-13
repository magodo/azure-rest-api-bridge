package swagger

import (
	"github.com/go-openapi/loads"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger/refutil"
)

// VariantMap maps the x-ms-discriminator-value to the model name in "/definitions".
// Note that the variant map is the plain translation of the swagger inheritance strucutre, it doesn't take
// cascaded variants into consideration. So always ensure use `Get()` to get the complete variant set of a model.
type VariantMap map[string]VariantInfo

type VariantInfo struct {
	Discriminator       string
	VariantValueToModel map[string]string
}

func (m VariantMap) Get(modelName string) (*VariantInfo, bool) {
	if _, ok := m[modelName]; !ok {
		return nil, false
	}
	wl := []string{}
	out := &VariantInfo{
		Discriminator:       m[modelName].Discriminator,
		VariantValueToModel: map[string]string{},
	}
	for vValue, vName := range m[modelName].VariantValueToModel {
		out.VariantValueToModel[vValue] = vName
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
			for vValue, vName := range mm.VariantValueToModel {
				out.VariantValueToModel[vValue] = vName
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
			m[modelName] = VariantInfo{
				Discriminator:       def.Discriminator,
				VariantValueToModel: map[string]string{},
			}
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
				if parentVariantInfo, ok := m[parent]; ok {
					m[modelName] = VariantInfo{
						Discriminator:       parentVariantInfo.Discriminator,
						VariantValueToModel: map[string]string{},
					}
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
			if varInfo, ok := m[parent]; ok {
				varInfo.VariantValueToModel[vname] = modelName
			}
		}
	}
	return m, nil
}
