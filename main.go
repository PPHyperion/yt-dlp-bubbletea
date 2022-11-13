package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
)

var program *tea.Program

type progressWriter struct {
	title      string
	downloaded int
	reader     io.ReadCloser
	onProgress func(float64)
	onFinish   func(string)
	command    exec.Cmd
}

func main() {
	url := flag.String("url", "", "url or id of the video to download")
	flag.Parse()

	if *url == "" {
		flag.Usage()
		os.Exit(1)
	}

	cmdName := "yt-dlp --no-colors --newline " + *url
	cmdArgs := strings.Fields(cmdName)

	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	stdout, _ := cmd.StdoutPipe()

	pw := &progressWriter{
		reader: stdout,
		onProgress: func(progress float64) {
			program.Send(progressUpdate(progress))
		},
		onFinish: func(message string) {
			program.Send(mergerMsg(message))
		},
		command:    *cmd,
		downloaded: 0,
	}

	m := model{
		pw:        pw,
		progress:  progress.New(progress.WithDefaultGradient()),
		stopwatch: stopwatch.NewWithInterval(time.Millisecond),
	}

	program = tea.NewProgram(m)

	go pw.Start()

	if _, err := program.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
}

func (pw *progressWriter) Start() {
	pw.command.Start()

	reader := bufio.NewReader(pw.reader)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		if strings.Contains(line, "%") {
			progress := getProgress(line)
			pw.onProgress(float64(progress))
		} else if strings.Contains(line, "[Merger]") {
			pw.onProgress(float64(100))
			pw.onFinish("Download finished, begin merge")
			pw.downloaded = 1
		}
	}
}

func getProgress(text string) float64 {
	regex, _ := regexp.Compile(`(\d{1,3}.\d{0,1})%`)
	result := regex.FindStringSubmatch(text)
	value, _ := strconv.ParseFloat(result[1], 64)
	return value / 100
}
