ldi sp, 30
ldi r2, 4096
ldi r3, 15
ldi r4, 9
ldi r1, 0xdead
shl r1, 16
ldi r20, 0xbeef
or r1, r20
xor r20,r20
push r1
sub r2,1
cmp r2, 0
beq [r3]
bne [r4]
nop
hlt
nop