package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/stopwatch"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding  = 2
	maxWidth = 80
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type mergerMsg string

type progressUpdate float64

type progressErrMsg struct{ err error }

type model struct {
	pw        *progressWriter
	progress  progress.Model
	err       error
	stopwatch stopwatch.Model
	completed bool
}

func (m model) Init() tea.Cmd {
	return m.stopwatch.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	case progressErrMsg:
		m.err = msg.err
		return m, tea.Quit

	case progressUpdate:
		var cmds []tea.Cmd

		cmds = append(cmds, m.progress.SetPercent(float64(msg)))
		return m, tea.Batch(cmds...)

	case mergerMsg:
		m.stopwatch.Stop()
		m.completed = true
		return m, tea.Sequence(finalPause(), tea.Quit)

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		var cmd tea.Cmd
		m.stopwatch, cmd = m.stopwatch.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	pad := strings.Repeat(" ", padding)

	s := m.stopwatch.View() + "\n"
	if !m.completed {
		s = "Elapsed: " + s
	} else {
		var style = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#73F59C"))

		s = style.Render("Finished in " + s)
	}

	return "\n" +
		pad + m.progress.View() + "\n\n" +
		pad + s + "\n" +
		pad + helpStyle("Press any key to quit")
}

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(_ time.Time) tea.Msg {
		return nil
	})
}
