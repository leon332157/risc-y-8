package decoder

import (
	"github.com/leon332157/risc-y-8/pkg/types"
)

func DecodeInstruction(encoded uint32) types.BaseInstruction {

	inst := types.BaseInstruction{}
	inst.Decode(encoded)

	return inst
}