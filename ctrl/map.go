package ctrl

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/go-openapi/jsonreference"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
	"github.com/magodo/jsonpointerpos"
	"golang.org/x/exp/maps"
)

// SingleModelMap maps a jsonpointer of a property in the application model to a property *definition* in the API model spec
// In case there is no such property defintion (e.g. some undefined properties appear in an object), or the property definition
// encountered a circular reference during its expansion, the value of the map is nil.
type SingleModelMap map[string]*swagger.JSONValuePos

// AddLink adds the LinkLocal and LinkGithuhub for each value (*swagger.JSONValuePos) of the SignleModelMap.
func (m SingleModelMap) AddLink(commit, specdir string) error {
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

func (m SingleModelMap) RelativeLocalLink(specdir string) error {
	for _, pos := range m {
		if pos.Ref.GetURL() != nil {
			path, err := filepath.Rel(specdir, pos.Ref.GetURL().Path)
			if err != nil {
				return err
			}
			pos.Ref = jsonreference.MustCreateRef(path + "#" + pos.Ref.GetPointer().String())
		}
		if pos.LinkLocal != "" {
			parts := strings.SplitN(pos.LinkLocal, ":", 2)
			path, err := filepath.Rel(specdir, parts[0])
			if err != nil {
				return err
			}
			pos.LinkLocal = path + ":" + parts[1]
		}
		if ref := pos.RootModel.PathRef; ref.GetURL() != nil {
			path, err := filepath.Rel(specdir, ref.GetURL().Path)
			if err != nil {
				return err
			}
			pos.RootModel.PathRef = jsonreference.MustCreateRef(path + "#" + ref.GetPointer().String())
		}
	}
	return nil
}

func MapSingleAppModel(appModel map[string]interface{}, apiModels ...swagger.JSONValue) (SingleModelMap, error) {
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

// ModelMap is same as SingleModelMap, but might maps one app model property to multiple API model properties.
// This is resulted from merging multiple SingleModelMap(s).
type ModelMap map[string][]*swagger.JSONValuePos

func NewModelMap(models []SingleModelMap) ModelMap {
	tmpM := map[string]map[string]*swagger.JSONValuePos{}
	for _, model := range models {
		for k, v := range model {
			m, ok := tmpM[k]
			if !ok {
				m = map[string]*swagger.JSONValuePos{}
				tmpM[k] = m
			}
			// We use API property address as the unique key
			m[v.Addr.String()] = v
		}
	}
	result := ModelMap{}
	for k, m := range tmpM {
		var l []*swagger.JSONValuePos
		for _, v := range m {
			l = append(l, v)
		}
		sort.Slice(l, func(i, j int) bool {
			return l[i].Addr.String() < l[j].Addr.String()
		})
		result[k] = l
	}
	return result
}

// jsonValueMap flattens a JSON object to a single level k-v map that mapps the jsonpointer to each property to the strings representation of its value, and reverse the keys and values to be a value map.
func jsonValueMap(m map[string]interface{}) map[string]string {
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

	m = flattenJSON(m)
	for k, val := range m {
		switch val := val.(type) {
		case float64:
			tryStore(strconv.FormatFloat(val, 'g', -1, 64), k)
		case string:
			tryStore(val, k)
		case bool:
			v := "FALSE"
			if val {
				v = "TRUE"
			}
			tryStore(v, k)
		}
	}
	return out
}

// flattenJSON flattens a JSON object to a single level k-v map that mapps the jsonpointer to each property's value,
func flattenJSON(m map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	fn := func(val interface{}, tks []string) {
		etks := make([]string, 0, len(tks))
		for _, tk := range tks {
			etks = append(etks, jsonpointer.Escape(tk))
		}
		ptr, _ := jsonpointer.New("/" + strings.Join(etks, "/"))
		switch val := val.(type) {
		case float64, string, bool:
			out[ptr.String()] = val
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

// compareFlattendJSON compares two flattend JSON object and returns:
// - Properties only exist in the 1st object
// - Properties only exist in the 2nd object
// - Properties exist in both objects, but their values are different
// The elements of all the returned slice are the JSON pointer of the properties.
func compareFlattendJSON(m1, m2 map[string]interface{}) (l1 []string, l2 []string, ldiff []string) {
	for _, key := range append(maps.Keys(m1), maps.Keys(m2)...) {
		v1, ok1 := m1[key]
		v2, ok2 := m2[key]
		if !ok1 {
			l2 = append(l2, key)
			continue
		}
		if !ok2 {
			l1 = append(l1, key)
			continue
		}
		if ok1 && ok2 && v1 != v2 {
			ldiff = append(ldiff, key)
			continue
		}
	}
	return
}
