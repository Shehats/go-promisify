package promise

import (
	"fmt"
	"reflect"
	"sync"
)

// Promise interface that defines
// methods that are similar to those
// of Javascript
type Promise[T any] struct {
	// successChannel is were the retun
	// of the promise is stored
	successChannel chan T
	// failureChannel is were the error
	// of the primise is stored
	failureChannel chan error
	// mutex that ensures that the promises
	// are executed in order
	mutex *sync.Mutex
	// promise's wait group
	wg *sync.WaitGroup
}

type meta struct {
	objectType  reflect.Type
	objectValue reflect.Value
}

// Creates a new Promise instance
func newPromise[T any]() *Promise[T] {
	return &Promise[T]{
		successChannel: make(chan T, 1),
		failureChannel: make(chan error, 1),
		mutex:          &sync.Mutex{},
		wg:             &sync.WaitGroup{},
	}
}

func fromPromise[T, S any](promise *Promise[T]) *Promise[S] {
	return &Promise[S]{
		successChannel: make(chan S, 1),
		failureChannel: make(chan error, 1),
		mutex:          promise.mutex,
		wg:             promise.wg,
	}
}

func isFunction(obj any) bool {
	return reflect.TypeOf(obj).Kind() == reflect.Func
}

// calls any function by reflection
func call(obj any, args ...meta) (any, error) {
	vargs := make([]reflect.Value, 0)
	for _, arg := range args {
		vargs = append(vargs, arg.objectValue)
	}
	function := reflect.ValueOf(obj)
	ret := function.Call(vargs)
	if !ret[1].IsNil() {
		err := ret[1].Interface().(error)
		return nil, err
	}
	return ret[0].Interface(), nil
}

func execute[T any](
	promise *Promise[T],
	f func(...any) (any, error),
	args ...any) {
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	obj, err := f(args...)
	if err == nil {
		promise.successChannel <- obj.(T)
		promise.failureChannel <- nil
	} else {
		promise.failureChannel <- err
	}
}

func executeThenCallback[T, S any](
	promise1 *Promise[T],
	promise2 *Promise[S],
	f func(T) (S, error),
) {
	defer promise2.mutex.Unlock()
	defer promise2.wg.Done()
	err := <-promise1.failureChannel
	if err == nil {
		arg := <-promise1.successChannel
		obj, err := f(arg)
		promise2.successChannel <- obj
		promise2.failureChannel <- err
	} else {
		promise2.failureChannel <- err
	}
}

func executeCatchCallback[T, S any](
	promise1 *Promise[T],
	promise2 *Promise[S],
	f func(error) (S, error),
) {
	defer promise2.mutex.Unlock()
	defer promise2.wg.Done()
	err := <-promise1.failureChannel
	if err != nil {
		obj, err := f(err)
		promise2.successChannel <- obj
		promise2.failureChannel <- err
	} else {
		promise2.failureChannel <- nil
	}
}

func executeFinally[T any](
	promise *Promise[T],
	f func(),
) {
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	promise.drainChannels()
	f()
}

func executeThen[T any](
	promise *Promise[T],
	f func(T),
) {
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	err := <-promise.failureChannel
	if err == nil {
		obj := <-promise.successChannel
		f(obj)
	} else {
		promise.failureChannel <- err
	}
}

func executeCatch[T any](
	promise *Promise[T],
	f func(error),
) {
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	err := <-promise.failureChannel
	if err != nil {
		f(err)
	}
}

func (promise *Promise[T]) drainChannels() {
	if len(promise.failureChannel) == 1 {
		err := <-promise.failureChannel
		errMsg := fmt.Sprintf("Promise execution has an unhandled error of %v\nPlease consider using a catch clause to handle errors", err)
		panic(errMsg)
	}
	if len(promise.successChannel) == 1 {
		<-promise.successChannel
	}
}

func Promisify[T any](obj any, args ...any) *Promise[T] {
	var promise *Promise[T]
	argsMeta := make([]meta, 0)
	for _, arg := range args {
		argsMeta = append(argsMeta, meta{
			objectType:  reflect.TypeOf(arg),
			objectValue: reflect.ValueOf(arg),
		})
	}
	if isFunction(obj) {
		runner := func(vargs ...any) (any, error) {
			return call(obj, argsMeta...)
		}
		promise = promisifyFunc[T](runner, args)
	}
	return promise
}

// PromisifyFunc
// Executes the function and creates a promise
// from the function's result.
// If the function returns an error, places the err in the failure channel
// If the function returns an object, puts the object in success channel
func promisifyFunc[T any](f func(...any) (any, error), args ...any) *Promise[T] {
	promise := newPromise[T]()
	promise.wg.Add(1)
	promise.mutex.Lock()
	go execute(promise, f, args)
	return promise
}

// Then
// Runs the then function using the
// object in the success channel
func Then[T, S any](promise *Promise[T], successFunc func(T) (S, error)) *Promise[S] {
	resultPromise := fromPromise[T, S](promise)
	promise.mutex.Lock()
	resultPromise.wg.Add(1)
	go executeThenCallback(promise, resultPromise, successFunc)
	return resultPromise
}

// Catch
// Runs the catch function using the
// error in the error channel
func Catch[T, S any](promise *Promise[T], catchFunc func(error) (S, error)) *Promise[S] {
	resultPromise := fromPromise[T, S](promise)
	promise.mutex.Lock()
	resultPromise.wg.Add(1)
	go executeCatchCallback(promise, resultPromise, catchFunc)
	return resultPromise
}

// Finally
// Runs the then function using the
// object in the success channel
// and the error in the error channel
func (promise *Promise[T]) Finally(finallyFunc func()) {
	promise.mutex.Lock()
	promise.wg.Add(1)
	go executeFinally(promise, finallyFunc)
}

func (promise *Promise[T]) Then(successFunc func(T)) {
	promise.mutex.Lock()
	promise.wg.Add(1)
	go executeThen(promise, successFunc)
}

func (promise *Promise[T]) Catch(errorFunc func(error)) {
	promise.mutex.Lock()
	promise.wg.Add(1)
	go executeCatch(promise, errorFunc)
}

// Exec
// Waits for all of the promises to
// execute
func (promise *Promise[T]) Exec() {
	promise.wg.Wait()
}

func (promise *Promise[T]) Await() (T, error) {
	promise.wg.Wait()
	obj := <-promise.successChannel
	err := <-promise.failureChannel
	return obj, err
}
