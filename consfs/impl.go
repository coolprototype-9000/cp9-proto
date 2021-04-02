package consfs

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/coolprototype-9000/cp9-proto/nine"
	"golang.org/x/net/websocket"
)

func (c *ConsFs) Register(devNo uint32) error {
	// Initialize dispatcher
	dispatcher = dispatch{c: make(chan *service), maxCon: 2}

	// Initialize ConsFs
	c.devNumber = devNo
	c.fidTable = make(map[uint64]map[fidDesc]uint64)
	c.fileTable = make(map[uint64]*service)
	c.statTable = make(map[uint64]nine.Stat)

	// Start up the HTTP server
	http.Handle("/", websocket.Handler(HandleWsCon))
	go http.ListenAndServe(fmt.Sprintf(":%d", servicePort), nil)
	return nil
}

func (c *ConsFs) Attach(conId uint64, f nine.Fid, uname string) (nine.Qid, error) {
	// "An error is returned if fid is already in use"
	if _, ok := c.getFid(conId, f); ok {
		return nine.Qid{}, errors.New("fid already in use for session")
	}

	// Make the root directory if it doesn't exist
	if !c.idExists(rootId) {
		c.statTable[rootId] = nine.Stat{
			DevType: nine.DevCons,
			DevNo:   c.devNumber,
			Q:       nine.Qid{Flags: nine.FDir},
			Mode:    nine.PUR | nine.PUW | nine.PUX | nine.PGR | nine.PGX | nine.POR | nine.POX,
			Atime:   uint32(time.Now().Unix()),
			Mtime:   uint32(time.Now().Unix()),
			Name:    "/",
			Uid:     "None", // yay anarchy
			Gid:     "None",
			Muid:    "None",
		}

		// If the root dir doesn't exist neither does listen
		c.statTable[listenId] = nine.Stat{
			DevType: nine.DevCons,
			DevNo:   c.devNumber,
			Q:       nine.Qid{},
			Mode:    nine.PUR | nine.PGR | nine.POR,
			Atime:   uint32(time.Now().Unix()),
			Mtime:   uint32(time.Now().Unix()),
			Name:    "listen",
			Uid:     "None", // yay anarchy
			Gid:     "None",
			Muid:    "None",
		}
	}

	// Attach the client
	fd := fidDesc{
		fid:   f,
		owner: uname,
	}

	if c.fidTable[conId] == nil {
		c.fidTable[conId] = make(map[fidDesc]uint64)
	}
	c.fidTable[conId][fd] = rootId
	return c.statTable[rootId].Q, nil

}

func (c *ConsFs) Walk(conId uint64, f nine.Fid, nf nine.Fid, wname []string) ([]nine.Qid, error) {
	// "the fid must be valid in the current session"
	if _, ok := c.getFid(conId, f); !ok {
		return []nine.Qid{}, errors.New("fid not already in use for session")
	}

	// "newfid may not be in use"
	if _, ok := c.getFid(conId, nf); ok {
		return []nine.Qid{}, errors.New("nfid already in use for session")
	}

	fd, _ := c.getFid(conId, f)
	nfd := fidDesc{
		fid:   nf,
		owner: fd.owner,
	}

	id := c.fidTable[conId][fd]

	// "the fid must represent a directory unless zero pathnames are specified"
	if len(wname) != 0 && id != rootId {
		return []nine.Qid{}, errors.New("starting fid is not a directory")
	}

	c.updateATime(rootId)

	// "the fid must not have been opened for i/o"
	if c.isOpenByMe(conId, f) {
		return []nine.Qid{}, errors.New("fid is open for i/o so can't walk")
	}

	// "it is legal for nwname to be zero, in which case the walk works and nf
	// represents the same file as f"
	// this just dups the fd
	if len(wname) == 0 {
		c.fidTable[conId][nfd] = id
		return []nine.Qid{}, nil
	}

	// "nwname path elements are walked in order, elementwise"
	// however, we can't walk more than one element because there's
	// only one directory. thus wname can only have one element
	if len(wname) > 1 {
		return []nine.Qid{}, errors.New("no such file or directory")
	} else if wname[0] == ".." {
		return []nine.Qid{}, errors.New("root directory has no parent")
	}

	// For the first walk to proceed, the file identified by f must be dir
	// and the implied user of the request must have permission to search the dir
	// The latter is a given, so if element cannot be walked at all, RError is returned
	wqid := make([]nine.Qid, 1)

	for id, srv := range c.fileTable {
		cStr := strconv.Itoa(int(srv.conNum))
		if cStr == wname[0] {
			// Walk returns RWalk containing QIDs corresponding to
			// successful descends
			wqid[0] = c.statTable[id].Q
			c.fidTable[conId][nfd] = id
			return wqid, nil
		}
	}

	if wname[0] == "listen" {
		wqid[0] = c.statTable[listenId].Q
		c.fidTable[conId][nfd] = listenId
		return wqid, nil
	}

	return []nine.Qid{}, errors.New("no such file or directory")

}

func (c *ConsFs) Create(conId uint64, f nine.Fid, name string, perm uint32, mode byte) (nine.Qid, error) {
	return nine.Qid{}, errors.New("permission denied")
}

func (c *ConsFs) Open(conId uint64, f nine.Fid, mode byte) (nine.Qid, error) {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return nine.Qid{}, errors.New("fid requested for open is not in use")
	}
	id := c.fidTable[conId][fd]
	c.updateATime(id)

	if fd.open {
		return nine.Qid{}, errors.New("fid is currently open, open is illegal")
	}

	// OK, let's get through the basics
	if mode&nine.OTRUNC > 0 {
		return nine.Qid{}, errors.New("cannot trunc this file")
	} else if mode&nine.ORCLOSE > 0 {
		return nine.Qid{}, errors.New("cannot rclose this file")
	}

	openTypeBits := mode & 0b11
	if id == rootId || id == listenId {
		if openTypeBits == nine.OWRITE || openTypeBits == nine.ORDWR {
			return nine.Qid{}, errors.New("it is illegal to write this file")
		} else if openTypeBits == nine.OEXEC && id == listenId {
			return nine.Qid{}, errors.New("it is illegal to exec this file")
		}

	} else {
		if openTypeBits == nine.OEXEC {
			return nine.Qid{}, errors.New("it is illegal to exec this file")
		}
	}

	// Permissions satisfied. It isn't until the read
	// that anything special happens. Commit.
	delete(c.fidTable[conId], fd)
	fd.open = true
	fd.openMode = mode
	c.fidTable[conId][fd] = id
	return c.statTable[id].Q, nil
}

func (c *ConsFs) Read(conId uint64, f nine.Fid, offset uint64, count uint32) ([]byte, error) {
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
		// "Seeking to other than the beginning is illegal in a directory"
		// We are going to be stricter about this. Practically this is limiting,
		// and definitely a TODO item due to limited message sizing, but tracking
		// the last return value is something I don't have time for right now.
		// - ramfs, we're the same here.
		if offset != 0 {
			return []byte{}, errors.New("directory seeking is illegal")
		}

		// Also from Ramfs:
		// Make a buffer to return things into
		scnt := int(count)

		// Add the listen file
		scnt -= len(nine.MarshalStat(c.statTable[listenId]))
		if scnt < 0 {
			return []byte{}, nil
		}
		ret := nine.MarshalStat(c.statTable[listenId])

		for _, srv := range c.fileTable {
			StatBytes := nine.MarshalStat(c.statTable[uint64(srv.conNum)])
			penalty := len(StatBytes)

			scnt -= penalty
			if scnt < 0 {
				break
			}
			ret = append(ret, StatBytes...)
		}

		return ret, nil
	}

	// Otherwise, this is a file. Is it listen?
	if id == listenId {
		// Completely ignore offset
		// If they don't want any bytes, idk what's wrong and don't waste my time
		// Specifically, you need 8 bytes to marshal the uint64
		if count == 0 {
			return []byte{}, nil
		}
		// Read in a new connection, might block
		nsrv := getService()
		nid := uint64(nsrv.conNum)

		// Make a new stat for this connection
		ns := nine.Stat{
			DevType: nine.DevCons,
			DevNo:   c.devNumber,
			Q:       nine.Qid{Flags: nine.FAppend},
			Mode:    nine.PUR | nine.PUW | nine.PGR | nine.PGW | nine.POR | nine.POW,
			Atime:   uint32(time.Now().Unix()),
			Mtime:   uint32(time.Now().Unix()),
			Name:    strconv.Itoa(int(nid)),
			Uid:     "None",
			Gid:     "None",
			Muid:    fd.owner,
		}

		c.statTable[nid] = ns
		c.fileTable[nid] = nsrv
		c.updateMTimeAs("None", rootId)

		// Creation is complete. Now read it back to them, if we can
		// Their loss if they don't make enough room
		ba := []byte(strconv.Itoa(int(nid)))
		if uint32(len(ba)) > count {
			ba = ba[:count+1]
		}
		return ba, nil
	}

	// Otherwise otherwise, this is a connection!
	s := c.fileTable[id]
	data, err := s.readWS()
	if uint32(len(data)) > count {
		data = data[:count+1]
	}
	return []byte(data), err
}

func (c *ConsFs) Write(conId uint64, f nine.Fid, offset uint64, data []byte) (uint32, error) {
	fd, ok := c.getFid(conId, f)
	if !ok {
		return 0, errors.New("fid requested for Write is not in use")
	}

	// Check perms
	if !fd.open || (fd.openMode&0b11 == nine.OREAD || fd.openMode&0b11 == nine.OEXEC) {
		return 0, errors.New("fid is not Opened for writing")
	}

	// The file cannot possibly be opened for writing unless it is a data file
	// These files are also not seekable/append only, so offset is ignored
	srv := c.fileTable[c.fidTable[conId][fd]]

	// If the data is zero, don't do a thing
	if len(data) == 0 {
		return 0, nil
	}

	// Let's get to it.
	c.updateMTimeAs(fd.owner, c.fidTable[conId][fd])
	err := srv.writeWS(string(data))
	if err != nil {
		return 0, err
	}
	return uint32(len(data)), nil
}

func (c *ConsFs) Clunk(conId uint64, f nine.Fid) error {
	// The fid must exist, of course
	fd, ok := c.getFid(conId, f)
	if !ok {
		return errors.New("fid requested for Clunk is not in use")
	}

	// So basically, if the file is open by ANYBODY, leave it
	// Otherwise, remove it...
	id := c.fidTable[conId][fd]
	if !c.isOpenByAnyone(id) {
		c.updateMTimeAs("None", rootId)
		delete(c.statTable, id)
		delete(c.fileTable, id)
	}
	delete(c.fidTable[conId], fd)
	return nil
}

func (c *ConsFs) Remove(conId uint64, f nine.Fid) error {
	// The fid must exist, of course
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

func (c *ConsFs) Stat(conId uint64, f nine.Fid) (nine.Stat, error) {
	// The fid must exist, of course
	fd, ok := c.getFid(conId, f)
	if !ok {
		return nine.Stat{}, errors.New("fid requested for Stat is not in use")
	}

	// "Stat request requires no special permissions"
	id := c.fidTable[conId][fd]
	return c.statTable[id], nil
}

func (c *ConsFs) Wstat(conId uint64, f nine.Fid, ns nine.Stat) error {
	return errors.New("permission denied")
}

func (c *ConsFs) Goodbye(conId uint64) error {
	// Clunk every connection
	for fd := range c.fidTable[conId] {
		c.Clunk(conId, fd.fid)
	}
	delete(c.fidTable, conId)
	return nil
}
