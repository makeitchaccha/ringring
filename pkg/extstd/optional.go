package extstd

// Option is a generic type that represents an optional value.
// idea from rust's Option<T> type.

type Option[T any] interface {
	Unwrap() T
	UnwrapOr(defaultValue T) T
	IsSome() bool
	IsNone() bool
}

var _ Option[any] = (*optionImpl[any])(nil)

type optionImpl[T any] struct {
	value    T
	hasValue bool
}

func (o *optionImpl[T]) Unwrap() T {
	if !o.hasValue {
		panic("Option is None")
	}
	return o.value
}

func (o *optionImpl[T]) UnwrapOr(defaultValue T) T {
	if o.hasValue {
		return o.value
	}
	return defaultValue
}

func (o *optionImpl[T]) IsSome() bool {
	return o.hasValue
}

func (o *optionImpl[T]) IsNone() bool {
	return !o.hasValue
}

// Some creates a new Option with a value.
func Some[T any](value T) Option[T] {
	return &optionImpl[T]{value: value, hasValue: true}
}

// None creates a new Option without a value.
func None[T any]() Option[T] {
	return &optionImpl[T]{hasValue: false}
}
