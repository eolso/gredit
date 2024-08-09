package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	args := []string{"-I"}
	args = append(args, os.Args[1:]...)

	execCmd := exec.Command("grep", args...)

	b, err := execCmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(b))
	scanner.Split(bufio.ScanLines)

	var lines [][2]string
	for scanner.Scan() {
		split := strings.SplitN(scanner.Text(), ":", 2)
		if len(split) != 2 {
			continue
		}

		lines = append(lines, [2]string{split[0], split[1]})
	}

	p := tea.NewProgram(newModel(lines))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
