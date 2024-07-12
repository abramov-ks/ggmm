package connection

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Connector struct {
	host    *string
	port    int
	timeout int
}

func NewConnector(host *string) *Connector {
	return &Connector{
		host:    host,
		port:    59152,
		timeout: 3,
	}
}

func (r Connector) Send(uri string, action string, request string) (string, error) {
	client := &http.Client{
		Timeout: time.Duration(r.timeout) * time.Second,
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/%s", *r.host, r.port, uri), strings.NewReader(request))
	if err != nil {
		return "", err
	}
	req.Header.Add("SoapAction", fmt.Sprintf("\"%s\"", action))
	req.Header.Add("Content-Type", "text/xml; charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)

	return string(body), nil
}
