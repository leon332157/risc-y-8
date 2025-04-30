ldi r1, 0xbeef
ldi sp, 16
ldi r2, 2048
ldi r3,5
ldi r4, 0x28
push r1
sub r2, 1
cmp r2, 1
bne [r3]
nop
nop
nop
hlt