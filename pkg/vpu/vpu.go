package vpu

func VPAdd(op1, op2 [4]uint32) [4]uint32 {
	return [4]uint32{
		op1[0] + op2[0],
		op1[1] + op2[1],
		op1[2] + op2[2],
		op1[3] + op2[3],
	}
}

func VPSub(op1, op2 [4]uint32) [4]uint32 {
	return [4]uint32{
		op1[0] - op2[0],
		op1[1] - op2[1],
		op1[2] - op2[2],
		op1[3] - op2[3],
	}
}

func VPMul(op1, op2 [4]uint32) [4]uint32 {
	return [4]uint32{
		op1[0] * op2[0],
		op1[1] * op2[1],
		op1[2] * op2[2],
		op1[3] * op2[3],
	}
}

func VPShl(op1, op2 [4]uint32) [4]uint32 {
	return [4]uint32{
		op1[0] << op2[0],
		op1[1] << op2[1],
		op1[2] << op2[2],
		op1[3] << op2[3],
	}
}

func VPXor(op1, op2 [4]uint32) [4]uint32 {
	return [4]uint32{
		op1[0] ^ op2[0],
		op1[1] ^ op2[1],
		op1[2] ^ op2[2],
		op1[3] ^ op2[3],
	}
}

func VPAnd(op1, op2 [4]uint32) [4]uint32 {
	return [4]uint32{
		op1[0] & op2[0],
		op1[1] & op2[1],
		op1[2] & op2[2],
		op1[3] & op2[3],
	}
}

func VPOr(op1, op2 [4]uint32) [4]uint32 {
	return [4]uint32{
		op1[0] | op2[0],
		op1[1] | op2[1],
		op1[2] | op2[2],
		op1[3] | op2[3],
	}
}
