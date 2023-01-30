[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![Build](https://github.com/drgkaleda/go-multiping/actions/workflows/build.yml/badge.svg)](https://github.com/drgkaleda/go-multiping/actions/workflows/build.yml)
[![Tests](https://github.com/drgkaleda/go-multiping/actions/workflows/test.yml/badge.svg)](https://github.com/drgkaleda/go-multiping/actions/workflows/test.yml)
[![GitHub release](https://img.shields.io/badge/release-releases-green)](https://github.com/drgkaleda/go-multiping/releases/)
[![GitHub license](http://img.shields.io/:license-mit-blue.svg?style=flat-square)](http://badges.mit-license.org)

# go-multiping
Ping library, that can ping and process multipple nodes at once

## The motivation for this multi-ping fork

There are quite a few Go pinger, but all of them have issues:
 * https://github.com/go-ping/ping works fine, but has problems when running
   several pingers in goroutines. When pinging ~300 hosts it looses ~1/3 packets.
  * https://github.com/caucy/batch_ping is umaintened for a long time and did not work for me at all.
  * https://github.com/rosenlo/go-MultiPing is a very young fork, has issues with logger, some parts
    of code are ineffective.

Also need to mention that all these pingers are periodic pingers, they try to mimmic shell ping command. They run
in internal loop, cancel that loop after timeout. They *can* be used, but you have to adjust your code to their style.
Instead I wanted a pinger, that can ping multipple hosts at a time and be robust. I don't think its a problem for
ping user to run it in a loop and don't want any hidden logic.

 So this ping is loosely based on above mentioned projects. It can ping multiple clients. And is cholesterol free.

## Note about concurency
The main package MultiPing has internal lock and can be reused in multiple threads with different PingData.

The PingData however is not tread safe. This mean that during Ping its content can change and thus the caller is responsible for locking.
