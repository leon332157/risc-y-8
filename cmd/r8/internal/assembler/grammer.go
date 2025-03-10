package assembler

import (
	"fmt"
	"log"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

)

// Define the lexer for extended assembly
var asmLexerDyn = lexer.MustStateful(lexer.Rules{
	"Root": {
		{"Comment", `#.*`, nil},
		{"Label", `\.\w{1,}:`, nil},
		{"Directive", `\.\w{2,}`, nil},
		//{"String", `"(\\"|[^"])*"`, nil},

		//{"Punct", `[!@#$%^&*()_={}\|:;"'<,>.?/]`, nil},

		{"Hex", `(?i)0x[0-9a-f]+`, nil},
		{"Decimal", `[-\+]?\d+`, nil},
		//{"Number", `[-+]?(\d*\.)?\d+`, nil},
		{"MemoryStart",`\[`,lexer.Push("Memory")},
		{"Comma",`,`,nil},
		{"Mnemonic",`(?i)[a-z]{2,}(\.[a-z]{1,})?`,nil},
		{"Ident", `(?i)[a-z0-9]\w*`, nil},
		{"EOL", `[\n\r]+`, nil},
		{"whitespace", `[ \t]+`, nil},
	}, "Memory" : {
		{"Operation",`\+|-`,nil},
		{"Hex", `(?i)0x[0-9a-f]+`, nil},
		{"Decimal", `[-\+]?\d+`, nil},
		{"Ident", `(?i)[a-z0-9]\w*`,nil},
		{"MemoryEnd",`]`,lexer.Pop()},
		//{"Displacement",`0x[0-9a-f]+|[-+]?\d+`,nil},
		//{"Hex", `0x[0-9a-f]+`, nil},
		//{"Decimal", `[-\+]?\d+`, nil},

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
	{"Punct", `[!@#$%^&*()_={}\|:;"'<,>.?/]|]`},
	//{"Operation",`\+`},
	//{"Hex",`0x[0-9a-f]+`},
	{"HexWithSign", `[-+]?0x[0-9a-f]+`},
	{"Decimal", `[-+]?\d+`},
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

type Program struct {
	Pos lexer.Position

	Lines []Line `@@*`
}

type Line struct {
	Pos lexer.Position

	Index int

	Comment   string `( @Comment`
	Directive *Directive `| @@`
	Label     *Label `| @@`
	Instruction *Instruction `| @@) EOL`
}

type Directive struct {
	//Pos lexer.Position

	Type string `@Directive`
	Operand Immediate `@@?`
}

type Label struct {
	//Pos lexer.Position

	Text string `@Label`
	Offset uint32
}

type Instruction struct {
	//Pos lexer.Position

	Mnemonic string    `@Mnemonic`
	Operads  []Operand `@@*` //`(@Ident","?)+`
}


type Immediate struct {
	Pos lexer.Position

	Value string ` @Decimal|@Hex`
}

type Memory struct {
	Pos lexer.Position
	Base string `"[" @Ident `
	Operation    string `@Operation? `
	Displacement Immediate `@@? "]"`
}

type Operand struct {
	Pos lexer.Position

	Register  string    `( @Ident ","?`
	Immediate *Immediate `| @@	","?`
	Memory    *Memory    `| @@ )`
}

func Gmain() {
	participle.Trace(log.Default().Writer())
	parser := participle.MustBuild[Program](
		participle.Lexer(asmLexerDyn),
		participle.Elide("Comment"),
		participle.UseLookahead(2),
	)

	example :=
		`.org 0x1
		#xor r1, r1
		hlt
		.L1:
		nop
		add r1, 1
		xor r2, r2
		add r2, -2
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
		VBEQ v2, v3, [pc-10]
		VLDALL.d v8,[r1+0xFFFF]
	`

	prog, err := parser.ParseString("example.asm", example)
	if err != nil {
		log.Printf("error: %v\n", err.Error())

	}
	fmt.Println(parser.Lexer().Symbols())
	for _, each := range prog.Lines {
		fmt.Printf("line: %+v\n", each)
		if each.Label != nil {
			fmt.Printf("label: %+v\n",each.Label)
		}
		if each.Directive != nil {
			fmt.Printf("directive: %+v\n",each.Directive)
		}
		if each.Instruction != nil {
			fmt.Printf("instr: %+v\n", each.Instruction)
		}
	}

}
