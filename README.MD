# GGMM CLI

CLI for managing GGMM hardware radio (WiiMu-based). Controls preset stations and playback via the device's UPnP/SOAP and HTTP API protocols.

## Usage

```
go run ./cmd/main.go [-host <IP>] <command>
```

| Flag / Command | Description |
|---|---|
| `-host <IP>` | Device IP address (default: `127.0.0.1`) |
| `info` | Show device settings |
| `list` | List preset stations (keys 1–6) |
| `set <N> <Title> <URL>` | Set station on key N (1–6) |
| `serve` | Start local web UI at http://localhost:8080 |

### Example

```
go run ./cmd/main.go -host 192.168.1.123 set 1 Radio http://radio.com/playlist.mp3
```

## Build

```bash
go build ./cmd/main.go
go test ./...
```

---

## Device Protocol

The device exposes two protocols.

### 1. HTTP API

Simple HTTP GET API on port **80**.

```
GET http://<host>/httpapi.asp?command=<cmd>
```

| Command | Description |
|---|---|
| `getStatusEx` | Device info (name, firmware, Wi-Fi, IP) |
| `getPlayerStatus` | Current player state (status, title, volume) |
| `setPlayerCmd:play:<url>` | Start playing stream URL |
| `setPlayerCmd:stop` | Stop playback |

String fields (`Title`, `Artist`, `uri`, `essid`) are returned as hex-encoded UTF-8.

**`getStatusEx` response:**
```json
{
  "DeviceName": "GGMM E5",
  "firmware": "4.6.415145",
  "essid": "486f6d65",
  "apcli0": "192.168.1.42",
  "RSSI": "-55"
}
```

**`getPlayerStatus` response:**
```json
{
  "status": "play",
  "Title": "526164696f",
  "Artist": "",
  "vol": "50",
  "mute": "0",
  "uri": "687474703a2f2f..."
}
```

`status` values: `play`, `stop`, `pause`, `load`.

---

### 2. UPnP / SOAP (Presets)

Preset management (keys 1–6) via UPnP on port **59152**.

```
POST http://<host>:59152/upnp/control/PlayQueue1
Content-Type: text/xml; charset=utf-8
SoapAction: "<action-urn>"
```

Service namespace: `urn:schemas-wiimu-com:service:PlayQueue:1`

#### GetKeyMapping — read presets

**SoapAction:** `urn:schemas-wiimu-com:service:PlayQueue:1#GetKeyMapping`

```xml
<u:GetKeyMapping xmlns:u="urn:schemas-wiimu-com:service:PlayQueue:1"></u:GetKeyMapping>
```

The response wraps a `KeyList` XML as HTML-escaped content inside `<QueueContext>` — two decode passes are required (SOAP XML → unescape → KeyList XML).

#### SetKeyMapping — write presets

**SoapAction:** `urn:schemas-wiimu-com:service:PlayQueue:1#SetKeyMapping`

```xml
<u:SetKeyMapping xmlns:u="urn:schemas-wiimu-com:service:PlayQueue:1">
  <QueueContext>&lt;KeyList&gt;...&lt;/KeyList&gt;</QueueContext>
</u:SetKeyMapping>
```

The full `KeyList` (Key0–Key6, `MaxNumber=6`) must be sent — the device replaces all presets at once.

#### KeyList structure

```xml
<KeyList>
  <ListName>Preset</ListName>
  <MaxNumber>6</MaxNumber>
  <Key1>
    <Name>Radio Jazz</Name>
    <Url>http://stream.example.com/jazz</Url>
    <PicUrl></PicUrl>
    <Source>newTuneIn</Source>
    <Metadata></Metadata>
  </Key1>
  <!-- Key2 … Key6 -->
</KeyList>
```

`Key0` is reserved by the device. Use `Source=newTuneIn` for internet radio streams.

---

### curl Examples

```bash
# Device info
curl "http://192.168.1.42/httpapi.asp?command=getStatusEx"

# Player status
curl "http://192.168.1.42/httpapi.asp?command=getPlayerStatus"

# Play a stream
curl "http://192.168.1.42/httpapi.asp?command=setPlayerCmd:play:http://stream.example.com/jazz"

# Stop
curl "http://192.168.1.42/httpapi.asp?command=setPlayerCmd:stop"

# Read presets (SOAP)
curl -X POST http://192.168.1.42:59152/upnp/control/PlayQueue1 \
  -H 'Content-Type: text/xml; charset=utf-8' \
  -H 'SoapAction: "urn:schemas-wiimu-com:service:PlayQueue:1#GetKeyMapping"' \
  -d '<?xml version="1.0" encoding="utf-8"?><s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><u:GetKeyMapping xmlns:u="urn:schemas-wiimu-com:service:PlayQueue:1"></u:GetKeyMapping></s:Body></s:Envelope>'
```