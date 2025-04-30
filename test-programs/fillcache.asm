ldi r1, 0xbeef
ldi sp, 20
ldi r2, 2048
ldi r3,4
push r2
sub r2, 1
cmp r2, 1
bne [r3]
ldi r2, 30
ldi r3, 10
pop r1
add r2, 1
cmp r2, 2040
bne [r3]
nop
nop
nop
hlt