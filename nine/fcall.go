package nine

// Qid is a server file identifier.
// Consists of 1 byte of flags,
// 4 bytes of version, and 4 bytes of id
type Qid struct {
	Flags   byte
	Version uint32
	Id      uint64
}

// Fid is just a uint32, nothing special
type Fid uint32

// Stat is a standard, plan 9 style stat
type Stat struct {
	Size    uint16
	DevType uint16
	DevNo   uint32
	Q       Qid
	Mode    uint32
	Atime   uint32
	Mtime   uint32
	Len     uint64
	Name    string
	Uid     string
	Gid     string
	Muid    string
}

// FCall describes a single file
// call to the server, in 9p parlance.
// It is composed of the union of all
// possible fields, in a format conducive
// to go programming. The type field determines
// the message type for marshalling purposes.
type FCall struct {
	// Common fields
	MsgType byte
	Tag     uint16

	// Negotiation
	MSize   uint32
	Version string

	// File IDs
	Af   Fid
	Aq   Qid
	F    Fid
	Q    Qid
	Newf Fid
	Wqid []Qid

	// Names
	Name  string
	Uname string
	Aname string
	Ename string
	Wname []string

	// Flushing
	OldTag uint16

	// Attributes
	Mode   byte
	IoUnit uint32
	Perm   uint32

	// Reading and writing
	Offset uint64
	Count  uint32
	Data   []byte

	// Stat structure
	St Stat
}

// FCall, but it has a conID attached to it
type csFCall struct {
	f  FCall
	id uint64
}
