ldi r5, 4
ldi r6, 9
ldi r2, 1
add sp, 20
add r4, 2
sub r2, 1
cmp r2, 0 
beq [r6]
bne [r5]
or r1, 0xbeef
#meow
push r1
push r1
push r1
pop r8
nop
nop
nop
meow