package assembler

// This grammar is generated using the Grammar.export_go() method and
// should be used with the goleri module.
//
// Source class: ASMGrammar
// Created at: 2025-02-24 22:12:24

import (
        "regexp"

        "github.com/cesbit/goleri"
)

// Element indentifiers
const (
        NoGid = iota
        GidASMOPTION = iota
        GidIMM = iota
        GidIMMDEC = iota
        GidIMMHEX = iota
        GidINST = iota
        GidINSTR = iota
        GidINSTRCTRL = iota
        GidINSTRIMMREG = iota
        GidINSTRMEM = iota
        GidINSTRMEMBRACKET = iota
        GidINSTRREGREG = iota
        GidLINE = iota
        GidLINES = iota
        GidNEWLINE = iota
        GidREGDEST = iota
        GidREGSRC = iota
        GidSTART = iota
)

// ASMGrammar returns a compiled goleri grammar.
func ASMGrammar() *goleri.Grammar {
        NEWLINE := goleri.NewRegex(GidNEWLINE, regexp.MustCompile(`^(\r\n|\r|\n)`))
        INST := goleri.NewRegex(GidINST, regexp.MustCompile(`(?i)^[a-z]{3}`))
        REGDEST := goleri.NewRegex(GidREGDEST, regexp.MustCompile(`(?i)^R[0-9]+`))
        REGSRC := goleri.NewRegex(GidREGSRC, regexp.MustCompile(`(?i)^R[0-9]+`))
        IMMHEX := goleri.NewRegex(GidIMMHEX, regexp.MustCompile(`(?i)^(0x)?[0-9a-f]+`))
        IMMDEC := goleri.NewRegex(GidIMMDEC, regexp.MustCompile(`^[0-9]+`))
        IMM := goleri.NewChoice(
                GidIMM,
                false,
                IMMHEX,
                IMMDEC,
        )
        ASMOPTION := goleri.NewSequence(
                GidASMOPTION,
                goleri.NewRegex(NoGid, regexp.MustCompile(`(?i)^.([a-zA-Z]{3})`)),
                goleri.NewOptional(NoGid, IMM),
        )
        INSTRREGREG := goleri.NewSequence(
                GidINSTRREGREG,
                INST,
                REGDEST,
                goleri.NewToken(NoGid, ","),
                REGSRC,
        )
        INSTRIMMREG := goleri.NewSequence(
                GidINSTRIMMREG,
                INST,
                REGDEST,
                goleri.NewToken(NoGid, ","),
                IMM,
        )
        INSTRCTRL := goleri.NewSequence(
                GidINSTRCTRL,
                INST,
                goleri.NewOptional(NoGid, goleri.NewChoice(
                        NoGid,
                        false,
                        IMM,
                        REGDEST,
                )),
        )
        INSTRMEM := goleri.NewSequence(
                GidINSTRMEM,
                INST,
                REGDEST,
                goleri.NewChoice(
                        NoGid,
                        false,
                        IMM,
                        REGSRC,
                ),
                goleri.NewOptional(NoGid, IMM),
        )
        INSTRMEMBRACKET := goleri.NewSequence(
                GidINSTRMEMBRACKET,
                INST,
                REGDEST,
                goleri.NewToken(NoGid, ","),
                goleri.NewSequence(
                        NoGid,
                        goleri.NewToken(NoGid, "["),
                        REGSRC,
                        goleri.NewTokens(NoGid, "+ - *"),
                        IMM,
                        goleri.NewToken(NoGid, "]"),
                ),
        )
        INSTR := goleri.NewChoice(
                GidINSTR,
                true,
                INSTRREGREG,
                INSTRIMMREG,
                INSTRCTRL,
                INSTRMEM,
                INSTRMEMBRACKET,
        )
        LINE := goleri.NewSequence(
                GidLINE,
                INSTR,
                goleri.NewOptional(NoGid, NEWLINE),
        )
        LINES := goleri.NewRepeat(GidLINES, LINE, 0, 0)
        START := goleri.NewSequence(
                GidSTART,
                goleri.NewOptional(NoGid, ASMOPTION),
                LINES,
        )
        return goleri.NewGrammar(START, regexp.MustCompile(`^\w+`))
}