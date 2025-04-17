package r8

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	table "github.com/charmbracelet/lipgloss/table"
	"github.com/leon332157/risc-y-8/pkg/cpu"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

const (
	ramLines    = 32
	ramWords    = 8
	cacheSets   = 8
	cacheWays   = 2
	pipelineLen = 5
	registers   = 32 // just int regs for now
)

var (
	ram      = memory.CreateRAM(32, 8, 5)
	cache    = memory.CreateCacheDefault(&ram)
	control  = cpu.CPU{}
	pipeline = cpu.NewPipeline(&control)
	fs       = &cpu.FetchStage{}
	ds       = &cpu.DecodeStage{}
	es       = &cpu.ExecuteStage{}
	ms       = &cpu.MemoryStage{}
	ws       = &cpu.WriteBackStage{}
)

func InitSystem() {
	control.Init(&cache, &ram, pipeline)
	fs.Init(pipeline, ds, nil)
	ds.Init(pipeline, es, fs)
	es.Init(pipeline, ms, ds)
	ms.Init(pipeline, ws, es)
	ws.Init(pipeline, nil, ms)
	pipeline.AddStages(ws, ms, es, ds, fs)
}

type model struct {
	instr     textinput.Model
	lastInstr string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "type an instruction . . ."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50

	return model{
		instr:     ti,
		lastInstr: "",
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

func getRegVals() [][]string {
	regVals := [][]string{}
	for i := range len(control.IntRegisters) {
		row := []string{fmt.Sprintf("R%d", i), fmt.Sprintf("%08X", control.IntRegisters[i].Value)}
		regVals = append(regVals, row)
	}
	return regVals
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			m.lastInstr = m.instr.Value()
			// Send instruction to be computed
			cache.Write(0x0, memory.FETCH_STAGE, 0xdeadbeef)
			m.instr.Reset()
			return m, nil
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
	clock := drawClock()
	lastInstr := m.drawLastInstruction()
	cmdLine := m.instr.View() + "\n"
	whitespace := lipgloss.Place(3, 3, lipgloss.Right, lipgloss.Bottom, "")

	// TODO: Show PC??

	clockAndInstr := lipgloss.JoinHorizontal(lipgloss.Center, clock, whitespace, lastInstr)
	column1 := lipgloss.JoinVertical(lipgloss.Top, pipeline, cache, clockAndInstr)
	regsCol := lipgloss.JoinHorizontal(lipgloss.Left, registerView, whitespace, column1)
	together := lipgloss.JoinHorizontal(lipgloss.Top, regsCol, whitespace, ram)
	ui := lipgloss.JoinVertical(lipgloss.Left, together, cmdLine)

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
	rows := getRegVals()
	regTable := table.New().Border(lipgloss.NormalBorder()).Rows(rows...)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render("IntRegisters\n" + regTable.Render() + "\n")
}

func (m model) drawLastInstruction() string {
	headers := []string{"Last Instruction"}
	row := []string{m.lastInstr}
	instrTable := table.New().Border(lipgloss.NormalBorder()).Headers(headers...).Rows(row)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render(instrTable.Render())
}

func drawClock() string {
	header := []string{"Clock"}
	row := []string{fmt.Sprintf("%d", control.Clock)}
	clockTable := table.New().Border(lipgloss.NormalBorder()).Headers(header...).Rows(row)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render(clockTable.Render())
}

func TUImain() {
	InitSystem()
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
