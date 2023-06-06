package ctrl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/go-openapi/jsonpointer"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
)

func Map(appModel interface{}, apiModels ...swagger.JSONValue) (map[string]string, error) {
	apiValueMap, err := swagger.JSONValueValueMap(apiModels...)
	if err != nil {
		return nil, fmt.Errorf("building value map for API models: %v", err)
	}
	m := map[string]string{}
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
