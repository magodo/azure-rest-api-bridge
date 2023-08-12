package swagger

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-openapi/jsonreference"
)

type primitiveType interface {
	bool | float64 | string
}

type JSONValue interface {
	JSONValue() interface{}

	// Only JSONPrimitive returns non-nil
	JSONValuePos() *JSONValuePos
}

type JSONObject struct {
	value map[string]JSONValue
	pos   *JSONValuePos
}

func (obj JSONObject) JSONValue() interface{} {
	m := map[string]interface{}{}
	for k, v := range obj.value {
		m[k] = v.JSONValue()
	}
	return m
}

func (obj JSONObject) JSONValuePos() *JSONValuePos {
	return nil
}

type JSONArray struct {
	value []JSONValue
	pos   *JSONValuePos
}

func (arr JSONArray) JSONValue() interface{} {
	l := make([]interface{}, 0, len(arr.value))
	for _, v := range arr.value {
		l = append(l, v.JSONValue())
	}
	return l
}

func (arr JSONArray) JSONValuePos() *JSONValuePos {
	return nil
}

type JSONPrimitive[T primitiveType] struct {
	value T
	pos   *JSONValuePos
}

func (p JSONPrimitive[T]) JSONValue() interface{} {
	return p.value
}

func (p JSONPrimitive[T]) JSONValuePos() *JSONValuePos {
	return p.pos
}

func walkJSONValue(val JSONValue, fn func(val JSONValue)) {
	switch val := val.(type) {
	case JSONArray:
		for _, v := range val.value {
			walkJSONValue(v, fn)
		}
	case JSONObject:
		for _, v := range val.value {
			walkJSONValue(v, fn)
		}
	default:
		fn(val)
	}
	return
}

type JSONValuePos struct {
	RootModel  RootModelInfo     `json:"root_model"`
	Ref        jsonreference.Ref `json:"ref"`
	Addr       PropertyAddr      `json:"addr"`
	LinkLocal  string            `json:"link_local,omitempty"`
	LinkGithub string            `json:"link_github,omitempty"`
}

func (pos JSONValuePos) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"root_model":  pos.RootModel,
		"ref":         pos.Ref.String(),
		"addr":        pos.Addr.String(),
		"link_local":  pos.LinkLocal,
		"link_github": pos.LinkGithub,
	}
	return json.Marshal(m)
}

func (pos *JSONValuePos) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	if v, ok := m["root_model"]; ok {
		b, err := json.Marshal(v)
		if err != nil {
			return nil
		}
		var rootModel RootModelInfo
		if err := json.Unmarshal(b, &rootModel); err != nil {
			return err
		}
		pos.RootModel = rootModel
	}
	if v, ok := m["ref"]; ok {
		pos.Ref = jsonreference.MustCreateRef(v.(string))
	}
	if v, ok := m["addr"]; ok {
		pos.Addr = ParseAddr(v.(string))
	}
	if v, ok := m["link_local"]; ok {
		pos.LinkLocal = v.(string)
	}
	if v, ok := m["link_github"]; ok {
		pos.LinkGithub = v.(string)
	}
	return nil
}

// JSONValueValueMap merges one or more JSONValue into a map whose key is the un-ambiguous leaf value of the input JSONValue(s).
// For the ambiguous leaf value (i.e. multiple properties among the JSONValue(s) have the same value), they are not included in the returning map.
func JSONValueValueMap(l ...JSONValue) (map[string]*JSONValuePos, error) {
	out := map[string]*JSONValuePos{}
	dupm := map[string]bool{}

	tryStore := func(k string, v *JSONValuePos) {
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

	fn := func(val JSONValue) {
		switch val := val.(type) {
		case JSONPrimitive[float64]:
			tryStore(strconv.FormatFloat(val.value, 'g', -1, 64), val.pos)
		case JSONPrimitive[string]:
			tryStore(val.value, val.pos)
		case JSONPrimitive[bool]:
			v := "FALSE"
			if val.value {
				v = "TRUE"
			}
			tryStore(v, val.pos)
		}
	}

	for i, v := range l {
		switch v := v.(type) {
		case JSONArray:
			walkJSONValue(v, fn)
		case JSONObject:
			walkJSONValue(v, fn)
		default:
			return nil, fmt.Errorf("%d-th element is not an JSONArray or JSONObject: %T", i, v)
		}
	}
	return out, nil
}

func FlattenJSONValueObjectByAddr(obj JSONObject) map[string]JSONValue {
	out := map[string]JSONValue{}
	fn := func(val JSONValue) {
		switch val := val.(type) {
		case JSONPrimitive[float64]:
			out[val.pos.Addr.String()] = val
		case JSONPrimitive[string]:
			out[val.pos.Addr.String()] = val
		case JSONPrimitive[bool]:
			out[val.pos.Addr.String()] = val
		}
	}
	walkJSONValue(obj, fn)
	return out
}

func UnmarshalJSONToJSONValue(b []byte, root *Property) (JSONValue, error) {
	var val interface{}

	if err := json.Unmarshal(b, &val); err != nil {
		return nil, err
	}

	var jsonVal func(v interface{}, prop *Property) (JSONValue, error)
	jsonVal = func(v interface{}, prop *Property) (JSONValue, error) {
		// In case the property is polymorphic, get the variant based on the input json value
		if prop != nil && len(prop.Variant) != 0 {
			v := v.(map[string]interface{})
			var randomVariant *Property
			for _, v := range prop.Variant {
				randomVariant = v
				break
			}
			discriminator := randomVariant.Discriminator
			dvalue, ok := v[discriminator].(string)
			if !ok {
				return nil, fmt.Errorf("value of the discriminator %q is not a string in JSON %v, got=%T", discriminator, v, v[discriminator])
			}
			prop = prop.Variant[dvalue]
		}

		var pos *JSONValuePos

		if prop != nil {
			pos = &JSONValuePos{
				Addr:      prop.addr,
				Ref:       prop.ref.Ref,
				RootModel: prop.RootModel,
			}
		}
		switch v := v.(type) {
		case float64:
			return JSONPrimitive[float64]{
				value: v,
				pos:   pos,
			}, nil
		case string:
			return JSONPrimitive[string]{
				value: v,
				pos:   pos,
			}, nil
		case bool:
			return JSONPrimitive[bool]{
				value: v,
				pos:   pos,
			}, nil
		case nil:
			return nil, nil
		case []interface{}:
			var p *Property
			if prop != nil {
				p = prop.Element
			}
			sv := JSONArray{
				pos: pos,
			}
			for i, elem := range v {
				nv, err := jsonVal(elem, p)
				if err != nil {
					return nil, fmt.Errorf("unmarshal %d-th array element: %v", i, err)
				}
				sv.value = append(sv.value, nv)
			}
			return sv, nil
		case map[string]interface{}:
			sv := JSONObject{
				value: map[string]JSONValue{},
				pos:   pos,
			}
			for k, elem := range v {
				var p *Property
				if prop != nil {
					if len(prop.Children) != 0 {
						// This is an object
						p = prop.Children[k]
					} else if prop.Element != nil {
						// This is a map
						p = prop.Element
					}
				}
				var err error
				sv.value[k], err = jsonVal(elem, p)
				if err != nil {
					return nil, fmt.Errorf("unmarshal object element (key=%s): %v", k, err)
				}
			}
			return sv, nil
		default:
			return nil, fmt.Errorf("invalid type: %v (type: %T)", v, v)
		}
	}

	return jsonVal(val, root)
}
