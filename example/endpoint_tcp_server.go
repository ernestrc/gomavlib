// +build ignore

package main

import (
	"fmt"

	"github.com/ernestrc/gomavlib"
	"github.com/ernestrc/gomavlib/dialects/ardupilotmega"
)

func main() {
	// create a node which
	// - communicates with a TCP endpoint in server mode
	// - understands ardupilotmega dialect
	// - writes messages with given system id
	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints: []gomavlib.EndpointConf{
			gomavlib.EndpointTcpServer{":5600"},
		},
		Dialect:     ardupilotmega.Dialect,
		OutSystemId: 10,
	})
	if err != nil {
		panic(err)
	}
	defer node.Close()

	for evt := range node.Events() {
		if frm, ok := evt.(*gomavlib.EventFrame); ok {
			fmt.Printf("received: id=%d, %+v\n", frm.Message().GetId(), frm.Message())
		}
	}
}
