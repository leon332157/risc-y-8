# Matrix multiplication calculator (can be useful in checking answers)

def createMatrix(size):
    matrix = []
    row = []
    for i in range(size):
        row.append(i)
    for i in range(size):
        matrix.append(row)
    return matrix

# assumes square matrices
def multMatrix(m1, m2):
    size = len(m1)
    result = [[0 for _ in range(size)] for _ in range(size)]

    for idx in range(size):
        for column in range(size):
            for row in range(size):
                result[idx][column] += m1[idx][row] * m2[row][column]

    return result    

def printMatrix(m):
    rows = len(m)
    for row in range(rows):
        print(m[row])
    print("\n")

A = createMatrix(50)
B = createMatrix(50)
C = multMatrix(A, B)
printMatrix(A)
printMatrix(B)
printMatrix(C)
    
    