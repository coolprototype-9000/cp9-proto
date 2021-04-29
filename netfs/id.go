package netfs

import (
	"fmt"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

type ptrType byte

const (
	dir    ptrType = 0
	ctl    ptrType = 1
	data   ptrType = 2
	listen ptrType = 3
)

type netPtr struct {
	tp      ptrType
	version uint32
	ats     uint32
	mts     uint32
	muid    string
}

const rootId = 0
const cloneId = 1
const baseId = 4

var maxId uint64 = baseId

func genId() uint64 {
	maxId += 4
	return maxId - 4
}

func reduceId(id uint64) (uint64, ptrType) {
	tp := ptrType(id % 4)
	ptrid := id - uint64(tp)
	return ptrid, tp
}

func genQid(n *netPtr, parent *netInst) nine.Qid {
	var flags byte
	if n.tp == ctl || n.tp == data {
		flags |= nine.FAppend
	} else if n.tp == dir {
		flags |= nine.FDir
	}

	return nine.Qid{
		Flags:   flags,
		Version: n.version,
		Id:      parent.id + uint64(n.tp),
	}
}

func ptrTypeToString(p ptrType) string {
	switch p {
	case dir:
		return "dir"
	case ctl:
		return "ctl"
	case data:
		return "data"
	case listen:
		return "listen"
	default:
		return "ILLEGAL"
	}
}

func (c *NetFs) genStat(id uint64) nine.Stat {
	cid, tp := reduceId(id)
	//	fmt.Printf("\tWALK: trying to get to %d/%d", cid, tp)
	parent := c.cons[cid]
	//	fmt.Printf("\tWALK: PARENT: %v", parent)
	n := parent.children[tp]

	if id == rootId {
		return *c.rootstat
	} else if id == cloneId {
		return *c.clonestat
	}

	ns := nine.Stat{
		DevType: nine.DevNet,
		DevNo:   c.devNumber,
		Q:       genQid(n, parent),
		Mode:    nine.PUR | nine.PUW | nine.PGR | nine.PGW,
		Atime:   n.ats,
		Mtime:   n.mts,
		Uid:     parent.owner,
		Gid:     parent.owner,
		Muid:    n.muid,
	}

	if n.tp == dir {
		ns.Mode = nine.PUR | nine.PUX | nine.PGR | nine.PGX
		ns.Name = fmt.Sprintf("%d", parent.id)
	} else {
		ns.Name = ptrTypeToString(tp)
	}

	if n.tp == listen {
		ns.Mode = nine.PUR | nine.PGR
	}

	return ns
}

func (c *NetFs) gc(id uint64) {
	cid, _ := reduceId(id)
	tps := []ptrType{dir, ctl, data, listen}
	for tp := range tps {
		nid := cid + uint64(tp)
		if c.isRefByAnyone(nid) {
			return
		}
	}

	// Can GC, so do it
	ni := c.cons[cid]
	if ni.c != nil {
		(*ni.c).Close()
	}
	delete(c.cons, cid)
}
