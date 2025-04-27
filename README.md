# Secure Shell Manager

> Terminal UI for SSH written in Go

[![version][version-badge]](changelog.md)
[![license][license-badge]](license)
[![go report card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)
[![contributors](https://img.shields.io/github/all-contributors/lfaoro/ssm?color=ee8449)](#contributors)
[![follow on x][x-badge]](https://twitter.com/intent/follow?screen_name=leonardofaoro)

[version-badge]: https://img.shields.io/badge/version-0.2.1-blue.svg
[license-badge]: https://img.shields.io/badge/license-BSD3-blue
[x-badge]: https://img.shields.io/twitter/follow/leonardofaoro?label=follow&style=social

SSM is an open source (MIT) SSH connection manager that helps engineers organize servers, connect, filter, tag, execute commands (soon), transfer files (soon), and much more from a simple terminal interface.

![demo](data/demo.png)

## Notable features
- vim keys navigation: jkhl, ctrl+d/u, g/G
- auto-reload SSH config on change
- filter through all your servers
- simple connect and return flow
- switch between SSH and MOSH with a tab
- quickly edit configs `<ctrl+e>`
- create free root servers for research
- extended config with `#tag:` keys, e.g.: `$ ssm admin` will load only hosts w/ `#tag: admin`
- `ssm --exit` will exit ssm once a conn is established
- `ssm --show` or `ctrl+v` in the UI will show selected host params

See [CHANGELOG](changelog.md) for more info. \
See [HELP](data/help) for CLI flags.

## Key-binds
```
<enter↵>       connect to selected host
<ctrl+v>       show config for selected host
<ctrl+e>       edit ssh configs
<tab>          switch between SSH/MOSH
</ or ?>       filter hosts
<q or esc>     quit

# under development (coming soon)
ctrl+r         run commands on the server without starting a pty 
ctrl+s         sftp upload/download files to/from server 
ctrl+p         port-forwarding UI 
space␣         select multiple hosts to interact with
```

## Install

Download `ssm` binary from [Releases](https://github.com/lfaoro/ssm/releases)
> available for Linux, MacOS, FreeBSD, NetBSD, OpenBSD, Solaris on x86_64, i386, arm64, arm architectures

```bash
# bash one-liner for linux/macos
curl -sSL https://raw.githubusercontent.com/lfaoro/ssm/refs/heads/main/scripts/get.sh | bash
wget -qO- https://raw.githubusercontent.com/lfaoro/ssm/refs/heads/main/scripts/get.sh | bash

# go install (requires Go)
go install github.com/lfaoro/ssm@latest
```

## Build from source

> requires [Go](https://go.dev/doc/install)

```bash
git clone https://github.com/lfaoro/ssm.git \
  && cd ssm \
  && make build \
  && bin/ssm
```

## Help
- [SSH config manual](https://man.openbsd.org/ssh_config.5)
- [SSH config example](data/config_example)
- [create SSH config script](scripts/create_config.sh)
- [message me on Telegram](https://t.me/leonarth)

## Road map
- [ ] add port-forwarding UI
- [ ] add run command on host
- [ ] add multiple hosts selection
- [ ] add run commands on multiple hosts asynchronously
- [ ] add sftp with interactive files selector
- [ ] add sftp to multiple hosts async

## Contributors

Pull requests are very welcome and will be merged.
Feature requests are also welcome, we're happy to implement your ideas.

### Support SSM

> If ssm is useful to you, kindly give us a star.

- **star the repo**
- **tell your friends**

- [GitHub sponsor](https://github.com/sponsors/lfaoro)
- [FIAT sponsor](https://checkout.revolut.com/pay/1122870b-1836-42e7-942b-90a99ef5e457)
- [BTC sponsor](https://mempool.space/address/bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs)
- [XMR sponsor](https://xmrchain.net/search?value=9XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c)
