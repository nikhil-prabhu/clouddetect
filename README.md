# clouddetect

![maintenance-status](https://img.shields.io/badge/maintenance-actively--developed-brightgreen.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/nikhil-prabhu/clouddetect.svg)](https://pkg.go.dev/github.com/nikhil-prabhu/clouddetect)
[![License: GPL v3](https://img.shields.io/badge/license-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/license/mit)
[![CI](https://github.com/nikhil-prabhu/clouddetect/actions/workflows/ci.yml/badge.svg)](https://github.com/nikhil-prabhu/clouddetect/actions)
[![CD](https://github.com/nikhil-prabhu/clouddetect/actions/workflows/cd.yml/badge.svg)](https://github.com/nikhil-prabhu/clouddetect/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikhil-prabhu/clouddetect)](https://goreportcard.com/report/github.com/nikhil-prabhu/clouddetect)

A Go library to detect the cloud service provider of a host.

This library is the Go version of my Rust crate
[cloud-detect](https://github.com/nikhil-prabhu/cloud-detect), which in itself is
inspired by the Python-based
[cloud-detect](https://github.com/dgzlopes/cloud-detect) and the Go-based
[satellite](https://github.com/banzaicloud/satellite) modules.

Like these modules, `clouddetect` uses a combination of checking vendor files
and metadata endpoints to accurately determine the cloud provider of a host.

*While this library is structured similarly to the Rust crate, it follows Go
conventions and idioms and is not a direct port.*

## Features

* Currently, this module supports the identification of the following providers:
  * Akamai Cloud (`akamai`)
  * Amazon Web Services (`aws`)
  * Microsoft Azure (`azure`)
  * Google Cloud Platform (`gcp`)
  * Alibaba Cloud (`alibaba`)
  * OpenStack (`openstack`)
  * DigitalOcean (`digitalocean`)
  * Oracle Cloud Infrastructure (`oci`)
  * Vultr (`vultr`)
* Fast, simple and extensible.
* Real-time console logging using the
[`zap`](https://pkg.go.dev/go.uber.org/zap) module.

## Usage

Add the library to your project by running:

```bash
go get github.com/nikhil-prabhu/clouddetect/v2@latest
```

Detect the cloud provider and print the result (with default timeout).

```go
package main

import (
 "fmt"

 "github.com/nikhil-prabhu/clouddetect/v2"
)

func main() {
 provider := clouddetect.Detect()

 // When tested on AWS:
 fmt.Println(provider) // "aws"

 // When tested on local/non-supported cloud environment:
 fmt.Println(provider) // "unknown"
}
```

Detect the cloud provider and print the result (with custom timeout and logging).

```go
package main

import (
 "fmt"

 "github.com/nikhil-prabhu/clouddetect/v2"
 "go.uber.org/zap"
)

func main() {
 // Use zap.NewDevelopment() for development mode
 logger := zap.Must(zap.NewProduction())
 defer logger.Sync()

 provider := clouddetect.Detect(
  clouddetect.WithTimeout(10),
  clouddetect.WithLogger(logger),
 )

 // When tested on AWS:
 fmt.Println(provider) // "aws"

 // When tested on local/non-supported cloud environment:
 fmt.Println(provider) // "unknown"
}
```

You can also check the list of currently supported cloud providers.

```go
package main

import (
 "fmt"

 "github.com/nikhil-prabhu/clouddetect/v2"
)

func main() {
 fmt.Println(clouddetect.SupportedProviders)
}
```

For more detailed documentation, please refer to
the [Module Documentation](https://pkg.go.dev/github.com/nikhil-prabhu/clouddetect).

## Contributing

Contributions are welcome and greatly appreciated! If you’d like to contribute
to clouddetect, here’s how you can help.

### 1. Report Issues

If you encounter a bug, unexpected behavior, or have a feature request, please open
an [issue](https://github.com/nikhil-prabhu/clouddetect/issues/new).
Be sure to include:

* A clear description of the issue.
* Steps to reproduce, if applicable.
* Details about your environment.

### 2. Submit Pull Requests

If you're submitting a
[pull request](https://github.com/nikhil-prabhu/clouddetect/compare), please
ensure the following.

* Your code is formatted using `go fmt`

```bash
go fmt ./...
```

* Code lints pass with (use `--fix` to autofix):

```bash
golangci-lint run -v --fix
```

**NOTE**: To install `golangci-lint`, follow the steps outlined
[on this page](https://golangci-lint.run/welcome/install/#local-installation)

* Your code contains sufficient unit tests and that all tests pass.

```bash
go test ./...
```

### 3. Improve Documentation

If you find areas in the documentation that are unclear or incomplete, feel free
to update the README or module-level documentation. Open a
[pull request](https://github.com/nikhil-prabhu/clouddetect/compare) with your improvements.

### 4. Review Pull Requests

You can also contribute by reviewing
[open pull requests](https://github.com/nikhil-prabhu/clouddetect/pulls?q=is%3Aopen+is%3Apr).
Providing constructive feedback helps maintain a high-quality codebase.
