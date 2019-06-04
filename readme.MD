# go-ewon

This package implements the client API for retrieving data from HMS EWON modules via their Talk2M Datamailbox.

## Purpose

This package should help you build applications that request data from EWON modules. At [Factry](https://www.factry.io), we use it to retrieve data from distributed assets such as boilers, electricity substations, etc.

## Example use

A simple example, without error handling, that prints your EWON's name, id and last synchronisation date.

```
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/factrylabs/go-ewon/dmweb"
)

func  main() {

	h  := http.Client{
		Timeout: 10  * time.Second,
	}

	n, _  := dmweb.New(&h, "accountID", "username", "password", "develeropID")
	es, _  := n.GetEwons()
	for  _, e  :=  range es {
		fmt.Println(e.Name, e.ID, e.LastSynchroDate)
	}
}
```

## Documentation

Run `godoc`.

## Missing methods

* delete
* clean

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request

## License

MIT
