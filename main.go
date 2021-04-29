package main

import (
	"fmt"
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

	myproc.Create("consfs", nine.OREAD, nine.PUR|nine.PUW|nine.PUX|nine.PGR|nine.PGX|(nine.FDir<<nine.FStatOffset))
	st = myproc.Create("testdir", nine.OREAD, nine.PUR|nine.PUW|nine.PUX|nine.PGR|nine.PGX|(nine.FDir<<nine.FStatOffset))

	if st < 0 {
		log.Fatalf("failure to create: %s", myproc.Errstr())
	} else {
		fmt.Printf("File %s created with fd %d\n", myproc.Fd2Path(st), st)
	}

	st = myproc.Bind("#c", "/consfs", client.Replace)
	if st < 0 {
		log.Fatalf("failure to bind consfs: %s", myproc.Errstr())
	}

	stat := myproc.Stat("/consfs/listen")
	if stat == nil {
		log.Fatalf("failure to get listen: %s", myproc.Errstr())
	}

	myproc.Create("testlisten", nine.ORDWR, nine.PUR|nine.PGR|nine.POR)
	st = myproc.Bind("/consfs/listen", "testlisten", client.Replace)
	if st < 0 {
		log.Fatalf("failure to bind file: %s", myproc.Errstr())
	}

	st = myproc.Chdir("consfs")
	if st < 0 {
		log.Fatalf("failure to chdir: %s", myproc.Errstr())
	}

	fmt.Printf("Changed directory!\n")

	stat = myproc.Stat("../testdir")
	if stat == nil {
		fmt.Printf("failure to get testdir: %s\n", myproc.Errstr())
	} else {
		fmt.Printf("WOA: %v\n", *stat)
	}

}
