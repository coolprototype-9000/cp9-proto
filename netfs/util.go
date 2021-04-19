package netfs

import (
	"errors"
	"strconv"
	"time"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// Fid utilities
func (c *NetFs) getFid(conId uint64, f nine.Fid) (fidDesc, bool) {
	ft := c.fidTable[conId]
	for fd := range ft {
		if fd.fid == f {
			return fd, true
		}
	}
	return fidDesc{}, false
}

func (c *NetFs) isOpenByMe(conId uint64, f nine.Fid) bool {
	ft := c.fidTable[conId]
	for fd := range ft {
		if fd.fid == f && fd.open {
			return true
		} else if fd.fid == f {
			return false
		}
	}
	return false
}

// big slow
func (c *NetFs) isOpenByAnyone(id uint64) bool {
	for _, ft := range c.fidTable {
		for fd, thisID := range ft {
			if fd.open && thisID == id {
				return true
			} else if thisID == id {
				return false
			}
		}
	}
	return false
}

func (c *NetFs) idIsDir(id uint64) bool {
	_, tp := reduceId(id)
	return tp == dir
}

// Id utilities
func (c *NetFs) idExists(id uint64) bool {
	if id == rootId {
		return c.rootstat != nil
	} else if id == cloneId {
		return c.clonestat != nil
	} else if id < baseId {
		return false
	}
	cid, _ := reduceId(id)
	if _, ok := c.cons[cid]; ok {
		return true
	}
	return false
}

func (c *NetFs) updateATime(id uint64) {
	cid, tp := reduceId(id)
	nf := c.cons[cid].children[tp]
	nf.ats = uint32(time.Now().Unix())
}

func (c *NetFs) updateMTimeAs(user string, id uint64) {
	cid, tp := reduceId(id)
	nf := c.cons[cid].children[tp]
	nf.mts = uint32(time.Now().Unix())
	nf.ats = nf.mts
	nf.muid = user
	nf.version++
}

// Descend once
func (c *NetFs) descend(startID uint64, user string, target string) (uint64, error) {
	cid, tp := reduceId(startID)
	if startID != rootId {
		if c.cons[cid].owner != user {
			return 0, errors.New("execute permission denied")
		}
	}

	if target == ".." {
		if startID == rootId {
			return 0, errors.New("no parent of root directory exists")
		} else if tp == dir {
			return rootId, nil
		} else {
			return cid, nil
		}
	}

	if startID == rootId {
		nid, err := strconv.ParseUint(target, 10, 64)
		if err != nil {
			goto failwalk
		} else if _, ok := c.cons[nid]; !ok {
			goto failwalk
		}
		return nid, nil
	} else {
		// Descend always gets called from a valid id
		switch target {
		case "ctl":
			return cid + uint64(ctl), nil
		case "data":
			return cid + uint64(data), nil
		case "listen":
			return cid + uint64(listen), nil
		default:
			goto failwalk
		}
	}
failwalk:
	return 0, errors.New("target of descend not found")
}

func (r *NetFs) checkPerms(id uint64, user string) []bool {
	res := make([]bool, 3)
	st := r.genStat(id)
	m := st.Mode

	if st.Uid == user {
		res[0] = (m & nine.PUR) > 0
		res[1] = (m & nine.PUW) > 0
		res[2] = (m & nine.PUX) > 0
	}
	if st.Gid == user {
		res[0] = res[0] || (m&nine.PGR) > 0
		res[1] = res[1] || (m&nine.PGW) > 0
		res[2] = res[2] || (m&nine.PGX) > 0
	}

	res[0] = res[0] || (m&nine.POR) > 0
	res[1] = res[1] || (m&nine.POW) > 0
	res[2] = res[2] || (m&nine.POX) > 0
	return res
}
