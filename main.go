package main

import (
	"github.com/coolprototype-9000/cp9-proto/client"
	"github.com/coolprototype-9000/cp9-proto/nine"
	"github.com/coolprototype-9000/cp9-proto/ramfs"
)

func main() {
	r := &ramfs.RamFs{}
	c := nine.MkConfig(r)
	go nine.ServeForever(&c)
	client.RunClient(&c)
}
