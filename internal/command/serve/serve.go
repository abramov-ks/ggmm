package serve

import (
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"ggmm/internal/dto"
	"ggmm/internal/ggmm/httpapi"
	"ggmm/internal/ggmm/service"
)

type CanConnect interface {
	Send(uri string, action string, request string) (string, error)
}

type Serve struct {
	connector CanConnect
	host      string
	tmpl      *template.Template
}

func New(connector CanConnect, host string) *Serve {
	return &Serve{
		connector: connector,
		host:      host,
		tmpl:      template.Must(template.New("").Parse(htmlTemplates)),
	}
}

type stationRow struct {
	N    int
	Name string
	Url  string
}

type deviceData struct {
	Name     string
	Firmware string
	Ssid     string
	IP       string
	RSSI     string
	Err      bool
}

type pageData struct {
	Host     string
	Device   deviceData
	Stations []stationRow
}

type statusData struct {
	PlayStatus string
	Title      string
	Artist     string
	Uri        string
	Vol        string
	Muted      bool
	Err        bool
}

func (s *Serve) rows(kl *dto.KeyList) []stationRow {
	keys := []dto.KeyData{kl.Key1, kl.Key2, kl.Key3, kl.Key4, kl.Key5, kl.Key6}
	rows := make([]stationRow, len(keys))
	for i, k := range keys {
		rows[i] = stationRow{
			N:    i + 1,
			Name: strings.TrimSpace(k.Name),
			Url:  strings.TrimSpace(k.Url),
		}
	}
	return rows
}

func keyURL(kl *dto.KeyList, n int) string {
	switch n {
	case 1:
		return strings.TrimSpace(kl.Key1.Url)
	case 2:
		return strings.TrimSpace(kl.Key2.Url)
	case 3:
		return strings.TrimSpace(kl.Key3.Url)
	case 4:
		return strings.TrimSpace(kl.Key4.Url)
	case 5:
		return strings.TrimSpace(kl.Key5.Url)
	case 6:
		return strings.TrimSpace(kl.Key6.Url)
	}
	return ""
}

func (s *Serve) deviceInfo() deviceData {
	info, err := httpapi.New(s.host).GetDeviceInfo()
	if err != nil {
		return deviceData{Err: true}
	}
	return deviceData{
		Name:     info.DeviceName,
		Firmware: info.Firmware,
		Ssid:     info.Ssid,
		IP:       info.IP,
		RSSI:     info.RSSI,
	}
}

func (s *Serve) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	svc := service.New(s.connector)
	kl, err := svc.GetList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.tmpl.ExecuteTemplate(w, "page", pageData{
		Host:     s.host,
		Device:   s.deviceInfo(),
		Stations: s.rows(kl),
	})
}

func (s *Serve) handleStatus(w http.ResponseWriter, r *http.Request) {
	client := httpapi.New(s.host)
	status, err := client.GetStatus()
	sd := statusData{Err: err != nil}
	if err == nil {
		sd.PlayStatus = status.Status
		sd.Title = status.Title
		sd.Artist = status.Artist
		sd.Uri = status.Uri
		sd.Vol = status.Vol
		sd.Muted = status.Mute == "1"
	}
	s.tmpl.ExecuteTemplate(w, "status", sd)
}

func (s *Serve) handleStation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nStr := strings.TrimPrefix(r.URL.Path, "/stations/")
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 || n > 6 {
		http.Error(w, "invalid station number", http.StatusBadRequest)
		return
	}

	r.ParseForm()
	name := r.FormValue("name")
	url := r.FormValue("url")

	svc := service.New(s.connector)
	kl, err := svc.GetList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	kl.Set(n, &dto.KeyData{Name: name, Url: url, Source: "newTuneIn"})

	if err := svc.SetStations(kl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	s.tmpl.ExecuteTemplate(w, "row", stationRow{N: n, Name: name, Url: url})
}

func (s *Serve) handlePlay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nStr := strings.TrimPrefix(r.URL.Path, "/play/")
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 || n > 6 {
		http.Error(w, "invalid preset", http.StatusBadRequest)
		return
	}

	svc := service.New(s.connector)
	kl, err := svc.GetList()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if url := keyURL(kl, n); url != "" {
		httpapi.New(s.host).Play(url)
		time.Sleep(time.Second)
	}

	s.handleStatus(w, r)
}

func (s *Serve) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	httpapi.New(s.host).Stop()
	time.Sleep(300 * time.Millisecond)
	s.handleStatus(w, r)
}

func (s *Serve) Handle() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/stations/", s.handleStation)
	mux.HandleFunc("/play/", s.handlePlay)
	mux.HandleFunc("/stop", s.handleStop)

	addr := ":8080"
	fmt.Printf("Server running at http://localhost%s\n", addr)
	go openBrowser("http://localhost" + addr)
	http.ListenAndServe(addr, mux)
}

func openBrowser(url string) {
	switch runtime.GOOS {
	case "darwin":
		exec.Command("open", url).Start()
	case "linux":
		exec.Command("xdg-open", url).Start()
	case "windows":
		exec.Command("cmd", "/c", "start", url).Start()
	}
}

const htmlTemplates = `
{{define "page"}}
<!DOCTYPE html>
<html lang="ru">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>GGMM Radio</title>
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
  <style>
    *, *::before, *::after { box-sizing: border-box; }
    body { font-family: system-ui, sans-serif; max-width: 800px; margin: 40px auto; padding: 0 20px; color: #1f2937; }
    h1 { font-size: 1.4rem; margin: 0; }
    h2 { font-size: .75rem; margin: 24px 0 10px; color: #9ca3af; text-transform: uppercase; letter-spacing: .08em; }
    .device-info { font-size: .8rem; color: #9ca3af; margin: 4px 0 0; }
    .device-info b { color: #6b7280; font-weight: 500; }
    #status { background: #f9fafb; border: 1px solid #e5e7eb; padding: 12px 16px; border-radius: 8px; margin: 16px 0; font-size: .9rem; min-height: 44px; line-height: 1.5; }
    .play  { color: #16a34a; font-weight: 600; }
    .stop, .pause, .load { color: #6b7280; }
    .err   { color: #dc2626; }
    .station { display: flex; align-items: center; gap: 8px; margin: 6px 0; }
    .num { width: 18px; text-align: right; color: #9ca3af; font-size: .85rem; flex-shrink: 0; }
    form { display: flex; gap: 8px; flex: 1; }
    input { padding: 7px 10px; border: 1px solid #d1d5db; border-radius: 6px; font-size: .9rem; font-family: inherit; }
    input[name="name"] { width: 180px; flex-shrink: 0; }
    input[name="url"]  { flex: 1; min-width: 0; font-family: monospace; font-size: .8rem; }
    button { padding: 7px 14px; background: #2563eb; color: #fff; border: none; border-radius: 6px; cursor: pointer; font-size: .9rem; flex-shrink: 0; }
    button:hover { background: #1d4ed8; }
    button[disabled] { background: #e5e7eb; color: #9ca3af; cursor: default; }
    .btn-play { background: #16a34a; padding: 7px 11px; }
    .btn-play:hover { background: #15803d; }
    .btn-stop { background: #dc2626; padding: 7px 12px; }
    .btn-stop:hover { background: #b91c1c; }
    button.htmx-request { opacity: .5; pointer-events: none; }
    .htmx-request button[type="submit"] { opacity: .5; pointer-events: none; }
  </style>
</head>
<body>
  <h1>GGMM Radio &middot; {{.Host}}</h1>
  {{if not .Device.Err}}
  <p class="device-info">
    <b>{{.Device.Name}}</b> &middot;
    прошивка {{.Device.Firmware}} &middot;
    {{.Device.Ssid}}{{if .Device.RSSI}} ({{.Device.RSSI}}&nbsp;dBm){{end}} &middot;
    {{.Device.IP}}
  </p>
  {{end}}
  <div id="status" hx-get="/status" hx-trigger="load, every 5s" hx-swap="outerHTML">
    Загрузка&hellip;
  </div>
  <h2>Пресеты</h2>
  {{range .Stations}}{{template "row" .}}{{end}}
</body>
</html>
{{end}}

{{define "row"}}
<div id="station-{{.N}}" class="station">
  <span class="num">{{.N}}</span>
  <form hx-put="/stations/{{.N}}" hx-target="#station-{{.N}}" hx-swap="outerHTML">
    <input name="name" value="{{.Name}}" placeholder="Название">
    <input name="url"  value="{{.Url}}"  placeholder="URL потока">
    {{if .Url}}
    <button type="button" class="btn-play"
            hx-post="/play/{{.N}}"
            hx-target="#status"
            hx-swap="outerHTML">&#9654;</button>
    {{else}}
    <button type="button" class="btn-play" disabled>&#9654;</button>
    {{end}}
    <button type="submit">Сохранить</button>
  </form>
</div>
{{end}}

{{define "status"}}
<div id="status" hx-get="/status" hx-trigger="every 5s" hx-swap="outerHTML" style="display:flex;align-items:center;gap:12px;">
  <div style="flex:1">
  {{if .Err}}
    <span class="err">&#9888; Нет связи с устройством</span>
  {{else}}
    <span class="{{.PlayStatus}}">{{.PlayStatus}}</span>
    {{if .Title}}
      &nbsp;&middot;&nbsp;<strong>{{.Title}}</strong>{{if .Artist}} &mdash; {{.Artist}}{{end}}
    {{else if .Uri}}
      &nbsp;&middot;&nbsp;<code style="font-size:.8rem">{{.Uri}}</code>
    {{end}}
    &nbsp;&middot;&nbsp;Vol:&nbsp;{{.Vol}}{{if .Muted}}&nbsp;&#128263;{{end}}
  {{end}}
  </div>
  {{if and (not .Err) (or (eq .PlayStatus "play") (eq .PlayStatus "load"))}}
  <button type="button" class="btn-stop"
          hx-post="/stop"
          hx-target="#status"
          hx-swap="outerHTML">&#9646;&#9646;</button>
  {{end}}
</div>
{{end}}
`
