package grammar

import (
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// Define the lexer for extended assembly
var asmLexerDyn = lexer.MustStateful(lexer.Rules{
	"Root": {
		{"Comment", `#.*`, nil},
		{"Label", `\w{1,}:`, nil},
		{"Directive", `\.\w{2,}`, nil},
		//{"String", `"(\\"|[^"])*"`, nil},
		//{"Punct", `[!@#$%^&*()_={}\|:;"'<,>.?/]`, nil},
		{"Hex", `(?i)0x[0-9a-f]+`, nil},
		{"Number", `[-]?\d+`, nil},
		//{"Number", `[-+]?(\d*\.)?\d+`, nil},
		{"MemoryStart", `\[`, lexer.Push("Memory")},
		{"Comma", `,`, nil},
		//{"Mnemonic", `[a-z]{1,}`, nil},
		{"Ident", `[a-z0-9]\w*`, nil},
		{"EOL", `[\n\r]+`, nil},
		{"Whitespace", `[ \t]+`, nil},
	}, "Memory": {
		{"Operation", `\+|-`, nil},
		{"Hex", `(?i)0x[0-9a-f]+`, nil},
		{"Decimal", `\d+`, nil},
		{"Ident", `[a-z0-9]\w*`, nil},
		{"whitespace", `[ \t]+`, nil},
		{"MemoryEnd", `]`, lexer.Pop()},
		//{"Displacement",`0x[0-9a-f]+|[-+]?\d+`,nil},
		//{"Hex", `0x[0-9a-f]+`, nil},
		//{"Decimal", `[-\+]?\d+`, nil},
	},
})

type Program struct {
	//Pos lexer.Position

	Lines []Line `@@*`
}

type Line struct {
	Pos lexer.Position

	Index int

	Comment     string       `EOL? ( @Comment`
	Directive   *Directive   `| @@`
	Label       *Label       `| @@`
	Instruction *Instruction `| @@) (EOL|EOF)`
}

type Directive struct {
	//Pos lexer.Position

	Type    string    `@Directive`
	Operand Immediate `@@?`
}

type Label struct {
	//Pos lexer.Position

	Text   string `@Label`
	Offset uint32
}

type Instruction struct {
	Pos *lexer.Position

	Mnemonic string    `@Ident " "?`
	Operands []Operand `@@*`
}

type Immediate struct {
	// Allows signed numbers and hex numbers
	//Pos *lexer.Position

	Value string ` @Number|@Hex`
}

type Displacement struct {
	// Allows only positive numbers (as decimal representation) and hex numbers
	//Pos *lexer.Position

	Value string `@Decimal|@Hex`
}

type Memory struct {
	//Pos *lexer.Position

	Base         string       `"[" @Ident " "? `
	Operation    string       `@Operation? " "? `
	Displacement Displacement `@@? " "? "]"`
}

type Operand interface {
}

type OperandRegister struct {
	//Pos *lexer.Position

	Value string `@Ident ","? " "?`
}
type OperandImmediate struct {
	//Pos *lexer.Position

	Value string ` (@Number|@Hex) ","?" "? `
}
type OperandMemory struct {
	//Pos *lexer.Position

	Value Memory `@@`
}

func toLower(token lexer.Token) (lexer.Token, error) {
	token.Value = strings.ToLower(token.Value)
	return token, nil
}

var Parser = participle.MustBuild[Program](
	participle.Lexer(asmLexerDyn),
	participle.Elide("Comment"),
	participle.UseLookahead(3),
	participle.Union[Operand](OperandRegister{}, OperandImmediate{}, OperandMemory{}),
	participle.Map(toLower, "Ident"), // lowercase all mnemonics and identifiers such as register names
)

func ParseString(name, input string) (*Program, error) {
	if name == "" {
		name = "<unknown>"
	}
	return Parser.ParseString(name, input)
}
