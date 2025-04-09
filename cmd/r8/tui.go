package r8

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	// Keyboard events
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c": // key to exit the tui
			return m, tea.Quit
		}
		// TODO: Click events
	}
	return m, nil
}

func (m Model) View() string {
	return "have a cup of tea"
}

func TUIMain() {
	// Log details to a file for debugging purposes
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("err: %w", err)
	}
	defer f.Close()

	// Run a new tea program
	program := tea.NewProgram(Model{}, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}

}
