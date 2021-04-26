package client

import (
	"errors"
	"net"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// Initiates a 9P connection with unbounded message size
func fVersion(onichan kchan, msize uint32, version string) {
	sf := nine.FCall{
		MsgType: nine.TVersion,
		Tag:     mkTag(),
		MSize:   msize,
		Version: version,
	}

	// Send, recv, print
	res := writeAndRead(onichan.c, &sf)
	checkMsg(res, nine.RVersion)
}

// TODO
func fAttach(c *net.Conn, newFid nine.Fid, uname string, prefix string) (kchan, error) {
	sf := nine.FCall{
		MsgType: nine.TAttach,
		Tag:     mkTag(),
		F:       newFid,
		Uname:   uname,
	}

	res := writeAndRead(c, &sf) //Not sure if I'm returning the nine FCall or JUST the dummy struct
	err := checkMsg(res, nine.RAttach)

	if err != nil {
		var dummy kchan
		dummy.c = c
		dummy.fid = newFid
		dummy.name = prefix
		return dummy, nil
	}
	return kchan{}, errors.New("Attach failed")
}

// TODO
func fWalk(onichan kchan, newFid nine.Fid, wname []string) (kchan, error) { //Is the wname same as the name for kchan?
	sf := nine.FCall{
		MsgType: nine.TWalk,
		Tag:     mkTag(),
		F:       onichan.fid,
		Newf:    newFid,
		Wname:   wname,
	}
	var onichan1 kchan
	onichan1.c = onichan.c
	onichan1.name = onichan.name
	onichan1.fid = newFid
	res := writeAndRead(onichan.c, &sf)
	err := checkMsg(res, nine.RRead)
	if err != nil && len(wname) == len(res.Wqid) {
		for i := 0; i < len(wname); i++ {
			onichan1.name = onichan1.name + "/" + wname[i]
		}
	} else {
		return onichan1, errors.New("walk failed")
	}
	return onichan1, nil
}

func fCreate(onichan kchan, name string, perm uint32, mode byte) (kchan, error) {
	sf := nine.FCall{
		MsgType: nine.TCreate,
		Tag:     mkTag(),
		F:       onichan.fid,
		Name:    name,
		Perm:    perm,
		Mode:    mode,
	}

	res := writeAndRead(onichan.c, &sf)

	err := checkMsg(res, nine.RCreate)
	if err != nil {
		onichan.name += "/" + name
		return onichan, nil
	}
	return onichan, errors.New("create failed")
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

func fRead(onichan kchan, off uint64, cnt uint32) ([]byte, error) {
	sf := nine.FCall{
		MsgType: nine.TRead,
		Tag:     mkTag(),
		F:       onichan.fid,
		Offset:  off,
		Count:   cnt,
	}

	res := writeAndRead(onichan.c, &sf)
	err := checkMsg(res, nine.RRead)

	return res.Data, err
}

func fWrite(onichan kchan, off uint64, data string) (uint32, error) {
	sf := nine.FCall{
		MsgType: nine.TWrite,
		Tag:     mkTag(),
		F:       onichan.fid,
		Offset:  off,
		Data:    []byte(data),
	}

	res := writeAndRead(onichan.c, &sf)
	err := checkMsg(res, nine.RWrite)
	return res.Count, err
}

func fClunk(onichan kchan) error {
	sf := nine.FCall{
		MsgType: nine.TClunk,
		Tag:     mkTag(),
		F:       onichan.fid,
	}

	res := writeAndRead(onichan.c, &sf)
	return checkMsg(res, nine.RClunk)
}

func fRemove(onichan kchan) error {
	sf := nine.FCall{
		MsgType: nine.TRemove,
		Tag:     mkTag(),
		F:       onichan.fid,
	}

	res := writeAndRead(onichan.c, &sf)
	return checkMsg(res, nine.RRemove)
}

func fStat(onichan kchan) (nine.Stat, error) {
	sf := nine.FCall{
		MsgType: nine.TStat,
		Tag:     mkTag(),
		F:       onichan.fid,
	}

	res := writeAndRead(onichan.c, &sf)
	err := checkMsg(res, nine.RStat)
	return res.St, err
}
