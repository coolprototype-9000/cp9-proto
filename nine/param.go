package nine

// Commonly used parameters, or constants
// Kicking this off with 9P message types

const (
	TVersion = iota + 100
	RVersion
	TAuth
	RAuth
	TAttach
	RAttach
	TError
	RError
	TFlush
	RFlush
	TWalk
	RWalk
	TOpen
	ROpen
	TCreate
	RCreate
	TRead
	RRead
	TWrite
	RWrite
	TClunk
	RClunk
	TRemove
	RRemove
	TStat
	RStat
	TWStat
	RWStat
	TGoodbye
)

// NineVersion is our own unique version string
const NineVersion = "9P2000.cp9.1"

// Byte flags
const (
	FDir        = 1 << 7
	FAppend     = 1 << 6
	FExcl       = 1 << 5
	FAuth       = 1 << 3
	FTmp        = 1 << 2
	FStatOffset = 24
)

// Permisssions
const (
	PUR = 1 << 10
	PUW = 1 << 9
	PUX = 1 << 8
	PGR = 1 << 6
	PGW = 1 << 5
	PGX = 1 << 4
	POR = 1 << 2
	POW = 1 << 1
	POX = 1 << 0
)

// Device types
const (
	DevRamFs uint16 = 0
	DevCons  uint16 = 1
)

// Mode bits
const (
	OREAD   = 0b00
	OWRITE  = 0b01
	ORDWR   = 0b10
	OEXEC   = 0b11
	OTRUNC  = 0x10
	ORCLOSE = 0x40
)
