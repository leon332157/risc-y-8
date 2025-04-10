package r8

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	BorderBackground(lipgloss.Color("240"))

// TODO: develop models to represents necessary components of our TUI
// RAM & Cache --> (configurable cache ??)
// Registers
// Pipeline Stages
// Current instruction
type Model struct {
	// embed other models on page
	ram       table.Model
	cache     table.Model
	regs      table.Model
	instr     textinput.Model
	fetch     table.Model //single column table
	decode    table.Model
	execute   table.Model
	memory    table.Model
	writeback table.Model
}

// type RamModel struct {
// 	ram table.Model
// }

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
		// TODO: Click events??
		// forward arrow advances the pipeline by one stage
		// ctrl+n advances the pipeline by 5 stages?? (1 cycle)
	}
	return m, nil
}

func (m Model) View() string {
	return baseStyle.Render(m.cache.View()) //+ baseStyle.Render(m.ram.View()) + "\n"
}

func TUIMain() {

	// RAM DISPLAY:
	ram := memory.CreateRAM(32, 8, 5)

	// loop thru contents, make RAM rows
	rRows := []table.Row{}
	addr := 0
	for range ram.NumLines {
		row := []string{}
		for range ram.WordsPerLine {
			row = append(row, fmt.Sprintf("%08X", ram.Contents[addr]))
			addr++
		}
		rRows = append(rRows, row)
	}

	rm := table.New(
		table.WithColumns([]table.Column{
			{Title: "RAM", Width: 8}, {Width: 8}, {Width: 8}, {Width: 8},
			{Width: 8}, {Width: 8}, {Width: 8}, {Width: 8},
		}),
		table.WithRows(rRows),
		table.WithFocused(true),
		table.WithHeight(int(ram.NumLines)+1),
	)

	// CACHE DISPLAY:
	cache := memory.CreateCacheDefault(&ram)

	cRows := []table.Row{}
	for i := range cache.Contents {
		for j := range cache.Ways {
			data := cache.Contents[i][j]
			row := []string{
				fmt.Sprintf("%05b", data.Tag),
				fmt.Sprintf("%03b", i),
				fmt.Sprintf("%08X", data.Data),
				fmt.Sprintf("%t", data.Valid),
				fmt.Sprintf("%d", data.LRU)}
			cRows = append(cRows, row)
		}
	}

	c := table.New(
		table.WithColumns([]table.Column{
			{Title: "Tag", Width: 5}, {Title: "Index", Width: 5}, {Title: "Data", Width: 8},
			{Title: "Valid", Width: 5}, {Title: "LRU", Width: 3},
		}),
		table.WithRows(cRows),
		table.WithFocused(true),
		table.WithHeight(int(cache.Sets*cache.Ways)+1),
	)

	rg := table.New(
	// How are registers created?
	)

	i := textinput.New()

	fet := table.New()
	dec := table.New()
	exe := table.New()
	mem := table.New()
	wb := table.New()

	m := Model{rm, c, rg, i, fet, dec, exe, mem, wb}

	// Log details to a file for debugging purposes
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		log.Fatalf("err: %w", err)
	}
	defer f.Close()

	// Run a new tea program
	program := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}
