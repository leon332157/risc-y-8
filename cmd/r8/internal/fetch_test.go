package internal

import (
	"testing"

	"github.com/leon332157/risc-y-8/pkg/cpu"
)

func TestFetchEmptyIR(t *testing.T) {

	pipeline := &cpu.Pipeline{}
	fetchStage := &cpu.FetchStage{}
	decodeStage := &cpu.DecodeStage{}

	fetchStage.Init(pipeline, decodeStage, nil)
	fetchStage.Execute()
	fetchStage.Advance(&cpu.InstructionIR{}, false)

}

func TestBadPipeline(t *testing.T) {

	fetchStage := &cpu.FetchStage{}

	fetchStage.Init(nil, nil, nil)
	fetchStage.Execute()

}

func TestFetchSkipExecute(t *testing.T) {

	// fails if advance panics
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	fetchStage := &cpu.FetchStage{}
	fetchStage.Advance(&cpu.InstructionIR{}, false)
}

func TestFetchName(t *testing.T) {

	fetchStage := &cpu.FetchStage{}

	if fetchStage.Name() != "Fetch" {
		t.Errorf("Expected Fetch, got %s", fetchStage.Name())
	}

}