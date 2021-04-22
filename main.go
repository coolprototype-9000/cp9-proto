package main

import (
	"github.com/coolprototype-9000/cp9-proto/consfs"
	"github.com/coolprototype-9000/cp9-proto/netfs"
	"github.com/coolprototype-9000/cp9-proto/nine"
)

func main() {

	c := &consfs.ConsFs{}
	n := &netfs.NetFs{}

	cConf := nine.MkConfig(c, 5640)
	rConf := nine.MkConfig(n, 5641)
	go nine.ServeForever(&cConf)
	go nine.ServeForever(&rConf)
	/*
		for {
			hardcoreclient.RunCli(&cConf, &rConf)
		}
	*/
}
