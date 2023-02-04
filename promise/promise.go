package promise

// Promise interface that defines
// methods that are similar to those
// of Javascript
type Promise[T any] interface {
	Then(func(T, ...any) (any, error)) Promise[any]
	Catch(func(error, ...any) (any, error)) Promise[any]
	Finally(func(T, error, ...any))
}

type promiseImpl[T any] struct {
	successChannel chan T
	failureChannel chan error
	argsChannel    chan []any
}

// Creates a new Promise instance
func newPromise[T any]() *promiseImpl[T] {
	return &promiseImpl[T]{
		successChannel: make(chan T, 1),
		failureChannel: make(chan error, 1),
		argsChannel:    make(chan []any, 1),
	}
}

// PromisifyObject
// Creates a promise from any type
// Like JS's Promise.from(any)
func PromisifyObject[T any](obj T, args ...any) Promise[T] {
	promise := newPromise[T]()
	promise.successChannel <- obj
	promise.argsChannel <- args
	return promise
}

// PromisifyFunc
// Executes the function and creates a promise
// from the function's result.
// If the function returns an error, places the err in the failure channel
// If the function returns an object, puts the object in success channel
func PromisifyFunc[T any](f func(...any) (T, error), args ...any) Promise[T] {
	promise := newPromise[T]()
	promise.argsChannel <- args
	go func() {
		obj, err := f(args)
		if err != nil {
			promise.failureChannel <- err
		} else {
			promise.successChannel <- obj
		}
	}()
	return promise
}

// Then
// Runs the then function using the
// object in the success channel
func (p *promiseImpl[T]) Then(successFunc func(T, ...any) (any, error)) Promise[any] {
	resultPromise := newPromise[any]()
	args := <-p.argsChannel
	resultPromise.argsChannel <- args
	go func() {
		arg := <-p.successChannel
		obj, err := successFunc(arg, args...)
		if err != nil {
			resultPromise.failureChannel <- err
		} else {
			resultPromise.successChannel <- obj
		}
	}()
	return resultPromise
}

// Catch
// Runs the catch function using the
// error in the error channel
func (p *promiseImpl[T]) Catch(catchFunc func(error, ...any) (any, error)) Promise[any] {
	resultPromise := newPromise[any]()
	args := <-p.argsChannel
	resultPromise.argsChannel <- args
	go func() {
		err := <-p.failureChannel
		obj, err := catchFunc(err, args...)
		if err != nil {
			resultPromise.failureChannel <- err
		} else {
			resultPromise.successChannel <- obj
		}
	}()
	return resultPromise
}

// Finally
// Runs the then function using the
// object in the success channel
// and the error in the error channel
func (p *promiseImpl[T]) Finally(finallyFunc func(T, error, ...any)) {
	go func() {
		obj := <-p.successChannel
		err := <-p.failureChannel
		args := <-p.argsChannel
		finallyFunc(obj, err, args...)
	}()
}
