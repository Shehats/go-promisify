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

// stores information about any
// object
type meta struct {
	objectType  reflect.Type
	objectValue reflect.Value
}

// newPromise
// Creates a new Promise instance
func newPromise[T any]() *Promise[T] {
	return &Promise[T]{
		successChannel: make(chan T, 1),
		failureChannel: make(chan error, 1),
		mutex:          &sync.Mutex{},
		wg:             &sync.WaitGroup{},
	}
}

// fromPromise
// Creates a promise from Promise
func fromPromise[T, S any](promise *Promise[T]) *Promise[S] {
	return &Promise[S]{
		successChannel: make(chan S, 1),
		failureChannel: make(chan error, 1),
		mutex:          promise.mutex,
		wg:             promise.wg,
	}
}

// isFunction
// checks whether an obj is a function
func isFunction(obj any) bool {
	return reflect.TypeOf(obj).Kind() == reflect.Func
}

// call
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
		return ret[0].Interface(), err
	}
	return ret[0].Interface(), nil
}

// execute
// calls the function that
// creates the promise and
// and runs a function and
// puts the function's return
// in the subsequent promise
func execute[T any](
	promise *Promise[T],
	f func(any, ...meta) (T, error),
	obj any,
	args ...meta) {
	defer promise.recover()
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	obj, err := f(obj, args...)
	if err == nil {
		promise.successChannel <- obj.(T)
		promise.failureChannel <- nil
	} else {
		promise.failureChannel <- err
	}
}

// executeObj
// locks the mutex while the promise is
// created from an object
func executeObj[T any](promise *Promise[T], obj T) {
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	promise.failureChannel <- nil
	promise.successChannel <- obj
}

// executeThenCallback
// executes then using two promises
func executeThenCallback[T, S any](
	promise1 *Promise[T],
	promise2 *Promise[S],
	f func(T) (S, error),
) {
	defer promise2.recover()
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

// executeCatchCallback
// executes catch using two promises
func executeCatchCallback[T, S any](
	promise1 *Promise[T],
	promise2 *Promise[S],
	f func(error) (S, error),
) {
	defer promise2.recover()
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

// executeFinally
// executes promise's finally
func executeFinally[T any](
	promise *Promise[T],
	f func(),
) {
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	promise.drainChannels()
	f()
}

// executeThen
// executes promise then
func executeThen[T any](
	promise *Promise[T],
	f func(T),
) {
	defer promise.recover()
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	err := <-promise.failureChannel
	if err == nil {
		obj := <-promise.successChannel
		promise.failureChannel <- nil
		f(obj)
	} else {
		promise.failureChannel <- err
	}
}

// executeCatch
// executes promise's catch
func executeCatch[T any](
	promise *Promise[T],
	f func(error),
) {
	defer promise.recover()
	defer promise.mutex.Unlock()
	defer promise.wg.Done()
	err := <-promise.failureChannel
	if err != nil {
		// there can be another then waiting to execute
		// putting nil into the failure channel allows
		// the following then to work correctly
		promise.failureChannel <- nil
		f(err)
	}
}

// drainChannels
// clears channels
func (promise *Promise[T]) drainChannels() {
	if len(promise.failureChannel) == 1 {
		err := <-promise.failureChannel
		if err != nil {
			errMsg := fmt.Sprintf("Promise execution has an unhandled error of %v\nPlease consider using a catch clause to handle errors", err)
			panic(errMsg)
		}
	}
	if len(promise.successChannel) == 1 {
		<-promise.successChannel
	}
}

// funcRunner
// casts call returns to the correct types
func funcRunner[T any](obj any, args ...meta) (T, error) {
	obj, err := call(obj, args...)
	return obj.(T), err
}

// recover
// recovers if the promise run causes
// a panic
func (promise *Promise[T]) recover() {
	if r := recover(); r != nil {
		err := fmt.Errorf("Promise entered an unhealth state due to panic:\n %v", r)
		if len(promise.failureChannel) == 1 {
			<-promise.failureChannel
		}
		promise.failureChannel <- err
	}
}

func Promisify[T any](obj any, args ...any) *Promise[T] {
	var promise *Promise[T]
	if isFunction(obj) {
		argsMeta := make([]meta, 0)
		for _, arg := range args {
			argsMeta = append(argsMeta, meta{
				objectType:  reflect.TypeOf(arg),
				objectValue: reflect.ValueOf(arg),
			})
		}
		promise = promisifyFunc(funcRunner[T], obj, argsMeta...)
	} else {
		promise = promisfyObj(obj.(T))
	}
	return promise
}

func promisfyObj[T any](obj T) *Promise[T] {
	promise := newPromise[T]()
	promise.wg.Add(1)
	promise.mutex.Lock()
	go executeObj(promise, obj)
	return promise
}

// promisifyFunc
// Executes the function and creates a promise
// from the function's result.
// If the function returns an error, places the err in the failure channel
// If the function returns an object, puts the object in success channel
func promisifyFunc[T any](f func(any, ...meta) (T, error), obj any, args ...meta) *Promise[T] {
	promise := newPromise[T]()
	promise.wg.Add(1)
	promise.mutex.Lock()
	go execute(promise, f, obj, args...)
	return promise
}

// Then
// runs a function following a promise
// success and creates a new promise from
// promise.
func Then[T, S any](promise *Promise[T], successFunc func(T) (S, error)) *Promise[S] {
	resultPromise := fromPromise[T, S](promise)
	promise.mutex.Lock()
	resultPromise.wg.Add(1)
	go executeThenCallback(promise, resultPromise, successFunc)
	return resultPromise
}

// Catch
// runs a function following a promise failure
// and creates a new promise from the failed
// promise.
func Catch[T, S any](promise *Promise[T], catchFunc func(error) (S, error)) *Promise[S] {
	resultPromise := fromPromise[T, S](promise)
	promise.mutex.Lock()
	resultPromise.wg.Add(1)
	go executeCatchCallback(promise, resultPromise, catchFunc)
	return resultPromise
}

// Finally
// runs a function after the promise and subsequent promises
// were executed.
// Ideal for clean up functions
func (promise *Promise[T]) Finally(finallyFunc func()) {
	promise.mutex.Lock()
	promise.wg.Add(1)
	go executeFinally(promise, finallyFunc)
}

// Then
// executes a function following a promise sucess
func (promise *Promise[T]) Then(successFunc func(T)) {
	promise.mutex.Lock()
	promise.wg.Add(1)
	go executeThen(promise, successFunc)
}

// Catch
// executes a function following a promise failure
func (promise *Promise[T]) Catch(errorFunc func(error)) {
	promise.mutex.Lock()
	promise.wg.Add(1)
	go executeCatch(promise, errorFunc)
}

// Exec
// Waits for all of the promises to
// execute without returning the
// value and cleans up the resources
// It's recommended to use it if neither Finally
// nor Await are used
func (promise *Promise[T]) Exec() {
	promise.wg.Wait()
	promise.drainChannels()
}

// Await
// Waits for all of the promises to
// execute and returns the computed value
// and the error if there is an error
func (promise *Promise[T]) Await() (T, error) {
	promise.wg.Wait()
	var obj T
	var err error
	if len(promise.successChannel) == 1 {
		obj = <-promise.successChannel
	}
	if len(promise.failureChannel) == 1 {
		err = <-promise.failureChannel
	}
	return obj, err
}
