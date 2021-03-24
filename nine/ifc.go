package nine

type FileSys interface {
	Register(uint32) error
	Attach(uint64, Fid, string) (Qid, error)
	Walk(uint64, Fid, Fid, []string) ([]Qid, error)
	Create(uint64, Fid, string, uint32, byte) (Qid, error)
	Open(uint64, Fid, byte) (Qid, error)
	Read(uint64, Fid, uint64, uint32) ([]byte, error)
	Write(uint64, Fid, uint64, []byte) (uint32, error)
	Clunk(uint64, Fid) error
	Remove(uint64, Fid) error
	Stat(uint64, Fid) (Stat, error)
	Wstat(uint64, Fid, Stat) error
	Goodbye(uint64) error
}
