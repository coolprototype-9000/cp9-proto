package client

import (
	"errors"
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

func checkMsg(f *nine.FCall, expected byte) error {
	var err error
	if f.MsgType == expected {
		fmt.Printf("Success! ")
		err = nil
	} else {
		fmt.Printf("FAILURE! ")
		err = errors.New("incorrect call gotten back, presuming failure")
	}

	fmt.Printf("Got message type %d\n", f.MsgType)
	fmt.Printf("Full struct: %v\n", f)
	return err
}
