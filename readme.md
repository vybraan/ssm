# Secure Shell Manager

SSM allows easy connection to SSH servers, hosts filtering, editing, tagging, command execution and file transfer.

[![version][version-badge]](changelog.md)
[![license][license-badge]](license.md)
[![go report card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)

## Features
- vim keys navigation: jkhl, ctrl+d/u, g/G
- auto-reload SSH config on change
- filter through all your servers
- simple connect and return flow
- switch between SSH and MOSH with a tab
- quickly edit configs
- create free root servers
- extended config with `#tag:` keys
- `$ ssm admin` will load only hosts w/ `#tag: admin`

[CHANGELOG](changelog.md) outlines all features.

## Key-binds
```
<enter↵>       connect to selected host
<ctrl+e>       edit ssh configs
<tab>          switch between SSH/MOSH
</>            filter hosts
<q>            quit

# under development (coming soon)
ctrl+r       run commands on the server without starting a pty 
ctrl+s       sftp upload/download files to/from server 
ctrl+p       port-forwarding UI 
space␣       select multiple hosts to interact with
```

## Install

```bash
# bash one-liner (linux/macos)
curl -sSL https://raw.githubusercontent.com/lfaoro/ssm/refs/heads/main/scripts/get.sh | bash
wget -qO- https://raw.githubusercontent.com/lfaoro/ssm/refs/heads/main/scripts/get.sh | bash

# brew (macos)
brew install lfaoro/tap/ssm

# go install
go install github.com/lfaoro/ssm@latest
```

## Build

> requires [Go](https://go.dev/doc/install)

```bash
git clone https://github.com/lfaoro/ssm.git \
  && cd ssm \
  && make build \
  && bin/ssm
```

## Road map
- [ ] refactor
- [ ] add port-forwarding UI
- [ ] add run command on host
- [ ] add multiple hosts selection
- [ ] add run commands on multiple hosts asynchronously
- [ ] add sftp with interactive files selector
- [ ] add sftp to multiple hosts async

## Contributing

I welcome pull & feature requests, don't hesitate. No rules, we discuss in the issue. Shoot me with your best ideas, bugs & honest feedback.

## Help ssm grow

> If ssm is useful to you, consider being useful to ssm.

- **star the repo**
- **tell your friends**

- [GitHub sponsor](https://github.com/sponsors/lfaoro)
- [FIAT sponsor](https://checkout.revolut.com/pay/1122870b-1836-42e7-942b-90a99ef5e457)
- [BTC sponsor](https://mempool.space/address/bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs)
- [XMR sponsor](https://xmrchain.net/search?value=9XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c)

## License
[MIT license](license.md)

[version-badge]: https://img.shields.io/badge/version-0.1.0-blue.svg
[license-badge]: https://img.shields.io/badge/license-MIT-blue
