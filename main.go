package main

import (
	"log"

	"github.com/coolprototype-9000/cp9-proto/client"
	"github.com/coolprototype-9000/cp9-proto/consfs"
	"github.com/coolprototype-9000/cp9-proto/netfs"
	"github.com/coolprototype-9000/cp9-proto/nine"
	"github.com/coolprototype-9000/cp9-proto/ramfs"
)

func main() {

	c := &consfs.ConsFs{}
	r := &ramfs.RamFs{}
	n := &netfs.NetFs{}

	cConf := nine.MkConfig(c, 5640)
	rConf := nine.MkConfig(r, 5641)
	nConf := nine.MkConfig(n, 5642)
	go nine.ServeForever(&cConf)
	go nine.ServeForever(&rConf)
	go nine.ServeForever(&nConf)

	// Let's try things!
	myproc := client.MkProc(nil, "jaytlang")
	st := myproc.Bind("#r", "/", client.Replace)
	if st < 0 {
		log.Fatalf("failure to bind ramfs: %s", myproc.Errstr())
	}

	st = myproc.Bind("#c", "/", client.After)
	if st < 0 {
		log.Fatalf("failure to bind consfs: %s", myproc.Errstr())
	}

	st = myproc.Create("testf", nine.ORDWR, nine.PUR|nine.PUW|nine.PGR|nine.PGW)
	if st < 0 {
		log.Fatalf("failure to create: %s", myproc.Errstr())
	}

}
