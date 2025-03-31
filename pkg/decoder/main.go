package decoder

import (
	"github.com/leon332157/risc-y-8/pkg/types"
)

var dtof types.DecodeToFetch

func DecodeStageToExecute(encoded types.FetchToDecode) types.DecodeToExe {

	dtoexe := types.DecodeToExe{}

	inst := types.BaseInstruction{}
	inst.Decode(encoded.MemInst)
	dtoexe.Instruction = inst
	dtof.Success = true

	return dtoexe
}

func DecodeStageToFetch() types.DecodeToFetch{
	return dtof
}