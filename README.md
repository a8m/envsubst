# envsubst [![Build status][travis-image]][travis-url] [![License][license-image]][license-url] [![GoDoc][godoc-img]][godoc-url]
> Environment variables substitution for Go. see docs [below](#docs)

#### Installation:
```sh
$ go get github.com/a8m/envsubst/cmd/envsubst
```

#### Using via cli
```sh
$ envsubst < input.tmpl > output.text
$ echo 'welcome $HOME ${USER:=a8m}' | envsubst
$ envsubst -help
```

#### Imposing restrictions
There are two command line flags with which you can cause the substitution to stop with an error code, should the restriction associated with the flag not be met. This can be handy if you want to avoid creating e.g. configuration files with unset or empty parameters. The flags and their restrictions are: 

|__Flag__     | __Meaning__    |
| ------------| -------------- |
|`-no-unset`  | fail if a variable is not set
|`-no-empty`  | fail if a variable is set but empty

These flags can be combined to form tighter restrictions. 

#### Using `envsubst` programmatically ?
You can take a look on [`_example/main`](https://github.com/a8m/envsubst/blob/master/_example/main.go) or see the example below.
```go
package main

import (
	"fmt"
	"github.com/a8m/envsubst"
)

func main() {
    input := "welcom $HOME"
    str, err := envsubst.String(input)
    // ...
    buf, err := envsubst.Bytes([]byte(input))
    // ...
    buf, err := envsubst.ReadFile("filename")
}
```
### Docs
> api docs here: [![GoDoc][godoc-img]][godoc-url]

|__Expression__     | __Meaning__    |
| ----------------- | -------------- |
|`${var}`	   | Value of var (same as `$var`)
|`${var-$DEFAULT}`  | If var not set, evaluate expression as $DEFAULT
|`${var:-$DEFAULT}` | If var not set or is empty, evaluate expression as $DEFAULT
|`${var=$DEFAULT}`  | If var not set, evaluate expression as $DEFAULT
|`${var:=$DEFAULT}` | If var not set or is empty, evaluate expression as $DEFAULT
|`${var+$OTHER}`	   | If var set, evaluate expression as $OTHER, otherwise as empty string
|`${var:+$OTHER}`   | If var set, evaluate expression as $OTHER, otherwise as empty string
<sub>table taken from [here](http://www.tldp.org/LDP/abs/html/refcards.html#AEN22728)</sub>

### See also

* `os.ExpandEnv(s string) string` - only supports `$var` and `${var}` notations

#### License
MIT

[godoc-url]: https://godoc.org/github.com/a8m/envsubst
[godoc-img]: https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square
[license-image]: https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square
[license-url]: LICENSE
[travis-image]: https://img.shields.io/travis/a8m/envsubst.svg?style=flat-square
[travis-url]: https://travis-ci.org/a8m/envsubst

