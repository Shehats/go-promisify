# go-promisify

![Alt text](https://github.com/Shehats/go-promisify/actions/workflows/codeql.yml/badge.svg) ![CI STATUS](https://github.com/Shehats/go-promisify/actions/workflows/test.yml/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/Shehats/go-promisify) [![Go Doc](https://godoc.org/github.com/Shehats/go-promisify?status.svg)](https://pkg.go.dev/github.com/Shehats/go-promisify)

A golang package that provides a ***Javascript/Typescript Promise like abstraction for Golang*** to help Node developers (like me) write concurrent/multithreaded code in Golang.

## Motivation

I developed this package while I was working a project that required me to create multiple go routines to asynchronously do CRUD operations. Coming from NodeJS/Java background, I wanted to manage threads in an abstracted manner.

## Similar packages

I have also found a few great packages that are pretty much the same thing, but not quite the usecase that I had when I was developing this.

Other Similar packages:

1- https://github.com/chebyrash/promise

I believe that there are more similar packages as this is a very common need.

## Requirements

- go1.19.x or newer

## Installing the package

```
go get github.com/Shehats/go-promisify
```

If you're using go.mod do:

```
go get -u github.com/Shehats/go-promisify
```

## Usage

### Importing a package

```go
import (
	"github.com/Shehats/go-promisify/promise"
)
```

## Creating a promise

### ***The different ways of creating a Promise***

We can create a `*Promise[T]` from an object of T (any type including pointers and functions). Think of it as `Promise.resolve` in Javascript.

eg:

```go
myStructInstance := MyStruct{
  // set the fields
}
p := promise.Promisify(myStructInstance) // returns *Promise[MyStruct]
```

```go
myStructInstance := &MyStruct{
  // set the fields
}
p := promise.Promisify(myStructInstance) // returns *Promise[*MyStruct]
```

```go
runner := func() *MyStruct {
   return &MyStruct{
   // set the fields
   }
}
p := promise.Promisify(runner()) // returns *Promise[*MyStruct]
```

We can create a promise using a function and its arguments. This will run the function in a go routine and return the output in a the promise `*Promise[T]`. There a couple of requirements for this to run correctly:

1- The function has to return `(T, error)`.
2- The function's arguments should be passed in the same order as they are defined in the function.

eg:

```go
func callAPI(method string, url string, obj any) (*http.Response, error) {
	// do stuff
}

p := promise.Promisify(callAPI, "GET", "https://myapi.com", nil) // That will return *Promise[*http.Response]

```

But changing the order of the parameters can cause a panic and not returning and error along side the type returned also causes a panic.

Note that we can get creative hear and pass the function directly in the first argument.

```go
p := promise.Promisify(func (method string, url string, obj any) (*http.Response, error) {
	// do stuff
}, "GET", "https://myapi.com", nil) // That will return *Promise[*http.Response]
```

## Creating a promise from a promise (mapping promises)

### Subscribing to a promise and map it to a different promise

`promise.Then[T,S](*Promise[T], func(T)(S,error))` creates a Promise of type `S` from the result of a Promise of type `T` like Javascript's `.then`.

eg:
```go
func getData(url string, name string, typearg string) (*http.Response, error) {
	resp, err := http.Get(url + "?" + "name=" + name+ "&" + "type="+ typearg)
	return resp, err
}
.......
url := "some/url/api/v1"
p1 := promise.Promisify[*http.Response](getData, "some/url/api/v1", "some_name", "some_type")
p2 := promise.Then(p1, func(r *http.Response) (MyObj, error) {
   var obj MyObj
   b, err := io.ReadAll(r.Body)
   if err != nil {
      return MyObj{}, err
   }
   err = json.Unmarshal(b, &obj)
   if err != nil {
      return MyObj{}, err
   }
   return obj, nil
}) // This creates *Promise[MyObj]
// promise can be chained
p2.Then(func (obj MyObj) {
  // Do something
})
p1.Catch(func (err error) {
  // Do something
})
```

### Subscribing to a promise on failure and mapping error to a different promse.

`promise.Catch[T,S](*Promise[T], func(error)(S, error)` creates a Promise of type `S` from the result of a Promise of type `T` if the promise returns an error or panics. Like Javascript's `.catch`.

eg:
```go
url := "some/url/api/v1"
p1 := promise.Promisify[*http.Response](getData, "some/url/api/v1", "some_name", "some_type")
p2 := promise.Then(p1, func(r *http.Response) (MyObj, error) {
   var obj MyObj
   b, err := io.ReadAll(r.Body)
   if err != nil {
      return MyObj{}, err
   }
   err = json.Unmarshal(b, &obj)
   if err != nil {
      return MyObj{}, err
   }
   return obj, nil
}) // This creates *Promise[MyObj]
p3 := promise.Catch(p, func(err error) (AnyObj, error) {
   // do something
   return obj, error
}) // This creates *Promise[AnyObj] if p promise fails
// promise can be chained
p2.Then(func (obj MyObj) {
  // Do something
})
p2.Catch(func (err error) {
  // Do something
})
// Also this this possible
// The promise from the catch can also be chained with another promises
p3.Then(func (obj MyObj) {
  // Do something
})
p3.Catch(func (err error) {
  // Do something
})
```

## Coding in Javascript like fashon

### Subscribing to a promise when the promise succeeds

If we'd like to subscribe to a promise and not to create a promise of it, while executing things asynchronously, we can use `*Promise[T].Then(func(T))`. This is pretty similar to `Promise.then` in Javascript, however it doesn't create a new promise. If we want to create a promise from a successful promise we can use `promise.Then` from earlier.

Using `*Promise[T].Then` is ideal to do something when a promise when the promise succeeds. Also `*Promise[T].Then` clears up the memory that was used so there is need to do `Await`, `Exec` or `Finally` afterwards. Also only `*Promise[T].Finally` can be used after `*Promise[T].Then`.

eg:

```go
p := promise.Promisify(func (method string, url string, obj any) (*http.Response, error) {
	// do stuff
}, "GET", "https://myapi.com", nil) // That will return *Promise[*http.Response]
p.Then(func(resp *http.Response) {
	// do something on reciving the response like calling another api or writing to the db
})
```

### Subscribing to a promise when the promise fails/errors

Using the `*Promise[T].Catch(func(error))`catches an error if the promise throw any errors and like the `*Promise[T].Then(func(T))` it will subscribe to the promise and clean the resources.

***Note that if a promise throws an error or panics and `*Promise[T].Catch` nor `Catch(*Promise[T], func(error)(S, error))` are defined will cause a panic.***

### Subscribing to a promise anyways *finally*

Using the `*Promise[T].Finally(func())` runs a function after all of the `Then` and `Catch` functions are done. Ideally it used to do clean ups. Also the `Finally` function cleans up the resources once it is run.

***Note that the `*Promise[T].Finally` always executes.***

### Using `*Promise[T].Await()`

Using `.Await()` returns the values from the promise. Like javascript's `await` keyword. Also cleans up resources.

eg:

```go
p := promise.Promisify(func (method string, url string, obj any) (*http.Response, error) {
	// do stuff
}, "GET", "https://myapi.com", nil) // That will return *Promise[*http.Response]

resp, err := p.Await()
```
## Notes

1. All promises can be chained unless `Exec` or `Finally` or `Await` are called.
2. You run promises inside other promises. see the [test](https://github.com/Shehats/go-promisify/blob/main/promise_web_test.go#L261)
3. You can have promises run in parallel by setting: `runtime.GOMAXPROCS(<SOME_NUMBER>)`

## Contributions are welcome

Just create a github issue and make a PR 🙏





Created with ❤️ by Saleh Shehata
