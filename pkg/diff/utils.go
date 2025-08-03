package diff

import (
	"github.com/n0madic/mysql-diff/pkg/parser"
)

// Helper functions for pointer and slice comparisons

// ptrToValue converts pointer to value, returning nil if pointer is nil
func ptrToValue[T any](ptr *T) any {
	if ptr == nil {
		return nil
	}
	return *ptr
}

// ptrEqual compares two pointers for equality
func ptrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// generatedColumnEqual compares two GeneratedColumn pointers
func generatedColumnEqual(a, b *parser.GeneratedColumn) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Type == b.Type
}
