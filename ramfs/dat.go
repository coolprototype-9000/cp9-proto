package ramfs

import (
	"fmt"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

type fidDesc struct {
	fid      nine.Fid
	owner    string
	open     bool
	openMode byte
}

type RamFs struct {
	fidTable  map[uint64]map[fidDesc]uint64
	devNumber uint32

	statTable map[uint64]nine.Stat
	dirTable  map[uint64][]uint64
	fileTable map[uint64][]byte

	// Root is by convention id 0
}

func (r *RamFs) DumpFs() {
	// LOLOL
	fmt.Println("--------------------")
	fmt.Printf("%v\n", *r)
	fmt.Println("--------------------")
}
