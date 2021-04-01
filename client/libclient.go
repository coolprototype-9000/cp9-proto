package client

import (
	"net"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

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

func fAttach(c *net.Conn, fid nine.Fid, uname string) {
	sf := nine.FCall{
		MsgType: nine.TAttach,
		Tag:     mkTag(),
		F:       fid,
		Uname:   uname,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RAttach)
}

func fWalk(c *net.Conn, fid nine.Fid, newFid nine.Fid, wname []string) {
	sf := nine.FCall{
		MsgType: nine.TWalk,
		Tag:     mkTag(),
		F:       fid,
		Newf:    newFid,
		Wname:   wname,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RWalk)
}

func fCreate(c *net.Conn, fid nine.Fid, name string, perm uint32, mode byte) {
	sf := nine.FCall{
		MsgType: nine.TCreate,
		Tag:     mkTag(),
		F:       fid,
		Name:    name,
		Perm:    perm,
		Mode:    mode,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RCreate)
}

func fOpen(c *net.Conn, fid nine.Fid, mode byte) {
	sf := nine.FCall{
		MsgType: nine.TOpen,
		Tag:     mkTag(),
		F:       fid,
		Mode:    mode,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.ROpen)
}

func fRead(c *net.Conn, fid nine.Fid, off uint64, cnt uint32) {
	sf := nine.FCall{
		MsgType: nine.TRead,
		Tag:     mkTag(),
		F:       fid,
		Offset:  off,
		Count:   cnt,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RRead)
}

func fWrite(c *net.Conn, fid nine.Fid, off uint64, data string) {
	sf := nine.FCall{
		MsgType: nine.TWrite,
		Tag:     mkTag(),
		F:       fid,
		Offset:  off,
		Data:    []byte(data),
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RWrite)
}

func fClunk(c *net.Conn, fid nine.Fid) {
	sf := nine.FCall{
		MsgType: nine.TClunk,
		Tag:     mkTag(),
		F:       fid,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RClunk)
}

func fRemove(c *net.Conn, fid nine.Fid) {
	sf := nine.FCall{
		MsgType: nine.TRemove,
		Tag:     mkTag(),
		F:       fid,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RRemove)
}

func fStat(c *net.Conn, fid nine.Fid) {
	sf := nine.FCall{
		MsgType: nine.TStat,
		Tag:     mkTag(),
		F:       fid,
	}

	res := writeAndRead(c, &sf)
	checkMsg(res, nine.RStat)
}
