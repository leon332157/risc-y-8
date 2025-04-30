ldi r1, 100# max array element
ldi bp, 0x50# array base address
mov sp, bp# store base address of array at 50
push r1
sub r1, 20# push array elements to stack starting at 100 (r1) at increments of 20
ldi r10, 3
cmp r1, 0
bne [r10]
xor r2, r2
ldi r3, 2
ldi r11, 34# OUTER LOOP START
cmp r3, r2
blt [r11]
ldi r4, 1
ldi r5, 3
ldi r12, 31# INNER LOOP START
cmp r5, r4
blt [r12]
mov r6, bp
add r6, r4
ldw r8, [r6]
add r6, 1
ldw r9, [r6]
ldi r13, 28
cmp r8, r9
blt [r13]
stw r7, [r8]# store current element value at next element's address
stw r6, [r9]# store next element value at current element's address
add r4, 1
ldi r14, 15
bunc [r14]
add r2, 1# END INNER LOOP
ldi r15, 10
bunc [r15]
hlt# END OUTER LOOP