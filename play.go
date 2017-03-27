package main

import (
	"os/exec"
)

func PlayMP3(fname string) error {
	cmd := exec.Command("mplayer", fname)
	return cmd.Run()
}
