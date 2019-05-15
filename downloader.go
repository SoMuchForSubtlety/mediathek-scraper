package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func downloadVOD(url string, title string) error {
	title = strings.Replace(title, ":", " -", -1)
	title += ".%(ext)s"
	cmd := exec.Command("youtube-dl", "-o", title, url)
	stdoutIn, _ := cmd.StdoutPipe()
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return err
	}
	monitorCommand(stdoutIn)
	return nil
}

func monitorCommand(output io.ReadCloser) {
	scanner := bufio.NewScanner(output)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}
