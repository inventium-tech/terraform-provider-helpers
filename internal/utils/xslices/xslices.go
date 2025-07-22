// Package xslices provides extended slice utility functions with generic type support.
package xslices

import (
	"cmp"
)

// SingleParamReturnBoolFunc is a generic function type that takes a single parameter
// of type E and returns a boolean value. This is commonly used as a predicate function
// for filtering or testing elements in collections.
type SingleParamReturnBoolFunc[E any] func(E) bool

// Every checks if all elements in the slice satisfy the given predicate function.
// It returns true if all elements pass the test implemented by the provided function,
// or true for an empty slice. It returns false as soon as any element fails the test.
//
// Parameters:
//   - s1: The slice to test
//   - f: The predicate function to test each element
//
// Returns:
//   - bool: true if all elements satisfy the predicate, false otherwise
//
// Example:
//
//	numbers := []int{2, 4, 6, 8}
//	allEven := Every(numbers, func(n int) bool { return n%2 == 0 })
//	// allEven will be true
func Every[S ~[]E, E cmp.Ordered | interface{}](s1 S, f SingleParamReturnBoolFunc[E]) bool {
	for _, v := range s1 {
		if !f(v) {
			return false
		}
	}
	return true
}

// Intersection returns a new slice containing elements that are present in both input slices.
// The function uses a hash map for efficient lookup, resulting in O(n+m) time complexity
// where n and m are the lengths of the input slices. The order of elements in the result
// follows the order they appear in the second slice (s2).
//
// Parameters:
//   - s1: The first slice
//   - s2: The second slice
//
// Returns:
//   - S: A new slice containing the intersection of s1 and s2
//
// Example:
//
//	slice1 := []int{1, 2, 3, 4}
//	slice2 := []int{3, 4, 5, 6}
//	result := Intersection(slice1, slice2)
//	// result will be []int{3, 4}
func Intersection[S ~[]E, E cmp.Ordered](s1, s2 S) S {
	result := make(S, 0)
	// Convert s1 to a hash map for O(1) lookup time complexity
	hash := make(map[E]struct{})

	for _, v := range s1 {
		hash[v] = struct{}{}
	}

	for _, v := range s2 {
		if _, found := hash[v]; found {
			result = append(result, v)
		}
	}

	return result
}

// MapTo transforms each element of the input slice using the provided mapping function
// and returns a new slice of the target type. The length of the result slice will be
// the same as the input slice.
//
// Parameters:
//   - s: The input slice to transform
//   - f: The mapping function that transforms elements from type E to type T
//
// Returns:
//   - []T: A new slice with transformed elements of type T
//
// Example:
//
//	numbers := []int{1, 2, 3, 4}
//	strings := MapTo(numbers, func(n int) string { return fmt.Sprintf("num_%d", n) })
//	// strings will be []string{"num_1", "num_2", "num_3", "num_4"}
func MapTo[S ~[]E, E cmp.Ordered, T any](s S, f func(E) T) []T {
	r := make([]T, len(s))
	for i, v := range s {
		r[i] = f(v)
	}
	return r
}
