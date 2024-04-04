package tool

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
