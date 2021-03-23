# cyclonedx-gomod

[![Build Status](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml/badge.svg)](https://github.com/CycloneDX/cyclonedx-gomod/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/CycloneDX/cyclonedx-gomod)](https://goreportcard.com/report/github.com/CycloneDX/cyclonedx-gomod)
[![License](https://img.shields.io/badge/license-Apache%202.0-brightgreen.svg)](LICENSE)
[![Website](https://img.shields.io/badge/https://-cyclonedx.org-blue.svg)](https://cyclonedx.org/)
[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack&labelColor=393939)](https://cyclonedx.org/slack/invite)
[![Group Discussion](https://img.shields.io/badge/discussion-groups.io-blue.svg)](https://groups.io/g/CycloneDX)
[![Twitter](https://img.shields.io/twitter/url/http/shields.io.svg?style=social&label=Follow)](https://twitter.com/CycloneDX_Spec)

*cyclonedx-gomod* creates CycloneDX Software Bill of Materials (SBOM) from Go modules

## Installation

Prebuilt binaries are available on the [releases](https://github.com/CycloneDX/cyclonedx-gomod/releases) page.

### From Source

```
go install github.com/CycloneDX/cyclonedx-gomod@latest
```

Building from source requires Go 1.16 or newer.

## Compatibility

*cyclonedx-gomod* will produce BOMs for the latest version of the CycloneDX specification 
[supported by cyclonedx-go](https://github.com/CycloneDX/cyclonedx-go#compatibility).  
You can use the [CycloneDX CLI](https://github.com/CycloneDX/cyclonedx-cli#convert-command) to convert between multiple 
BOM formats or specification versions. 

## Usage

```
Usage of cyclonedx-gomod:
  -json
        Output in JSON format
  -module string
        Path to Go module (default ".")
  -noserial
        Omit serial number
  -output string
        Output path (default "-")
  -serial string
        Serial number (default [random UUID])
  -type string
        Type of the main component (default "application")
  -version
        Show version
```

In order to be able to calculate hashes, all modules have to be present in Go's module cache.  
Make sure to run `go mod download` before generating BOMs with *cyclonedx-gomod*.

### Example

```
$ go mod tidy
$ go mod download
$ cyclonedx-gomod -output bom.xml 
```

Checkout the [`examples`](./examples) directory for examples of BOMs generated by *cyclonedx-gomod*.

### Replacements

By using the [`replace` directive](https://golang.org/ref/mod#go-mod-file-replace), users of Go modules can replace the 
content of a given module, e.g.:

```
require github.com/jameskeane/bcrypt v0.0.0-20170924085257-7509ea014998
replace github.com/jameskeane/bcrypt => github.com/ProtonMail/bcrypt v0.0.0-20170924085257-7509ea014998
```

We consider the replaced module (`github.com/jameskeane/bcrypt`) to be the ancestor of the replacement 
(`github.com/ProtonMail/bcrypt`) and include it in the replacement's [pedigree](https://cyclonedx.org/use-cases/#pedigree):

```xml
<component bom-ref="pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998" type="library">
  <name>github.com/ProtonMail/bcrypt</name>
  <version>v0.0.0-20170924085257-7509ea014998</version>
  <scope>required</scope>
  <hashes>
    <hash alg="SHA-256">613dae57042245067109a69a8707dc813ab68f78faeb0d349ffdbb81bff3b9bb</hash>
  </hashes>
  <purl>pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998</purl>
  <pedigree>
    <ancestors>
      <component bom-ref="pkg:golang/github.com/jameskeane/bcrypt@v0.0.0-20170924085257-7509ea014998" type="library">
        <name>github.com/jameskeane/bcrypt</name>
        <version>v0.0.0-20170924085257-7509ea014998</version>
        <hashes>
          <hash alg="SHA-256">c510a93977f0fe9cf70bc2b8ec586828f64b985128d88a1f5d2e355b7e895f9f</hash>
        </hashes>
        <purl>pkg:golang/github.com/jameskeane/bcrypt@v0.0.0-20170924085257-7509ea014998</purl>
      </component>
    </ancestors>
  </pedigree>
</component>
```

The [dependency graph](https://cyclonedx.org/use-cases/#dependency-graph) will also reference the replacement, 
not the replaced module:

```xml
<dependencies>
    <dependency ref="pkg:golang/github.com/ProtonMail/proton-bridge@v0.0.0-20210210160947-565c0b6ddf0f">
        <dependency ref="pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998"></dependency>
        <!-- ... -->
    </dependency>
    <dependency ref="pkg:golang/github.com/ProtonMail/bcrypt@v0.0.0-20170924085257-7509ea014998"></dependency>
</dependencies>
```

### Hashes

*cyclonedx-gomod* uses the same hashing algorithm Go uses for its module integrity checks.  
[`vikyd/go-checksum`](https://github.com/vikyd/go-checksum#calc-checksum-of-module-directory) does a great job of
explaining what exactly that entails. In essence, the hash you see in a BOM should be the same as in your `go.sum` file,
just in a different format. This is because the CycloneDX specification enforces hashes to be provided in hex encoding,
while Go uses base64 encoded values.

To verify a hash found in a BOM, do the following:

1. Hex decode the value
2. Base64 encode the value
3. Prefix the value with `h1:`

Given the hex encoded hash `a8962d5e72515a6a5eee6ff75e5ca1aec2eb11446a1d1336931ce8c57ab2503b`, we'd end up with a
module checksum of `h1:qJYtXnJRWmpe7m/3XlyhrsLrEURqHRM2kxzoxXqyUDs=`. 
Now, query your [checksum database](https://go.googlesource.com/proposal/+/master/design/25530-sumdb.md#checksum-database) 
for the expected checksum and compare the values.

## License

Permission to modify and redistribute is granted under the terms of the Apache 2.0 license.  
See the [LICENSE](./LICENSE) file for the full license.

## Contributing

Pull requests are welcome. But please read the
[CycloneDX contributing guidelines](https://github.com/CycloneDX/.github/blob/master/CONTRIBUTING.md) first.

It is generally expected that pull requests will include relevant tests. Tests are automatically run against all
supported Go versions for every pull request.
