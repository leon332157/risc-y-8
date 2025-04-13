package r8

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	table "github.com/charmbracelet/lipgloss/table"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

const (
	ramLines    = 32
	ramWords    = 8
	cacheSets   = 8
	cacheWays   = 2
	pipelineLen = 5
	registers   = 8 // ?????? Int, FP, Vector
)

var ram = memory.CreateRAM(32, 8, 5)
var cache = memory.CreateCacheDefault(&ram)

// TODO: get the rest connected and working
// var pipeline =
// var alu =
// var cpu =

type model struct {
	instr textinput.Model
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Instruction . . ."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return model{
		instr: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func getCacheRows() [][]string {
	cRows := [][]string{}
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
	return cRows
}

func getRAMRows() [][]string {
	rRows := [][]string{}
	addr := 0
	for i := 0; i < int(ram.NumLines); i++ {
		row := []string{}
		row = append(row, fmt.Sprintf("%d", i))
		for range ram.WordsPerLine {
			row = append(row, fmt.Sprintf("%08X", ram.Contents[addr]))
			addr++
		}
		rRows = append(rRows, row)
	}
	return rRows
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			// TODO: text input is saved and instruction is sent
			// call view?
		}
	case tea.WindowSizeMsg:
		// handle resize if needed
	}

	var cmd tea.Cmd
	m.instr, cmd = m.instr.Update(msg)
	return m, cmd
}

func (m model) View() string {
	ram := drawRAM()
	cache := drawCache()
	pipeline := drawPipeline()
	registerView := drawRegisters()
	cmdLine := m.instr.View() + "\n"

	// TODO: Show Clock and PC

	column1 := lipgloss.JoinVertical(lipgloss.Top, pipeline)
	regCache := lipgloss.JoinHorizontal(lipgloss.Left, registerView, cache, "\n")
	connect := lipgloss.JoinVertical(lipgloss.Top, column1, regCache)
	row1 := lipgloss.JoinHorizontal(lipgloss.Top, connect, "\n", ram)
	ui := lipgloss.JoinVertical(lipgloss.Left, row1, cmdLine)

	return "\n" + ui + "\n" + "---- ctrl+c or q to quit ----" + "\n"
}

func drawRAM() string {
	rows := getRAMRows()
	ramTable := table.New().Border(lipgloss.NormalBorder()).Rows(rows...)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("63")).Render("RAM\n" + ramTable.Render())
}

func drawCache() string {
	headers := []string{"Tag", "Index", "Data", "Valid", "LRU"}
	rows := getCacheRows()
	cacheTable := table.New().Border(lipgloss.NormalBorder()).Headers(headers...).Rows(rows...)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("129")).Render("Cache\n" + cacheTable.Render())
}

func drawPipeline() string {
	// TODO: create pipeline instance along with cpu
	labels := []string{" Fetch ", " Decode ", " Execute ", " Memory ", " Writeback "}
	row := []string{"", "", "", "", ""}
	// TODO: show stage result in row?? SUCCESS, STALL, FAILURE, NOOP, etc.
	pipelineTable := table.New().Border(lipgloss.NormalBorder()).Headers(labels...).Rows(row)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("33")).Render("Pipeline\n" + pipelineTable.Render())
}

func drawRegisters() string {
	// TODO: create a cpu instance and make int registers
	headers := []string{"Register", "Value"}
	rows := [][]string{}
	for i := 0; i < registers; i++ {
		rows = append(rows, []string{fmt.Sprintf("R%d", i), "00"})
	}
	regTable := table.New().Border(lipgloss.NormalBorder()).Headers(headers...).Rows(rows...)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render("Registers\n" + regTable.Render())
}

func TUImain() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
