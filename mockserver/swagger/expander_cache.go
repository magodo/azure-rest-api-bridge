package swagger

type expanderCache struct {
	m map[string]*Property
}

func NewExpanderCache() *expanderCache {
	return &expanderCache{m: map[string]*Property{}}
}

func (cache *expanderCache) save(exp *Expander) {
	cache.m[exp.cacheKey()] = exp.root
}

func (cache *expanderCache) load(exp *Expander) bool {
	prop, ok := cache.m[exp.cacheKey()]
	if !ok {
		return false
	}
	exp.root = prop
	return true
}
