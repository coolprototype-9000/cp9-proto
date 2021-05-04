package user

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/coolprototype-9000/cp9-proto/nine"
)

func setupMapReduce() {
	matches, err := filepath.Glob("mr/sherlock-holmes-*")
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range matches {
		content, err := ioutil.ReadFile(m)
		if err != nil {
			log.Fatal(err)
		}

		sc := string(content)
		fd := p.Create(path.Base(m), nine.ORDWR, nine.PUR|nine.PUW|nine.PGR|nine.PGW|nine.POR|nine.POW)
		if fd < 0 {
			log.Fatal(p.Errstr())
		}

		ei := p.Write(fd, sc)
		if ei < 0 {
			log.Fatal(p.Errstr())
		}
		p.Close(fd)
	}
}

func setupConsFs() {
	var cs string
	fd := p.Open("/cons/listen", nine.OREAD)
	if fd < 0 {
		goto err
	}

	Printf("Waiting for consfs to connect...")
	cs = p.Read(fd, maxQt)
	if len(cs) == 0 {
		goto err
	}
	Printf("Connected.\n")

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
	setupMapReduce()
	// setupConsFs()

	backupIn = p.Dup(0, -1)
	backupOut = p.Dup(1, -1)
	if backupIn < 0 || backupOut < 0 {
		Printf("couldn't dup stdout/stdin for future use\n")
		os.Exit(1)
	}

	for {
		cmd := strings.TrimSuffix(p.Read(0, maxQt), "\n")
		args := strings.Split(cmd, " ")
		if len(args) == 0 {
			Printf("no command received\n")
			continue
		}

		lcmd := []string{}
		var rcmd string
		t := noBinary
		for _, arg := range args {
			if arg == ">" || arg == "<" || arg == ">>" {
				if len(lcmd) == 0 {
					Printf("syntax error\n")
					continue
				} else if t != noBinary {
					Printf("only binary commands supported\n")
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
				Printf("right side of rcmd should be a single file!\n")
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
					Printf("Failed to create %s: %s\n", rcmd, p.Errstr())
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
					Printf("Failed to create %s: %s\n", rcmd, p.Errstr())
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
				Printf("Failed to open %s: %s\n", rcmd, p.Errstr())
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
	case "mrc":
		mrc(args...)
	case "mrw":
		mrw(args...)
	default:
		Printf("unknown command\n")
	}
}
