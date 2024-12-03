package xslices

import "golang.org/x/exp/constraints"

type SingleParamReturnBoolFunc[E any] func(E) bool

func Every[E constraints.Ordered | interface{}](s1 []E, f SingleParamReturnBoolFunc[E]) bool {
	for _, v := range s1 {
		if !f(v) {
			return false
		}
	}
	return true
}
