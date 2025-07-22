package xslices

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestEvery(t *testing.T) {
	tests := []struct {
		name      string
		slice     []int
		predicate SingleParamReturnBoolFunc[int]
		expected  bool
	}{
		{
			name:      "empty slice returns true",
			slice:     []int{},
			predicate: func(n int) bool { return n > 0 },
			expected:  true,
		},
		{
			name:      "all elements satisfy predicate",
			slice:     []int{2, 4, 6, 8},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  true,
		},
		{
			name:      "not all elements satisfy predicate",
			slice:     []int{2, 3, 4, 6},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  false,
		},
		{
			name:      "no elements satisfy predicate",
			slice:     []int{1, 3, 5, 7},
			predicate: func(n int) bool { return n%2 == 0 },
			expected:  false,
		},
		{
			name:      "single element satisfies predicate",
			slice:     []int{10},
			predicate: func(n int) bool { return n > 5 },
			expected:  true,
		},
		{
			name:      "single element does not satisfy predicate",
			slice:     []int{3},
			predicate: func(n int) bool { return n > 5 },
			expected:  false,
		},
		{
			name:      "all positive numbers",
			slice:     []int{1, 2, 3, 4, 5},
			predicate: func(n int) bool { return n > 0 },
			expected:  true,
		},
		{
			name:      "contains negative number",
			slice:     []int{1, 2, -1, 4, 5},
			predicate: func(n int) bool { return n > 0 },
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Every(tt.slice, tt.predicate)
			if result != tt.expected {
				t.Errorf("Every() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestEveryWithStrings(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		predicate SingleParamReturnBoolFunc[string]
		expected  bool
	}{
		{
			name:      "empty string slice returns true",
			slice:     []string{},
			predicate: func(s string) bool { return len(s) > 0 },
			expected:  true,
		},
		{
			name:      "all strings non-empty",
			slice:     []string{"hello", "world", "test"},
			predicate: func(s string) bool { return len(s) > 0 },
			expected:  true,
		},
		{
			name:      "contains empty string",
			slice:     []string{"hello", "", "test"},
			predicate: func(s string) bool { return len(s) > 0 },
			expected:  false,
		},
		{
			name:      "all strings start with uppercase",
			slice:     []string{"Hello", "World", "Test"},
			predicate: func(s string) bool { return len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z' },
			expected:  true,
		},
		{
			name:      "not all strings start with uppercase",
			slice:     []string{"Hello", "world", "Test"},
			predicate: func(s string) bool { return len(s) > 0 && s[0] >= 'A' && s[0] <= 'Z' },
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Every(tt.slice, tt.predicate)
			if result != tt.expected {
				t.Errorf("Every() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []int
		slice2   []int
		expected []int
	}{
		{
			name:     "empty slices",
			slice1:   []int{},
			slice2:   []int{},
			expected: []int{},
		},
		{
			name:     "first slice empty",
			slice1:   []int{},
			slice2:   []int{1, 2, 3},
			expected: []int{},
		},
		{
			name:     "second slice empty",
			slice1:   []int{1, 2, 3},
			slice2:   []int{},
			expected: []int{},
		},
		{
			name:     "no intersection",
			slice1:   []int{1, 2, 3},
			slice2:   []int{4, 5, 6},
			expected: []int{},
		},
		{
			name:     "partial intersection",
			slice1:   []int{1, 2, 3, 4},
			slice2:   []int{3, 4, 5, 6},
			expected: []int{3, 4},
		},
		{
			name:     "complete intersection",
			slice1:   []int{1, 2, 3},
			slice2:   []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name:     "single element intersection",
			slice1:   []int{1, 2, 3},
			slice2:   []int{3, 4, 5},
			expected: []int{3},
		},
		{
			name:     "duplicate elements in first slice",
			slice1:   []int{1, 1, 2, 3},
			slice2:   []int{1, 2},
			expected: []int{1, 2},
		},
		{
			name:     "duplicate elements in second slice",
			slice1:   []int{1, 2, 3},
			slice2:   []int{1, 1, 2, 2},
			expected: []int{1, 1, 2, 2},
		},
		{
			name:     "different order",
			slice1:   []int{3, 1, 4, 2},
			slice2:   []int{2, 4, 1},
			expected: []int{2, 4, 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.slice1, tt.slice2)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Intersection() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIntersectionWithStrings(t *testing.T) {
	tests := []struct {
		name     string
		slice1   []string
		slice2   []string
		expected []string
	}{
		{
			name:     "empty string slices",
			slice1:   []string{},
			slice2:   []string{},
			expected: []string{},
		},
		{
			name:     "no string intersection",
			slice1:   []string{"apple", "banana"},
			slice2:   []string{"cherry", "date"},
			expected: []string{},
		},
		{
			name:     "partial string intersection",
			slice1:   []string{"apple", "banana", "cherry"},
			slice2:   []string{"banana", "date", "elderberry"},
			expected: []string{"banana"},
		},
		{
			name:     "complete string intersection",
			slice1:   []string{"apple", "banana"},
			slice2:   []string{"apple", "banana"},
			expected: []string{"apple", "banana"},
		},
		{
			name:     "case sensitive intersection",
			slice1:   []string{"Apple", "banana"},
			slice2:   []string{"apple", "banana"},
			expected: []string{"banana"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Intersection(tt.slice1, tt.slice2)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Intersection() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMapTo(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		mapper   func(int) string
		expected []string
	}{
		{
			name:     "empty slice",
			slice:    []int{},
			mapper:   func(n int) string { return strconv.Itoa(n) },
			expected: []string{},
		},
		{
			name:     "int to string conversion",
			slice:    []int{1, 2, 3, 4},
			mapper:   func(n int) string { return strconv.Itoa(n) },
			expected: []string{"1", "2", "3", "4"},
		},
		{
			name:     "int to formatted string",
			slice:    []int{1, 2, 3},
			mapper:   func(n int) string { return fmt.Sprintf("num_%d", n) },
			expected: []string{"num_1", "num_2", "num_3"},
		},
		{
			name:     "single element",
			slice:    []int{42},
			mapper:   func(n int) string { return fmt.Sprintf("value: %d", n) },
			expected: []string{"value: 42"},
		},
		{
			name:     "mathematical transformation",
			slice:    []int{1, 2, 3, 4},
			mapper:   func(n int) string { return strconv.Itoa(n * n) },
			expected: []string{"1", "4", "9", "16"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapTo(tt.slice, tt.mapper)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapTo() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMapToIntToInt(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		mapper   func(int) int
		expected []int
	}{
		{
			name:     "empty slice int to int",
			slice:    []int{},
			mapper:   func(n int) int { return n * 2 },
			expected: []int{},
		},
		{
			name:     "double values",
			slice:    []int{1, 2, 3, 4},
			mapper:   func(n int) int { return n * 2 },
			expected: []int{2, 4, 6, 8},
		},
		{
			name:     "square values",
			slice:    []int{1, 2, 3, 4},
			mapper:   func(n int) int { return n * n },
			expected: []int{1, 4, 9, 16},
		},
		{
			name:     "add constant",
			slice:    []int{1, 2, 3},
			mapper:   func(n int) int { return n + 10 },
			expected: []int{11, 12, 13},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapTo(tt.slice, tt.mapper)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapTo() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestMapToStringToInt(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		mapper   func(string) int
		expected []int
	}{
		{
			name:     "empty string slice",
			slice:    []string{},
			mapper:   func(s string) int { return len(s) },
			expected: []int{},
		},
		{
			name:     "string length",
			slice:    []string{"a", "bb", "ccc", "dddd"},
			mapper:   func(s string) int { return len(s) },
			expected: []int{1, 2, 3, 4},
		},
		{
			name:  "string to uppercase count",
			slice: []string{"Hello", "WORLD", "test"},
			mapper: func(s string) int {
				count := 0
				for _, r := range s {
					if r >= 'A' && r <= 'Z' {
						count++
					}
				}
				return count
			},
			expected: []int{1, 5, 0},
		},
		{
			name:  "count vowels",
			slice: []string{"hello", "world", "aeiou"},
			mapper: func(s string) int {
				vowels := "aeiouAEIOU"
				count := 0
				for _, r := range s {
					if strings.ContainsRune(vowels, r) {
						count++
					}
				}
				return count
			},
			expected: []int{2, 1, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapTo(tt.slice, tt.mapper)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("MapTo() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Benchmark tests for performance validation.
func BenchmarkEvery(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i * 2 // all even numbers
	}
	predicate := func(n int) bool { return n%2 == 0 }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Every(slice, predicate)
	}
}

func BenchmarkIntersection(b *testing.B) {
	slice1 := make([]int, 1000)
	slice2 := make([]int, 1000)
	for i := range slice1 {
		slice1[i] = i
		slice2[i] = i + 500 // 50% overlap
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Intersection(slice1, slice2)
	}
}

func BenchmarkMapTo(b *testing.B) {
	slice := make([]int, 1000)
	for i := range slice {
		slice[i] = i
	}
	mapper := func(n int) string { return strconv.Itoa(n) }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MapTo(slice, mapper)
	}
}
