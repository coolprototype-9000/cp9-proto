package user

import (
	"fmt"
	"os"
	"strings"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

func setupConsFs() {
	var cs string
	fd := p.Open("/cons/listen", nine.OREAD)
	if fd < 0 {
		goto err
	}

	printf("Waiting for consfs to connect...")
	cs = p.Read(fd, maxQt)
	if len(cs) == 0 {
		goto err
	}
	printf("Connected.\n")

	p.Close(fd)
	fd = p.Open(fmt.Sprintf("/cons/%s", cs), nine.ORDWR)
	if fd < 0 {
		goto err
	}

	if p.Dup(fd, 0) < 0 || p.Dup(fd, 1) < 0 || p.Dup(fd, 2) < 0 {
		goto err
	}
	p.Close(fd)
	return

err:
	p.Write(2, fmt.Sprintf("failed to configure consfs: %s\n", p.Errstr()))
	os.Exit(1)
}

var backupIn int
var backupOut int

const (
	noBinary = iota
	inRedir
	outRedir
	outAppendRedir
)

func sh() {
	setupConsFs()

	backupIn = p.Dup(0, -1)
	backupOut = p.Dup(1, -1)
	if backupIn < 0 || backupOut < 0 {
		printf("couldn't dup stdout/stdin for future use\n")
		os.Exit(1)
	}

	for {
		cmd := p.Read(0, maxQt)
		args := strings.Split(cmd, " ")
		if len(args) == 0 {
			printf("no command received\n")
			continue
		}

		lcmd := []string{}
		var rcmd string
		t := noBinary
		for _, arg := range args {
			if arg == ">" || arg == "<" || arg == ">>" {
				if len(lcmd) == 0 {
					printf("syntax error\n")
					continue
				} else if t != noBinary {
					printf("only binary commands supported\n")
					continue
				}
				switch arg {
				case ">":
					t = outRedir
				case "<":
					t = inRedir
				case ">>":
					t = outAppendRedir
				}
			} else if t == noBinary {
				lcmd = append(lcmd, arg)
			} else if rcmd != "" {
				printf("right side of rcmd should be a single file!\n")
				continue
			} else {
				rcmd = arg
			}
		}

		switch t {
		case noBinary:
			execCmd(lcmd)
		case outRedir:
			fd := p.Open(rcmd, nine.OWRITE|nine.OTRUNC)
			if fd < 0 {
				fd = p.Create(rcmd, nine.OWRITE, nine.PUR|nine.PUW|nine.PGR|nine.PGW)
				if fd < 0 {
					printf("Failed to create %s: %s\n", rcmd, p.Errstr())
					continue
				}
			}

			p.Dup(fd, 1)
			execCmd(lcmd)
			p.Dup(backupOut, 1)
			p.Close(fd)
		case outAppendRedir:
			fd := p.Open(rcmd, nine.OWRITE)
			if fd < 0 {
				fd = p.Create(rcmd, nine.OWRITE, nine.PUR|nine.PUW|nine.PGR|nine.PGW)
				if fd < 0 {
					printf("Failed to create %s: %s\n", rcmd, p.Errstr())
					continue
				}
			}

			p.Dup(fd, 1)
			execCmd(lcmd)
			p.Dup(backupOut, 1)
			p.Close(fd)
		case inRedir:
			fd := p.Open(rcmd, nine.OREAD)
			if fd < 0 {
				printf("Failed to open %s: %s\n", rcmd, p.Errstr())
				continue
			}

			p.Dup(fd, 0)
			execCmd(lcmd)
			p.Dup(backupIn, 0)
			p.Close(fd)
		}
	}
}

func execCmd(args []string) {
	switch args[0] {
	case "ls":
		ls(args...)
	case "cd":
		cd(args...)
	case "echo":
		echo(args...)
	case "cat":
		cat(args...)
	case "bind":
		bind(args...)
	case "unmount":
		unmount(args...)
	case "rm":
		rm(args...)
	case "pwd":
		pwd(args...)
	case "mkdir":
		mkdir(args...)
	case "touch":
		touch(args...)
	default:
		printf("unknown command")
	}
}
