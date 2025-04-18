package r8

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/leon332157/risc-y-8/cmd/r8/simulator"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	table "github.com/charmbracelet/lipgloss/table"
	"github.com/leon332157/risc-y-8/pkg/cpu"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

var (
	tuiCmd = &cobra.Command{
		Use: "tui <flags> [binary file]",
		//Aliases: []string{},
		Short: "Simulate with TUI RISC-Y-8 binary",
		//Long:    "Assemble RISC-Y-8 assembly code into machine code",
		RunE:    runTui,
		Args:    cobra.ExactArgs(1),
		Example: "r8 tui input.bin",
	}
)

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTui(cmd *cobra.Command, args []string) error {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	infile := args[0]
	f, err := os.Open(infile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer f.Close()
	buffer, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}
	if len(buffer) == 0 {
		return fmt.Errorf("input file is empty: %v", err)
	}
	program := make([]uint32, len(buffer)/4)
	err = binary.Read(bytes.NewReader(buffer), binary.LittleEndian, &program)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}
	model := initialModel()
	system := simulator.NewSystem(program)
	model.system = &system
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}

type model struct {
	instr     textinput.Model
	lastInstr string

	system *simulator.System
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

func getCacheRows(ca *memory.CacheType) [][]string {
	cRows := [][]string{}
	for i := range ca.Contents {
		for j := range ca.Ways {
			data := ca.Contents[i][j]
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

func getRAMRows(ram *memory.RAM) [][]string {
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

func getRegVals(control *cpu.CPU) [][]string {
	regVals := [][]string{}
	for i := range len(control.IntRegisters) {
		row := []string{fmt.Sprintf("R%d", i), fmt.Sprintf("%08X", control.ReadIntRNoBlock(uint8(i)))}
		regVals = append(regVals, row)
	}
	return regVals
}

func (m model) ExecuteCommand() {
	switch m.lastInstr {
	case "step", "s":
		if m.system.CPU.Halted {
			return
		}
		m.system.CPU.Pipeline.RunOneClock()
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			temp := m.instr.Value()
			if temp != "" {
				m.lastInstr = m.instr.Value()
			}
			m.ExecuteCommand()
			// Send instruction to be computed
			//cache.Write(0x0, memory.FETCH_STAGE, 0xdeadbeef)
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
	ram := m.drawRAM()
	cache := m.drawCache()
	pipeline := m.drawPipeline()
	registerView := m.drawRegisters()
	clock := m.drawClock()
	pc := m.drawPC()
	lastInstr := m.drawLastInstruction()
	cmdLine := m.instr.View() + "\n"
	whitespace := lipgloss.Place(3, 3, lipgloss.Right, lipgloss.Bottom, "")

	// TODO: Show PC??

	clockAndInstr := lipgloss.JoinHorizontal(lipgloss.Center, clock, pc, whitespace, lastInstr)
	column1 := lipgloss.JoinVertical(lipgloss.Top, pipeline, clockAndInstr)
	regsCol := lipgloss.JoinHorizontal(lipgloss.Left, registerView, whitespace, column1)
	together := lipgloss.JoinHorizontal(lipgloss.Top, regsCol, whitespace, cache, ram)
	ui := lipgloss.JoinVertical(lipgloss.Left, together, cmdLine)

	return "\n" + ui //+ "\n" + "---- ctrl+c or q to quit ----" + "\n"
}

func (m model) drawRAM() string {
	rows := getRAMRows(m.system.RAM)
	ramTable := table.New().Border(lipgloss.NormalBorder()).Rows(rows...)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("63")).Render("RAM\n" + ramTable.Render())
}

func (m model) drawCache() string {
	headers := []string{"Tag", "Index", "Data", "Valid", "LRU"}
	rows := getCacheRows(m.system.Cache)
	cacheTable := table.New().Border(lipgloss.NormalBorder()).Headers(headers...).Rows(rows...)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("129")).Render("Cache\n" + cacheTable.Render())
}

func (m model) drawPipeline() string {
	// TODO: create pipeline instance along with cpu
	//labels := []string{" Fetch ", " Decode ", " Execute ", " Memory ", " Writeback "}
	labels := []string{" WB ", " MEM ", " EXE ", " DEC ", " FET "}
	row := make([]string, 0)
	for i := range m.system.CPU.Pipeline.Stages {
		row = append(row, m.system.CPU.Pipeline.Stages[i].FormatInstruction())
	}
	// TODO: show stage result in row?? SUCCESS, STALL, FAILURE, NOOP, etc.
	pipelineTable := table.New().Border(lipgloss.NormalBorder()).Headers(labels...).Rows(row)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("33")).Render("Pipeline\n" + pipelineTable.Render())
}

func (m model) drawRegisters() string {
	// TODO: create a cpu instance and make int registers
	rows := getRegVals(m.system.CPU)
	regTable := table.New().Border(lipgloss.NormalBorder()).Rows(rows...)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render("IntRegisters\n" + regTable.Render() + "\n")
}

func (m model) drawLastInstruction() string {
	headers := []string{"Last Input"}
	row := []string{m.lastInstr}
	instrTable := table.New().Border(lipgloss.NormalBorder()).Headers(headers...).Rows(row)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render(instrTable.Render())
}

func (m model) drawClock() string {
	header := []string{"Clock"}
	row := []string{fmt.Sprintf("%d", m.system.CPU.Clock)}
	clockTable := table.New().Border(lipgloss.NormalBorder()).Headers(header...).Rows(row)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render(clockTable.Render())
}

func (m model) drawPC() string {
	header := []string{"PC"}
	row := []string{fmt.Sprintf("%d", m.system.CPU.ProgramCounter)}
	clockTable := table.New().Border(lipgloss.NormalBorder()).Headers(header...).Rows(row)
	return lipgloss.NewStyle().BorderForeground(lipgloss.Color("207")).Render(clockTable.Render())
}
