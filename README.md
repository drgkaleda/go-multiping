[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![Build](https://github.com/drgkaleda/go-multiping/actions/workflows/build.yml/badge.svg)](https://github.com/drgkaleda/go-multiping/actions/workflows/build.yml)
[![Tests](https://github.com/drgkaleda/go-multiping/actions/workflows/test.yml/badge.svg)](https://github.com/drgkaleda/go-multiping/actions/workflows/test.yml)
[![GitHub release](https://img.shields.io/badge/release-releases-green)](https://github.com/drgkaleda/go-multiping/releases/)
[![GitHub license](http://img.shields.io/:license-mit-blue.svg?style=flat-square)](http://badges.mit-license.org)

# go-multiping
Ping library, that can ping and process multipple nodes at once

## Note about concurency
The main package MultiPing has internal lock and can be reused in multiple threads with different PingData.

The PingData however is not tread safe. This mean that during Ping its content can change and thus the caller is responsible for locking.