### go-promisify
A golang package that provides a Javascript/Typescript Promise like go routine management.

### Assumptions

### Usage

#### Importing a package

```
import (
	"github.com/Shehats/go-promisify/promise"
)
```

#### Creating a promise using `promise.Promisify(any, ...any)`


Using a function that should be run in the background
```
func getData(url string) (*http.Response, error) {
	resp, err := http.Get(url)
	return resp, err
}
```
This code would run `getData` concurrently or in parallel depending on the app scope and return a promise of `*Promise[*http.Response]`
```
  url := "some/url/api/v1"
	p1 := promise.Promisify[*http.Response](getData, url)
```
Also we can write the function like that
