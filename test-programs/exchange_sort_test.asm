ldi r1, 100
ldi bp, 0x50# array base address
mov sp, bp# store base address of array at 1000
push r1
sub r1, 20# push array elements to stack starting at 100 (r1) at increments of 20
ldi r10, 3# jump to line 5
cmp r1, 0
bne [r10]
xor r2, r2# zero out outer count
ldi r3, 3 # outer loop limit = size - 2
cmp r3, r2# compare limit vs outer count
ldi r20, 32
beq [r20]
cpy r4, r2# r4 = j
add r4, 1#j=i+1
ldi r5, 4#inner loop size - 1
cpy r12,bp
add r12,r2
cpy r14,bp
add r14,r4
ldw r8, [r12]#i
ldw r9, [r14]#j
cmp r9, r8
ldi r20,30#skip
blt [r20]#skip 
stw r9, [r12]#i
stw r8, [r14]#j
add r4, 1
cmp r4, r5
ldi r20,13
blt [r20]
add r2, 1# END INNER LOOP
ldi r20,10 
bunc [r20]
hlt