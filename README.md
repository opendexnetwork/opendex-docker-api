opendex-docker-api
==================

[![Discord](https://img.shields.io/discord/628640072748761118.svg)](https://discord.gg/RnXFHpn)
[![Build](https://github.com/opendexnetwork/opendex-docker-api/workflows/Build/badge.svg)](https://github.com/opendexnetwork/opendex-docker-api/actions?query=workflow%3ABuild)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

This is an API gateway (proxy) for all OpenDEX services. It provides REST + Socket.IO endpoints which are easier to integrate into frontend projects than gRPC. 


### Compile gRPC proto files

```sh
make proto
```

### Run in [opendex-docker](https://github.com/opendexnetwork/opendex-docker) environment

```bash
scripts/run.sh testnet
```
