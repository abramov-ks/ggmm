package httpapi

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type rawStatus struct {
	Status string `json:"status"`
	Title  string `json:"Title"`
	Artist string `json:"Artist"`
	Vol    string `json:"vol"`
	Mute   string `json:"mute"`
	Uri    string `json:"uri"`
}

type Status struct {
	Status string
	Title  string
	Artist string
	Vol    string
	Mute   string
	Uri    string
}

func decodeHex(s string) string {
	if len(s)%2 != 0 {
		return s
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return s
	}
	return string(b)
}

type Client struct {
	host   string
	client *http.Client
}

func New(host string) *Client {
	return &Client{
		host:   host,
		client: &http.Client{Timeout: 3 * time.Second},
	}
}

func (c *Client) command(cmd string) ([]byte, error) {
	resp, err := c.client.Get(fmt.Sprintf("http://%s/httpapi.asp?command=%s", c.host, cmd))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (c *Client) Play(streamURL string) error {
	_, err := c.command("setPlayerCmd:play:" + streamURL)
	return err
}

func (c *Client) Stop() error {
	_, err := c.command("setPlayerCmd:stop")
	return err
}

type DeviceInfo struct {
	DeviceName string
	Firmware   string
	Ssid       string
	IP         string
	RSSI       string
}

func (c *Client) GetDeviceInfo() (*DeviceInfo, error) {
	body, err := c.command("getStatusEx")
	if err != nil {
		return nil, err
	}
	raw := struct {
		DeviceName string `json:"DeviceName"`
		Firmware   string `json:"firmware"`
		Essid      string `json:"essid"`
		Apcli0     string `json:"apcli0"`
		RSSI       string `json:"RSSI"`
	}{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return &DeviceInfo{
		DeviceName: raw.DeviceName,
		Firmware:   raw.Firmware,
		Ssid:       decodeHex(raw.Essid),
		IP:         raw.Apcli0,
		RSSI:       raw.RSSI,
	}, nil
}

func (c *Client) GetStatus() (*Status, error) {
	body, err := c.command("getPlayerStatus")
	if err != nil {
		return nil, err
	}
	raw := &rawStatus{}
	if err := json.Unmarshal(body, raw); err != nil {
		return nil, err
	}
	return &Status{
		Status: raw.Status,
		Title:  decodeHex(raw.Title),
		Artist: decodeHex(raw.Artist),
		Vol:    raw.Vol,
		Mute:   raw.Mute,
		Uri:    decodeHex(raw.Uri),
	}, nil
}
