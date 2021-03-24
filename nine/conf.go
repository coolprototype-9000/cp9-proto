package nine

// Conf represents a set of options for the
// server to operate on. Fixed after first
// initialization, it is passed around other
// initialization routines so they know which
// parameters to use
type Conf struct {
	Port    int
	Version string
	Fs      FileSys
}

// MkDefConfig generates a configuration with
// sane defaults for the 9P server to use
func MkConfig(f FileSys) Conf {
	return Conf{
		Port:    564,         // The typical 9P port of choice, times ten
		Version: nineVersion, // Our 9P protocol variant
		Fs:      f,
	}
}
