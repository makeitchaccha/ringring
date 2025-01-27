package extstd

import "time"

type Cache[T any] interface {
	Valid() bool
	Unwrap() T
}

func NewCache[T any](value T, duration time.Duration) Cache[T] {
	return &normalCacheImpl[T]{
		value:     value,
		expiredAt: time.Now().Add(duration),
	}
}

type normalCacheImpl[T any] struct {
	value     T
	expiredAt time.Time
}

func (c *normalCacheImpl[T]) Valid() bool {
	return c.expiredAt.After(time.Now())
}

func (c *normalCacheImpl[T]) Unwrap() T {
	if !c.Valid() {
		panic("cache expired")
	}
	return c.value
}

type neverExpireCacheImpl[T any] struct {
	value T
}

func (c *neverExpireCacheImpl[T]) Valid() bool {
	return true
}

func (c *neverExpireCacheImpl[T]) Unwrap() T {
	return c.value
}
