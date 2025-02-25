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

WORDSIZE = 4 # bytes

class ASMGrammar(Grammar):

    NEW_LINE = Regex(r"(\r\n|\r|\n)")
    COMMA = ","
    # LABEL = Regex(r'([a-zA-Z_][a-zA-Z0-9_]*):')
    # END = Regex(r'')
    INST = Regex(r"[a-z]{3}", IGNORECASE)
    REG_DEST = Regex(r"R[0-9]+", IGNORECASE)
    REG_SRC = Regex(r"R[0-9]+", IGNORECASE)
    IMM_HEX = Regex(r"^(0x)?[0-9a-f]+", IGNORECASE)
    IMM_DEC = Regex(r"^-?[0-9]+")
    IMM = Choice(IMM_HEX, IMM_DEC, most_greedy=False)

    DIRECTIVES = Sequence(Regex(r".[a-zA-Z]+", IGNORECASE), Optional(IMM))
    ORIGIN = Sequence(Regex(r".org", IGNORECASE), Optional(IMM))

    INSTR_REG_REG = Sequence(INST, REG_DEST, COMMA, REG_SRC)
    INSTR_IMM_REG = Sequence(INST, REG_DEST, COMMA, IMM)
    INSTR_CTRL = Sequence(INST, Optional(Choice(IMM, REG_DEST, most_greedy=False)))
    INSTR_MEM = Sequence(
        INST, REG_DEST, Choice(IMM, REG_SRC, most_greedy=False), Optional(IMM)
    )
    INSTR_MEM_BRACKET = Sequence(
        INST, REG_DEST, COMMA, Sequence("[", REG_SRC, Tokens("+ - *"), IMM, "]")
    )

    INSTRUCTION = Choice(
        INSTR_REG_REG, INSTR_IMM_REG, INSTR_CTRL, INSTR_MEM, INSTR_MEM_BRACKET
    )
    # LINES = Ref()
    LINE = Sequence(
        Choice(DIRECTIVES, INSTRUCTION, most_greedy=False), Optional(NEW_LINE)
    )
    LINES = Repeat(LINE)
    START = Sequence(Optional(ORIGIN), LINES)

class Operand:
    def __init__(self, value, size = WORDSIZE):
        self.size = size
        if not isinstance(value, int):
            raise TypeError("Value of operand must be an integer")
        if value > 2**(size*8):
            raise ValueError("Value of operand is too large")
        self.value = value
        
class InstructionEncoding:
    def __init__(self,inst:Operand, op1:Operand, op2:Operand):
        self.inst = inst
    def encode(self)->bytes[4]:
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
            else 'String is NOT valid. \nAfter "{}" expected: '.format(
                node.tree.string[: node.pos]
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


# Recursive method to get the children of a node object:
def get_children(children):
    return [node_props(c, get_children(c.children)) for c in children]


# View the parse tree:
def view_parse_tree(res):
    start = res.tree.children[0] if res.tree.children else res.tree
    return node_props(start, get_children(start.children))


TEST = """
.ORG 0x1000
add r1, 1
"""

gmr = asm_grammar.parse(TEST)

# print(asm_grammar.export_go())
pp = pprint.PrettyPrinter(indent=1, sort_dicts=True, compact=True)
pp.pprint(view_parse_tree(gmr))


def instruction_matcher():
    pass


if not gmr.is_valid:
    auto_correction(TEST, asm_grammar)
