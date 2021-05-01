package nine

import (
	"errors"
)

// marshalUint is used frequently
func marshalUint(num uint64, byteCnt uint8) []byte {
	buf := make([]byte, byteCnt)
	for i := uint8(0); i < byteCnt; i++ {
		buf[i] = uint8(num >> (8 * i))
	}
	return buf
}

func unmarshalUint8(b []byte) (uint8, []byte) {
	return b[0], b[1:]
}

func unmarshalUint16(b []byte) (uint16, []byte) {
	lt := uint16(b[0])
	ut := uint16(b[1]) << 8
	return ut | lt, b[2:]
}

func unmarshalUint32(b []byte) (uint32, []byte) {
	lt, b := unmarshalUint16(b)
	ht, b := unmarshalUint16(b)
	return uint32(ht)<<16 | uint32(lt), b
}

func unmarshalUint64(b []byte) (uint64, []byte) {
	lt, b := unmarshalUint32(b)
	ht, b := unmarshalUint32(b)
	return uint64(ht)<<32 | uint64(lt), b
}

func marshalString(s string) []byte {
	l := uint16(len(s))
	nbuf := marshalUint(uint64(l), 2)
	sbuf := []byte(s)
	return append(nbuf, sbuf...)
}

func unmarshalString(b []byte) (string, []byte) {
	l, b := unmarshalUint16(b)
	return string(b[0:l]), b[l:]
}

func marshalStrList(s []string) []byte {
	ovrLen := uint16(len(s))
	buf := marshalUint(uint64(ovrLen), 2)
	for i := 0; i < int(ovrLen); i++ {
		buf = append(buf, marshalString(s[i])...)
	}
	return buf
}

func unmarshalStrList(b []byte) ([]string, []byte) {
	l, b := unmarshalUint16(b)
	sb := make([]string, 0)
	for i := uint16(0); i < l; i++ {
		var s string
		s, b = unmarshalString(b)
		sb = append(sb, s)
	}
	return sb, b
}

func marshalQid(q Qid) []byte {
	buf := marshalUint(uint64(q.Flags), 1)
	buf = append(buf, marshalUint(uint64(q.Version), 4)...)
	buf = append(buf, marshalUint(q.Id, 8)...)
	return buf
}

func unmarshalQid(b []byte) (Qid, []byte) {
	f, b := unmarshalUint8(b)
	vr, b := unmarshalUint32(b)
	id, b := unmarshalUint64(b)
	return Qid{f, vr, id}, b
}

func marshalQidList(q []Qid) []byte {
	ovrLen := uint16(len(q))
	buf := marshalUint(uint64(ovrLen), 2)
	for i := 0; i < int(ovrLen); i++ {
		buf = append(buf, marshalQid(q[i])...)
	}
	return buf
}

func unmarshalQidList(b []byte) ([]Qid, []byte) {
	ovrLen, b := unmarshalUint16(b)
	buf := make([]Qid, 0)
	for i := 0; i < int(ovrLen); i++ {
		var q Qid
		q, b = unmarshalQid(b)
		buf = append(buf, q)
	}

	return buf, b
}

func marshalFid(f Fid) []byte {
	return marshalUint(uint64(f), 4)
}

func unmarshalFid(b []byte) (Fid, []byte) {
	f, b := unmarshalUint32(b)
	return Fid(f), b
}

// MarshalStat is needed for our read implementation,
// so tis exported
func MarshalStat(s Stat) []byte {
	buf := marshalUint(uint64(s.DevType), 2)
	buf = append(buf, marshalUint(uint64(s.DevNo), 4)...)
	buf = append(buf, marshalQid(s.Q)...)
	buf = append(buf, marshalUint(uint64(s.Mode), 4)...)
	buf = append(buf, marshalUint(uint64(s.Atime), 4)...)
	buf = append(buf, marshalUint(uint64(s.Mtime), 4)...)
	buf = append(buf, marshalUint(uint64(s.Len), 8)...)
	buf = append(buf, marshalString(s.Name)...)
	buf = append(buf, marshalString(s.Uid)...)
	buf = append(buf, marshalString(s.Gid)...)
	buf = append(buf, marshalString(s.Muid)...)

	sz := uint16(2 + 4 + 13 + 4 + 4 + 4 + 8 + len(s.Name) + len(s.Uid) + len(s.Gid) + len(s.Muid))
	s.Size = sz + 2
	buf = append(marshalUint(uint64(s.Size), 2), buf...)

	// Now, encapsulate one more time
	buf = append(marshalUint(uint64(s.Size+2), 2), buf...)
	return buf
}

func UnmarshalStat(b []byte) (Stat, []byte) {
	_, b = unmarshalUint16(b)
	sz, b := unmarshalUint16(b)
	dt, b := unmarshalUint16(b)
	dn, b := unmarshalUint32(b)
	q, b := unmarshalQid(b)
	mode, b := unmarshalUint32(b)
	atime, b := unmarshalUint32(b)
	mtime, b := unmarshalUint32(b)
	l, b := unmarshalUint64(b)
	name, b := unmarshalString(b)
	uid, b := unmarshalString(b)
	gid, b := unmarshalString(b)
	muid, b := unmarshalString(b)

	return Stat{
		Size:    sz,
		DevType: dt,
		DevNo:   dn,
		Q:       q,
		Mode:    mode,
		Atime:   atime,
		Mtime:   mtime,
		Len:     l,
		Name:    name,
		Uid:     uid,
		Gid:     gid,
		Muid:    muid,
	}, b
}

func marshalFCall(f FCall) ([]byte, error) {
	// Do size last. Here's the other common stuff.
	buf := marshalUint(uint64(f.MsgType), 1)
	buf = append(buf, marshalUint(uint64(f.Tag), 2)...)

	// Now we move around based on message type
	switch f.MsgType {
	case TVersion, RVersion:
		buf = append(buf, marshalUint(uint64(f.MSize), 4)...)
		buf = append(buf, marshalString(f.Version)...)
	case TAuth:
		buf = append(buf, marshalFid(f.Af)...)
		buf = append(buf, marshalString(f.Uname)...)
		buf = append(buf, marshalString(f.Aname)...)
	case RAuth:
		buf = append(buf, marshalQid(f.Aq)...)
	case RError:
		buf = append(buf, marshalString(f.Ename)...)
	case TFlush:
		buf = append(buf, marshalUint(uint64(f.OldTag), 2)...)
	case TAttach:
		buf = append(buf, marshalFid(f.F)...)
		buf = append(buf, marshalFid(f.Af)...)
		buf = append(buf, marshalString(f.Uname)...)
		buf = append(buf, marshalString(f.Aname)...)
	case RAttach:
		buf = append(buf, marshalQid(f.Q)...)
	case TWalk:
		buf = append(buf, marshalFid(f.F)...)
		buf = append(buf, marshalFid(f.Newf)...)
		buf = append(buf, marshalStrList(f.Wname)...)
	case RWalk:
		buf = append(buf, marshalQidList(f.Wqid)...)
	case TOpen:
		buf = append(buf, marshalFid(f.F)...)
		buf = append(buf, marshalUint(uint64(f.Mode), 1)...)
	case ROpen, RCreate:
		buf = append(buf, marshalQid(f.Q)...)
		buf = append(buf, marshalUint(uint64(f.IoUnit), 4)...)
	case TCreate:
		buf = append(buf, marshalFid(f.F)...)
		buf = append(buf, marshalString(f.Name)...)
		buf = append(buf, marshalUint(uint64(f.Perm), 4)...)
		buf = append(buf, marshalUint(uint64(f.Mode), 1)...)
	case TRead:
		buf = append(buf, marshalFid(f.F)...)
		buf = append(buf, marshalUint(f.Offset, 8)...)
		buf = append(buf, marshalUint(uint64(f.Count), 4)...)
	case RRead:
		buf = append(buf, marshalUint(uint64(f.Count), 4)...)
		buf = append(buf, f.Data...)
	case TWrite:
		buf = append(buf, marshalFid(f.F)...)
		buf = append(buf, marshalUint(f.Offset, 8)...)
		buf = append(buf, marshalUint(uint64(f.Count), 4)...)
		buf = append(buf, f.Data...)
	case RWrite:
		buf = append(buf, marshalUint(uint64(f.Count), 4)...)
	case TClunk, TRemove, TStat:
		buf = append(buf, marshalFid(f.F)...)
	case RStat:
		buf = append(buf, MarshalStat(f.St)...)
	case TWStat:
		buf = append(buf, marshalFid(f.F)...)
		buf = append(buf, MarshalStat(f.St)...)
	}

	ovrLen := len(buf)
	buf = append(marshalUint(uint64(ovrLen)+4, 4), buf...)

	// Cursory validation
	if len(buf) > int(^uint32(0)) {
		return nil, errors.New("message is oversized")
	}

	return buf, nil
}

func unmarshalFCall(b []byte) FCall {
	// Golang doesn't require the size parameter afaik,
	// so i'll toss it out for now
	_, b = unmarshalUint32(b)
	msgType, b := unmarshalUint8(b)
	tag, b := unmarshalUint16(b)

	//	fmt.Printf("Message metadata: %d, %d\n", msgType, tag)

	f := FCall{MsgType: msgType, Tag: tag}

	switch msgType {
	case TVersion, RVersion:
		f.MSize, b = unmarshalUint32(b)
		f.Version, _ = unmarshalString(b)
	case TAuth:
		f.Af, b = unmarshalFid(b)
		f.Uname, b = unmarshalString(b)
		f.Aname, _ = unmarshalString(b)
	case RAuth:
		f.Aq, _ = unmarshalQid(b)
	case RError:
		f.Ename, _ = unmarshalString(b)
	case TFlush:
		f.OldTag, _ = unmarshalUint16(b)
	case TAttach:
		f.F, b = unmarshalFid(b)
		f.Af, b = unmarshalFid(b)
		f.Uname, b = unmarshalString(b)
		f.Aname, _ = unmarshalString(b)
	case RAttach:
		f.Q, _ = unmarshalQid(b)
	case TWalk:
		f.F, b = unmarshalFid(b)
		f.Newf, b = unmarshalFid(b)
		f.Wname, _ = unmarshalStrList(b)
	case RWalk:
		f.Wqid, _ = unmarshalQidList(b)
	case TOpen:
		f.F, b = unmarshalFid(b)
		f.Mode, _ = unmarshalUint8(b)
	case ROpen, RCreate:
		f.Q, b = unmarshalQid(b)
		f.IoUnit, _ = unmarshalUint32(b)
	case TCreate:
		f.F, b = unmarshalFid(b)
		f.Name, b = unmarshalString(b)
		f.Perm, b = unmarshalUint32(b)
		f.Mode, _ = unmarshalUint8(b)
	case TRead:
		f.F, b = unmarshalFid(b)
		f.Offset, b = unmarshalUint64(b)
		f.Count, _ = unmarshalUint32(b)
	case RRead:
		f.Count, b = unmarshalUint32(b)
		f.Data = b
	case TWrite:
		f.F, b = unmarshalFid(b)
		f.Offset, b = unmarshalUint64(b)
		f.Count, b = unmarshalUint32(b)
		f.Data = b
	case RWrite:
		f.Count, _ = unmarshalUint32(b)
	case TClunk, TRemove, TStat:
		f.F, _ = unmarshalFid(b)
	case RStat:
		f.St, _ = UnmarshalStat(b)
	case TWStat:
		f.F, b = unmarshalFid(b)
		f.St, _ = UnmarshalStat(b)
	}

	return f
}
