package main

import (
	"flag"
	list2 "ggmm/internal/command/info"
	"ggmm/internal/command/list"
	list3 "ggmm/internal/command/set"
	"ggmm/internal/ggmm/connection"
	"os"
)

func main() {
	host := flag.String("host", "127.0.0.1", "ggmm host")

	flag.Parse()

	if *host == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	tailArgs := flag.Args()

	if len(tailArgs) < 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	cmd := tailArgs[0]

	connector := connection.NewConnector(host)

	switch cmd {
	case "list":
		cmd := list.NewList(connector)
		cmd.Handle()
	case "info":
		cmd := list2.NewInfo(connector)
		cmd.Handle()
	case "set":
		cmd := list3.NewSet(connector)
		cmd.Handle(tailArgs[1:])
	}

}
