// Code generated from TsvsheetLexer.g4 by ANTLR 4.13.2. DO NOT EDIT.

package tsvsheetgrammar

import (
	"fmt"
	"sync"
	"unicode"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import error
var (
	_ = fmt.Printf
	_ = sync.Once{}
	_ = unicode.IsLetter
)

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
		"DEFAULT_MODE", "META",
	}
	staticData.LiteralNames = []string{
		"", "'>='", "'<='", "'<>'", "'>'", "'<'", "'TRUE'", "'FALSE'", "", "'='",
		"", "", "':'", "", "'$'", "'*'", "'+'", "'-'", "'/'", "'%'", "'^'",
		"'&'", "'!'", "'|@'", "'|'",
	}
	staticData.SymbolicNames = []string{
		"", "GE", "LE", "NE", "GT", "LT", "TRUE", "FALSE", "ERRORCONST", "EQ",
		"LPAREN", "RPAREN", "COLON", "COMMA", "DOLLAR", "STAR", "PLUS", "DASH",
		"SLASH", "PERCENT", "CARET", "AMP", "BANG", "PIPEMETA", "PIPE", "NUMBER",
		"COL", "NAME", "NAMEREF", "STRING", "WS", "META_IDENT", "META_LPAREN",
		"META_RPAREN", "META_COMMA", "META_WS",
	}
	staticData.RuleNames = []string{
		"GE", "LE", "NE", "GT", "LT", "TRUE", "FALSE", "ERRORCONST", "EQ", "LPAREN",
		"RPAREN", "COLON", "COMMA", "DOLLAR", "STAR", "PLUS", "DASH", "SLASH",
		"PERCENT", "CARET", "AMP", "BANG", "PIPEMETA", "PIPE", "NUMBER", "COL",
		"NAME", "NAMEREF", "STRING", "WS", "META_IDENT", "META_LPAREN", "META_RPAREN",
		"META_COMMA", "META_WS",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 35, 260, 6, -1, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3,
		7, 3, 2, 4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9,
		7, 9, 2, 10, 7, 10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7,
		14, 2, 15, 7, 15, 2, 16, 7, 16, 2, 17, 7, 17, 2, 18, 7, 18, 2, 19, 7, 19,
		2, 20, 7, 20, 2, 21, 7, 21, 2, 22, 7, 22, 2, 23, 7, 23, 2, 24, 7, 24, 2,
		25, 7, 25, 2, 26, 7, 26, 2, 27, 7, 27, 2, 28, 7, 28, 2, 29, 7, 29, 2, 30,
		7, 30, 2, 31, 7, 31, 2, 32, 7, 32, 2, 33, 7, 33, 2, 34, 7, 34, 1, 0, 1,
		0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 4, 1, 4, 1,
		5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 6, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1,
		7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 3, 7, 155, 8, 7, 1,
		8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 1, 11, 1, 11, 1, 12, 1, 12, 1, 13, 1,
		13, 1, 14, 1, 14, 1, 15, 1, 15, 1, 16, 1, 16, 1, 17, 1, 17, 1, 18, 1, 18,
		1, 19, 1, 19, 1, 20, 1, 20, 1, 21, 1, 21, 1, 22, 1, 22, 1, 22, 1, 22, 1,
		22, 1, 23, 1, 23, 1, 24, 4, 24, 193, 8, 24, 11, 24, 12, 24, 194, 1, 24,
		1, 24, 4, 24, 199, 8, 24, 11, 24, 12, 24, 200, 3, 24, 203, 8, 24, 1, 25,
		4, 25, 206, 8, 25, 11, 25, 12, 25, 207, 1, 26, 4, 26, 211, 8, 26, 11, 26,
		12, 26, 212, 1, 27, 1, 27, 1, 27, 5, 27, 218, 8, 27, 10, 27, 12, 27, 221,
		9, 27, 1, 28, 1, 28, 5, 28, 225, 8, 28, 10, 28, 12, 28, 228, 9, 28, 1,
		28, 1, 28, 1, 29, 4, 29, 233, 8, 29, 11, 29, 12, 29, 234, 1, 29, 1, 29,
		1, 30, 1, 30, 5, 30, 241, 8, 30, 10, 30, 12, 30, 244, 9, 30, 1, 31, 1,
		31, 1, 32, 1, 32, 1, 32, 1, 32, 1, 33, 1, 33, 1, 34, 4, 34, 255, 8, 34,
		11, 34, 12, 34, 256, 1, 34, 1, 34, 0, 0, 35, 2, 1, 4, 2, 6, 3, 8, 4, 10,
		5, 12, 6, 14, 7, 16, 8, 18, 9, 20, 10, 22, 11, 24, 12, 26, 13, 28, 14,
		30, 15, 32, 16, 34, 17, 36, 18, 38, 19, 40, 20, 42, 21, 44, 22, 46, 23,
		48, 24, 50, 25, 52, 26, 54, 27, 56, 28, 58, 29, 60, 30, 62, 31, 64, 32,
		66, 33, 68, 34, 70, 35, 2, 0, 1, 6, 1, 0, 48, 57, 1, 0, 65, 90, 2, 0, 65,
		90, 97, 122, 3, 0, 65, 90, 95, 95, 97, 122, 4, 0, 48, 57, 65, 90, 95, 95,
		97, 122, 3, 0, 10, 10, 13, 13, 34, 34, 278, 0, 2, 1, 0, 0, 0, 0, 4, 1,
		0, 0, 0, 0, 6, 1, 0, 0, 0, 0, 8, 1, 0, 0, 0, 0, 10, 1, 0, 0, 0, 0, 12,
		1, 0, 0, 0, 0, 14, 1, 0, 0, 0, 0, 16, 1, 0, 0, 0, 0, 18, 1, 0, 0, 0, 0,
		20, 1, 0, 0, 0, 0, 22, 1, 0, 0, 0, 0, 24, 1, 0, 0, 0, 0, 26, 1, 0, 0, 0,
		0, 28, 1, 0, 0, 0, 0, 30, 1, 0, 0, 0, 0, 32, 1, 0, 0, 0, 0, 34, 1, 0, 0,
		0, 0, 36, 1, 0, 0, 0, 0, 38, 1, 0, 0, 0, 0, 40, 1, 0, 0, 0, 0, 42, 1, 0,
		0, 0, 0, 44, 1, 0, 0, 0, 0, 46, 1, 0, 0, 0, 0, 48, 1, 0, 0, 0, 0, 50, 1,
		0, 0, 0, 0, 52, 1, 0, 0, 0, 0, 54, 1, 0, 0, 0, 0, 56, 1, 0, 0, 0, 0, 58,
		1, 0, 0, 0, 0, 60, 1, 0, 0, 0, 1, 62, 1, 0, 0, 0, 1, 64, 1, 0, 0, 0, 1,
		66, 1, 0, 0, 0, 1, 68, 1, 0, 0, 0, 1, 70, 1, 0, 0, 0, 2, 72, 1, 0, 0, 0,
		4, 75, 1, 0, 0, 0, 6, 78, 1, 0, 0, 0, 8, 81, 1, 0, 0, 0, 10, 83, 1, 0,
		0, 0, 12, 85, 1, 0, 0, 0, 14, 90, 1, 0, 0, 0, 16, 96, 1, 0, 0, 0, 18, 156,
		1, 0, 0, 0, 20, 158, 1, 0, 0, 0, 22, 160, 1, 0, 0, 0, 24, 162, 1, 0, 0,
		0, 26, 164, 1, 0, 0, 0, 28, 166, 1, 0, 0, 0, 30, 168, 1, 0, 0, 0, 32, 170,
		1, 0, 0, 0, 34, 172, 1, 0, 0, 0, 36, 174, 1, 0, 0, 0, 38, 176, 1, 0, 0,
		0, 40, 178, 1, 0, 0, 0, 42, 180, 1, 0, 0, 0, 44, 182, 1, 0, 0, 0, 46, 184,
		1, 0, 0, 0, 48, 189, 1, 0, 0, 0, 50, 192, 1, 0, 0, 0, 52, 205, 1, 0, 0,
		0, 54, 210, 1, 0, 0, 0, 56, 214, 1, 0, 0, 0, 58, 222, 1, 0, 0, 0, 60, 232,
		1, 0, 0, 0, 62, 238, 1, 0, 0, 0, 64, 245, 1, 0, 0, 0, 66, 247, 1, 0, 0,
		0, 68, 251, 1, 0, 0, 0, 70, 254, 1, 0, 0, 0, 72, 73, 5, 62, 0, 0, 73, 74,
		5, 61, 0, 0, 74, 3, 1, 0, 0, 0, 75, 76, 5, 60, 0, 0, 76, 77, 5, 61, 0,
		0, 77, 5, 1, 0, 0, 0, 78, 79, 5, 60, 0, 0, 79, 80, 5, 62, 0, 0, 80, 7,
		1, 0, 0, 0, 81, 82, 5, 62, 0, 0, 82, 9, 1, 0, 0, 0, 83, 84, 5, 60, 0, 0,
		84, 11, 1, 0, 0, 0, 85, 86, 5, 84, 0, 0, 86, 87, 5, 82, 0, 0, 87, 88, 5,
		85, 0, 0, 88, 89, 5, 69, 0, 0, 89, 13, 1, 0, 0, 0, 90, 91, 5, 70, 0, 0,
		91, 92, 5, 65, 0, 0, 92, 93, 5, 76, 0, 0, 93, 94, 5, 83, 0, 0, 94, 95,
		5, 69, 0, 0, 95, 15, 1, 0, 0, 0, 96, 154, 5, 35, 0, 0, 97, 98, 5, 78, 0,
		0, 98, 99, 5, 47, 0, 0, 99, 155, 5, 65, 0, 0, 100, 101, 5, 82, 0, 0, 101,
		102, 5, 69, 0, 0, 102, 103, 5, 70, 0, 0, 103, 155, 5, 33, 0, 0, 104, 105,
		5, 86, 0, 0, 105, 106, 5, 65, 0, 0, 106, 107, 5, 76, 0, 0, 107, 108, 5,
		85, 0, 0, 108, 109, 5, 69, 0, 0, 109, 155, 5, 33, 0, 0, 110, 111, 5, 78,
		0, 0, 111, 112, 5, 65, 0, 0, 112, 113, 5, 77, 0, 0, 113, 114, 5, 69, 0,
		0, 114, 155, 5, 63, 0, 0, 115, 116, 5, 68, 0, 0, 116, 117, 5, 73, 0, 0,
		117, 118, 5, 86, 0, 0, 118, 119, 5, 47, 0, 0, 119, 120, 5, 48, 0, 0, 120,
		155, 5, 33, 0, 0, 121, 122, 5, 78, 0, 0, 122, 123, 5, 85, 0, 0, 123, 124,
		5, 77, 0, 0, 124, 155, 5, 33, 0, 0, 125, 126, 5, 78, 0, 0, 126, 127, 5,
		85, 0, 0, 127, 128, 5, 76, 0, 0, 128, 129, 5, 76, 0, 0, 129, 155, 5, 33,
		0, 0, 130, 131, 5, 83, 0, 0, 131, 132, 5, 80, 0, 0, 132, 133, 5, 73, 0,
		0, 133, 134, 5, 76, 0, 0, 134, 135, 5, 76, 0, 0, 135, 155, 5, 33, 0, 0,
		136, 137, 5, 67, 0, 0, 137, 138, 5, 73, 0, 0, 138, 139, 5, 82, 0, 0, 139,
		140, 5, 67, 0, 0, 140, 155, 5, 33, 0, 0, 141, 142, 5, 73, 0, 0, 142, 143,
		5, 77, 0, 0, 143, 144, 5, 80, 0, 0, 144, 145, 5, 79, 0, 0, 145, 146, 5,
		82, 0, 0, 146, 147, 5, 84, 0, 0, 147, 155, 5, 33, 0, 0, 148, 149, 5, 76,
		0, 0, 149, 150, 5, 73, 0, 0, 150, 151, 5, 77, 0, 0, 151, 152, 5, 73, 0,
		0, 152, 153, 5, 84, 0, 0, 153, 155, 5, 33, 0, 0, 154, 97, 1, 0, 0, 0, 154,
		100, 1, 0, 0, 0, 154, 104, 1, 0, 0, 0, 154, 110, 1, 0, 0, 0, 154, 115,
		1, 0, 0, 0, 154, 121, 1, 0, 0, 0, 154, 125, 1, 0, 0, 0, 154, 130, 1, 0,
		0, 0, 154, 136, 1, 0, 0, 0, 154, 141, 1, 0, 0, 0, 154, 148, 1, 0, 0, 0,
		155, 17, 1, 0, 0, 0, 156, 157, 5, 61, 0, 0, 157, 19, 1, 0, 0, 0, 158, 159,
		5, 40, 0, 0, 159, 21, 1, 0, 0, 0, 160, 161, 5, 41, 0, 0, 161, 23, 1, 0,
		0, 0, 162, 163, 5, 58, 0, 0, 163, 25, 1, 0, 0, 0, 164, 165, 5, 44, 0, 0,
		165, 27, 1, 0, 0, 0, 166, 167, 5, 36, 0, 0, 167, 29, 1, 0, 0, 0, 168, 169,
		5, 42, 0, 0, 169, 31, 1, 0, 0, 0, 170, 171, 5, 43, 0, 0, 171, 33, 1, 0,
		0, 0, 172, 173, 5, 45, 0, 0, 173, 35, 1, 0, 0, 0, 174, 175, 5, 47, 0, 0,
		175, 37, 1, 0, 0, 0, 176, 177, 5, 37, 0, 0, 177, 39, 1, 0, 0, 0, 178, 179,
		5, 94, 0, 0, 179, 41, 1, 0, 0, 0, 180, 181, 5, 38, 0, 0, 181, 43, 1, 0,
		0, 0, 182, 183, 5, 33, 0, 0, 183, 45, 1, 0, 0, 0, 184, 185, 5, 124, 0,
		0, 185, 186, 5, 64, 0, 0, 186, 187, 1, 0, 0, 0, 187, 188, 6, 22, 0, 0,
		188, 47, 1, 0, 0, 0, 189, 190, 5, 124, 0, 0, 190, 49, 1, 0, 0, 0, 191,
		193, 7, 0, 0, 0, 192, 191, 1, 0, 0, 0, 193, 194, 1, 0, 0, 0, 194, 192,
		1, 0, 0, 0, 194, 195, 1, 0, 0, 0, 195, 202, 1, 0, 0, 0, 196, 198, 5, 46,
		0, 0, 197, 199, 7, 0, 0, 0, 198, 197, 1, 0, 0, 0, 199, 200, 1, 0, 0, 0,
		200, 198, 1, 0, 0, 0, 200, 201, 1, 0, 0, 0, 201, 203, 1, 0, 0, 0, 202,
		196, 1, 0, 0, 0, 202, 203, 1, 0, 0, 0, 203, 51, 1, 0, 0, 0, 204, 206, 7,
		1, 0, 0, 205, 204, 1, 0, 0, 0, 206, 207, 1, 0, 0, 0, 207, 205, 1, 0, 0,
		0, 207, 208, 1, 0, 0, 0, 208, 53, 1, 0, 0, 0, 209, 211, 7, 2, 0, 0, 210,
		209, 1, 0, 0, 0, 211, 212, 1, 0, 0, 0, 212, 210, 1, 0, 0, 0, 212, 213,
		1, 0, 0, 0, 213, 55, 1, 0, 0, 0, 214, 215, 5, 64, 0, 0, 215, 219, 7, 3,
		0, 0, 216, 218, 7, 4, 0, 0, 217, 216, 1, 0, 0, 0, 218, 221, 1, 0, 0, 0,
		219, 217, 1, 0, 0, 0, 219, 220, 1, 0, 0, 0, 220, 57, 1, 0, 0, 0, 221, 219,
		1, 0, 0, 0, 222, 226, 5, 34, 0, 0, 223, 225, 8, 5, 0, 0, 224, 223, 1, 0,
		0, 0, 225, 228, 1, 0, 0, 0, 226, 224, 1, 0, 0, 0, 226, 227, 1, 0, 0, 0,
		227, 229, 1, 0, 0, 0, 228, 226, 1, 0, 0, 0, 229, 230, 5, 34, 0, 0, 230,
		59, 1, 0, 0, 0, 231, 233, 5, 32, 0, 0, 232, 231, 1, 0, 0, 0, 233, 234,
		1, 0, 0, 0, 234, 232, 1, 0, 0, 0, 234, 235, 1, 0, 0, 0, 235, 236, 1, 0,
		0, 0, 236, 237, 6, 29, 1, 0, 237, 61, 1, 0, 0, 0, 238, 242, 7, 3, 0, 0,
		239, 241, 7, 4, 0, 0, 240, 239, 1, 0, 0, 0, 241, 244, 1, 0, 0, 0, 242,
		240, 1, 0, 0, 0, 242, 243, 1, 0, 0, 0, 243, 63, 1, 0, 0, 0, 244, 242, 1,
		0, 0, 0, 245, 246, 5, 40, 0, 0, 246, 65, 1, 0, 0, 0, 247, 248, 5, 41, 0,
		0, 248, 249, 1, 0, 0, 0, 249, 250, 6, 32, 2, 0, 250, 67, 1, 0, 0, 0, 251,
		252, 5, 44, 0, 0, 252, 69, 1, 0, 0, 0, 253, 255, 5, 32, 0, 0, 254, 253,
		1, 0, 0, 0, 255, 256, 1, 0, 0, 0, 256, 254, 1, 0, 0, 0, 256, 257, 1, 0,
		0, 0, 257, 258, 1, 0, 0, 0, 258, 259, 6, 34, 1, 0, 259, 71, 1, 0, 0, 0,
		13, 0, 1, 154, 194, 200, 202, 207, 212, 219, 226, 234, 242, 256, 3, 5,
		1, 0, 6, 0, 0, 4, 0, 0,
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
	TsvsheetLexerGE          = 1
	TsvsheetLexerLE          = 2
	TsvsheetLexerNE          = 3
	TsvsheetLexerGT          = 4
	TsvsheetLexerLT          = 5
	TsvsheetLexerTRUE        = 6
	TsvsheetLexerFALSE       = 7
	TsvsheetLexerERRORCONST  = 8
	TsvsheetLexerEQ          = 9
	TsvsheetLexerLPAREN      = 10
	TsvsheetLexerRPAREN      = 11
	TsvsheetLexerCOLON       = 12
	TsvsheetLexerCOMMA       = 13
	TsvsheetLexerDOLLAR      = 14
	TsvsheetLexerSTAR        = 15
	TsvsheetLexerPLUS        = 16
	TsvsheetLexerDASH        = 17
	TsvsheetLexerSLASH       = 18
	TsvsheetLexerPERCENT     = 19
	TsvsheetLexerCARET       = 20
	TsvsheetLexerAMP         = 21
	TsvsheetLexerBANG        = 22
	TsvsheetLexerPIPEMETA    = 23
	TsvsheetLexerPIPE        = 24
	TsvsheetLexerNUMBER      = 25
	TsvsheetLexerCOL         = 26
	TsvsheetLexerNAME        = 27
	TsvsheetLexerNAMEREF     = 28
	TsvsheetLexerSTRING      = 29
	TsvsheetLexerWS          = 30
	TsvsheetLexerMETA_IDENT  = 31
	TsvsheetLexerMETA_LPAREN = 32
	TsvsheetLexerMETA_RPAREN = 33
	TsvsheetLexerMETA_COMMA  = 34
	TsvsheetLexerMETA_WS     = 35
)

// TsvsheetLexerMETA is the TsvsheetLexer mode.
const TsvsheetLexerMETA = 1
