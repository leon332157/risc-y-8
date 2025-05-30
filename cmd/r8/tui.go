package r8

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"math/bits"
	"os"
	"strconv"
	"strings"

	"github.com/leon332157/risc-y-8/cmd/r8/simulator"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	table "github.com/charmbracelet/lipgloss/table"
	"github.com/leon332157/risc-y-8/pkg/cpu"
	"github.com/leon332157/risc-y-8/pkg/memory"
)

var desiredHeight = 9
var Message string = "none"

var (
	tuiCmd = &cobra.Command{
		Use:     "tui <flags> [binary file]",
		Short:   "Simulate with TUI RISC-Y-8 binary",
		RunE:    runTui,
		Args:    cobra.ExactArgs(1),
		Example: "r8 tui input.bin",
	}
	disableCache    bool
	disablePipeline bool
	NumInstructions = 0
)

func init() {
	tuiCmd.Flags().BoolVar(&disableCache, "disable-cache", false, "Disable cache")
	tuiCmd.Flags().BoolVar(&disablePipeline, "disable-pipeline", false, "Disable pipeline")
	rootCmd.AddCommand(tuiCmd)
}

func runTui(cmd *cobra.Command, args []string) error {
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
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
	NumInstructions = len(buffer) / 4
	err = binary.Read(bytes.NewReader(buffer), binary.LittleEndian, &program)
	if err != nil {
		return fmt.Errorf("failed to read input file: %v", err)
	}
	system := simulator.NewSystem(program, disableCache, disablePipeline)
	model := initialModel(&system)
	// model.system = &system
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

	system              *simulator.System
	ramViewport         viewport.Model
	cacheViewport       viewport.Model
	cacheHeaderViewport viewport.Model
}

func initialModel(s *simulator.System) model {
	ti := textinput.New()
	ti.Placeholder = "type a command . . ."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 50
	ti.Cursor.SetMode(cursor.CursorStatic)

	tableHeight := 34
	headerSize := 3

	// +2 for 0x prefix
	// +1 for left vertical line
	ramLinesSize := uint(math.Log(float64(s.RAM.SizeWords()))/math.Log(16)) + 2 + 1 + 1
	ramDataSize := s.RAM.WordsPerLine*10 + s.RAM.WordsPerLine + 1
	ramVPWidth := ramDataSize + ramLinesSize
	ramVP := viewport.New(int(ramVPWidth), tableHeight)

	offsetBits := bits.Len32(uint32(s.Cache.WordsPerLine)) - 1
	indexBits := bits.Len32(uint32(s.Cache.Sets)) - 1
	// memSize := s.RAM.SizeWords()
	totalBits := 32

	sizeTag := max(uint(totalBits-indexBits-int(offsetBits)), 3) + 2
	sizeIndex := max(uint(indexBits), 3) + 1
	sizeData := (s.Cache.WordsPerLine * 8) + (s.Cache.WordsPerLine - 1) + 2 + 1
	sizeValid := uint(5 + 1)
	sizeLRU := uint(max(math.Log2(float64(s.Cache.Ways)), 3)) + 1

	cacheVPWidth := sizeTag + sizeIndex + sizeData + sizeValid + sizeLRU + 5

	cacheHeaderVP := viewport.New(int(cacheVPWidth), headerSize)
	cacheVP := viewport.New(int(cacheVPWidth), tableHeight-headerSize)
	return model{
		instr:               ti,
		lastInstr:           "",
		system:              s,
		ramViewport:         ramVP,
		cacheViewport:       cacheVP,
		cacheHeaderViewport: cacheHeaderVP,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func getCacheRows(ca *memory.CacheType) [][]string {
	cRows := [][]string{}

	offsetBits := bits.Len32(uint32(ca.WordsPerLine)) - 1
	indexBits := bits.Len32(uint32(ca.Sets)) - 1
	//memSize := ca.LowerLevel.SizeWords()
	//totalBits := int(math.Log2(float64(memSize)))
	totalBits := 32
	sizeTag := uint(totalBits - indexBits - int(offsetBits)) //+ 1
	sizeIndex := uint(indexBits)                             //+ 1

	tagStr := fmt.Sprintf("%%0%db", sizeTag)
	idxStr := fmt.Sprintf("%%0%db", sizeIndex)

	//waysSize := max(math.Log2(float64(ca.Ways)), 3)
	//waysStr := fmt.Sprintf("%%%db", int(waysSize))

	for i := range ca.Contents {
		for j := range ca.Ways {
			data := ca.Contents[i][j]

			validStr := "%t"
			if data.Valid {
				validStr = validStr + strings.Repeat(" ", 1)
			}

			row := []string{
				fmt.Sprintf(tagStr, data.Tag),
				fmt.Sprintf(idxStr, i),
				fmt.Sprintf("%08X", data.Data),
				fmt.Sprintf(validStr, data.Valid),
				fmt.Sprintf(" %d", data.LRU)}
			cRows = append(cRows, row)
		}
	}
	return cRows
}

func getRAMRows(ram *memory.RAM) [][]string {
	rRows := [][]string{}
	addr := 0
	for i := range int(ram.NumLines) {
		row := []string{}
		row = append(row, fmt.Sprintf("0x%X", i*int(ram.WordsPerLine)))
		for range ram.WordsPerLine {
			row = append(row, fmt.Sprintf("0x%08X", ram.Contents[addr]))
			addr++
		}
		rRows = append(rRows, row)
	}
	return rRows
}

func getRegVals(control *cpu.CPU) [][]string {
	regVals := [][]string{}

	for i := range len(control.IntRegisters) {
		var style = lipgloss.NewStyle()
		if !control.IntRegisters[i].ReadEnable {
			style = style.Foreground(lipgloss.Color("#FF0000"))
		} else {
			style = style.Foreground(lipgloss.Color("#04B575"))
		}
		var row = []string{}
		if i == 0 {
			flags := make([]string, 4)
			if control.ALU.GetCF() {
				flags = append(flags, style.Foreground(lipgloss.Color("#CC6CE7")).Render("CF"))
			}
			if control.ALU.GetOVF() {
				flags = append(flags, style.Foreground(lipgloss.Color("#CC6CE7")).Render("OF"))
			}
			if control.ALU.GetSF() {
				flags = append(flags, style.Foreground(lipgloss.Color("#CC6CE7")).Render("SF"))
			}
			if control.ALU.GetZF() {
				flags = append(flags, style.Foreground(lipgloss.Color("#CC6CE7")).Render("ZF"))
			}
			row = []string{style.Render("Rflag"), strings.Join(flags, "")}
		} else {
			row = []string{style.Render(fmt.Sprintf("R%d", i)), fmt.Sprintf("%08X", control.ReadIntRNoBlock(uint8(i)))}
		}
		regVals = append(regVals, row)
	}
	return regVals
}

func (m *model) ExecuteCommand() {
	args := strings.Split(m.lastInstr, " ")

	switch args[0] {
	case "step", "s", "next", "n":
		if m.system.CPU.ProgramCounter >= uint32(NumInstructions)+6 {
			m.system.CPU.Halted = true
			Message = "Program finished"
			return
		}
		m.system.RunOneClock(nil)
		/*if !m.system.CPU.Halted {
			m.system.CPU.Pipeline.RunOneClock()
		} else {
			m.system.CPU.Halted = false
		}*/
	case "run", "r":
		if len(args) > 1 {
			if args[1] == "complete" {
				Message = "Running to end . . ."
				m.system.RunToEndTUI(nil)
				m.system.CPU.Halted = true
				Message = "Program finished"
				return
			}
			cycles, err := strconv.Atoi(args[1])
			if err != nil {
				Message = fmt.Sprintf("Invalid number of cycles %v", err)
				return
			}
			Message = fmt.Sprintf("Running for %d cycles", cycles)
			for _ = range cycles {
				if m.system.CPU.ProgramCounter >= uint32(NumInstructions)+6 {
					m.system.CPU.Halted = true
					Message = "Program finished"
					return
				}
				if !m.system.CPU.Halted {
					m.system.CPU.Pipeline.RunOneClock()
				} else {
					m.system.CPU.Halted = false
				}
			}
		} else {
			Message = "Invalid command, please use 'run <cycles>' or run complete"
		}
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
			m.ramViewport.SetContent(m.drawRAMTable())
			m.cacheViewport.SetContent(m.drawCacheBodyTable())
			// Send instruction to be computed
			//cache.Write(0x0, memory.FETCH_STAGE, 0xdeadbeef)
			m.instr.Reset()
			return m, nil
		case "d":
			m.ramViewport.ScrollDown(16)
			m.instr.Reset()
		case "f":
			m.ramViewport.ScrollUp(16)
			m.instr.Reset()
		case "j":
			m.cacheViewport.ScrollDown(8)
			m.instr.Reset()
		case "k":
			m.cacheViewport.ScrollUp(8)
			m.instr.Reset()
		}
	case tea.WindowSizeMsg:
		// handle resize if needed
		m.ramViewport.SetContent(m.drawRAMTable())
		m.cacheViewport.SetContent(m.drawCacheBodyTable())
		m.cacheHeaderViewport.SetContent(m.drawCacheHeaderTable())
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
	msg := m.drawMsg()
	lastInstr := m.drawLastInstruction()
	cmdLine := m.instr.View() + "\n"
	whitespace := lipgloss.Place(3, 3, lipgloss.Right, lipgloss.Bottom, "")
	SimAndCPU := lipgloss.JoinHorizontal(lipgloss.Center, clock, pc, whitespace, lastInstr, msg)
	pipelineAndCPU := lipgloss.JoinHorizontal(lipgloss.Top, pipeline, whitespace, SimAndCPU)
	regsCol := lipgloss.JoinHorizontal(lipgloss.Left, registerView, whitespace, ram, whitespace, cache)
	together := lipgloss.JoinVertical(lipgloss.Top, pipelineAndCPU, regsCol)
	ui := lipgloss.JoinVertical(lipgloss.Left, together, cmdLine)

	return "\n" + ui //+ "\n" + "---- ctrl+c or q to quit ----" + "\n"
}

func (m model) drawRAM() string {

	content := m.ramViewport.View()
	title := "RAM\n"
	style := lipgloss.NewStyle()

	if m.system.RAM.Requester() == memory.NONE {
		title = "RAM - FREE"
		style = style.Foreground(lipgloss.Color("#04B575"))
	} else {
		title = "RAM - BUSY " + fmt.Sprintf("%d cycles left", m.system.RAM.CyclesLeft)
		style = style.Foreground(lipgloss.Color("#FF0000"))
	}

	return style.Render(title) + "\n" + content
}

func (m model) drawRAMTable() string {

	rows := getRAMRows(m.system.RAM)

	ramTable := table.New().
		Border(lipgloss.NormalBorder()).
		Rows(rows...)

	return ramTable.Render()
}

func (m model) getCacheSize() []uint {

	offsetBits := bits.Len32(uint32(m.system.Cache.WordsPerLine)) - 1
	indexBits := bits.Len32(uint32(m.system.Cache.Sets)) - 1
	totalBits := 32

	sizeTag := max(uint(totalBits-indexBits-int(offsetBits)), 3)
	sizeIndex := max(uint(indexBits), 3)
	sizeData := (m.system.Cache.WordsPerLine * 8) + (m.system.Cache.WordsPerLine - 1) + 2

	return []uint{
		sizeTag,
		sizeIndex,
		sizeData,
	}
}

func (m model) drawCache() string {
	headerStr := m.cacheHeaderViewport.View()
	content := m.cacheViewport.View()
	style := lipgloss.NewStyle()
	title := "Cache"

	if m.system.Cache.Requester() == memory.NONE {
		title = "Cache - FREE"
		style = style.Foreground(lipgloss.Color("#04B575"))
	} else if m.system.Cache.MemoryRequestState.WaitNext {
		title = "Cache - WAITING ON RAM"
		style = style.Foreground(lipgloss.Color("#FFA500"))
	} else {
		title = "Cache - BUSY " + fmt.Sprintf("%d cycles left", m.system.Cache.CyclesLeft)
		style = style.Foreground(lipgloss.Color("#FF0000"))
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		style.Render(title),
		headerStr,
		content,
	)
}

func (m model) drawCacheHeaderTable() string {
	if m.system.Cache.Sets == 0 {
		return "Cache Disabled"
	}
	sizeInfo := m.getCacheSize()
	tagSize := sizeInfo[0]
	indexSize := sizeInfo[1]
	dataSize := sizeInfo[2]

	tagHeader := "Tag" + strings.Repeat(" ", int(tagSize)-3)
	indexHeader := "Idx" + strings.Repeat(" ", int(indexSize)-3)
	dataHeader := "Data" + strings.Repeat(" ", int(dataSize)-4)

	header := table.New().
		Headers(tagHeader, indexHeader, dataHeader, "Valid", "LRU").
		Border(lipgloss.NormalBorder())

	return header.Render()

}

func (m model) drawCacheBodyTable() string {

	rows := getCacheRows(m.system.Cache)

	cacheTable := table.New().
		Border(lipgloss.NormalBorder()).
		Rows(rows...)

	return cacheTable.Render()
}

// desiredHeight := 5 // or however many stages you want to display
// row := make([]string, 0, desiredHeight)

// for i := range m.system.CPU.Pipeline.Stages {
// 	row = append(row, m.system.CPU.Pipeline.Stages[i].FormatInstruction())
// }

// // Pad with "<bubble>" or empty instructions if needed
// for len(row) < desiredHeight {
// 	row = append(row, "<bubble>")
// }

func (m model) checkNewlines(instr string, height int, i int) int {

	if strings.Count(m.system.CPU.Pipeline.Stages[i].FormatInstruction(), "\n") > height {
		return 0
	}

	return height - strings.Count(m.system.CPU.Pipeline.Stages[i].FormatInstruction(), "\n")

}

func (m model) drawPipeline() string {
	
	labels := []string{" WB ", " MEM ", " EXE ", " DEC ", " FET "}
	row := make([]string, 0)

	for i := range m.system.CPU.Pipeline.Stages {
		row = append(row, m.system.CPU.Pipeline.Stages[i].FormatInstruction()+strings.Repeat("\n", m.checkNewlines(m.system.CPU.Pipeline.Stages[i].FormatInstruction(), desiredHeight, i)))
	}

	// TODO: show stage result in row?? SUCCESS, STALL, FAILURE, NOOP, etc.
	pipelineTable := table.New().
		Width(100).
		Border(lipgloss.NormalBorder()).
		Headers(labels...).
		Rows(row)

	return lipgloss.NewStyle().
		// Padding(10, 10).
		BorderForeground(lipgloss.Color("33")).
		Render("Pipeline\n" + pipelineTable.Render())
}

func (m model) drawRegisters() string {
	rows := getRegVals(m.system.CPU)

	regTable := table.New().
		Border(lipgloss.NormalBorder()).
		Rows(rows...)

	return lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("207")).
		Render("IntRegisters\n" + regTable.Render() + "\n")
}

func (m model) drawLastInstruction() string {

	headers := []string{"Last Input"}
	row := []string{m.lastInstr}

	instrTable := table.New().
		Border(lipgloss.NormalBorder()).
		Headers(headers...).
		Rows(row)

	return lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("207")).
		Render(instrTable.Render())
}

func (m *model) drawMsg() string {
	/* 	if m.system.CPU.Halted {
	   		m.msg = "CPU is halted"
	   	} else {
	   		m.msg = "CPU is running"
	   	} */
	headers := []string{"Message"}
	row := []string{Message}

	msgTable := table.New().
		Border(lipgloss.NormalBorder()).
		Headers(headers...).Rows(row)

	return lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("207")).
		Render(msgTable.Render())
}

func (m model) drawClock() string {
	header := []string{"Clock"}
	row := []string{fmt.Sprintf("%d", m.system.CPU.Clock)}

	clockTable := table.New().
		Border(lipgloss.NormalBorder()).
		Headers(header...).
		Rows(row)

	return lipgloss.NewStyle().
		Padding(desiredHeight/2, 0).
		BorderForeground(lipgloss.Color("207")).
		Render("CPU\n" + clockTable.Render())
}

func (m model) drawPC() string {

	header := []string{"PC", "Total"}
	row := []string{fmt.Sprintf("%d", m.system.CPU.ProgramCounter), fmt.Sprintf("%d", NumInstructions)}

	clockTable := table.New().
		Border(lipgloss.NormalBorder()).
		Headers(header...).
		Rows(row)

	return lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("207")).
		Render(clockTable.Render())
}
