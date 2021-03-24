package ramfs

import (
	"errors"
	"math/rand"
	"time"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// Fid utilities
func (r *RamFs) getFid(conId uint64, f nine.Fid) (fidDesc, bool) {
	ft := r.fidTable[conId]
	for fd := range ft {
		if fd.fid == f {
			return fd, true
		}
	}
	return fidDesc{}, false
}

func (r *RamFs) isOpenByMe(conId uint64, f nine.Fid) bool {
	ft := r.fidTable[conId]
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
func (r *RamFs) isOpenByAnyone(id uint64) bool {
	for _, ft := range r.fidTable {
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

// Qid utilities

// Get the parent directory of a node. Not permission checked.
func (r *RamFs) getParent(startID uint64) (uint64, error) {
	if startID == 0 {
		return 0, errors.New("root directory has no parent")
	}
	for p, c := range r.dirTable {
		for _, child := range c {
			if child == startID {
				return p, nil
			}
		}
	}
	return 0, errors.New("very bad error condition: startID not in any directory?")
}

// Descend once. startID must correspond to a directory.
// Returns id corresponding to target, or error
func (r *RamFs) descend(startID uint64, user string, target string) (uint64, error) {
	// Need permission to execute directory
	if !r.checkPerms(startID, user)[2] {
		return 0, errors.New("execute permission denied")
	}

	// Parent directory?
	if target == ".." {
		parent, err := r.getParent(startID)
		return parent, err
	}

	// Usual case
	ids := r.dirTable[startID]
	for _, nid := range ids {
		name := r.statTable[nid].Name
		if name == target {
			return nid, nil
		}
	}
	return 0, errors.New("target of descend not found")
}

// Helper, returns RWX in a triple of booleans
// Simplifying assumption: users are their own groups
func (r *RamFs) checkPerms(id uint64, user string) []bool {
	res := make([]bool, 3)
	st := r.statTable[id]
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

func (r *RamFs) idExists(id uint64) bool {
	if _, ok := r.statTable[id]; ok {
		return true
	}
	return false
}

func (r *RamFs) genId() uint64 {
	var nid uint64
generate:
	nid = rand.Uint64()
	if r.idExists(nid) {
		goto generate
	}
	return nid
}

func (r *RamFs) idIsDir(id uint64) bool {
	if _, ok := r.dirTable[id]; ok {
		return true
	}
	return false
}

// Time generation
func mkTimeStamp() uint32 {
	return uint32(time.Now().Unix())
}

func (r *RamFs) updateATime(id uint64) {
	st := r.statTable[id]
	st.Atime = mkTimeStamp()
	r.statTable[id] = st
}

func (r *RamFs) updateMTimeAs(user string, id uint64) {
	st := r.statTable[id]
	st.Mtime = mkTimeStamp()
	st.Atime = mkTimeStamp()
	st.Muid = user
	st.Q.Version++
	r.statTable[id] = st
}
