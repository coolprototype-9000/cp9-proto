package main

import (
	"github.com/coolprototype-9000/cp9-proto/client"
	"github.com/coolprototype-9000/cp9-proto/consfs"
	"github.com/coolprototype-9000/cp9-proto/nine"
	"github.com/coolprototype-9000/cp9-proto/ramfs"
)

func main() {

	c := &consfs.ConsFs{}
	r := &ramfs.RamFs{}

	cConf := nine.MkConfig(c, 5640)
	rConf := nine.MkConfig(r, 5641)
	go nine.ServeForever(&cConf)
	go nine.ServeForever(&rConf)
	client.RunCli(&cConf, &rConf)
}
