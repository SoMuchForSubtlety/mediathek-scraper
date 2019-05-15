package main

import (
	"os/exec"
	"strings"
)

func download(url string, title string) error {
	title = strings.Replace(title, ":", " -", -1)
	title += ".%(ext)s"
	cmd := exec.Command("youtube-dl", "-o", title, url)
	err := cmd.Start()
	if err != nil {
		return err
	}
	return nil
}
