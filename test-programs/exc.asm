ldi r1, 100# max array element
ldi bp, 0x50# array base address
mov sp, bp# store base address of array at 50
push r1
sub r1, 20# push array elements to stack starting at 100 (r1) at increments of 20
ldi r10, 3
cmp r1, 0
bne [r10]
xor r2, r2
ldi r3, 3
ldi r11, 36# OUTER LOOP START
cmp r3, r2
blt [r11]
ldi r4, 0
ldi r5, 3
sub r5, r2
ldi r12, 33# INNER LOOP START
cmp r5, r4
blt [r12]
mov r6, bp
add r6, r4
ldw r8, [r6]
mov r7, r6
add r7, 1
ldw r9, [r7]
ldi r13, 30
cmp r8, r9
blt [r13]
stw r8, [r7]
stw r9, [r6]
add r4, 1
ldi r14, 16
bunc [r14]
add r2, 1# END INNER LOOP
ldi r15, 10
bunc [r15]
nop
nop
nop
nop
nop