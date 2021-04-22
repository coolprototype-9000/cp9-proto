package client

import (
	"net"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// Initiates a 9P connection with unbounded message size
func fVersion(c *net.Conn, msize uint32, version string) {
	sf := nine.FCall{
		MsgType: nine.TVersion,
		Tag:     mkTag(),
		MSize:   msize,
		Version: version,
	}

	// Send, recv, print
	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RVersion)
}

// TODO
func fAttach(c *net.Conn, fid nine.Fid, uname string) *nine.FCall {
	sf := nine.FCall{
		MsgType: nine.TAttach,
		Tag:     mkTag(),
		F:       fid,
		Uname:   uname,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RAttach)
	return res
}

// TODO
func fWalk(c *net.Conn, fid nine.Fid, newFid nine.Fid, wname []string) *nine.FCall {
	sf := nine.FCall{
		MsgType: nine.TWalk,
		Tag:     mkTag(),
		F:       fid,
		Newf:    newFid,
		Wname:   wname,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RWalk)
	return res
}

func fCreate(c *net.Conn, fid nine.Fid, name string, perm uint32, mode byte) error {
	sf := nine.FCall{
		MsgType: nine.TCreate,
		Tag:     mkTag(),
		F:       fid,
		Name:    name,
		Perm:    perm,
		Mode:    mode,
	}

	res := writeAndRead(c, &sf)
	return checkMsg(res, nine.RCreate)
}

func fOpen(c *net.Conn, fid nine.Fid, mode byte) error {
	sf := nine.FCall{
		MsgType: nine.TOpen,
		Tag:     mkTag(),
		F:       fid,
		Mode:    mode,
	}

	res := writeAndRead(c, &sf)
	return checkMsg(res, nine.ROpen)
}

func fRead(c *net.Conn, fid nine.Fid, off uint64, cnt uint32) ([]byte, error) {
	sf := nine.FCall{
		MsgType: nine.TRead,
		Tag:     mkTag(),
		F:       fid,
		Offset:  off,
		Count:   cnt,
	}

	res := writeAndRead(c, &sf)
	err := checkMsg(res, nine.RRead)
	return res.Data, err
}

func fWrite(c *net.Conn, fid nine.Fid, off uint64, data string) (uint32, error) {
	sf := nine.FCall{
		MsgType: nine.TWrite,
		Tag:     mkTag(),
		F:       fid,
		Offset:  off,
		Data:    []byte(data),
	}

	res := writeAndRead(c, &sf)
	err := checkMsg(res, nine.RWrite)
	return res.Count, err
}

func fClunk(c *net.Conn, fid nine.Fid) error {
	sf := nine.FCall{
		MsgType: nine.TClunk,
		Tag:     mkTag(),
		F:       fid,
	}

	res := writeAndRead(c, &sf)
	return checkMsg(res, nine.RClunk)
}

func fRemove(c *net.Conn, fid nine.Fid) error {
	sf := nine.FCall{
		MsgType: nine.TRemove,
		Tag:     mkTag(),
		F:       fid,
	}

	res := writeAndRead(c, &sf)
	return checkMsg(res, nine.RRemove)
}

func fStat(c *net.Conn, fid nine.Fid) (nine.Stat, error) {
	sf := nine.FCall{
		MsgType: nine.TStat,
		Tag:     mkTag(),
		F:       fid,
	}

	res := writeAndRead(c, &sf)
	err := checkMsg(res, nine.RStat)
	return res.St, err
}
