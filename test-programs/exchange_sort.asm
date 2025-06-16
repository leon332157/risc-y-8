ldi r1, 100 # max array element
mov r16, r1
sub r16, 2
ldi bp, 0x50 # array base address
mov sp, bp # store base address of array at 50
push r1
sub r1, 1 # push array elements to stack starting at 100 (r1) at increments of 20
ldi r10, 5
cmp r1, 0
bne [r10]
xor r2, r2
mov r3, r16
ldi r11, 38 # OUTER LOOP START
cmp r3, r2
blt [r11]
ldi r4, 0
mov r5, r16
sub r5, r2
ldi r12, 35 # INNER LOOP START
cmp r5, r4
blt [r12]
mov r6, bp
add r6, r4
ldw r8, [r6]
mov r7, r6
add r7, 1
ldw r9, [r7]
ldi r13, 32
cmp r8, r9
blt [r13]
stw r8, [r7]
stw r9, [r6]
add r4, 1
ldi r14, 18
bunc [r14]
add r2, 1 # END INNER LOOP
ldi r15, 12
bunc [r15]
nop
nop
nop
nop
nop