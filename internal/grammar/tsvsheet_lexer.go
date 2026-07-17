// Code generated from TsvsheetLexer.g4 by ANTLR 4.13.2. DO NOT EDIT.

package tsvsheetgrammar

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type TsvsheetLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var TsvsheetLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func tsvsheetlexerLexerInit() {
	staticData := &TsvsheetLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'>='", "'<='", "'<>'", "'>'", "'<'", "'TRUE'", "'FALSE'", "", "'='",
		"'('", "')'", "':'", "','", "'$'", "'*'", "'+'", "'-'", "'/'", "'%'",
		"'^'", "'&'", "'!'", "'|'",
	}
	staticData.SymbolicNames = []string{
		"", "GE", "LE", "NE", "GT", "LT", "TRUE", "FALSE", "ERRORCONST", "EQ",
		"LPAREN", "RPAREN", "COLON", "COMMA", "DOLLAR", "STAR", "PLUS", "DASH",
		"SLASH", "PERCENT", "CARET", "AMP", "BANG", "PIPE", "NUMBER", "COL",
		"NAME", "STRING", "WS",
	}
	staticData.RuleNames = []string{
		"GE", "LE", "NE", "GT", "LT", "TRUE", "FALSE", "ERRORCONST", "EQ", "LPAREN",
		"RPAREN", "COLON", "COMMA", "DOLLAR", "STAR", "PLUS", "DASH", "SLASH",
		"PERCENT", "CARET", "AMP", "BANG", "PIPE", "NUMBER", "COL", "NAME",
		"STRING", "WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 28, 197, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 2, 15,
		7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19, 2, 20, 7,
		20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2, 25, 7, 25,
		2, 26, 7, 26, 2, 27, 7, 27, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 2, 1,
		2, 1, 2, 1, 3, 1, 3, 1, 4, 1, 4, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1,
		6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 3, 7, 127, 8, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11,
		1, 11, 1, 12, 1, 12, 1, 13, 1, 13, 1, 14, 1, 14, 1, 15, 1, 15, 1, 16, 1,
		16, 1, 17, 1, 17, 1, 18, 1, 18, 1, 19, 1, 19, 1, 20, 1, 20, 1, 21, 1, 21,
		1, 22, 1, 22, 1, 23, 4, 23, 160, 8, 23, 11, 23, 12, 23, 161, 1, 23, 1,
		23, 4, 23, 166, 8, 23, 11, 23, 12, 23, 167, 3, 23, 170, 8, 23, 1, 24, 4,
		24, 173, 8, 24, 11, 24, 12, 24, 174, 1, 25, 4, 25, 178, 8, 25, 11, 25,
		12, 25, 179, 1, 26, 1, 26, 5, 26, 184, 8, 26, 10, 26, 12, 26, 187, 9, 26,
		1, 26, 1, 26, 1, 27, 4, 27, 192, 8, 27, 11, 27, 12, 27, 193, 1, 27, 1,
		27, 0, 0, 28, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 9,
		19, 10, 21, 11, 23, 12, 25, 13, 27, 14, 29, 15, 31, 16, 33, 17, 35, 18,
		37, 19, 39, 20, 41, 21, 43, 22, 45, 23, 47, 24, 49, 25, 51, 26, 53, 27,
		55, 28, 1, 0, 4, 1, 0, 48, 57, 1, 0, 65, 90, 2, 0, 65, 90, 97, 122, 3,
		0, 10, 10, 13, 13, 34, 34, 211, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0,
		5, 1, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0,
		13, 1, 0, 0, 0, 0, 15, 1, 0, 0, 0, 0, 17, 1, 0, 0, 0, 0, 19, 1, 0, 0, 0,
		0, 21, 1, 0, 0, 0, 0, 23, 1, 0, 0, 0, 0, 25, 1, 0, 0, 0, 0, 27, 1, 0, 0,
		0, 0, 29, 1, 0, 0, 0, 0, 31, 1, 0, 0, 0, 0, 33, 1, 0, 0, 0, 0, 35, 1, 0,
		0, 0, 0, 37, 1, 0, 0, 0, 0, 39, 1, 0, 0, 0, 0, 41, 1, 0, 0, 0, 0, 43, 1,
		0, 0, 0, 0, 45, 1, 0, 0, 0, 0, 47, 1, 0, 0, 0, 0, 49, 1, 0, 0, 0, 0, 51,
		1, 0, 0, 0, 0, 53, 1, 0, 0, 0, 0, 55, 1, 0, 0, 0, 1, 57, 1, 0, 0, 0, 3,
		60, 1, 0, 0, 0, 5, 63, 1, 0, 0, 0, 7, 66, 1, 0, 0, 0, 9, 68, 1, 0, 0, 0,
		11, 70, 1, 0, 0, 0, 13, 75, 1, 0, 0, 0, 15, 81, 1, 0, 0, 0, 17, 128, 1,
		0, 0, 0, 19, 130, 1, 0, 0, 0, 21, 132, 1, 0, 0, 0, 23, 134, 1, 0, 0, 0,
		25, 136, 1, 0, 0, 0, 27, 138, 1, 0, 0, 0, 29, 140, 1, 0, 0, 0, 31, 142,
		1, 0, 0, 0, 33, 144, 1, 0, 0, 0, 35, 146, 1, 0, 0, 0, 37, 148, 1, 0, 0,
		0, 39, 150, 1, 0, 0, 0, 41, 152, 1, 0, 0, 0, 43, 154, 1, 0, 0, 0, 45, 156,
		1, 0, 0, 0, 47, 159, 1, 0, 0, 0, 49, 172, 1, 0, 0, 0, 51, 177, 1, 0, 0,
		0, 53, 181, 1, 0, 0, 0, 55, 191, 1, 0, 0, 0, 57, 58, 5, 62, 0, 0, 58, 59,
		5, 61, 0, 0, 59, 2, 1, 0, 0, 0, 60, 61, 5, 60, 0, 0, 61, 62, 5, 61, 0,
		0, 62, 4, 1, 0, 0, 0, 63, 64, 5, 60, 0, 0, 64, 65, 5, 62, 0, 0, 65, 6,
		1, 0, 0, 0, 66, 67, 5, 62, 0, 0, 67, 8, 1, 0, 0, 0, 68, 69, 5, 60, 0, 0,
		69, 10, 1, 0, 0, 0, 70, 71, 5, 84, 0, 0, 71, 72, 5, 82, 0, 0, 72, 73, 5,
		85, 0, 0, 73, 74, 5, 69, 0, 0, 74, 12, 1, 0, 0, 0, 75, 76, 5, 70, 0, 0,
		76, 77, 5, 65, 0, 0, 77, 78, 5, 76, 0, 0, 78, 79, 5, 83, 0, 0, 79, 80,
		5, 69, 0, 0, 80, 14, 1, 0, 0, 0, 81, 126, 5, 35, 0, 0, 82, 83, 5, 78, 0,
		0, 83, 84, 5, 47, 0, 0, 84, 127, 5, 65, 0, 0, 85, 86, 5, 82, 0, 0, 86,
		87, 5, 69, 0, 0, 87, 88, 5, 70, 0, 0, 88, 127, 5, 33, 0, 0, 89, 90, 5,
		86, 0, 0, 90, 91, 5, 65, 0, 0, 91, 92, 5, 76, 0, 0, 92, 93, 5, 85, 0, 0,
		93, 94, 5, 69, 0, 0, 94, 127, 5, 33, 0, 0, 95, 96, 5, 78, 0, 0, 96, 97,
		5, 65, 0, 0, 97, 98, 5, 77, 0, 0, 98, 99, 5, 69, 0, 0, 99, 127, 5, 63,
		0, 0, 100, 101, 5, 68, 0, 0, 101, 102, 5, 73, 0, 0, 102, 103, 5, 86, 0,
		0, 103, 104, 5, 47, 0, 0, 104, 105, 5, 48, 0, 0, 105, 127, 5, 33, 0, 0,
		106, 107, 5, 78, 0, 0, 107, 108, 5, 85, 0, 0, 108, 109, 5, 77, 0, 0, 109,
		127, 5, 33, 0, 0, 110, 111, 5, 78, 0, 0, 111, 112, 5, 85, 0, 0, 112, 113,
		5, 76, 0, 0, 113, 114, 5, 76, 0, 0, 114, 127, 5, 33, 0, 0, 115, 116, 5,
		83, 0, 0, 116, 117, 5, 80, 0, 0, 117, 118, 5, 73, 0, 0, 118, 119, 5, 76,
		0, 0, 119, 120, 5, 76, 0, 0, 120, 127, 5, 33, 0, 0, 121, 122, 5, 67, 0,
		0, 122, 123, 5, 73, 0, 0, 123, 124, 5, 82, 0, 0, 124, 125, 5, 67, 0, 0,
		125, 127, 5, 33, 0, 0, 126, 82, 1, 0, 0, 0, 126, 85, 1, 0, 0, 0, 126, 89,
		1, 0, 0, 0, 126, 95, 1, 0, 0, 0, 126, 100, 1, 0, 0, 0, 126, 106, 1, 0,
		0, 0, 126, 110, 1, 0, 0, 0, 126, 115, 1, 0, 0, 0, 126, 121, 1, 0, 0, 0,
		127, 16, 1, 0, 0, 0, 128, 129, 5, 61, 0, 0, 129, 18, 1, 0, 0, 0, 130, 131,
		5, 40, 0, 0, 131, 20, 1, 0, 0, 0, 132, 133, 5, 41, 0, 0, 133, 22, 1, 0,
		0, 0, 134, 135, 5, 58, 0, 0, 135, 24, 1, 0, 0, 0, 136, 137, 5, 44, 0, 0,
		137, 26, 1, 0, 0, 0, 138, 139, 5, 36, 0, 0, 139, 28, 1, 0, 0, 0, 140, 141,
		5, 42, 0, 0, 141, 30, 1, 0, 0, 0, 142, 143, 5, 43, 0, 0, 143, 32, 1, 0,
		0, 0, 144, 145, 5, 45, 0, 0, 145, 34, 1, 0, 0, 0, 146, 147, 5, 47, 0, 0,
		147, 36, 1, 0, 0, 0, 148, 149, 5, 37, 0, 0, 149, 38, 1, 0, 0, 0, 150, 151,
		5, 94, 0, 0, 151, 40, 1, 0, 0, 0, 152, 153, 5, 38, 0, 0, 153, 42, 1, 0,
		0, 0, 154, 155, 5, 33, 0, 0, 155, 44, 1, 0, 0, 0, 156, 157, 5, 124, 0,
		0, 157, 46, 1, 0, 0, 0, 158, 160, 7, 0, 0, 0, 159, 158, 1, 0, 0, 0, 160,
		161, 1, 0, 0, 0, 161, 159, 1, 0, 0, 0, 161, 162, 1, 0, 0, 0, 162, 169,
		1, 0, 0, 0, 163, 165, 5, 46, 0, 0, 164, 166, 7, 0, 0, 0, 165, 164, 1, 0,
		0, 0, 166, 167, 1, 0, 0, 0, 167, 165, 1, 0, 0, 0, 167, 168, 1, 0, 0, 0,
		168, 170, 1, 0, 0, 0, 169, 163, 1, 0, 0, 0, 169, 170, 1, 0, 0, 0, 170,
		48, 1, 0, 0, 0, 171, 173, 7, 1, 0, 0, 172, 171, 1, 0, 0, 0, 173, 174, 1,
		0, 0, 0, 174, 172, 1, 0, 0, 0, 174, 175, 1, 0, 0, 0, 175, 50, 1, 0, 0,
		0, 176, 178, 7, 2, 0, 0, 177, 176, 1, 0, 0, 0, 178, 179, 1, 0, 0, 0, 179,
		177, 1, 0, 0, 0, 179, 180, 1, 0, 0, 0, 180, 52, 1, 0, 0, 0, 181, 185, 5,
		34, 0, 0, 182, 184, 8, 3, 0, 0, 183, 182, 1, 0, 0, 0, 184, 187, 1, 0, 0,
		0, 185, 183, 1, 0, 0, 0, 185, 186, 1, 0, 0, 0, 186, 188, 1, 0, 0, 0, 187,
		185, 1, 0, 0, 0, 188, 189, 5, 34, 0, 0, 189, 54, 1, 0, 0, 0, 190, 192,
		5, 32, 0, 0, 191, 190, 1, 0, 0, 0, 192, 193, 1, 0, 0, 0, 193, 191, 1, 0,
		0, 0, 193, 194, 1, 0, 0, 0, 194, 195, 1, 0, 0, 0, 195, 196, 6, 27, 0, 0,
		196, 56, 1, 0, 0, 0, 9, 0, 126, 161, 167, 169, 174, 179, 185, 193, 1, 6,
		0, 0,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// TsvsheetLexerInit initializes any static state used to implement TsvsheetLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewTsvsheetLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func TsvsheetLexerInit() {
	staticData := &TsvsheetLexerLexerStaticData
	staticData.once.Do(tsvsheetlexerLexerInit)
}

// NewTsvsheetLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewTsvsheetLexer(input antlr.CharStream) *TsvsheetLexer {
	TsvsheetLexerInit()
	l := new(TsvsheetLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &TsvsheetLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "TsvsheetLexer.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// TsvsheetLexer tokens.
const (
	TsvsheetLexerGE         = 1
	TsvsheetLexerLE         = 2
	TsvsheetLexerNE         = 3
	TsvsheetLexerGT         = 4
	TsvsheetLexerLT         = 5
	TsvsheetLexerTRUE       = 6
	TsvsheetLexerFALSE      = 7
	TsvsheetLexerERRORCONST = 8
	TsvsheetLexerEQ         = 9
	TsvsheetLexerLPAREN     = 10
	TsvsheetLexerRPAREN     = 11
	TsvsheetLexerCOLON      = 12
	TsvsheetLexerCOMMA      = 13
	TsvsheetLexerDOLLAR     = 14
	TsvsheetLexerSTAR       = 15
	TsvsheetLexerPLUS       = 16
	TsvsheetLexerDASH       = 17
	TsvsheetLexerSLASH      = 18
	TsvsheetLexerPERCENT    = 19
	TsvsheetLexerCARET      = 20
	TsvsheetLexerAMP        = 21
	TsvsheetLexerBANG       = 22
	TsvsheetLexerPIPE       = 23
	TsvsheetLexerNUMBER     = 24
	TsvsheetLexerCOL        = 25
	TsvsheetLexerNAME       = 26
	TsvsheetLexerSTRING     = 27
	TsvsheetLexerWS         = 28
)
