package main

import (
	"os"
	"os/exec"
)

func PlayMP3s(req Req, resp Resp) error {
	for _, item := range resp.Items {
		path := req.CacheMP3Fname(item.PathMP3)
		if _, err := os.Stat(path); err != nil {
			// missing (download error); ignore
			continue
		}
		cmd := exec.Command("mplayer", path)
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
