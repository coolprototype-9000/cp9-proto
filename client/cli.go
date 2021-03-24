package client

import (
	"fmt"
	"log"
	"net"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// RunClient spins a very simple
// CLI, useful for debugging the 9P
// server. It assumes a port of 5640, and
// that files are being served from localhost.
func RunClient(c *nine.Conf) {
	address := fmt.Sprintf("localhost:%d", c.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Failed to dial local server:", err)
	}

	tf := nine.FCall{
		MsgType: nine.TVersion,
		Tag:     5,
		Version: (*c).Version,
	}
	rf := writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RVersion)

	// Make me a new fid!
	var rootFid nine.Fid = 10
	tf = nine.FCall{
		MsgType: nine.TAttach,
		Tag:     6,
		F:       rootFid,
		Uname:   "jaytlang",
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RAttach)

	// Dup it
	var anotherRootFid nine.Fid = 11
	tf = nine.FCall{
		MsgType: nine.TWalk,
		Tag:     9,
		F:       rootFid,
		Newf:    anotherRootFid,
		Wname:   make([]string, 0),
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RWalk)

	// Make me another one
	tf = nine.FCall{
		MsgType: nine.TCreate,
		Tag:     7,
		F:       anotherRootFid,
		Name:    "snoop",
		Perm:    (nine.FDir << nine.FStatOffset) | nine.PUR | nine.PUW | nine.PUX | nine.PGR | nine.PGX | nine.POR | nine.POX,
		Mode:    nine.OREAD,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RCreate)

	// Clunk this one
	tf = nine.FCall{
		MsgType: nine.TClunk,
		Tag:     14,
		F:       anotherRootFid,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RClunk)

	// Walk to the new directory
	var dirFid nine.Fid = 420
	tf = nine.FCall{
		MsgType: nine.TWalk,
		Tag:     15,
		F:       rootFid,
		Newf:    dirFid,
		Wname:   []string{"snoop"},
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RWalk)

	// Stat it
	tf = nine.FCall{
		MsgType: nine.TStat,
		Tag:     20,
		F:       dirFid,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RStat)

	// Make a file in it
	tf = nine.FCall{
		MsgType: nine.TCreate,
		Tag:     70,
		F:       dirFid,
		Name:    "snoop.txt",
		Perm:    (nine.FAppend << nine.FStatOffset) | nine.PUR | nine.PUW | nine.PGR | nine.PGW,
		Mode:    nine.ORDWR,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RCreate)

	// Write to it
	tf = nine.FCall{
		MsgType: nine.TWrite,
		Tag:     25,
		F:       dirFid,
		Offset:  5,
		Data:    []byte("SMOKE WEED ERY DAY"),
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RWrite)

	// Stat the new file that has a changed length
	tf = nine.FCall{
		MsgType: nine.TStat,
		Tag:     21,
		F:       dirFid,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RStat)

	// Read 30 bytes from our file
	tf = nine.FCall{
		MsgType: nine.TRead,
		Tag:     230,
		F:       dirFid,
		Offset:  0,
		Count:   30,
	}
	rf = writeAndRead(&conn, &tf)
	checkMsg(rf, nine.RRead)

	conn.Close()
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
