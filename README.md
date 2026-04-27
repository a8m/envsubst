# envsubst
[![GoDoc][godoc-img]][godoc-url]
[![License][license-image]][license-url]
[![Build status][travis-image]][travis-url]
[![Github All Releases][releases-image]][releases]

> Environment variables substitution for Go. see docs [below](#docs)

## About this fork

This is a fork of [a8m/envsubst](https://github.com/a8m/envsubst) tuned for use as
an [ArgoCD ConfigManagementPlugin](https://argo-cd.readthedocs.io/en/stable/operator-manual/config-management-plugins/)
rendering step. Two deliberate deviations from upstream:

1. **`-prefix` flag.** Restrict substitution to variables matching a given prefix
   (e.g. `-prefix ARGOCD_ENV_`). Variable references that don't match the prefix
   are left as literal text. This keeps the plugin from accidentally substituting
   anything that happens to look like a shell variable in rendered Kubernetes
   manifests.

2. **`$$` is preserved as literal text, not collapsed to `$`.** Upstream a8m
   treats `$$` as a shell-style escape (so `$$VAR` produces literal `$VAR`).
   That semantics silently mangles Kubernetes manifests whose contents quote
   shell or kubelet idioms verbatim — most notably the KEDA Helm chart, which
   embeds Kubernetes' own env-var-expansion docs in CRD schema descriptions
   ("Double `$$` are reduced to a single `$`"). With upstream behavior, ArgoCD
   shows perpetual `$$ → $` drift on synced resources. This fork emits `$$` as
   literal text.

The fork is intentionally narrow: same parser, same library API, same
restrictions (`-no-unset`, `-no-empty`, `-fail-fast`). **Shell-style parameter
expansion operators are preserved** — `${VAR:-default}`, `${VAR:=default}`,
`${VAR:+alt}`, `${VAR:?err}`, `${VAR-default}`, `${VAR=default}`,
`${VAR+alt}`, `${VAR?err}` all work as in upstream. Only the `$$` escape
behavior is dropped. If you want shell-style `$$` escapes, use upstream
[a8m/envsubst](https://github.com/a8m/envsubst) instead.

#### Combining literal `$$` with a substitution

After this change, `$$VAR` does **not** substitute `VAR` — both `$` characters
are consumed as literal text and `VAR` becomes plain text without a leading `$`
to mark it as a variable reference. The same applies to `$${VAR}`: the `$$`
consumes both dollars, leaving `{VAR}` as plain text rather than a substitution.

To emit literal `$$` followed by a substituted variable, use one of:

```
input:  $$$ARGOCD_ENV_FOO    →  output: $$<value-of-FOO>   (triple-dollar; third $ starts the variable)
input:  $$ ${ARGOCD_ENV_FOO} →  output: $$ <value-of-FOO>  (space-separated)

input:  $$ARGOCD_ENV_FOO     →  output: $$ARGOCD_ENV_FOO   (no substitution)
input:  $${ARGOCD_ENV_FOO}   →  output: $${ARGOCD_ENV_FOO} (no substitution)
```

#### Installation:

##### From binaries
Latest stable `envsubst` [prebuilt binaries for 64-bit Linux, or Mac OS X][releases] are available via Github releases.

###### Linux and MacOS
```console
curl -L https://github.com/a8m/envsubst/releases/download/v1.2.0/envsubst-`uname -s`-`uname -m` -o envsubst
chmod +x envsubst
sudo mv envsubst /usr/local/bin
```

###### Windows
Download the latest prebuilt binary from [releases page][releases], or if you have curl installed:
```console
curl -L https://github.com/a8m/envsubst/releases/download/v1.2.0/envsubst.exe
```

##### With go
You can install via `go get` (provided you have installed go):
```console
go get github.com/a8m/envsubst/cmd/envsubst
```


#### Using via cli
```sh
envsubst < input.tmpl > output.text
echo 'welcome $HOME ${USER:=a8m}' | envsubst
envsubst -help
```

#### Imposing restrictions
There are three command line flags with which you can cause the substitution to stop with an error code, should the restriction associated with the flag not be met. This can be handy if you want to avoid creating e.g. configuration files with unset or empty parameters.
Setting a `-fail-fast` flag in conjunction with either no-unset or no-empty or both will result in a faster feedback loop, this can be especially useful when running through a large file or byte array input, otherwise a list of errors is returned.

The flags and their restrictions are: 

|__Option__     | __Meaning__    | __Type__ | __Default__  |
| ------------| -------------- | ------------ | ------------ |
|`-i`  | input file  | `string \| stdin` | `stdin`
|`-o`  | output file | `string \| stdout` |  `stdout`
|`-prefix`  | only substitute variables with this prefix; others are left as literal text (e.g. `-prefix ARGOCD_ENV_`) | `string` | `""` (substitute all)
|`-no-digit`  | do not replace variables starting with a digit, e.g. $1 and ${1} | `flag` |  `false` 
|`-no-unset`  | fail if a variable is not set | `flag` |  `false` 
|`-no-empty`  | fail if a variable is set but empty | `flag` | `false`
|`-fail-fast`  | fails at first occurrence of an error, if `-no-empty` or `-no-unset` flags were **not** specified this is ignored | `flag` | `false`

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
    input := "welcome $HOME"
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
|`${var}`           | Value of var (same as `$var`)
|`${var-$DEFAULT}`  | If var not set, evaluate expression as $DEFAULT
|`${var:-$DEFAULT}` | If var not set or is empty, evaluate expression as $DEFAULT
|`${var=$DEFAULT}`  | If var not set, evaluate expression as $DEFAULT
|`${var:=$DEFAULT}` | If var not set or is empty, evaluate expression as $DEFAULT
|`${var+$OTHER}`    | If var set, evaluate expression as $OTHER, otherwise as empty string
|`${var:+$OTHER}`   | If var set, evaluate expression as $OTHER, otherwise as empty string
|`$$`               | Preserved as literal `$$`. **Differs from upstream a8m**, which treats `$$` as a shell-style escape collapsing to `$`. See [About this fork](#about-this-fork).

<sub>Most of the rows in this table were taken from [here](http://www.tldp.org/LDP/abs/html/refcards.html#AEN22728)</sub>

### See also

* `os.ExpandEnv(s string) string` - only supports `$var` and `${var}` notations

#### Creating Releases

This project uses automated workflows to create releases with prebuilt binaries for multiple platforms.

##### Release Workflows

**`create-release.yml`**: Creates git tags and GitHub releases
- **Trigger**: Manual via GitHub Actions UI
- **Inputs**: Tag name, optional release title and description
- **Features**: 
  - Creates git tag (skips if already exists)
  - Creates GitHub release
  - Handles existing tags gracefully (rerunnable)

**`binaries.yml`**: Builds and uploads binaries
- **Triggers**: 
  - Automatically on release creation
  - Manual dispatch with optional tag name
- **Platforms**: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
- **Features**: Builds from specific tag or latest release

##### Release Procedure

1. Go to the **Actions** tab in the GitHub repository
2. Select the **"Create Release"** workflow
3. Click **"Run workflow"**
4. Enter the tag name following semantic versioning (e.g., `v1.4.5`)
5. Optionally provide a release title and description
6. Click **"Run workflow"**
7. Wait for the workflow to complete successfully
8. Go back to the **Actions** tab and select the **"Release Binaries"** workflow
9. Click **"Run workflow"** to build and upload binaries for the new release
10. Enter the same tag name in step (4) (or select `Use workflow from` released tag)
11. Click **"Run workflow"**

##### What Happens During Release

1. **Tag Creation**: Creates a git tag with the specified version
2. **Release Creation**: Creates a GitHub release with optional title/description
3. **Binary Building**: Automatically triggers binary builds for all platforms:
   - `envsubst-Linux-x86_64` (Linux AMD64)
   - `envsubst-Linux-arm64` (Linux ARM64)
   - `envsubst-Darwin-x86_64` (macOS Intel)
   - `envsubst-Darwin-arm64` (macOS Apple Silicon)
   - `envsubst` (Windows AMD64)
4. **Asset Upload**: Binaries are automatically attached to the release

##### Rerunning Releases

- **Same Tag**: Both workflows can be rerun with the same tag name
- **Tag Exists**: The create-release workflow will skip tag creation if it already exists
- **Release Exists**: The workflow continues even if the release already exists
- **Binary Rebuild**: Use the binaries workflow to rebuild assets for existing releases

##### Supported Platforms

| Platform | Architecture  | Binary Name              |
|----------|---------------|--------------------------|
| Linux    | AMD64         | `envsubst-Linux-x86_64`  |
| Linux    | ARM64         | `envsubst-Linux-arm64`   |
| macOS    | Intel         | `envsubst-Darwin-x86_64` |
| macOS    | Apple Silicon | `envsubst-Darwin-arm64`  |
| Windows  | AMD64         | `envsubst`               |

#### License
MIT

[releases]: https://github.com/a8m/envsubst/releases
[releases-image]: https://img.shields.io/github/downloads/a8m/envsubst/total.svg?style=for-the-badge
[godoc-url]: https://godoc.org/github.com/a8m/envsubst
[godoc-img]: https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge
[license-image]: https://img.shields.io/badge/license-MIT-blue.svg?style=for-the-badge
[license-url]: LICENSE
[travis-image]: https://img.shields.io/travis/a8m/envsubst.svg?style=for-the-badge
[travis-url]: https://travis-ci.org/a8m/envsubst
