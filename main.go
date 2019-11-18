package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
)

const (
	targetDir = "/watch"
	uid       = 1000
	gid       = 100
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	cmd := exec.Command("inotifywait", "-m", "-r", targetDir)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return errors.Wrap(err, "failed to create stdout pipe")
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "failed to start command")
	}
	log.Println("watching...")

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		if err := process(scanner.Text()); err != nil {
			log.Println("error:", err)
		}
	}

	return errors.Wrap(scanner.Err(), "reading stdout error")
}

func process(line string) error {
	attributes := strings.Split(line, " ")
	if len(attributes) < 2 {
		return fmt.Errorf("strange line format: %s", line)
	}

	action := attributes[1]

	if strings.HasPrefix(action, "CREATE") {
		if len(attributes) < 3 {
			return fmt.Errorf("strange line format: %s", line)
		}

		dir := attributes[0]
		file := attributes[2]
		file = filepath.Join(dir, file)

		return chown(file)
	}

	return nil
}

func chown(file string) error {
	fuid, fgid, err := getUIDGID(file)
	if err == nil && (fuid == 0 || fgid == 0) {
		return os.Chown(file, uid, gid)
	}

	return nil
}

func getUIDGID(file string) (int, int, error) {
	fstat, err := os.Stat(file)
	if err != nil {
		return 0, 0, err
	}
	if stat, ok := fstat.Sys().(*syscall.Stat_t); ok {
		return int(stat.Uid), int(stat.Gid), nil
	}

	return 0, 0, errors.New("failed to get stat")
}
