package user

// args[0] is "rm"
// p.Remove(filename) -> 0 if success, -1 otherwise
// p.Errstr() -> returns string describing most recent error
func rm(args ...string) {
	if len(args) == 1 {
		printf("too few arguments")
		return
	}
	for i := 1; i < len(args); i++ {
		if p.Remove(args[i]) == 0 {
			printf("Yeeted file succesfully")
		} else {
			printf("rm: error: %s\n", p.Errstr())
			break
		}
	}
}
