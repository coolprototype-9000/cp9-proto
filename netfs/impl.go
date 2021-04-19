package netfs

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

func (c *NetFs) Register(devNo uint32) error {
	c.devNumber = devNo
	c.fidTable = make(map[uint64]map[fidDesc]uint64)
	c.cons = make(map[uint64]*netInst)
	return nil
}

func (c *NetFs) Attach(conId uint64, f nine.Fid, uname string) (nine.Qid, error) {
	if _, ok := c.getFid(conId, f); ok {
		return nine.Qid{}, errors.New("fid already in use for session")
	}

	if !c.idExists(rootId) {
		c.rootstat = &nine.Stat{
			DevType: nine.DevNet,
			DevNo:   c.devNumber,
			Q:       nine.Qid{Flags: nine.FDir, Id: rootId},
			Mode:    nine.PUR | nine.PUW | nine.PUX | nine.PGR | nine.PGX | nine.POR | nine.POX,
			Atime:   uint32(time.Now().Unix()),
			Mtime:   uint32(time.Now().Unix()),
			Name:    "/",
			Uid:     "None", // yay anarchy
			Gid:     "None",
			Muid:    "None",
		}

		c.clonestat = &nine.Stat{
			DevType: nine.DevNet,
			DevNo:   c.devNumber,
			Q:       nine.Qid{Id: cloneId},
			Mode:    nine.PUR | nine.PGR | nine.POR,
			Atime:   uint32(time.Now().Unix()),
			Mtime:   uint32(time.Now().Unix()),
			Name:    "clone",
			Uid:     "None", // yay anarchy
			Gid:     "None",
			Muid:    "None",
		}
	}

	fd := fidDesc{
		fid:   f,
		owner: uname,
	}

	if c.fidTable[conId] == nil {
		c.fidTable[conId] = make(map[fidDesc]uint64)
	}
	c.fidTable[conId][fd] = rootId
	return c.rootstat.Q, nil
}

func (c *NetFs) Walk(conId uint64, f nine.Fid, nf nine.Fid, wname []string) ([]nine.Qid, error) {
	if _, ok := c.getFid(conId, f); !ok {
		return []nine.Qid{}, errors.New("fid not in use for session")
	}

	if _, ok := c.getFid(conId, nf); ok {
		return []nine.Qid{}, errors.New("nfid already in use for session")
	}

	fd, _ := c.getFid(conId, f)
	nfd := fidDesc{
		fid:   nf,
		owner: fd.owner,
	}

	id := c.fidTable[conId][fd]
	_, tp := reduceId(id)
	if len(wname) != 0 && id != rootId && tp != dir {
		return []nine.Qid{}, errors.New("starting fid is not a directory")
	}

	if c.isOpenByMe(conId, f) {
		return []nine.Qid{}, errors.New("fid is open for i/o so can't walk")
	}

	c.updateATime(id)

	if len(wname) == 0 {
		c.fidTable[conId][nfd] = id
		return []nine.Qid{}, nil
	}

	wqid := make([]nine.Qid, 0)
	for i, wn := range wname {
		c.updateATime(id)
		if !c.idIsDir(id) {
			return wqid, nil
		}
		var err error
		id, err = c.descend(id, fd.owner, wn)
		if err != nil {
			if i == 0 {
				return make([]nine.Qid, 0), err
			} else {
				return wqid, nil
			}
		}
		if id == rootId {
			wqid = append(wqid, c.rootstat.Q)
		} else if id == cloneId {
			wqid = append(wqid, c.clonestat.Q)
		} else {
			wqid = append(wqid, c.genStat(id).Q)
		}
	}

	c.fidTable[conId][nfd] = id
	return wqid, nil
}

func (c *NetFs) Create(conId uint64, f nine.Fid, name string, perm uint32, mode byte) (nine.Qid, error) {
	return nine.Qid{}, errors.New("permission denied")
}

func (c *NetFs) Open(conId uint64, f nine.Fid, mode byte) (nine.Qid, error) {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return nine.Qid{}, errors.New("fid requested for open is not in use")
	}

	id := c.fidTable[conId][fd]
	c.updateATime(id)

	// Basic checks
	if fd.open {
		return nine.Qid{}, errors.New("fid is currently open, can't double open")
	} else if mode&nine.OTRUNC > 0 {
		return nine.Qid{}, errors.New("cannot trunc this file")
	} else if mode&nine.ORCLOSE > 0 {
		return nine.Qid{}, errors.New("cannot rclose this file")
	}

	// Make initial permission validation check
	fperm := c.checkPerms(id, fd.owner)
	fmt.Printf("!!!!!!!!!! HELLO PRERMISSIONS ARE %v\n", fperm)
	openTypeBits := mode & 0b11

	if !fperm[0] && (openTypeBits == nine.OREAD || openTypeBits == nine.ORDWR) {
		return nine.Qid{}, errors.New("permission denied, no Read access given")
	} else if !fperm[1] && (openTypeBits == nine.OWRITE || openTypeBits == nine.ORDWR) {
		return nine.Qid{}, errors.New("permission denied, no Write access given")
	} else if !fperm[2] && (openTypeBits == nine.OEXEC) {
		return nine.Qid{}, errors.New("permission denied, no exec access given")
	}

	if c.idIsDir(id) {
		if openTypeBits == nine.OWRITE || openTypeBits == nine.ORDWR {
			return nine.Qid{}, errors.New("it is illegal to Write a directory")
		}
	}

	delete(c.fidTable[conId], fd)
	fd.open = true
	fd.openMode = mode
	c.fidTable[conId][fd] = id

	if id == rootId {
		return c.rootstat.Q, nil
	} else if id == cloneId {
		return c.clonestat.Q, nil
	}
	return c.genStat(id).Q, nil
}

func (c *NetFs) Read(conId uint64, f nine.Fid, offset uint64, count uint32) ([]byte, error) {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return []byte{}, errors.New("fid requested for read not in use")
	}

	openTypeBits := fd.openMode & 0b11
	if !fd.open || (openTypeBits == nine.OWRITE || openTypeBits == nine.OEXEC) {
		return []byte{}, errors.New("fid not opened for reading")
	}

	id := c.fidTable[conId][fd]
	c.updateATime(id)
	if id == rootId {
		if offset != 0 {
			return []byte{}, errors.New("directory seeking is illegal")
		}

		scnt := int(count)
		scnt -= len(nine.MarshalStat(*c.clonestat))
		if scnt < 0 {
			return []byte{}, nil
		}
		ret := nine.MarshalStat(*c.clonestat)
		for id := range c.cons {
			sb := nine.MarshalStat(c.genStat(id))
			penalty := len(sb)

			scnt -= penalty
			if scnt < 0 {
				break
			}
			ret = append(ret, sb...)
		}

		fmt.Printf("_____________READ_____________\n")
		return ret, nil

	} else if id == cloneId {
		if count == 0 {
			return []byte{}, nil
		}

		// Make a new connection, but don't
		// do anything with it yet. It is idle.
		id := genId()
		c.cons[id] = mkEmptyNetInst(fd.owner, id)

		// Creation is complete. Now read the new
		// line dir back at them
		ba := []byte(strconv.Itoa(int(id)))
		if uint32(len(ba)) > count {
			ba = ba[:count+1]
		}
		return ba, nil
	}

	// Some other id
	cid, tp := reduceId(id)
	ni := c.cons[cid]

	switch tp {
	case dir:
		scnt := int(count)
		tps := []ptrType{ctl, data, listen}
		ret := []byte{}

		for _, tp := range tps {
			sb := nine.MarshalStat(c.genStat(cid + uint64(tp)))
			penalty := len(sb)

			scnt -= penalty
			if scnt < 0 {
				break
			}
			ret = append(ret, sb...)
		}
		return ret, nil
	case ctl:
		// Ctl is never read
		// and doing so helps nobody
		if ni.checkDead() {
			err := []byte("error: dead connection")
			if uint32(len(err)) > count {
				err = err[:count]
			}
			return err, nil
		}
		return []byte{}, nil
	case data:
		// If the state is connecting, open
		// the connection for the first time.
		if ni.checkDead() {
			return []byte("error: dead connection")[:count], nil
		} else if ni.s == idle {
			err := ni.openCon()
			if err != nil {
				err := []byte(err.Error())
				if uint32(len(err)) > count {
					err = err[:count]
				}
				return err, nil
			}
		}

		if ni.s == connected {
			// Reconciling this is ur problem
			rb := make([]byte, count)
			_, err := (*ni.c).Read(rb)
			if err != nil {
				ni.s = dead
				err := []byte(err.Error())
				if uint32(len(err)) > count {
					err = err[:count]
				}
				return err, nil
			}
			return rb, nil
		} else {
			// If you are idle or listening, this makes
			// approximately zero sense to read
			err := []byte("error: you're probably listening - unsupported")
			if uint32(len(err)) > count {
				err = err[:count]
			}
			return err, nil
		}
	case listen:
		if ni.checkDead() {
			err := []byte("error: dead connection")
			if uint32(len(err)) > count {
				err = err[:count]
			}
			return err, nil
		}
		if ni.s == idle {
			err := ni.enterListeningState()
			if err != nil {
				err := []byte(err.Error())
				if uint32(len(err)) > count {
					err = err[:count]
				}
				return err, nil
			}
		}

		if ni.s == listening {
			cn, err := ni.acceptCon()
			if err != nil {
				ni.s = dead
				err := []byte(err.Error())
				if uint32(len(err)) > count {
					err = err[:count]
				}
				return err, nil
			}

			id := genId()
			c.cons[id] = mkEmptyNetInst(fd.owner, id)
			c.cons[id].s = connected
			c.cons[id].c = cn
			return []byte(fmt.Sprintf("%d", id))[:count], nil
		} else {
			err := []byte("error: you are probably connected - unsupported")
			if uint32(len(err)) > count {
				err = err[:count]
			}
			return err, nil
		}
	}
	return []byte("THIS IS A BUG")[:count], nil
}

func (c *NetFs) Write(conId uint64, f nine.Fid, offset uint64, dta []byte) (uint32, error) {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return 0, errors.New("fid requested for write is not in use")
	}

	if !fd.open || (fd.openMode&0b11 == nine.OREAD || fd.openMode&0b11 == nine.OEXEC) {
		return 0, errors.New("fid is not Opened for writing")
	}

	// The file cannot possibly be opened for writing unless it is
	// - ctl
	// - data
	id := c.fidTable[conId][fd]
	cid, tp := reduceId(id)
	ni := c.cons[cid]
	switch tp {
	case ctl:
		// Accept the written string into cs if we
		// are idle, otherwise do nothing
		if ni.s != idle {
			return 0, nil
		}
		c.updateMTimeAs(fd.owner, id)
		ni.cs = string(data)
		return uint32(len(dta)), nil
	case data:
		// If the state is connecting, open
		// the connection for the first time.
		if ni.checkDead() {
			return 0, nil
		} else if ni.s == idle {
			err := ni.openCon()
			if err != nil {
				return 0, nil
			}
		}

		if ni.s == connected {
			// Reconciling this is ur problem
			_, err := (*ni.c).Write(dta)
			if err != nil {
				ni.s = dead
				return 0, nil
			}
			c.updateMTimeAs(fd.owner, id)
			return uint32(len(dta)), nil
		} else {
			// If you are idle or listening, this makes
			// approximately zero sense to read
			return 0, nil
		}
	default:
		return 0, nil
	}
}

func (c *NetFs) Clunk(conId uint64, f nine.Fid) error {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return errors.New("fid requested for Clunk is not in use")
	}

	id := c.fidTable[conId][fd]
	delete(c.fidTable[conId], fd)
	if id != rootId && id != cloneId {
		c.gc(id)
	}
	return nil
}

func (c *NetFs) Remove(conId uint64, f nine.Fid) error {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return errors.New("fid requested for Remove is not in use")
	}

	oldID := c.fidTable[conId][fd]
	err := c.Clunk(conId, f)
	if err != nil {
		if c.idExists(oldID) {
			return errors.New("permission denied")
		} else {
			return nil
		}
	}
	return err
}

func (c *NetFs) Stat(conId uint64, f nine.Fid) (nine.Stat, error) {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return nine.Stat{}, errors.New("fid requested for Stat is not in use")
	}

	id := c.fidTable[conId][fd]
	if id == rootId {
		return *c.rootstat, nil
	} else if id == cloneId {
		return *c.clonestat, nil
	} else {
		return c.genStat(id), nil
	}
}

func (c *NetFs) Wstat(conId uint64, f nine.Fid, ns nine.Stat) error {
	return errors.New("permission denied")
}

func (c *NetFs) Goodbye(conId uint64) error {
	for fd := range c.fidTable[conId] {
		c.Clunk(conId, fd.fid)
	}
	delete(c.fidTable, conId)
	return nil
}
