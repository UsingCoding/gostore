package orderedmap

import (
	"cmp"
	"slices"
)

func New[K comparable, V any](sort SortFunc[K]) *Map[K, V] {
	return &Map[K, V]{
		keys: nil,
		v:    map[K]V{},
		sort: sort,
	}
}

func NewStable[K cmp.Ordered, V any]() *Map[K, V] {
	return New[K, V](slices.Sort[[]K, K])
}

type SortFunc[K comparable] func(s []K)

type Map[K comparable, V any] struct {
	keys []K
	v    map[K]V

	sort func(s []K)
}

type Pair[K comparable, V any] struct {
	K K
	V V
}

func (o *Map[K, V]) Keys() []K {
	keys := slices.Clone(o.keys)
	o.sort(keys)
	return keys
}

func (o *Map[K, V]) Pairs() []Pair[K, V] {
	res := make([]Pair[K, V], 0, len(o.keys))
	for _, k := range o.Keys() {
		v := o.v[k]
		res = append(res, Pair[K, V]{
			K: k,
			V: v,
		})
	}
	return res
}

func (o *Map[K, V]) Add(k K, v V) {
	_, exists := o.v[k]
	if !exists {
		o.keys = append(o.keys, k)
	}
	o.v[k] = v
}
