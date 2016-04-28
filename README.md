# envsubst [![Build status][travis-image]][travis-url] [![License][license-image]][license-url] [![GoDoc][godoc-img]][godoc-url]
> Environment variables substitution for Go.

#### Installation:
```sh
$ go get github.com/a8m/envsubst/cmd/envsubst
```

#### Using via cli
```sh
$ envsubst < input.tmpl > output.text
$ echo 'welcom $HOME' | substenv
$ substenv -help
```

#### Using `envsubst` programmatically ?
You can take a look on `\_example/main` or see the example below.
```go
package main

import (
	"fmt"
	"github.com/a8m/envsubst"
)

func main() {
    input := "welcom $HOME"
    res, err := envsubst.String(input)
    // ...
    bres, err := envsubst.Bytes([]byte(input))
    // ...
    bres, err := envsubst.ReadFile("filename")
}
```

#### License
MIT

[godoc-url]: https://godoc.org/github.com/a8m/envsubst
[godoc-img]: https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square
[license-image]: https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square
[license-url]: LICENSE
[travis-image]: https://img.shields.io/travis/a8m/envsubst.svg?style=flat-square
[travis-url]: https://travis-ci.org/a8m/envsubst

