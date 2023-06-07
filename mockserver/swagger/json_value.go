package swagger

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/go-openapi/jsonreference"
)

type JSONValue interface {
	JSONValue() interface{}
}

type JSONObject struct {
	value map[string]JSONValue
	pos   *JSONValuePos
}

type primitiveType interface {
	bool | float64 | string
}

func (obj JSONObject) JSONValue() interface{} {
	m := map[string]interface{}{}
	for k, v := range obj.value {
		m[k] = v.JSONValue()
	}
	return m
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

type JSONPrimitive[T primitiveType] struct {
	value T
	pos   *JSONValuePos
}

func (p JSONPrimitive[T]) JSONValue() interface{} {
	return p.value
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
	Ref  jsonreference.Ref `json:"ref"`
	Addr PropertyAddr      `json:"addr"`
}

func (pos JSONValuePos) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"ref":  pos.Ref.String(),
		"addr": pos.Addr.String(),
	}
	return json.Marshal(m)
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
			return nil, fmt.Errorf("%d-th element is not an JSONArray or JSONObject", i)
		}
	}
	return out, nil
}

func UnmarshalJSONToJSONValue(b []byte, root *Property) (JSONValue, error) {
	var val interface{}
	var jsonVal func(v interface{}, prop *Property) (JSONValue, error)
	jsonVal = func(v interface{}, prop *Property) (JSONValue, error) {
		var pos *JSONValuePos
		if prop != nil {
			pos = &JSONValuePos{
				Addr: prop.addr,
				Ref:  prop.ref.Ref,
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

	if err := json.Unmarshal(b, &val); err != nil {
		return nil, err
	}

	return jsonVal(val, root)
}
