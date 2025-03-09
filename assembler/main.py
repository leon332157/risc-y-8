from pyleri import (
    Grammar,
    Keyword,
    Regex,
    Optional,
    Choice,
    Ref,
    Sequence,
    List,
    Tokens,
    Prio,
    Repeat,
)


from pyleri import end_of_statement
import pprint
from re import IGNORECASE

WORDSIZE = 4  # bytes


class ASMGrammar(Grammar):
    NEW_LINE = Regex(r"(\r\n|\r|\n)")
    COMMA = ","

    """
    Base set GP registers: r0 (zero), r1-r16, bp, sp, pc
    """
    REG_ZERO = Regex(r"^r0", IGNORECASE)
    "0x00"
    REG_GP = Regex(r"^r[1-9][0-6]?", IGNORECASE)
    "0x01 - 0x10 respectively"
    REG_BP = Regex(r"^bp|^rbp", IGNORECASE)
    "0x0E"
    REG_SP = Regex(r"sp|rsp", IGNORECASE)
    "0x0F"
    REG_LR = Regex(r"lr", IGNORECASE)
    "0x10"
    REG_INT_FLAG = Regex(r"rflag", IGNORECASE)
    "0x11"
    REG_PC = Regex(r"pc", IGNORECASE)
    "0x1F"
    REG_ALL_GP = Choice(
        REG_ZERO, REG_GP, REG_BP, REG_SP, REG_LR, REG_PC, most_greedy=False
    )
    "all registers in base set"

    REG_FP = Regex(r"f[1-8]", IGNORECASE)
    "FP registers: f1-f8"

    REG_VEC = Regex(r"v[1-8]", IGNORECASE)
    "Vector registers: v1-v8"
    REG_VFLAG = Regex(r"vflag", IGNORECASE)  # 0x1D
    REG_VTYPE = Regex(r"vtype", IGNORECASE)  # 0x1E

    REG_ALL = Choice(
        REG_ALL_GP,
        REG_FP,
        REG_VEC,
        REG_INT_FLAG,
        REG_VFLAG,
        REG_VTYPE,
        most_greedy=False,
    )
    "all registers"

    """
    Immediate notation: 0x1234 (hex), 1234 (decimal)
    """
    IMM_HEX = Regex(r"^0x?[0-9a-f]+", IGNORECASE)  # hex
    IMM_DEC = Regex(r"^-?[0-9]+")  # decimal
    IMM = Choice(
        IMM_HEX, IMM_DEC, most_greedy=False
    )  # Immediate notation: 0x1234 or 1234 (hex or decimal)

    """
    Memory notation: [0x100],[r1], [r1+0x10]
    """
    MEM_IMM = Sequence("[", IMM, "]")
    MEM_REG = Sequence("[", REG_ALL_GP, "]")
    MEM_REG_DISP = Sequence(Sequence("[", REG_ALL_GP, Tokens("+ -"), IMM, "]"))
    MEM = Choice(MEM_IMM, MEM_REG, MEM_REG_DISP, most_greedy=False)
    "all memory notations"

    """
    Instructions
    """
    INST = Regex(r"^[a-z]+", IGNORECASE)
    INST_ONLY = Sequence(INST)
    INST_REG = Sequence(INST, REG_ALL)
    INST_MEM = Sequence(INST, MEM)
    INST_IMM = Sequence(INST, IMM)
    INST_REG_IMM = Sequence(INST, REG_ALL_GP, COMMA, IMM)
    INST_REG_REG = Sequence(INST, REG_ALL_GP, COMMA, REG_ALL_GP)
    INST_REG_MEM = Sequence(INST, REG_ALL_GP, COMMA, MEM)
    INST_BASE = Choice(
        INST_ONLY,
        INST_REG,
        INST_MEM,
        INST_IMM,
        INST_REG_IMM,
        INST_REG_REG,
        INST_REG_MEM,
        most_greedy=True
    )

    INST_FR = Sequence(INST, REG_FP)
    INST_FR_FR = Sequence(INST, REG_FP, COMMA, REG_FP)
    INST_FR_FR_FR = Sequence(INST, REG_FP, COMMA, REG_FP, COMMA, REG_FP)
    INST_FR_MEM = Sequence(INST, REG_FP, COMMA, MEM)
    INST_FR_FR_MEM = Sequence(INST, REG_FP, COMMA, REG_FP, COMMA, MEM)
    INST_FLOAT = Choice(
        INST_FR,
        INST_FR_FR,
        INST_FR_FR_FR,
        INST_FR_MEM,
        INST_FR_FR_MEM,
        most_greedy=True
    )
    
    VEC_TYPE = Regex(r"i|f|d|b", IGNORECASE)
    INST_VREG = Sequence(INST, REG_VEC)
    INST_VT_VREG_MEM =  Sequence(INST,'.',VEC_TYPE, REG_VEC, COMMA, COMMA, MEM)
    INST_VREG_VREG_VREG = Sequence(INST, REG_VEC, COMMA, REG_VEC, COMMA, REG_VEC)
    INST_VREG_VREG_IMM = Sequence(INST, REG_VEC, COMMA, REG_VEC, COMMA, IMM)
    INST_VEC = Choice(
        INST_VREG,
        INST_VT_VREG_MEM,
        INST_VREG_VREG_VREG,
        INST_VREG_VREG_IMM,
        most_greedy=True
    )
    
    INSTRUCTION = Choice(
        INST_BASE,
        INST_FLOAT,
        INST_VEC,
        most_greedy=True,
    )
    LABEL = Regex(r"\.[a-zA-Z]+")
    DIRECTIVES = Sequence(LABEL, Optional(IMM))
    # ORIGIN = Sequence(Regex(r".org", IGNORECASE), Optional(IMM))
    CODE_LABEL = Sequence(LABEL,":")
    # LINES = Ref()
    # LINE_DIRECTIVE = Sequence(DIRECTIVES,NEW_LINE)
    # LINE_LABEL = Sequence(LABEL, NEW_LINE)
    LINE_COMMENT = Regex(r"^#.*")
    LINE = Choice(LINE_COMMENT, INSTRUCTION, most_greedy=False)

    START = Repeat(LINE)  # Optional(ORIGIN), LINES)


class Operand:
    def __init__(self, value, size=WORDSIZE):
        self.size = size
        if not isinstance(value, int):
            raise TypeError("Value of operand must be an integer")
        if value > 2 ** (size * 8):
            raise ValueError("Value of operand is too large")
        self.value = value


class InstructionEncoding:
    def __init__(self, inst: Operand, op1: Operand, op2: Operand):
        self.inst = inst

    def encode(self) -> bytearray:
        pass


asm_grammar = ASMGrammar()


def print_expecting(node_expecting, string_expecting):
    for loop, e in enumerate(node_expecting):
        string_expecting = "{}\n\t({}) {}".format(string_expecting, loop, e)
    print(string_expecting)


# Complete a string until it is valid according to the grammar.
def auto_correction(string, my_grammar):
    node = my_grammar.parse(string)
    print("\nParsed string: {}".format(node.tree.string))

    if node.is_valid:
        string_expecting = "String is valid. \nExpected: "
        print_expecting(node.expecting, string_expecting)

    else:
        string_expecting = (
            "String is NOT valid.\nExpected: "
            if not node.pos
            else 'String is NOT valid. \nAfter "{}" {} expected: '.format(
                node.tree.string[: node.pos],node.pos
            )
        )
        print_expecting(node.expecting, string_expecting)
        # auto_correction(string, my_grammar)


# Returns properties of a node object as a dictionary:
def node_props(node, children):
    return {
        "start": node.start,
        "end": node.end,
        "name": node.element.name if hasattr(node.element, "name") else None,
        "element": node.element.__class__.__name__,
        "string": node.string,
        "children": children,
    }

def parse_node(node):
    return {
        "name": node.element.name if hasattr(node.element, "name") else None,
        "element": node.element.__class__.__name__,
        "string": node.string,
    }
LINES = []

# Recursive method to get the children of a node object:
def get_children(children):
    for c in children:
      if c:
        parsed = parse_node(c)
        if parsed['element'] not in ['Choice','Token']:
            if parsed['name'] != "LINE_COMMENT":
               LINES.append(parsed)
    return [node_props(c, get_children(c.children)) for c in children]


# View the parse tree:
def view_parse_tree(res):
    start = res.tree.children[0] if res.tree.children else res.tree
    return node_props(start, get_children(start.children))


TEST = """
#.ORG 0x1000
#xor r1, r1
add r1, 1
xor r2, r2
add r2, 2
add r1, r2
ldr r2, [r1+10]
ldr r2, [r1-0x20]
cmp r2, 0x10
nop
xor r3,r3
orr r3, 0x3f80
shl r3, 16
push r3
fldr f1, [sp]
pop r3
xor r3, r3
add r3, 0x4
shl r3, 28
push r3
fldr f2, [sp]
fadd f3, f1, f2
fmul f4, f3, f2
fsqrt f5, f4
bunc 0x1000
hlt
VZEROALL v0
VZEROUPPER v1
VBEQ v2, v3, -10
"""

gmr = asm_grammar.parse(TEST)

print(asm_grammar.export_go())
pp = pprint.PrettyPrinter(indent=1, sort_dicts=True, compact=True)
pp.pprint(view_parse_tree(gmr))

for each in LINES:
    match each['name']:
        case 'INSTRUCTION':
            print("Instruction: ", each['string'])
        case 'LABEL':
            print("Label: ", each['string'])
        case 'DIRECTIVES':
            print("Directive: ", each['string'])
        case _:
            print("Unknown: ", each['string'])
if not gmr.is_valid:
    auto_correction(TEST, asm_grammar)
