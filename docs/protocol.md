# GGMM Device Protocol

Устройство GGMM (WiiMu-based радио) поддерживает два независимых протокола.

---

## 1. HTTP API (`httpapi.asp`)

Простой HTTP GET API. Порт **80** (стандартный).

```
GET http://<host>/httpapi.asp?command=<cmd>
```

Ответы — JSON.

### Команды

| Команда | Описание | Ответ |
|---|---|---|
| `getStatusEx` | Информация об устройстве | JSON-объект |
| `getPlayerStatus` | Текущее состояние плеера | JSON-объект |
| `setPlayerCmd:play:<url>` | Запустить поток по URL | — |
| `setPlayerCmd:stop` | Остановить воспроизведение | — |

### `getStatusEx` — поля ответа

```json
{
  "DeviceName": "GGMM E5",
  "firmware": "4.6.415145",
  "essid": "486f6d65",   // hex-encoded SSID
  "apcli0": "192.168.1.42",
  "RSSI": "-55"
}
```

> `essid` — hex-строка в UTF-8. Декодируется парами байт: `486f6d65` → `Home`.

### `getPlayerStatus` — поля ответа

```json
{
  "status": "play",
  "Title": "526164696f",  // hex-encoded
  "Artist": "",
  "vol": "50",
  "mute": "0",
  "uri": "687474703a2f2f..."  // hex-encoded
}
```

Поля `Title`, `Artist`, `uri` — hex-encoded UTF-8 строки. Значение `status`: `play`, `stop`, `pause`, `load`.

---

## 2. UPnP / SOAP (`PlayQueue1`)

Используется для управления пресетами (кнопки 1–6). Порт **59152**.

### Транспорт

```
POST http://<host>:59152/upnp/control/PlayQueue1
Content-Type: text/xml; charset=utf-8
SoapAction: "<action-urn>"
```

Тело — SOAP-конверт:

```xml
<?xml version="1.0" encoding="utf-8"?>
<s:Envelope
  s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"
  xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <!-- action element -->
  </s:Body>
</s:Envelope>
```

Namespace сервиса: `urn:schemas-wiimu-com:service:PlayQueue:1`

---

### GetKeyMapping — читать пресеты

**SoapAction:** `urn:schemas-wiimu-com:service:PlayQueue:1#GetKeyMapping`

**Запрос (тело `<s:Body>`):**

```xml
<u:GetKeyMapping xmlns:u="urn:schemas-wiimu-com:service:PlayQueue:1"></u:GetKeyMapping>
```

**Ответ:**

```xml
<s:Envelope ...>
  <s:Body>
    <u:GetKeyMappingResponse xmlns:u="urn:schemas-wiimu-com:service:PlayQueue:1">
      <QueueContext>&lt;KeyList&gt;...&lt;/KeyList&gt;</QueueContext>
    </u:GetKeyMappingResponse>
  </s:Body>
</s:Envelope>
```

`QueueContext` содержит HTML-escaped XML со списком пресетов. Нужно два прохода декодирования: сначала распарсить SOAP, затем `html.UnescapeString` и снова распарсить XML.

**Структура `KeyList` (после unescaping):**

```xml
<KeyList>
  <ListName>Preset</ListName>
  <MaxNumber>6</MaxNumber>
  <Key0><!-- зарезервировано, не используется --></Key0>
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

| Поле | Описание |
|---|---|
| `Name` | Название станции (отображается на экране) |
| `Url` | URL аудиопотока |
| `PicUrl` | URL обложки (опционально) |
| `Source` | Тип источника; для интернет-радио — `newTuneIn` |
| `Metadata` | Дополнительные метаданные (обычно пусто) |

---

### SetKeyMapping — записать пресеты

**SoapAction:** `urn:schemas-wiimu-com:service:PlayQueue:1#SetKeyMapping`

**Запрос:**

```xml
<u:SetKeyMapping xmlns:u="urn:schemas-wiimu-com:service:PlayQueue:1">
  <QueueContext>&lt;KeyList&gt;...&lt;/KeyList&gt;</QueueContext>
</u:SetKeyMapping>
```

`QueueContext` — HTML-escaped XML полного `KeyList` (включая все 7 ключей Key0–Key6 и `MaxNumber=6`). Устройство заменяет весь список целиком.

---

## Примеры curl

```bash
# Информация об устройстве
curl "http://192.168.1.42/httpapi.asp?command=getStatusEx"

# Статус плеера
curl "http://192.168.1.42/httpapi.asp?command=getPlayerStatus"

# Запустить поток
curl "http://192.168.1.42/httpapi.asp?command=setPlayerCmd:play:http://stream.example.com/jazz"

# Остановить
curl "http://192.168.1.42/httpapi.asp?command=setPlayerCmd:stop"

# Получить пресеты (SOAP)
curl -X POST http://192.168.1.42:59152/upnp/control/PlayQueue1 \
  -H 'Content-Type: text/xml; charset=utf-8' \
  -H 'SoapAction: "urn:schemas-wiimu-com:service:PlayQueue:1#GetKeyMapping"' \
  -d '<?xml version="1.0" encoding="utf-8"?><s:Envelope s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/" xmlns:s="http://schemas.xmlsoap.org/soap/envelope/"><s:Body><u:GetKeyMapping xmlns:u="urn:schemas-wiimu-com:service:PlayQueue:1"></u:GetKeyMapping></s:Body></s:Envelope>'
```

---

## Особенности и ограничения

- **Timeout соединения** — 3 секунды (оба протокола).
- **Key0** — зарезервирован устройством, CLI использует только Key1–Key6.
- **Двойное XML-кодирование** — `QueueContext` всегда передаётся как HTML-escaped XML внутри XML. Это не ошибка, а особенность прошивки WiiMu.
- **Source=newTuneIn** — стандартное значение для интернет-радиостанций; другие значения прошивкой не документированы.
- **Порты**: HTTP API — 80, SOAP — 59152. Оба принимают только локальные подключения (устройство не открывает порты наружу).