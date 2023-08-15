package swagger

type ExpanderCache struct {
	m map[string]*Property
}

func NewExpanderCache() *ExpanderCache {
	return &ExpanderCache{m: map[string]*Property{}}
}

func (cache *ExpanderCache) save(exp *Expander) {
	cache.m[exp.cacheKey()] = exp.root
}

func (cache *ExpanderCache) load(exp *Expander) bool {
	prop, ok := cache.m[exp.cacheKey()]
	if !ok {
		return false
	}
	exp.root = prop
	return true
}
