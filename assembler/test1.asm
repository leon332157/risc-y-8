
		#xor r1, r1
		hlt
		nop
		add r1, 1
		xor r2, r2
		add r2, -2
		add r1, r2
		ldw r2, [r1+10]
		ldw r2, [r1-0x20]
		bunc [r13]
		cmp r2, 0x10
		nop
		xor r3,r3
		orr r3, 0x3f80
		shl r3, 16
		push r3