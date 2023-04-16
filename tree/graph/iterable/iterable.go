package iterable

// Iterable represents a type that can be iterated on
type Iterable interface {
	Iterate() <-chan interface{}
}

// ToSlice takes an iterable, and converts it into a slice
func ToSlice(i Iterable) []interface{} {
	result := []interface{}{}

	for v := range i.Iterate() {
		result = append(result, v)
	}

	return result
}
