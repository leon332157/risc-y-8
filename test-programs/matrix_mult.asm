ldi r1, 50# matrix size
ldi r30, 160# base address for matrices
mov r2, r1
mul r2, r2# max value for elements r1 * r1
mov r17, r2# r17 = size * size
mul r2, 2# for two matrices
mov r20, r30# matrix A base address
mov r21, r20# matrix B base address
add r21, r17# matrix B base address = matrix A base address + size * 2
mov r22, r21# matrix C base address
add r22, r17# matrix C base address = matrix B base address + size * 2
mov sp, r22# store base address of matrix C
xor r28, r28# matrix count register, counts up to 2500
mov r24, r20# set r24 to base address of matrix A
stw r28, [r24]# populate matrices | store count at matrix A base address + count
add r28, 1# increment count
add r24, 1# increment base address to next element
ldi r10, 14# jump to start of populate matrices
cmp r28, r2#
blt [r10]#
xor r3 r3# i = 0
ldi r11, 61# LOOP_I | r11 = END_I
cmp r3, r1# compare i with matrix size
bge [r11]# branch if greater than or equal to
xor r4, r4# j = 0
ldi r12, 57# LOOP_J | r12 = END_J
cmp r4, r1# compare j with matrix size
bge [r12]# branch if greater than or equal to
xor r5, r5# accumulator = 0
xor r6, r6# k = 0
ldi r13, 49# LOOP_K | r13 = END_K
cmp r6, r1# compare k with matrix size
bge [r13]# branch if greater than or equal to
mov r7, r3# compute A[i][k] = base_A + (i * 50 + k)
mul r7, r1# r7 = i * 50
add r7, r6# r7 = i * 50 + k
add r7, r20# r7 = base_A + (i * 50 + k)
ldw r8, [r7]# load A[i][k] into r8
mov r9, r6# compute B[k][j] = base_B + (k * 50 + j)
mul r9, r1# r9 = k * 50
add r9, r4# r9 = k * 50 + j
add r9, r21# r9 = base_B + (k * 50 + j)
ldw r18, [r9]# load B[k][j] into r18
mov r25, r8# multiply and accumulate | move A[i][k] to r25
mul r25, r18# r25 = A[i][k] * B[k][j]
add r5, r25# accumulator += A[i][k] * B[k][j]
add r6, 1# increment k
ldi r14, 30# r14 = LOOP_K
bunc [r14]# branch to LOOP_K 
mov r26, r3# END_K | compute C[i][j] = base_C + (i * 50 + j)
mul r26, r1# r26 = i * 50
add r26, r4# r26 = i * 50 + j
add r26, r22# r26 = base_C + (i * 50 + j)
stw r5, [r26]# store accumulator in C[i][j]
add r4, 1# increment j
ldi r15, 25# r15 = LOOP_J
bunc [r15]# branch to LOOP_J
add r3, 1# END_J | increment i
ldi r16, 21# r16 = LOOP_I
bunc [r16]# branch to LOOP_I
hlt# END_I | end of program
nop
nop
nop