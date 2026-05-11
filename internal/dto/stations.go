package dto

import (
	"fmt"
	"reflect"
)

type KeyData struct {
	Metadata string `xml:"Metadata,omitempty"`
	Name     string `xml:"Name,omitempty"`
	PicUrl   string `xml:"PicUrl,omitempty"`
	Source   string `xml:"Source,omitempty"`
	Url      string `xml:"Url,omitempty"`
}

type KeyList struct {
	ListName  string  `xml:"ListName"`
	MaxNumber int     `xml:"MaxNumber"`
	Key0      KeyData `xml:"Key0"`
	Key1      KeyData `xml:"Key1"`
	Key2      KeyData `xml:"Key2"`
	Key3      KeyData `xml:"Key3"`
	Key4      KeyData `xml:"Key4"`
	Key5      KeyData `xml:"Key5"`
	Key6      KeyData `xml:"Key6"`
}

func (receiver *KeyList) Set(i int, data *KeyData) {
	reflect.ValueOf(receiver).Elem().FieldByName(fmt.Sprintf("Key%d", i)).FieldByName("Name").Set(reflect.ValueOf(data.Name))
	reflect.ValueOf(receiver).Elem().FieldByName(fmt.Sprintf("Key%d", i)).FieldByName("PicUrl").Set(reflect.ValueOf(data.PicUrl))
	reflect.ValueOf(receiver).Elem().FieldByName(fmt.Sprintf("Key%d", i)).FieldByName("Source").Set(reflect.ValueOf(data.Source))
	reflect.ValueOf(receiver).Elem().FieldByName(fmt.Sprintf("Key%d", i)).FieldByName("Url").Set(reflect.ValueOf(data.Url))
}
