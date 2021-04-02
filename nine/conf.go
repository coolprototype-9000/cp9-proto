package nine

// Conf represents a set of options for the
// server to operate on. Fixed after first
// initialization, it is passed around other
// initialization routines so they know which
// parameters to use
type Conf struct {
	Port int
	Fs   FileSys
}

// MkDefConfig generates a configuration with
// sane defaults for the 9P server to use
func MkConfig(f FileSys, port int) Conf {
	return Conf{
		Port: port,
		Fs:   f,
	}
}
