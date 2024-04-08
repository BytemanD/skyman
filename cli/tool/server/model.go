package server

import "fmt"

type Interface struct {
	NetId  string
	PortId string
	Name   string
}

func (i Interface) String() string {
	if i.PortId != "" {
		return fmt.Sprintf("port: %s", i.PortId)
	} else {
		return fmt.Sprintf("net: %s", i.NetId)
	}
}

type Volume struct {
	Id   string
	Name string
}

func (i Volume) String() string {
	if i.Name != "" {
		return fmt.Sprintf("id: %s", i.Name)
	} else {
		return fmt.Sprintf("id: %s", i.Id)
	}
}
