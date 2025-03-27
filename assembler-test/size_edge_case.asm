        xor r16, r16
        add r16, 0xffff
        add r16, 1
        nop
        xor r15, r15
        sub r15, 0x1
        nop
        xor r14, r14
        add r14, 0x0
        nop
        xor r13, r13
        add r13, 0xffffff
        nop
        xor r12, r12
        mul r12, r16
        rem r12, r16
        nop
        xor r11, r11
        ldw r11, [r11 + 0xffff]
        nop
        xor r10, r10
        stw r10, [r10 - 0xffff]
        nop
        xor r9, r9
        ldw r9, [r9 + 0xffffff]
        nop
        xor r8, r8
        stw r8, [r8 + 0xffffff]
        nop
        xor r7, r7
        ldw r7, [r7 + 0x0]
        nop
        xor r6, r6
        stw r6, [r6 - 0x1]
        nop
        xor r5, r5
        BEQ[r5]
        nop
        xor r4, r4
        BNE[r4 + 0xffff]
        nop
        xor r3, r3
        BLT[r3 + 0xffffff]
        nop
        xor r2, r2
        BGE[r2 + 0x0]
        nop
        xor r1, r1
        BL[r1 - 0x1]