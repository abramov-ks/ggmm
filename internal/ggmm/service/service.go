package service

import (
	"encoding/xml"
	"fmt"
	"ggmm/internal/dto"
	"html"
)

const infoUri = "upnp/control/PlayQueue1"
const requestWrapper = "<?xml version=\"1.0\" encoding=\"utf-8\"?><s:Envelope s:encodingStyle=\"http://schemas.xmlsoap.org/soap/encoding/\" xmlns:s=\"http://schemas.xmlsoap.org/soap/envelope/\"><s:Body>%s</s:Body></s:Envelope>"

type Response struct {
	Body struct {
		GetKeyMappingResponse struct {
			U            string `xml:"u,attr"`
			QueueContext string `xml:"QueueContext"`
		} `xml:"GetKeyMappingResponse"`
	} `xml:"Body"`
}

type CanConnect interface {
	Send(uri string, action string, request string) (string, error)
}

type Service struct {
	connector CanConnect
}

func New(connector CanConnect) *Service {
	return &Service{connector: connector}
}

func (l Service) GetList() (*dto.KeyList, error) {
	action := "urn:schemas-wiimu-com:service:PlayQueue:1#GetKeyMapping"
	request := fmt.Sprintf(requestWrapper, "<u:GetKeyMapping xmlns:u=\"urn:schemas-wiimu-com:service:PlayQueue:1\"></u:GetKeyMapping>")
	resp, err := l.connector.Send(infoUri, action, request)

	if err != nil {
		return nil, fmt.Errorf("send command error: %s", err)
	}

	data := &Response{}
	err = xml.Unmarshal([]byte(resp), &data)
	if err != nil {
		return nil, fmt.Errorf("xml decode error")
	}

	keyList := &dto.KeyList{}
	err = xml.Unmarshal([]byte(html.UnescapeString(data.Body.GetKeyMappingResponse.QueueContext)), &keyList)
	if err != nil {
		return nil, fmt.Errorf("xml decode error")
	}

	return keyList, nil
}

func convertKeyListToXml(list *dto.KeyList) (string, error) {
	xmlData, err := xml.Marshal(list)
	if err != nil {
		return "", err
	}

	return xml.Header + string(xmlData), nil
}

func (l Service) SetStations(list *dto.KeyList) error {
	action := "urn:schemas-wiimu-com:service:PlayQueue:1#SetKeyMapping"
	keyListXml, err := convertKeyListToXml(list)
	if err != nil {
		return err
	}
	request := fmt.Sprintf(requestWrapper, "<u:SetKeyMapping xmlns:u=\"urn:schemas-wiimu-com:service:PlayQueue:1\"><QueueContext>"+html.EscapeString(keyListXml)+"</QueueContext></u:SetKeyMapping>")
	resp, err := l.connector.Send(infoUri, action, request)

	if err != nil {
		return fmt.Errorf("send command error: %s", err)
	}

	fmt.Sprintf("%s", resp)
	return nil
}

//
//func (l Service) GetStations() *KeyList {
//
//}
