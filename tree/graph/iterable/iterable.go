package iterable

// Iterable represents a type that can be iterated on
type Iterable[V any] interface {
	Iterate() <-chan V
}

// ToSlice takes an iterable, and converts it into a slice
func ToSlice[V any](i Iterable[V]) []interface{} {
	result := []interface{}{}

	for v := range i.Iterate() {
		result = append(result, v)
	}

	return result
}
