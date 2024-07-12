package list

import (
	"fmt"
	"ggmm/internal/dto"
	"ggmm/internal/ggmm/service"
	"strings"
)

type CanConnect interface {
	Send(uri string, action string, request string) (string, error)
}

type List struct {
	connector CanConnect
}

func NewList(connector CanConnect) *List {
	return &List{connector: connector}
}

func printKeyData(keyData dto.KeyData) string {

	return fmt.Sprintf("%s (%s)", strings.Trim(keyData.Name, " \n"), strings.Trim(keyData.Url, " \n"))
}

func (l List) Handle() {

	fmt.Println("List stations")

	ggmmService := service.New(l.connector)
	keyList, err := ggmmService.GetList()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	fmt.Printf("1. %s\n", printKeyData(keyList.Key1))
	fmt.Printf("2. %s\n", printKeyData(keyList.Key2))
	fmt.Printf("3. %s\n", printKeyData(keyList.Key3))
	fmt.Printf("4. %s\n", printKeyData(keyList.Key4))
	fmt.Printf("5. %s\n", printKeyData(keyList.Key5))
	fmt.Printf("6. %s\n", printKeyData(keyList.Key6))
}
