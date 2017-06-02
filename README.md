# unitsvc

[![Build Status](https://travis-ci.org/studiously/unitsvc.svg?branch=master)](https://travis-ci.org/studiously/unitsvc) [![Coverage Status](https://coveralls.io/repos/github/studiously/unitsvc/badge.svg)](https://coveralls.io/github/studiously/unitsvc)

Unit service for Studiously.

>unitsvc is responsible for data and functionality pertaining to units and their relation to classes.

## Quickstart

To start a unitsvc instance, you can quickly grab the CLI:

```bash
go get github.com/studiously/unitsvc
```

Then, to see what configuration options are necessary, run:
```bash
unitsvc host --help
```

There is also a minimal image available at [https://hub.docker.com/r/studiously/unitsvc/](https://hub.docker.com/r/studiously/unitsvc/).

## Requirements

- Go 1.8+ (or use the Docker image)
- postgres database
- [classsvc](https://github.com/studiously/classsvc) 
- [Hydra](https://github.com/ory/hydra)
- [NATS](https://github.com/nats-io/go-nats)