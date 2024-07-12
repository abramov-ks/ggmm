package list

import (
	"fmt"
	"ggmm/internal/dto"
	"ggmm/internal/ggmm/service"
	"strconv"
	"strings"
)

const infoUri = "upnp/control/PlayQueue1"

type CanConnect interface {
	Send(uri string, action string, request string) (string, error)
}

type Response struct {
	Body struct {
		GetKeyMappingResponse struct {
			U            string `xml:"u,attr"`
			QueueContext string `xml:"QueueContext"`
		} `xml:"GetKeyMappingResponse"`
	} `xml:"Body"`
}

type Set struct {
	connector CanConnect
}

func NewSet(connector CanConnect) *Set {
	return &Set{connector: connector}
}

func printKeyData(keyData dto.KeyData) string {

	return fmt.Sprintf("%s (%s)", strings.Trim(keyData.Name, " \n"), strings.Trim(keyData.Url, " \n"))
}

func (l Set) Handle(params []string) {

	fmt.Printf("Set station %s %s\n", params[0], params[1])

	stationNo, err := strconv.Atoi(params[0])
	if err != nil {
		fmt.Printf("Wrong station number")
		return

	}

	ggmmService := service.New(l.connector)
	keyList, err := ggmmService.GetList()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	keyList.Set(stationNo, &dto.KeyData{
		Name:   params[1],
		Url:    params[2],
		Source: "newTuneIn",
	})

	ggmmService.SetStations(keyList)
}
