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

	names := []string{"disks", "home", "home/rob", "home/ken", "disks/disk1", "disks/disk2", "/disks/disk1/bin"}
	for _, name := range names {
		fd := myproc.Create(name, nine.OREAD, nine.PUR|nine.PUW|nine.PUX|nine.PGR|nine.PGX|(nine.FDir<<nine.FStatOffset))
		if fd < 0 {
			log.Fatalf("failed to create file: %d\n", fd)
		}
		fd = myproc.Close(fd)
		if fd < 0 {
			log.Fatalf("failed to close file: %s\n", myproc.Errstr())
		}
	}

	fmt.Printf("Files created\n")
	myproc.Bind("/disks/disk1", "/home/rob", client.Replace)
	myproc.Bind("/disks/disk2", "/home/ken", client.Replace)

	if myproc.Chdir("home/rob/bin") < 0 {
		log.Fatalf("chdir failed")
	}
	fmt.Printf("Currently here: %v\n---------------------\n", *myproc.Stat("."))

	if myproc.Chdir("../../ken") < 0 {
		log.Fatalf("chdir failed")
	}
	fmt.Printf("Currently here: %v\n---------------------\n", *myproc.Stat("."))

	// Try to unmount, should fail
	if myproc.Unmount("/disks/disk1", "/home/rob") < 0 {
		fmt.Printf("Failed as we should NOT have: %s\n", myproc.Errstr())
	}
	if myproc.Unmount("/disks/disk1", "/home/ken") < 0 {
		fmt.Printf("Failed as we should NOT have: %s\n", myproc.Errstr())
	}
	log.Fatal("umount did not fail!!")
}
