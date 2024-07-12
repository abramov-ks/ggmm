package list

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
)

const infoUri = "upnp/control/rendercontrol1"

type CanConnect interface {
	Send(uri string, action string, request string) (string, error)
}

type Info struct {
	connector CanConnect
}

type Response struct {
	Body struct {
		GetControlDeviceInfoResponse struct {
			U              string `xml:"u,attr"`
			CurrentChannel int    `xml:"CurrentChannel"`
			CurrentMute    int    `xml:"CurrentMute"`
			CurrentVolume  int    `xml:"CurrentVolume"`
			MultiType      int    `xml:"MultiType"`
			Router         string `xml:"Router"`
			SlaveList      string `xml:"SlaveList"`
			SlaveMask      int    `xml:"SlaveMask"`
			Ssid           string `xml:"Ssid"`
			Status         string `xml:"Status"`
		} `xml:"GetControlDeviceInfoResponse"`
	} `xml:"Body"`
}

func NewInfo(connector CanConnect) *Info {
	return &Info{connector: connector}
}

func (l Info) Handle() {

	fmt.Println("Info about device")
	action := "urn:schemas-upnp-org:service:RenderingControl:1#GetControlDeviceInfo"
	request := "<?xml version=\"1.0\" encoding=\"utf-8\"?><s:Envelope s:encodingStyle=\"http://schemas.xmlsoap.org/soap/encoding/\" xmlns:s=\"http://schemas.xmlsoap.org/soap/envelope/\"><s:Body><u:GetControlDeviceInfo xmlns:u=\"urn:schemas-upnp-org:service:RenderingControl:1\"><InstanceID>0</InstanceID></u:GetControlDeviceInfo></s:Body></s:Envelope>"
	resp, err := l.connector.Send(infoUri, action, request)
	if err != nil {
		fmt.Printf("Send command error: %s", err)
		return
	}

	data := &Response{}
	err = xml.Unmarshal([]byte(resp), &data)
	if err != nil {
		fmt.Printf("Xml decode error")
		return
	}

	if len(data.Body.GetControlDeviceInfoResponse.Status) < 1 {
		fmt.Println("Empty info data")
	}

	infoPairs := make(map[string]interface{})
	err = json.Unmarshal([]byte(data.Body.GetControlDeviceInfoResponse.Status), &infoPairs)

	if err != nil {
		fmt.Println("wrong data json")
		return
	}

	for k, v := range infoPairs {
		fmt.Printf("%s: %s\n", k, v)
	}
}
