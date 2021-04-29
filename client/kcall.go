package client

import (
	"errors"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

// Initiates a 9P connection with unbounded message size
func fVersion(onichan *kchan, msize uint32, version string) (*kchan, error) {
	sf := nine.FCall{
		MsgType: nine.TVersion,
		Tag:     mkTag(),
		MSize:   msize,
		Version: version,
	}

	// Send, recv, print
	res := writeAndRead(onichan.c, &sf)
	err := checkMsg(res, nine.RVersion)

	if err == nil {
		var nc kchan
		nc.c = onichan.c
		return &nc, nil
	}
	return &kchan{}, errors.New("version failed")
}

// TODO
func fAttach(onichan *kchan, newFid nine.Fid, uname string, prefix string) error {
	sf := nine.FCall{
		MsgType: nine.TAttach,
		Tag:     mkTag(),
		F:       newFid,
		Uname:   uname,
	}

	res := writeAndRead(onichan.c, &sf) //Not sure if I'm returning the nine FCall or JUST the nc struct
	err := checkMsg(res, nine.RAttach)

	if err == nil {
		onichan.fid = newFid
		onichan.name = prefix
		return nil
	}
	return errors.New("attach failed")
}

// TODO
func fWalk(onichan *kchan, newFid nine.Fid, wname []string) (*kchan, error) { //Is the wname same as the name for kchan?
	sf := nine.FCall{
		MsgType: nine.TWalk,
		Tag:     mkTag(),
		F:       onichan.fid,
		Newf:    newFid,
		Wname:   wname,
	}

	res := writeAndRead(onichan.c, &sf)
	err := checkMsg(res, nine.RWalk)

	if err == nil && len(wname) == len(res.Wqid) {
		nc := &kchan{
			c:    onichan.c,
			name: onichan.name,
			fid:  newFid,
		}

		for i := 0; i < len(wname); i++ {
			nc.name += "/" + wname[i]
		}
		nc.name = cleanPath(nc.name)
		return nc, nil
	}

	return &kchan{}, errors.New("walk failed")
}

func fCreate(onichan *kchan, name string, perm uint32, mode byte) error {
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
	if err == nil {
		onichan.name = cleanPath(onichan.name + "/" + name)
		return nil
	}
	return errors.New("create failed")
}

func fOpen(onichan *kchan, mode byte) error {
	sf := nine.FCall{
		MsgType: nine.TOpen,
		Tag:     mkTag(),
		F:       onichan.fid,
		Mode:    mode,
	}

	res := writeAndRead(onichan.c, &sf)
	return checkMsg(res, nine.ROpen)
}

func fRead(onichan *kchan, off uint64, cnt uint32) ([]byte, error) {
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

func fWrite(onichan *kchan, off uint64, data string) (uint32, error) {
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

func fClunk(onichan *kchan) error {
	sf := nine.FCall{
		MsgType: nine.TClunk,
		Tag:     mkTag(),
		F:       onichan.fid,
	}

	res := writeAndRead(onichan.c, &sf)
	return checkMsg(res, nine.RClunk)
}

func fRemove(onichan *kchan) error {
	sf := nine.FCall{
		MsgType: nine.TRemove,
		Tag:     mkTag(),
		F:       onichan.fid,
	}

	res := writeAndRead(onichan.c, &sf)
	return checkMsg(res, nine.RRemove)
}

func fStat(onichan *kchan) (nine.Stat, error) {
	sf := nine.FCall{
		MsgType: nine.TStat,
		Tag:     mkTag(),
		F:       onichan.fid,
	}

	res := writeAndRead(onichan.c, &sf)
	err := checkMsg(res, nine.RStat)
	return res.St, err
}
