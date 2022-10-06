package main

import (
	"fmt"
	"github.com/mhazley/mblue-toolz/btmgmt"
)

func main() {
	// Try to open a connection to mgmt socket
	mgmt, err := btmgmt.NewBtMgmt()
	if err != nil {
		panic(err)
	}

	cntrlInf, err := mgmt.ReadControllerInformation(0)
	if err != nil {
		panic(err)
	}

	addr := cntrlInf.Address.Addr.String()
	fmt.Println(addr)

	cnList, err := mgmt.ReadConnectionList(0)
	fmt.Println(cnList.String())
}
