package main

import (
	"os/exec"
)

func PlayMP3(fname string) error {
	return exec.Command("afplay", fname).Run()
}
