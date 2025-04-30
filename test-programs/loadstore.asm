ldi r2, 30
ldi r3, 2
stw r2, [r2]
add r2, 1
cmp r2, 512
bne [r3]
nop
ldi r2, 30
ldi r4, 0x201
ldi r3, 10
ldw r5, [r2]
stw r5, [r4]
add r2, 1
add r4, 1
cmp r2, 0x300
blt [r3]
nop
nop
nop
hlt