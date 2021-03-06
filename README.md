[![CircleCI Image]](https://circleci.com/gh/nimona/go-nimona)
[![Coveralls Image]](https://coveralls.io/github/nimona/go-nimona)
[![License Image]](https://github.com/nimona/go-nimona/blob/master/LICENSE)
[![Issues Image]](https://waffle.io/nimona/go-nimona)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fnimona%2Fgo-nimona.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fnimona%2Fgo-nimona?ref=badge_shield)

# Nimona

Nimona’s main goal is to provide a number of layers/components to help with the challenges presented when dealing with decentralized and peer to peer applications.

## Architecture

For a technical overview, please refer to the [documentation introduction](https://nimona.io).

## Development

### Installation from Source

Assuming you have a working [Go environment] with Go 1.10 or higher:

```
go get -d nimona.io/go
cd $GOPATH/src/nimona.io/go
dep ensure 
```

### Running

You can either `go install nimona.io/go/cmd/nimona` or run it from 
source every time with `go run nimona.io/go/cmd/nimona`.

### Installation in Provider

You can install the daemon in a supported provider.
```
nimona daemon install --platform do --token <> --ssh-fingerprint <> --hostname <>
```

#### Supported Flags
* **--platform** the provider to be used for the deployment
* **--hostname** the hostname that nimona will use, if defined the dns will also be updated
* **--token** the access token required to authenticate with the provider
* **--ssh-fingerprint** the ssh fingerprint for the key that will be added to the server (needs to exist in the provider)
* **--size** size of the server, default for DO *s-1vcpu-1gb*
* **--region** region that the server will be deployed, default *lon1*

#### Suppored Providers
* do - DigitalOcean

#### Commands

[CircleCI Image]: https://img.shields.io/circleci/project/github/nimona/go-nimona.svg?style=flat-square
[Coveralls Image]: https://img.shields.io/coveralls/github/nimona/go-nimona.svg?style=flat-square
[License Image]: https://img.shields.io/github/license/nimona/go-nimona.svg?style=flat-square
[Issues Image]: https://img.shields.io/waffle/label/nimona/go-nimona/in%20progress.svg?style=flat-square

[Go environment]: https://golang.org/doc/install


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fnimona%2Fgo-nimona.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fnimona%2Fgo-nimona?ref=badge_large)