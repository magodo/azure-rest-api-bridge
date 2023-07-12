package ctrl

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
	"github.com/magodo/jsonpointerpos"
)

// ModelMap maps a jsonpointer of a property in the application model to a property *definition* in the API model spec
// In case there is no such property defintion (e.g. some undefined properties appear in an object), or the property definition
// encountered a circular reference during its expansion, the value of the map is nil.
type ModelMap map[string]*swagger.JSONValuePos

func (m ModelMap) AddLink(commit, specdir string) error {
	pm := map[string][]jsonpointer.Pointer{}
	for k, v := range m {
		if v == nil {
			// We deliberately not nil checking `jsonpos` here, as the response is generated with the guarantee that no circular/undefined property will be generated.
			// In other words, the jsonpos here must be non-nil. Otherwise, it indicates a bug in the code.
			return fmt.Errorf("unexpected nil JSONValuePos got for %s, this is either a code bug or user usage error", k)
		}
		filepath := v.Ref.GetURL().Path
		pm[filepath] = append(pm[filepath], *v.Ref.GetPointer())
	}
	posm := map[string]map[string]jsonpointerpos.JSONPointerPosition{}
	for fpath, ptrs := range pm {
		b, err := os.ReadFile(fpath)
		if err != nil {
			return fmt.Errorf("reading %s: %v", fpath, err)
		}
		posResult, err := jsonpointerpos.GetPositions(string(b), ptrs)
		if err != nil {
			return err
		}
		posm[fpath] = posResult
	}
	for _, v := range m {
		fpath := v.Ref.GetURL().Path
		relFile, err := filepath.Rel(specdir, fpath)
		if err != nil {
			return err
		}
		pos, ok := posm[fpath][v.Ref.GetPointer().String()]
		if !ok {
			return fmt.Errorf("can't find file position for %s", &v.Ref)
		}
		v.LinkLocal = fmt.Sprintf("%s:%d:%d", fpath, pos.Line, pos.Column)
		if commit != "" {
			v.LinkGithub = "https://github.com/Azure/azure-rest-api-specs/blob/" + commit + "/specification/" + relFile + "#L" + strconv.Itoa(pos.Line)
		}
	}
	return nil
}

func MapModels(appModel interface{}, apiModels ...swagger.JSONValue) (ModelMap, error) {
	apiValueMap, err := swagger.JSONValueValueMap(apiModels...)
	if err != nil {
		return nil, fmt.Errorf("building value map for API models: %v", err)
	}
	m := map[string]*swagger.JSONValuePos{}
	appValueMap := jsonValueMap(appModel)
	for val, appAddr := range appValueMap {
		apiAddr, ok := apiValueMap[val]
		if !ok {
			continue
		}
		m[appAddr] = apiAddr
	}
	return m, nil
}

func jsonValueMap(m interface{}) map[string]string {
	out := map[string]string{}
	dupm := map[string]bool{}

	tryStore := func(k, v string) {
		if dupm[k] {
			return
		}
		if _, ok := out[k]; ok {
			delete(out, k)
			dupm[k] = true
			return
		}
		out[k] = v
	}

	fn := func(val interface{}, tks []string) {
		etks := make([]string, 0, len(tks))
		for _, tk := range tks {
			etks = append(etks, jsonpointer.Escape(tk))
		}
		ptr, _ := jsonpointer.New("/" + strings.Join(etks, "/"))
		switch val := val.(type) {
		case float64:
			tryStore(strconv.FormatFloat(val, 'g', -1, 64), ptr.String())
		case string:
			tryStore(val, ptr.String())
		case bool:
			v := "FALSE"
			if val {
				v = "TRUE"
			}
			tryStore(v, ptr.String())
		}
	}

	walkJSON(m, []string{}, fn)

	return out
}

func walkJSON(node interface{}, ptks []string, fn func(node interface{}, ptrtks []string)) {
	switch node := node.(type) {
	case map[string]interface{}:
		for k, v := range node {
			tks := make([]string, len(ptks))
			copy(tks, ptks)
			walkJSON(v, append(tks, k), fn)
		}
	case []interface{}:
		for i, v := range node {
			tks := make([]string, len(ptks))
			copy(tks, ptks)
			walkJSON(v, append(tks, strconv.Itoa(i)), fn)
		}
	default:
		fn(node, ptks)
	}
	return
}
