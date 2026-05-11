# CLAUDE.md

## About

This is CLI for configure GGMM player (hardware radio, playing streams from web)
GGMM uses DLNA protocol

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build ./cmd/main.go   # build the CLI binary
go run ./cmd/main.go     # run without building
go test ./...            # run tests
go vet ./...             # static analysis
go fmt ./...             # format code
```

**CLI usage:**
```
go run ./cmd/main.go [-host <IP>] <command>

Commands:
  info                        list device settings
  list                        list preset stations (keys 1–6)
  set <N> <Title> <URL>       set station on key N (1–6)
```

Default host is `127.0.0.1`; device listens on port `59152`.

## Architecture

Four layers, each in its own package:

**`cmd/main.go`** — parses `-host` flag and routes to a command handler.

**`internal/command/{info,list,set}/`** — one package per CLI subcommand. Each exports a `Handle()` function and receives a `CanConnect` interface via its constructor for dependency injection.

**`internal/ggmm/service/`** — `Service` encapsulates SOAP communication. `GetList()` fetches the current station XML; `SetStations()` pushes an updated `KeyList` back. The device returns station XML embedded as escaped HTML inside a SOAP envelope, so parsing involves two XML decode passes.

**`internal/ggmm/connection/`** — `Connector` is a thin HTTP wrapper that POSTs SOAP requests to the device. Implements the `CanConnect` interface consumed by the service layer. Hard-coded to port 59152 with a 3-second timeout.

**`internal/dto/stations.go`** — `KeyList` (7 slots: Key0–Key6) and `KeyData` (Name, Url, PicUrl, Source, Metadata). `KeyList.Set(index, KeyData)` uses reflection to set fields by index, keeping the XML serialization order intact.

The device uses UPnP/SOAP over HTTP; the target service is `PlayQueue1`. There are no external Go dependencies — everything is pure standard library.