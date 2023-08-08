package swagger

import "sort"

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
