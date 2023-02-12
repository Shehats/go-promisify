### go-promisify
A golang package that provides a Javascript/Typescript Promise like go routine management.

### Requirements

- go1.18 or newer

### Usage

#### Importing a package

```go
import (
	"github.com/Shehats/go-promisify/promise"
)
```

#### Creating a promise using `promise.Promisify(any, ...any)`

The first argument of `promise.Promisify` can be:

1. An object like any sruct or pointer instance; in that case the resulting promise will be `*Promise[<YOUR_OBJECT_TYPE>]`

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
2. A function call and in that case the resulting promise with `*Promise[<YOUR_FUNCTION_RETURN_TYPE>`.

eg:
```go
p := promise.Promisify(func() *MyStruct {
    return &MyStruct{
    // set the fields
    }
  }()) // returns *Promise[*MyStruct]
```
3. A function instance and the arguments of the function. This will run the function with the arguments provided in a go routine and return a `*Promise[<YOUR_FUNCTION_RETURN_TYPE>]`. The function should return `(YOUR_TYPE, error)`, otherwise a panic will be raised. Also it's worth noting that the args should be provided in the same order the arguments are defined in the function, otherwise a panic will be raised.

eg:

This example will call this api in the a go routine and returns a `Promise[*http.Response]`

```go 
p1 := promise.Promisify[*http.Response](func(url string, name string, typearg string) (*http.Response, error) {
	resp, err := http.Get(url + "?" + "name=" + name+ "&" + "type="+ typearg)
	return resp, err
}, "some/url/api/v1", "some_name", "some_type")
```

OR


```go
func getData(url string, name string, typearg string) (*http.Response, error) {
	resp, err := http.Get(url + "?" + "name=" + name+ "&" + "type="+ typearg)
	return resp, err
}
url := "some/url/api/v1"
p1 := promise.Promisify[*http.Response](getData, "some/url/api/v1", "some_name", "some_type")
```

### Using `*Promise[T].Then(func(T))`

Using the `.Then` runs a function in a go routine with the result of the promise, if the promise didn't error.

### Using `*Promise[T].Catch(func(error))`

Using the `.Catch`catches an error if the promise throw any errors. Note that if a promise throws an error or panics and `*Promise[T].Catch` nor `Catch(*Promise[T], func(error)(S, error))` are defined will cause a panic.

### Using `*Promise[T].Finally(func())`

Using the `.Finally` runs a function after all of the `Then` and `Catch` functions are done. Ideally it used to do clean ups. Also the `Finally` function cleans up the resources once it is run.

### Using `*Promise[T].Await()`

Using `.Await()` returns the values from the promise. Like javascript's `await` keyword. Also cleans up resources.

### Using `*Promise[T].Exec()`

Using `.Exec()` frees up resources. It is recommended to use `.Exec` if neither `Finally` or `Await` are used to free up resouces so that your app won't use more memory that it needs.

### Creating a `*Promise[S]` from `*Promise[T]`

#### Using `promise.Then[T, S any](promise *Promise[T], successFunc func(T) (S, error)) *Promise[S]`

`promise.Then` creates a Promise of type `S` from the result of a Promise of type `T` like Javascript's `.then`.

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

#### Using `promise.Catch[T, S any](promise *Promise[T], catchFunc func(error) (S, error)) *Promise[S]`

`promise.Catch` creates a Promise of type `S` from the result of a Promise of type `T` if the promise returns an error or panics. Like Javascript's `.catch`.

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
p1.Exec()
// Also this this possible
// The promise from the catch can also be chained with another promises
p3.Then(func (obj MyObj) {
  // Do something
})
p3.Catch(func (err error) {
  // Do something
})
```

### Notes

1. All promises can be chained unless `Exec` or `Finally` or `Await` are called.
2. You run promises inside other promises.
3. You can have promises run in parallel by setting: `runtime.GOMAXPROCS(<SOME_NUMBER>)`

### Contributions are welcome

Just create a github issue and make a PR 🙏



Created with ❤️ by Saleh Shehata

