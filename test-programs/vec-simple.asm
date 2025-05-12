ldx sp, 0x20
ldi r1, 1
push r1
ldi r1, 2
push r1
ldi r1, 3
push r1
ldi r1, 4
push r1
vldp v1, [sp-4]
ldi r1, 5
push r1
ldi r1, 6
push r1
ldi r1, 7
push r1
ldi r1, 8
vldp v2, [sp-4]
vpadd v3, v1, v2
nop
nop
nop
nop
meow