package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/coolprototype-9000/cp9-proto/client"
	"github.com/coolprototype-9000/cp9-proto/consfs"
	"github.com/coolprototype-9000/cp9-proto/netfs"
	"github.com/coolprototype-9000/cp9-proto/nine"
	"github.com/coolprototype-9000/cp9-proto/ramfs"
	"github.com/coolprototype-9000/cp9-proto/user"
)

func main() {

	// Initialize the kernel
	c := &consfs.ConsFs{}
	r := &ramfs.RamFs{}
	n := &netfs.NetFs{}

	// Get bootargs
	fmt.Printf("CP9 loader -- 2021\n")
	fmt.Printf("Running mountless: all filesystems local\n")
uprompt:
	fmt.Printf("Log in as? ")
	rdr := bufio.NewReader(os.Stdin)
	uname, _ := rdr.ReadString('\n')
	uname = uname[:len(uname)-1]
	if uname == "" {
		goto uprompt
	}

	// Initialize all file servers
	cConf := nine.MkConfig(c, 5640)
	rConf := nine.MkConfig(r, 5641)
	nConf := nine.MkConfig(n, 5642)
	go nine.ServeForever(&cConf)
	go nine.ServeForever(&rConf)
	go nine.ServeForever(&nConf)

	// Create init and enter userland
	init := client.MkProc(nil, uname)
	user.Init(init)
}
