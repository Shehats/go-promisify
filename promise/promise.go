package promise

import (
	"context"
	"log"
)

// Promise interface that defines
// methods that are similar to those
// of Javascript
type Promise[T any] interface {
	Then(func(T) any) Promise[T]
	ThenWithContext(func(T, *context.Context) any) Promise[T]
	Catch(func(error) (any, error)) Promise[T]
	CatchWithContext(func(error, *context.Context) (any, error)) Promise[T]
	Finally(func(T, error) (any, error))
	FinallyWithContext(func(T, error, *context.Context) (any, error))
}

type promiseImpl[T any] struct {
	logger         *log.Logger
	context        *context.Context
	successChannel chan T
	failureChannel chan error
}

// Creates a new Promise instance
func newPromise[T any]() *promiseImpl[T] {
	return &promiseImpl[T]{
		successChannel: make(chan T, 1),
		failureChannel: make(chan error, 1),
	}
}

// PromisifyObject
// Creates a promise from any type
// Like JS's Promise.from(any)
func PromisifyObject[T any](obj T) Promise[T] {
	promise := newPromise[T]()
	promise.successChannel <- obj
	return promise
}

// PromisifyFunc
// Executes the function and creates a promise
// from the function's result.
// If the function returns an error, places the err in the failure channel
// If the function returns an object, puts the object in success channel
func PromisifyFunc[T any](f func(...any) (T, error), args ...any) Promise[T] {
	promise := newPromise[T]()
	go func() {
		obj, err := f(args)
		if err != nil {
			promise.successChannel <- obj
		} else {
			promise.failureChannel <- err
		}
	}()
	return promise
}

// Then
// Runs the then function using the
// object in the success channel
func (p *promiseImpl[T]) Then(successFunc func(T) any) Promise[T] {
	return nil
}

// ThenWithContext
// Runs the then function using the
// object in the success channel and
// also gives access to the context
// in case this is used with database
// connection or web application
func (p *promiseImpl[T]) ThenWithContext(func(T, *context.Context) any) Promise[T] {
	return nil
}

// Catch
// Runs the catch function using the
// error in the error channel
func (p *promiseImpl[T]) Catch(catchFunc func(error) (any, error)) Promise[T] {
	return nil
}

// CatchWithContext
// Runs the catch function using the
// error in the error channel
// also gives access to the context
// in case this is used with database
// connection or web application
func (p *promiseImpl[T]) CatchWithContext(func(error, *context.Context) (any, error)) Promise[T] {
	return nil
}

// Finally
// Runs the then function using the
// object in the success channel
// and the error in the error channel
func (p *promiseImpl[T]) Finally(finallyFunc func(T, error) (any, error)) {}

// FinallyWithContext
// Runs the finally function using the
// object in the success channel
// and the error in the error channel
// also gives access to the context
// in case this is used with database
// connection or web application
func (p *promiseImpl[T]) FinallyWithContext(func(T, error, *context.Context) (any, error)) {}
