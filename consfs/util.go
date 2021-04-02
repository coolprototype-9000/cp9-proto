package consfs

import (
	"time"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// Fid utilities
func (c *ConsFs) getFid(conId uint64, f nine.Fid) (fidDesc, bool) {
	ft := c.fidTable[conId]
	for fd := range ft {
		if fd.fid == f {
			return fd, true
		}
	}
	return fidDesc{}, false
}

func (c *ConsFs) isOpenByMe(conId uint64, f nine.Fid) bool {
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
func (c *ConsFs) isOpenByAnyone(id uint64) bool {
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

// Id utilities
func (c *ConsFs) idExists(id uint64) bool {
	if _, ok := c.statTable[id]; ok {
		return true
	}
	return false
}

func (c *ConsFs) updateATime(id uint64) {
	st := c.statTable[id]
	st.Atime = uint32(time.Now().Unix())
	c.statTable[id] = st
}

func (c *ConsFs) updateMTimeAs(user string, id uint64) {
	st := c.statTable[id]
	st.Mtime = uint32(time.Now().Unix())
	st.Atime = uint32(time.Now().Unix())
	st.Muid = user
	st.Q.Version++
	c.statTable[id] = st
}
