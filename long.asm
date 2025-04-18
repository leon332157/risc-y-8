ldi r5, 3
ldi r6, 7
ldi r2, 10
add r4, 2
sub r2, 1
cmp r2, 0 
beq [r6]
bne [r5]
or r1, 0xbeef
xor r1, 0xdead
nsa r2, r1
add r1, r2
sub r1, r2
shl r1, 2
shr r1, 2
shl r1, 2
shr r1, 2
shl r1, 2
shr r1, 2
ror r1, 2
rol r1, 2
