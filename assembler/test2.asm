        xor r1, r1
        xor r2, r2
        xor r3, r3
        stw r4, 4
        stw r1, 0xff
        stw r2, 256
        stw r3, [r3 + 0x10]
        cpy r4, r3
        push r5
        pop r6
        ldw r7, [r8 + 0x20]
        nop
        cmp r9, 0x30
        bne [r1 + 0x800]
        cmp r10, 0
        beq [r3]
        not r11, r12
        and r13, 0x1
        orr r14, 2
        rem r15, r16