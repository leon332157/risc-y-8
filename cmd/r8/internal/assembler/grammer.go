package assembler

import (
	"fmt"
	"log"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

// Define the lexer for extended assembly
var asmLexerDyn = lexer.MustStateful(lexer.Rules{
	"Root": {
		{`Line`, `(?i).*`, lexer.Push("Line")},
	},
	"Line": {
		{`whitespace`, `\s*`, nil},
		{`Instruction`, `(?i).`, lexer.Push("Instruction")},
		{"Directive", `^\.[a-z]+$`, nil},
		{"Label", `^\.[a-z]+:$`, nil},
		{"Comment", `(?im)^#.*$`, nil},
		{"lineEnd", `(?i)(\n|\r|\r\n)`, lexer.Pop()},
		//lexer.Return(),
	},

	"Instruction": {
		{"Mnemonic", `(?i)[a-z]{2,}`, nil},
		{"Operands", `.+`, lexer.Pop()},
		//{"Operands",`.`,lexer.Push("Operands")}
	},
})

var asmLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"Comment", `^#.*$`},
	{"Line", `(?im)^.+$`},
	{"Mnemonic", `[a-z]{2,}`},
	{"EOL", `[\n\r]+`},
	{"Whitespace", `\s*`},
})

var basicLexer = lexer.MustSimple([]lexer.SimpleRule{
	{"Comment", `#.*`},
	{"Directive", `\.[a-z]{1,}`},
	{"Label", `\.[a-z]{1,}[0-9]{1,}:`},
	{"String", `"(\\"|[^"])*"`},
	{"Punct", `[-[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
	{"Immediate", `0x?[0-9a-f]+|-?\d+`},
	{"Number", `[-+]?(\d*\.)?\d+`},
	{"Ident", `(?i)[a-z0-9]\w*`},
	{"EOL", `[\n\r]+`},
	{"whitespace", `[ \t]+`},
})

/*
type Directive struct {
	Type string `@Identifier`
}
*/

/*
	 type Instruction struct {
		Mnemonic  string   `@Mnemonic`
		Operands string `@Operands`
	}
*/
type comment struct {
	Pos lexer.Position

	Comment string `@Comment`
}

type Program struct {
	Pos lexer.Position

	Lines []Line `@@*`
}

type Line struct {
	Pos lexer.Position

	Index int

	Comment     string       `(   @Comment`
	Directive   string       `| @Directive`
	Label       string       `| @Label`
	Input       *Input       `  | @@`
	Let         *Let         `  | @@`
	Goto        *Goto        `  | @@`
	If          *If          `  | @@`
	Print       *Print       `  | @@`
	Instruction *Instruction `|@@) EOL`
}

type Memory struct {
	BaseReg      string `"[" @Ident `
	Operation    string `@("+" | "-")? `
	Displacement string `@Immediate? "]"?`
}

type Operand struct {
	Register  string `( @Ident","?`
	Immediate uint32 `| @Immediate	","?`
	Memory    Memory `| @@ )`
}

type Instruction struct {
	Pos lexer.Position

	Mnemonic string    `@Ident`
	Operads  []Operand `@@*` //`(@Ident","?)+`
	//Args []*Expression `"(" ( @@ ( "," @@ )* )? ")"`
}

type Print struct {
	Pos lexer.Position

	Expression *Expression `"PRINT" @@`
}

type Input struct {
	Pos lexer.Position

	Variable string `"INPUT" @Ident`
}

type Let struct {
	Pos lexer.Position

	Variable string      `"LET" @Ident`
	Value    *Expression `"=" @@`
}

type Goto struct {
	Pos lexer.Position

	Line int `"GOTO" @Number`
}

type If struct {
	Pos lexer.Position

	Condition *Expression `"IF" @@`
	Line      int         `"THEN" @Number`
}

type Operator string

func (o *Operator) Capture(s []string) error {
	*o = Operator(strings.Join(s, ""))
	return nil
}

type Value struct {
	Pos lexer.Position

	Number        *float64    `  @Number`
	Variable      *string     `| @Ident`
	String        *string     `| @String`
	Subexpression *Expression `| "(" @@ ")"`
}

type Factor struct {
	Pos lexer.Position

	Base     *Value `@@`
	Exponent *Value `( "^" @@ )?`
}

type OpFactor struct {
	Pos lexer.Position

	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
}

type Term struct {
	Pos lexer.Position

	Left  *Factor     `@@`
	Right []*OpFactor `@@*`
}

type OpTerm struct {
	Pos lexer.Position

	Operator Operator `@("+" | "-")`
	Term     *Term    `@@`
}

type Cmp struct {
	Pos lexer.Position

	Left  *Term     `@@`
	Right []*OpTerm `@@*`
}

type OpCmp struct {
	Pos lexer.Position

	Operator Operator `@("=" | "<" "=" | ">" "=" | "<" | ">" | "!" "=")`
	Cmp      *Cmp     `@@`
}

type Expression struct {
	Pos lexer.Position

	Left  *Cmp     `@@`
	Right []*OpCmp `@@*`
}

func Gmain() {
	participle.Trace(log.Default().Writer())
	parser := participle.MustBuild[Program](
		participle.Lexer(basicLexer),
		participle.Elide("Comment"),
		participle.UseLookahead(2),
	)

	example :=
		`#xor r1, r1
		hlt
		nop
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
	`

	prog, err := parser.ParseString("example.asm", example)
	if err != nil {
		log.Printf("error: %v\n", err.Error())

	}
	fmt.Println(parser.Lexer().Symbols())
	for _, each := range prog.Lines {
		//fmt.Printf("line: %+v\n", each)
		if each.Instruction != nil {
			fmt.Printf("instr: %+v\n", each.Instruction)
		}
	}

}
