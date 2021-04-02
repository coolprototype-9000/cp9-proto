package client

import (
	"fmt"
	"log"
	"net"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

var tag uint16 = 0

func mkTag() uint16 {
	tag++
	return tag - 1
}

func writeAndRead(c *net.Conn, f *nine.FCall) *nine.FCall {
	if err := nine.Write9P(c, f); err != nil {
		log.Fatal(err)
	}
	rf, err := nine.Read9P(c)
	if err != nil {
		log.Fatal(err)
	}
	return &rf
}

func checkMsg(f *nine.FCall, expected byte) {
	if f.MsgType == expected {
		fmt.Printf("Success! ")
	} else {
		fmt.Printf("FAILURE! ")
	}

	fmt.Printf("Got message type %d\n", f.MsgType)
	fmt.Printf("Full struct: %v\n", f)
}

// Configure consfs with an open connection
func setupConsFs(c *net.Conn) nine.Fid {
	root := nine.Fid(5)
	listen := nine.Fid(6)
	res := nine.Fid(7)
	fVersion(c, 0, nine.NineVersion)
	fAttach(c, root, "snoop_dogg")
	fWalk(c, root, listen, []string{"listen"})
	fOpen(c, listen, nine.OREAD)
	fRead(c, listen, 0, 20)
	fClunk(c, listen)
	fWalk(c, root, res, []string{"2"})
	fOpen(c, res, nine.ORDWR)
	return res
}
