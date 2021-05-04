package user

import (
	"github.com/coolprototype-9000/cp9-proto/client"
	"github.com/coolprototype-9000/cp9-proto/nine"
)

var p *client.Proc

func Init(tp *client.Proc) {
	// Set up all thew ay
	p = tp

	// Configure initial filesystems
	baseFlags := nine.PUR | nine.PUW | nine.PUX | nine.PGR | nine.PGW | nine.PGX | nine.POR | nine.POW | nine.POX | (nine.FDir << nine.FStatOffset)
	p.Bind("#r", "/", client.Replace)

	fd := p.Create("/cons", nine.OREAD, uint32(baseFlags))
	p.Bind("#c", "/cons", client.Replace)
	p.Close(fd)

	fd = p.Create("/net", nine.OREAD, uint32(baseFlags))
	p.Bind("#n", "/net", client.Replace)
	p.Close(fd)

	Printf("3 filesystems created\nStarting sh..\n")
	sh()

}
