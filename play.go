package main

import (
	"os/exec"
)

func PlayMP3(fname string) error {
	return exec.Command("mplayer", fname).Run()
}
