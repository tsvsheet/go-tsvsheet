// Code generated from TsvsheetParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package tsvsheetgrammar // TsvsheetParser
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type TsvsheetParser struct {
	*antlr.BaseParser
}

var TsvsheetParserParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func tsvsheetparserParserInit() {
	staticData := &TsvsheetParserParserStaticData
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
		"expression", "functionCall", "argList", "reference", "sheetQualifier",
		"cellRef", "directiveValue", "axisCall", "itemList", "item", "rangeCall",
		"countCall", "span", "endpoint", "offset",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 28, 161, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 2, 11, 7, 11, 2, 12, 7, 12, 2, 13, 7, 13, 2, 14, 7, 14, 1, 0, 1, 0,
		1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 3, 0,
		44, 8, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1,
		0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 5, 0, 66, 8, 0,
		10, 0, 12, 0, 69, 9, 0, 1, 1, 1, 1, 3, 1, 73, 8, 1, 1, 1, 1, 1, 3, 1, 77,
		8, 1, 1, 1, 1, 1, 3, 1, 81, 8, 1, 1, 2, 1, 2, 1, 2, 5, 2, 86, 8, 2, 10,
		2, 12, 2, 89, 9, 2, 1, 3, 3, 3, 92, 8, 3, 1, 3, 1, 3, 1, 3, 3, 3, 97, 8,
		3, 1, 4, 1, 4, 1, 4, 1, 5, 3, 5, 103, 8, 5, 1, 5, 1, 5, 3, 5, 107, 8, 5,
		1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 1, 7, 1, 7, 1, 8, 1, 8,
		1, 8, 5, 8, 122, 8, 8, 10, 8, 12, 8, 125, 9, 8, 1, 9, 1, 9, 3, 9, 129,
		8, 9, 1, 10, 1, 10, 1, 10, 1, 10, 1, 10, 5, 10, 136, 8, 10, 10, 10, 12,
		10, 139, 9, 10, 1, 10, 1, 10, 1, 11, 1, 11, 1, 11, 1, 11, 1, 11, 1, 12,
		1, 12, 1, 12, 1, 12, 1, 13, 1, 13, 3, 13, 154, 8, 13, 1, 14, 3, 14, 157,
		8, 14, 1, 14, 1, 14, 1, 14, 0, 1, 0, 15, 0, 2, 4, 6, 8, 10, 12, 14, 16,
		18, 20, 22, 24, 26, 28, 0, 5, 1, 0, 16, 17, 1, 0, 6, 7, 2, 0, 15, 15, 18,
		18, 2, 0, 1, 5, 9, 9, 1, 0, 25, 26, 172, 0, 43, 1, 0, 0, 0, 2, 80, 1, 0,
		0, 0, 4, 82, 1, 0, 0, 0, 6, 91, 1, 0, 0, 0, 8, 98, 1, 0, 0, 0, 10, 102,
		1, 0, 0, 0, 12, 110, 1, 0, 0, 0, 14, 113, 1, 0, 0, 0, 16, 118, 1, 0, 0,
		0, 18, 128, 1, 0, 0, 0, 20, 130, 1, 0, 0, 0, 22, 142, 1, 0, 0, 0, 24, 147,
		1, 0, 0, 0, 26, 153, 1, 0, 0, 0, 28, 156, 1, 0, 0, 0, 30, 31, 6, 0, -1,
		0, 31, 32, 5, 10, 0, 0, 32, 33, 3, 0, 0, 0, 33, 34, 5, 11, 0, 0, 34, 44,
		1, 0, 0, 0, 35, 36, 7, 0, 0, 0, 36, 44, 3, 0, 0, 12, 37, 44, 3, 2, 1, 0,
		38, 44, 3, 6, 3, 0, 39, 44, 5, 24, 0, 0, 40, 44, 5, 27, 0, 0, 41, 44, 7,
		1, 0, 0, 42, 44, 5, 8, 0, 0, 43, 30, 1, 0, 0, 0, 43, 35, 1, 0, 0, 0, 43,
		37, 1, 0, 0, 0, 43, 38, 1, 0, 0, 0, 43, 39, 1, 0, 0, 0, 43, 40, 1, 0, 0,
		0, 43, 41, 1, 0, 0, 0, 43, 42, 1, 0, 0, 0, 44, 67, 1, 0, 0, 0, 45, 46,
		10, 13, 0, 0, 46, 47, 5, 20, 0, 0, 47, 66, 3, 0, 0, 13, 48, 49, 10, 11,
		0, 0, 49, 50, 7, 2, 0, 0, 50, 66, 3, 0, 0, 12, 51, 52, 10, 10, 0, 0, 52,
		53, 7, 0, 0, 0, 53, 66, 3, 0, 0, 11, 54, 55, 10, 9, 0, 0, 55, 56, 5, 21,
		0, 0, 56, 66, 3, 0, 0, 10, 57, 58, 10, 8, 0, 0, 58, 59, 7, 3, 0, 0, 59,
		66, 3, 0, 0, 9, 60, 61, 10, 14, 0, 0, 61, 66, 5, 19, 0, 0, 62, 63, 10,
		7, 0, 0, 63, 64, 5, 23, 0, 0, 64, 66, 3, 2, 1, 0, 65, 45, 1, 0, 0, 0, 65,
		48, 1, 0, 0, 0, 65, 51, 1, 0, 0, 0, 65, 54, 1, 0, 0, 0, 65, 57, 1, 0, 0,
		0, 65, 60, 1, 0, 0, 0, 65, 62, 1, 0, 0, 0, 66, 69, 1, 0, 0, 0, 67, 65,
		1, 0, 0, 0, 67, 68, 1, 0, 0, 0, 68, 1, 1, 0, 0, 0, 69, 67, 1, 0, 0, 0,
		70, 72, 7, 4, 0, 0, 71, 73, 5, 24, 0, 0, 72, 71, 1, 0, 0, 0, 72, 73, 1,
		0, 0, 0, 73, 74, 1, 0, 0, 0, 74, 76, 5, 10, 0, 0, 75, 77, 3, 4, 2, 0, 76,
		75, 1, 0, 0, 0, 76, 77, 1, 0, 0, 0, 77, 78, 1, 0, 0, 0, 78, 81, 5, 11,
		0, 0, 79, 81, 7, 4, 0, 0, 80, 70, 1, 0, 0, 0, 80, 79, 1, 0, 0, 0, 81, 3,
		1, 0, 0, 0, 82, 87, 3, 0, 0, 0, 83, 84, 5, 13, 0, 0, 84, 86, 3, 0, 0, 0,
		85, 83, 1, 0, 0, 0, 86, 89, 1, 0, 0, 0, 87, 85, 1, 0, 0, 0, 87, 88, 1,
		0, 0, 0, 88, 5, 1, 0, 0, 0, 89, 87, 1, 0, 0, 0, 90, 92, 3, 8, 4, 0, 91,
		90, 1, 0, 0, 0, 91, 92, 1, 0, 0, 0, 92, 93, 1, 0, 0, 0, 93, 96, 3, 10,
		5, 0, 94, 95, 5, 12, 0, 0, 95, 97, 3, 10, 5, 0, 96, 94, 1, 0, 0, 0, 96,
		97, 1, 0, 0, 0, 97, 7, 1, 0, 0, 0, 98, 99, 5, 27, 0, 0, 99, 100, 5, 22,
		0, 0, 100, 9, 1, 0, 0, 0, 101, 103, 5, 14, 0, 0, 102, 101, 1, 0, 0, 0,
		102, 103, 1, 0, 0, 0, 103, 104, 1, 0, 0, 0, 104, 106, 5, 25, 0, 0, 105,
		107, 5, 14, 0, 0, 106, 105, 1, 0, 0, 0, 106, 107, 1, 0, 0, 0, 107, 108,
		1, 0, 0, 0, 108, 109, 5, 24, 0, 0, 109, 11, 1, 0, 0, 0, 110, 111, 3, 14,
		7, 0, 111, 112, 5, 0, 0, 1, 112, 13, 1, 0, 0, 0, 113, 114, 5, 26, 0, 0,
		114, 115, 5, 10, 0, 0, 115, 116, 3, 16, 8, 0, 116, 117, 5, 11, 0, 0, 117,
		15, 1, 0, 0, 0, 118, 123, 3, 18, 9, 0, 119, 120, 5, 13, 0, 0, 120, 122,
		3, 18, 9, 0, 121, 119, 1, 0, 0, 0, 122, 125, 1, 0, 0, 0, 123, 121, 1, 0,
		0, 0, 123, 124, 1, 0, 0, 0, 124, 17, 1, 0, 0, 0, 125, 123, 1, 0, 0, 0,
		126, 129, 3, 20, 10, 0, 127, 129, 3, 22, 11, 0, 128, 126, 1, 0, 0, 0, 128,
		127, 1, 0, 0, 0, 129, 19, 1, 0, 0, 0, 130, 131, 5, 26, 0, 0, 131, 132,
		5, 10, 0, 0, 132, 137, 3, 24, 12, 0, 133, 134, 5, 13, 0, 0, 134, 136, 3,
		24, 12, 0, 135, 133, 1, 0, 0, 0, 136, 139, 1, 0, 0, 0, 137, 135, 1, 0,
		0, 0, 137, 138, 1, 0, 0, 0, 138, 140, 1, 0, 0, 0, 139, 137, 1, 0, 0, 0,
		140, 141, 5, 11, 0, 0, 141, 21, 1, 0, 0, 0, 142, 143, 5, 26, 0, 0, 143,
		144, 5, 10, 0, 0, 144, 145, 3, 28, 14, 0, 145, 146, 5, 11, 0, 0, 146, 23,
		1, 0, 0, 0, 147, 148, 3, 26, 13, 0, 148, 149, 5, 12, 0, 0, 149, 150, 3,
		26, 13, 0, 150, 25, 1, 0, 0, 0, 151, 154, 5, 25, 0, 0, 152, 154, 3, 28,
		14, 0, 153, 151, 1, 0, 0, 0, 153, 152, 1, 0, 0, 0, 154, 27, 1, 0, 0, 0,
		155, 157, 5, 17, 0, 0, 156, 155, 1, 0, 0, 0, 156, 157, 1, 0, 0, 0, 157,
		158, 1, 0, 0, 0, 158, 159, 5, 24, 0, 0, 159, 29, 1, 0, 0, 0, 16, 43, 65,
		67, 72, 76, 80, 87, 91, 96, 102, 106, 123, 128, 137, 153, 156,
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

// TsvsheetParserInit initializes any static state used to implement TsvsheetParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewTsvsheetParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func TsvsheetParserInit() {
	staticData := &TsvsheetParserParserStaticData
	staticData.once.Do(tsvsheetparserParserInit)
}

// NewTsvsheetParser produces a new parser instance for the optional input antlr.TokenStream.
func NewTsvsheetParser(input antlr.TokenStream) *TsvsheetParser {
	TsvsheetParserInit()
	this := new(TsvsheetParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &TsvsheetParserParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "TsvsheetParser.g4"

	return this
}

// TsvsheetParser tokens.
const (
	TsvsheetParserEOF        = antlr.TokenEOF
	TsvsheetParserGE         = 1
	TsvsheetParserLE         = 2
	TsvsheetParserNE         = 3
	TsvsheetParserGT         = 4
	TsvsheetParserLT         = 5
	TsvsheetParserTRUE       = 6
	TsvsheetParserFALSE      = 7
	TsvsheetParserERRORCONST = 8
	TsvsheetParserEQ         = 9
	TsvsheetParserLPAREN     = 10
	TsvsheetParserRPAREN     = 11
	TsvsheetParserCOLON      = 12
	TsvsheetParserCOMMA      = 13
	TsvsheetParserDOLLAR     = 14
	TsvsheetParserSTAR       = 15
	TsvsheetParserPLUS       = 16
	TsvsheetParserDASH       = 17
	TsvsheetParserSLASH      = 18
	TsvsheetParserPERCENT    = 19
	TsvsheetParserCARET      = 20
	TsvsheetParserAMP        = 21
	TsvsheetParserBANG       = 22
	TsvsheetParserPIPE       = 23
	TsvsheetParserNUMBER     = 24
	TsvsheetParserCOL        = 25
	TsvsheetParserNAME       = 26
	TsvsheetParserSTRING     = 27
	TsvsheetParserWS         = 28
)

// TsvsheetParser rules.
const (
	TsvsheetParserRULE_expression     = 0
	TsvsheetParserRULE_functionCall   = 1
	TsvsheetParserRULE_argList        = 2
	TsvsheetParserRULE_reference      = 3
	TsvsheetParserRULE_sheetQualifier = 4
	TsvsheetParserRULE_cellRef        = 5
	TsvsheetParserRULE_directiveValue = 6
	TsvsheetParserRULE_axisCall       = 7
	TsvsheetParserRULE_itemList       = 8
	TsvsheetParserRULE_item           = 9
	TsvsheetParserRULE_rangeCall      = 10
	TsvsheetParserRULE_countCall      = 11
	TsvsheetParserRULE_span           = 12
	TsvsheetParserRULE_endpoint       = 13
	TsvsheetParserRULE_offset         = 14
)

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) CopyAll(ctx *ExpressionContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type ErrorExprContext struct {
	ExpressionContext
}

func NewErrorExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ErrorExprContext {
	var p = new(ErrorExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ErrorExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ErrorExprContext) ERRORCONST() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserERRORCONST, 0)
}

func (s *ErrorExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterErrorExpr(s)
	}
}

func (s *ErrorExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitErrorExpr(s)
	}
}

func (s *ErrorExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitErrorExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type PipeExprContext struct {
	ExpressionContext
}

func NewPipeExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PipeExprContext {
	var p = new(PipeExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *PipeExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PipeExprContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *PipeExprContext) PIPE() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserPIPE, 0)
}

func (s *PipeExprContext) FunctionCall() IFunctionCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionCallContext)
}

func (s *PipeExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterPipeExpr(s)
	}
}

func (s *PipeExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitPipeExpr(s)
	}
}

func (s *PipeExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitPipeExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type NumberExprContext struct {
	ExpressionContext
}

func NewNumberExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *NumberExprContext {
	var p = new(NumberExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *NumberExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *NumberExprContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNUMBER, 0)
}

func (s *NumberExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterNumberExpr(s)
	}
}

func (s *NumberExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitNumberExpr(s)
	}
}

func (s *NumberExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitNumberExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type ParenExprContext struct {
	ExpressionContext
}

func NewParenExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ParenExprContext {
	var p = new(ParenExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ParenExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ParenExprContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserLPAREN, 0)
}

func (s *ParenExprContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ParenExprContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserRPAREN, 0)
}

func (s *ParenExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterParenExpr(s)
	}
}

func (s *ParenExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitParenExpr(s)
	}
}

func (s *ParenExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitParenExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type ConcatExprContext struct {
	ExpressionContext
}

func NewConcatExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *ConcatExprContext {
	var p = new(ConcatExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *ConcatExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ConcatExprContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ConcatExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ConcatExprContext) AMP() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserAMP, 0)
}

func (s *ConcatExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterConcatExpr(s)
	}
}

func (s *ConcatExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitConcatExpr(s)
	}
}

func (s *ConcatExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitConcatExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type StringExprContext struct {
	ExpressionContext
}

func NewStringExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *StringExprContext {
	var p = new(StringExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *StringExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *StringExprContext) STRING() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserSTRING, 0)
}

func (s *StringExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterStringExpr(s)
	}
}

func (s *StringExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitStringExpr(s)
	}
}

func (s *StringExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitStringExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type UnaryExprContext struct {
	ExpressionContext
	op antlr.Token
}

func NewUnaryExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *UnaryExprContext {
	var p = new(UnaryExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *UnaryExprContext) GetOp() antlr.Token { return s.op }

func (s *UnaryExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *UnaryExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *UnaryExprContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *UnaryExprContext) PLUS() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserPLUS, 0)
}

func (s *UnaryExprContext) DASH() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserDASH, 0)
}

func (s *UnaryExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterUnaryExpr(s)
	}
}

func (s *UnaryExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitUnaryExpr(s)
	}
}

func (s *UnaryExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitUnaryExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type AddExprContext struct {
	ExpressionContext
	op antlr.Token
}

func NewAddExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *AddExprContext {
	var p = new(AddExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *AddExprContext) GetOp() antlr.Token { return s.op }

func (s *AddExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *AddExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AddExprContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *AddExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *AddExprContext) PLUS() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserPLUS, 0)
}

func (s *AddExprContext) DASH() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserDASH, 0)
}

func (s *AddExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterAddExpr(s)
	}
}

func (s *AddExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitAddExpr(s)
	}
}

func (s *AddExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitAddExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type RefExprContext struct {
	ExpressionContext
}

func NewRefExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *RefExprContext {
	var p = new(RefExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *RefExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RefExprContext) Reference() IReferenceContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IReferenceContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IReferenceContext)
}

func (s *RefExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterRefExpr(s)
	}
}

func (s *RefExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitRefExpr(s)
	}
}

func (s *RefExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitRefExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type MulExprContext struct {
	ExpressionContext
	op antlr.Token
}

func NewMulExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *MulExprContext {
	var p = new(MulExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *MulExprContext) GetOp() antlr.Token { return s.op }

func (s *MulExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *MulExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *MulExprContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *MulExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *MulExprContext) STAR() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserSTAR, 0)
}

func (s *MulExprContext) SLASH() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserSLASH, 0)
}

func (s *MulExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterMulExpr(s)
	}
}

func (s *MulExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitMulExpr(s)
	}
}

func (s *MulExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitMulExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type PercentExprContext struct {
	ExpressionContext
}

func NewPercentExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PercentExprContext {
	var p = new(PercentExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *PercentExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PercentExprContext) Expression() IExpressionContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *PercentExprContext) PERCENT() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserPERCENT, 0)
}

func (s *PercentExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterPercentExpr(s)
	}
}

func (s *PercentExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitPercentExpr(s)
	}
}

func (s *PercentExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitPercentExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type CallExprContext struct {
	ExpressionContext
}

func NewCallExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CallExprContext {
	var p = new(CallExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *CallExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CallExprContext) FunctionCall() IFunctionCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IFunctionCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IFunctionCallContext)
}

func (s *CallExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterCallExpr(s)
	}
}

func (s *CallExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitCallExpr(s)
	}
}

func (s *CallExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitCallExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type BoolExprContext struct {
	ExpressionContext
}

func NewBoolExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *BoolExprContext {
	var p = new(BoolExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *BoolExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *BoolExprContext) TRUE() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserTRUE, 0)
}

func (s *BoolExprContext) FALSE() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserFALSE, 0)
}

func (s *BoolExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterBoolExpr(s)
	}
}

func (s *BoolExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitBoolExpr(s)
	}
}

func (s *BoolExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitBoolExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type PowExprContext struct {
	ExpressionContext
}

func NewPowExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PowExprContext {
	var p = new(PowExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *PowExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PowExprContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *PowExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *PowExprContext) CARET() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCARET, 0)
}

func (s *PowExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterPowExpr(s)
	}
}

func (s *PowExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitPowExpr(s)
	}
}

func (s *PowExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitPowExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

type CompareExprContext struct {
	ExpressionContext
	op antlr.Token
}

func NewCompareExprContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *CompareExprContext {
	var p = new(CompareExprContext)

	InitEmptyExpressionContext(&p.ExpressionContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExpressionContext))

	return p
}

func (s *CompareExprContext) GetOp() antlr.Token { return s.op }

func (s *CompareExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *CompareExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CompareExprContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *CompareExprContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *CompareExprContext) EQ() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserEQ, 0)
}

func (s *CompareExprContext) NE() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNE, 0)
}

func (s *CompareExprContext) LT() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserLT, 0)
}

func (s *CompareExprContext) LE() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserLE, 0)
}

func (s *CompareExprContext) GT() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserGT, 0)
}

func (s *CompareExprContext) GE() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserGE, 0)
}

func (s *CompareExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterCompareExpr(s)
	}
}

func (s *CompareExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitCompareExpr(s)
	}
}

func (s *CompareExprContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitCompareExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) Expression() (localctx IExpressionContext) {
	return p.expression(0)
}

func (p *TsvsheetParser) expression(_p int) (localctx IExpressionContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExpressionContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 0
	p.EnterRecursionRule(localctx, 0, TsvsheetParserRULE_expression, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(43)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 0, p.GetParserRuleContext()) {
	case 1:
		localctx = NewParenExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(31)
			p.Match(TsvsheetParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(32)
			p.expression(0)
		}
		{
			p.SetState(33)
			p.Match(TsvsheetParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewUnaryExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(35)

			var _lt = p.GetTokenStream().LT(1)

			localctx.(*UnaryExprContext).op = _lt

			_la = p.GetTokenStream().LA(1)

			if !(_la == TsvsheetParserPLUS || _la == TsvsheetParserDASH) {
				var _ri = p.GetErrorHandler().RecoverInline(p)

				localctx.(*UnaryExprContext).op = _ri
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		{
			p.SetState(36)
			p.expression(12)
		}

	case 3:
		localctx = NewCallExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(37)
			p.FunctionCall()
		}

	case 4:
		localctx = NewRefExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(38)
			p.Reference()
		}

	case 5:
		localctx = NewNumberExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(39)
			p.Match(TsvsheetParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 6:
		localctx = NewStringExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(40)
			p.Match(TsvsheetParserSTRING)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 7:
		localctx = NewBoolExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(41)
			_la = p.GetTokenStream().LA(1)

			if !(_la == TsvsheetParserTRUE || _la == TsvsheetParserFALSE) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	case 8:
		localctx = NewErrorExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(42)
			p.Match(TsvsheetParserERRORCONST)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(67)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(65)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext()) {
			case 1:
				localctx = NewPowExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(45)

				if !(p.Precpred(p.GetParserRuleContext(), 13)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 13)", ""))
					goto errorExit
				}
				{
					p.SetState(46)
					p.Match(TsvsheetParserCARET)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(47)
					p.expression(13)
				}

			case 2:
				localctx = NewMulExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(48)

				if !(p.Precpred(p.GetParserRuleContext(), 11)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 11)", ""))
					goto errorExit
				}
				{
					p.SetState(49)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*MulExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == TsvsheetParserSTAR || _la == TsvsheetParserSLASH) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*MulExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(50)
					p.expression(12)
				}

			case 3:
				localctx = NewAddExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(51)

				if !(p.Precpred(p.GetParserRuleContext(), 10)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 10)", ""))
					goto errorExit
				}
				{
					p.SetState(52)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*AddExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == TsvsheetParserPLUS || _la == TsvsheetParserDASH) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*AddExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(53)
					p.expression(11)
				}

			case 4:
				localctx = NewConcatExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(54)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
					goto errorExit
				}
				{
					p.SetState(55)
					p.Match(TsvsheetParserAMP)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(56)
					p.expression(10)
				}

			case 5:
				localctx = NewCompareExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(57)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
					goto errorExit
				}
				{
					p.SetState(58)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*CompareExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&574) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*CompareExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(59)
					p.expression(9)
				}

			case 6:
				localctx = NewPercentExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(60)

				if !(p.Precpred(p.GetParserRuleContext(), 14)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 14)", ""))
					goto errorExit
				}
				{
					p.SetState(61)
					p.Match(TsvsheetParserPERCENT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case 7:
				localctx = NewPipeExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(62)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(63)
					p.Match(TsvsheetParserPIPE)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(64)
					p.FunctionCall()
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(69)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IFunctionCallContext is an interface to support dynamic dispatch.
type IFunctionCallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	LPAREN() antlr.TerminalNode
	RPAREN() antlr.TerminalNode
	NAME() antlr.TerminalNode
	COL() antlr.TerminalNode
	NUMBER() antlr.TerminalNode
	ArgList() IArgListContext

	// IsFunctionCallContext differentiates from other interfaces.
	IsFunctionCallContext()
}

type FunctionCallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyFunctionCallContext() *FunctionCallContext {
	var p = new(FunctionCallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_functionCall
	return p
}

func InitEmptyFunctionCallContext(p *FunctionCallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_functionCall
}

func (*FunctionCallContext) IsFunctionCallContext() {}

func NewFunctionCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *FunctionCallContext {
	var p = new(FunctionCallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_functionCall

	return p
}

func (s *FunctionCallContext) GetParser() antlr.Parser { return s.parser }

func (s *FunctionCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserLPAREN, 0)
}

func (s *FunctionCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserRPAREN, 0)
}

func (s *FunctionCallContext) NAME() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNAME, 0)
}

func (s *FunctionCallContext) COL() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOL, 0)
}

func (s *FunctionCallContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNUMBER, 0)
}

func (s *FunctionCallContext) ArgList() IArgListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArgListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArgListContext)
}

func (s *FunctionCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionCallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *FunctionCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterFunctionCall(s)
	}
}

func (s *FunctionCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitFunctionCall(s)
	}
}

func (s *FunctionCallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitFunctionCall(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) FunctionCall() (localctx IFunctionCallContext) {
	localctx = NewFunctionCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, TsvsheetParserRULE_functionCall)
	var _la int

	p.SetState(80)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(70)
			_la = p.GetTokenStream().LA(1)

			if !(_la == TsvsheetParserCOL || _la == TsvsheetParserNAME) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		p.SetState(72)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == TsvsheetParserNUMBER {
			{
				p.SetState(71)
				p.Match(TsvsheetParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(74)
			p.Match(TsvsheetParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(76)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&251872704) != 0 {
			{
				p.SetState(75)
				p.ArgList()
			}

		}
		{
			p.SetState(78)
			p.Match(TsvsheetParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(79)
			_la = p.GetTokenStream().LA(1)

			if !(_la == TsvsheetParserCOL || _la == TsvsheetParserNAME) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArgListContext is an interface to support dynamic dispatch.
type IArgListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllExpression() []IExpressionContext
	Expression(i int) IExpressionContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsArgListContext differentiates from other interfaces.
	IsArgListContext()
}

type ArgListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArgListContext() *ArgListContext {
	var p = new(ArgListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_argList
	return p
}

func InitEmptyArgListContext(p *ArgListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_argList
}

func (*ArgListContext) IsArgListContext() {}

func NewArgListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArgListContext {
	var p = new(ArgListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_argList

	return p
}

func (s *ArgListContext) GetParser() antlr.Parser { return s.parser }

func (s *ArgListContext) AllExpression() []IExpressionContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExpressionContext); ok {
			len++
		}
	}

	tst := make([]IExpressionContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExpressionContext); ok {
			tst[i] = t.(IExpressionContext)
			i++
		}
	}

	return tst
}

func (s *ArgListContext) Expression(i int) IExpressionContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExpressionContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExpressionContext)
}

func (s *ArgListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserCOMMA)
}

func (s *ArgListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOMMA, i)
}

func (s *ArgListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArgListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArgListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterArgList(s)
	}
}

func (s *ArgListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitArgList(s)
	}
}

func (s *ArgListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitArgList(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) ArgList() (localctx IArgListContext) {
	localctx = NewArgListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, TsvsheetParserRULE_argList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(82)
		p.expression(0)
	}
	p.SetState(87)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == TsvsheetParserCOMMA {
		{
			p.SetState(83)
			p.Match(TsvsheetParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(84)
			p.expression(0)
		}

		p.SetState(89)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IReferenceContext is an interface to support dynamic dispatch.
type IReferenceContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllCellRef() []ICellRefContext
	CellRef(i int) ICellRefContext
	SheetQualifier() ISheetQualifierContext
	COLON() antlr.TerminalNode

	// IsReferenceContext differentiates from other interfaces.
	IsReferenceContext()
}

type ReferenceContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyReferenceContext() *ReferenceContext {
	var p = new(ReferenceContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_reference
	return p
}

func InitEmptyReferenceContext(p *ReferenceContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_reference
}

func (*ReferenceContext) IsReferenceContext() {}

func NewReferenceContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ReferenceContext {
	var p = new(ReferenceContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_reference

	return p
}

func (s *ReferenceContext) GetParser() antlr.Parser { return s.parser }

func (s *ReferenceContext) AllCellRef() []ICellRefContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ICellRefContext); ok {
			len++
		}
	}

	tst := make([]ICellRefContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ICellRefContext); ok {
			tst[i] = t.(ICellRefContext)
			i++
		}
	}

	return tst
}

func (s *ReferenceContext) CellRef(i int) ICellRefContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICellRefContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICellRefContext)
}

func (s *ReferenceContext) SheetQualifier() ISheetQualifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISheetQualifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISheetQualifierContext)
}

func (s *ReferenceContext) COLON() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOLON, 0)
}

func (s *ReferenceContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ReferenceContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ReferenceContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterReference(s)
	}
}

func (s *ReferenceContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitReference(s)
	}
}

func (s *ReferenceContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitReference(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) Reference() (localctx IReferenceContext) {
	localctx = NewReferenceContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, TsvsheetParserRULE_reference)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(91)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserSTRING {
		{
			p.SetState(90)
			p.SheetQualifier()
		}

	}
	{
		p.SetState(93)
		p.CellRef()
	}
	p.SetState(96)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(94)
			p.Match(TsvsheetParserCOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(95)
			p.CellRef()
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISheetQualifierContext is an interface to support dynamic dispatch.
type ISheetQualifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	BANG() antlr.TerminalNode

	// IsSheetQualifierContext differentiates from other interfaces.
	IsSheetQualifierContext()
}

type SheetQualifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySheetQualifierContext() *SheetQualifierContext {
	var p = new(SheetQualifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_sheetQualifier
	return p
}

func InitEmptySheetQualifierContext(p *SheetQualifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_sheetQualifier
}

func (*SheetQualifierContext) IsSheetQualifierContext() {}

func NewSheetQualifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SheetQualifierContext {
	var p = new(SheetQualifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_sheetQualifier

	return p
}

func (s *SheetQualifierContext) GetParser() antlr.Parser { return s.parser }

func (s *SheetQualifierContext) STRING() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserSTRING, 0)
}

func (s *SheetQualifierContext) BANG() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserBANG, 0)
}

func (s *SheetQualifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SheetQualifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SheetQualifierContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterSheetQualifier(s)
	}
}

func (s *SheetQualifierContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitSheetQualifier(s)
	}
}

func (s *SheetQualifierContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitSheetQualifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) SheetQualifier() (localctx ISheetQualifierContext) {
	localctx = NewSheetQualifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, TsvsheetParserRULE_sheetQualifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(98)
		p.Match(TsvsheetParserSTRING)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(99)
		p.Match(TsvsheetParserBANG)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICellRefContext is an interface to support dynamic dispatch.
type ICellRefContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	COL() antlr.TerminalNode
	NUMBER() antlr.TerminalNode
	AllDOLLAR() []antlr.TerminalNode
	DOLLAR(i int) antlr.TerminalNode

	// IsCellRefContext differentiates from other interfaces.
	IsCellRefContext()
}

type CellRefContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCellRefContext() *CellRefContext {
	var p = new(CellRefContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_cellRef
	return p
}

func InitEmptyCellRefContext(p *CellRefContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_cellRef
}

func (*CellRefContext) IsCellRefContext() {}

func NewCellRefContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CellRefContext {
	var p = new(CellRefContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_cellRef

	return p
}

func (s *CellRefContext) GetParser() antlr.Parser { return s.parser }

func (s *CellRefContext) COL() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOL, 0)
}

func (s *CellRefContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNUMBER, 0)
}

func (s *CellRefContext) AllDOLLAR() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserDOLLAR)
}

func (s *CellRefContext) DOLLAR(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserDOLLAR, i)
}

func (s *CellRefContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CellRefContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CellRefContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterCellRef(s)
	}
}

func (s *CellRefContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitCellRef(s)
	}
}

func (s *CellRefContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitCellRef(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) CellRef() (localctx ICellRefContext) {
	localctx = NewCellRefContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 10, TsvsheetParserRULE_cellRef)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(102)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserDOLLAR {
		{
			p.SetState(101)
			p.Match(TsvsheetParserDOLLAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(104)
		p.Match(TsvsheetParserCOL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(106)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserDOLLAR {
		{
			p.SetState(105)
			p.Match(TsvsheetParserDOLLAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(108)
		p.Match(TsvsheetParserNUMBER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IDirectiveValueContext is an interface to support dynamic dispatch.
type IDirectiveValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AxisCall() IAxisCallContext
	EOF() antlr.TerminalNode

	// IsDirectiveValueContext differentiates from other interfaces.
	IsDirectiveValueContext()
}

type DirectiveValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyDirectiveValueContext() *DirectiveValueContext {
	var p = new(DirectiveValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_directiveValue
	return p
}

func InitEmptyDirectiveValueContext(p *DirectiveValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_directiveValue
}

func (*DirectiveValueContext) IsDirectiveValueContext() {}

func NewDirectiveValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *DirectiveValueContext {
	var p = new(DirectiveValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_directiveValue

	return p
}

func (s *DirectiveValueContext) GetParser() antlr.Parser { return s.parser }

func (s *DirectiveValueContext) AxisCall() IAxisCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IAxisCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IAxisCallContext)
}

func (s *DirectiveValueContext) EOF() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserEOF, 0)
}

func (s *DirectiveValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *DirectiveValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *DirectiveValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterDirectiveValue(s)
	}
}

func (s *DirectiveValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitDirectiveValue(s)
	}
}

func (s *DirectiveValueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitDirectiveValue(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) DirectiveValue() (localctx IDirectiveValueContext) {
	localctx = NewDirectiveValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, TsvsheetParserRULE_directiveValue)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(110)
		p.AxisCall()
	}
	{
		p.SetState(111)
		p.Match(TsvsheetParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IAxisCallContext is an interface to support dynamic dispatch.
type IAxisCallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NAME() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	ItemList() IItemListContext
	RPAREN() antlr.TerminalNode

	// IsAxisCallContext differentiates from other interfaces.
	IsAxisCallContext()
}

type AxisCallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyAxisCallContext() *AxisCallContext {
	var p = new(AxisCallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_axisCall
	return p
}

func InitEmptyAxisCallContext(p *AxisCallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_axisCall
}

func (*AxisCallContext) IsAxisCallContext() {}

func NewAxisCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *AxisCallContext {
	var p = new(AxisCallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_axisCall

	return p
}

func (s *AxisCallContext) GetParser() antlr.Parser { return s.parser }

func (s *AxisCallContext) NAME() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNAME, 0)
}

func (s *AxisCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserLPAREN, 0)
}

func (s *AxisCallContext) ItemList() IItemListContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IItemListContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IItemListContext)
}

func (s *AxisCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserRPAREN, 0)
}

func (s *AxisCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *AxisCallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *AxisCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterAxisCall(s)
	}
}

func (s *AxisCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitAxisCall(s)
	}
}

func (s *AxisCallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitAxisCall(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) AxisCall() (localctx IAxisCallContext) {
	localctx = NewAxisCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, TsvsheetParserRULE_axisCall)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(113)
		p.Match(TsvsheetParserNAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(114)
		p.Match(TsvsheetParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(115)
		p.ItemList()
	}
	{
		p.SetState(116)
		p.Match(TsvsheetParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IItemListContext is an interface to support dynamic dispatch.
type IItemListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllItem() []IItemContext
	Item(i int) IItemContext
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsItemListContext differentiates from other interfaces.
	IsItemListContext()
}

type ItemListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyItemListContext() *ItemListContext {
	var p = new(ItemListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_itemList
	return p
}

func InitEmptyItemListContext(p *ItemListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_itemList
}

func (*ItemListContext) IsItemListContext() {}

func NewItemListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ItemListContext {
	var p = new(ItemListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_itemList

	return p
}

func (s *ItemListContext) GetParser() antlr.Parser { return s.parser }

func (s *ItemListContext) AllItem() []IItemContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IItemContext); ok {
			len++
		}
	}

	tst := make([]IItemContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IItemContext); ok {
			tst[i] = t.(IItemContext)
			i++
		}
	}

	return tst
}

func (s *ItemListContext) Item(i int) IItemContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IItemContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IItemContext)
}

func (s *ItemListContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserCOMMA)
}

func (s *ItemListContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOMMA, i)
}

func (s *ItemListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ItemListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ItemListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterItemList(s)
	}
}

func (s *ItemListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitItemList(s)
	}
}

func (s *ItemListContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitItemList(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) ItemList() (localctx IItemListContext) {
	localctx = NewItemListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, TsvsheetParserRULE_itemList)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(118)
		p.Item()
	}
	p.SetState(123)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == TsvsheetParserCOMMA {
		{
			p.SetState(119)
			p.Match(TsvsheetParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(120)
			p.Item()
		}

		p.SetState(125)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IItemContext is an interface to support dynamic dispatch.
type IItemContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	RangeCall() IRangeCallContext
	CountCall() ICountCallContext

	// IsItemContext differentiates from other interfaces.
	IsItemContext()
}

type ItemContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyItemContext() *ItemContext {
	var p = new(ItemContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_item
	return p
}

func InitEmptyItemContext(p *ItemContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_item
}

func (*ItemContext) IsItemContext() {}

func NewItemContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ItemContext {
	var p = new(ItemContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_item

	return p
}

func (s *ItemContext) GetParser() antlr.Parser { return s.parser }

func (s *ItemContext) RangeCall() IRangeCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRangeCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IRangeCallContext)
}

func (s *ItemContext) CountCall() ICountCallContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ICountCallContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ICountCallContext)
}

func (s *ItemContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ItemContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ItemContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterItem(s)
	}
}

func (s *ItemContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitItem(s)
	}
}

func (s *ItemContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitItem(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) Item() (localctx IItemContext) {
	localctx = NewItemContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, TsvsheetParserRULE_item)
	p.SetState(128)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(126)
			p.RangeCall()
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(127)
			p.CountCall()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IRangeCallContext is an interface to support dynamic dispatch.
type IRangeCallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NAME() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	AllSpan() []ISpanContext
	Span(i int) ISpanContext
	RPAREN() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsRangeCallContext differentiates from other interfaces.
	IsRangeCallContext()
}

type RangeCallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRangeCallContext() *RangeCallContext {
	var p = new(RangeCallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_rangeCall
	return p
}

func InitEmptyRangeCallContext(p *RangeCallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_rangeCall
}

func (*RangeCallContext) IsRangeCallContext() {}

func NewRangeCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RangeCallContext {
	var p = new(RangeCallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_rangeCall

	return p
}

func (s *RangeCallContext) GetParser() antlr.Parser { return s.parser }

func (s *RangeCallContext) NAME() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNAME, 0)
}

func (s *RangeCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserLPAREN, 0)
}

func (s *RangeCallContext) AllSpan() []ISpanContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISpanContext); ok {
			len++
		}
	}

	tst := make([]ISpanContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISpanContext); ok {
			tst[i] = t.(ISpanContext)
			i++
		}
	}

	return tst
}

func (s *RangeCallContext) Span(i int) ISpanContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISpanContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISpanContext)
}

func (s *RangeCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserRPAREN, 0)
}

func (s *RangeCallContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserCOMMA)
}

func (s *RangeCallContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOMMA, i)
}

func (s *RangeCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RangeCallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RangeCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterRangeCall(s)
	}
}

func (s *RangeCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitRangeCall(s)
	}
}

func (s *RangeCallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitRangeCall(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) RangeCall() (localctx IRangeCallContext) {
	localctx = NewRangeCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, TsvsheetParserRULE_rangeCall)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(130)
		p.Match(TsvsheetParserNAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(131)
		p.Match(TsvsheetParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(132)
		p.Span()
	}
	p.SetState(137)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == TsvsheetParserCOMMA {
		{
			p.SetState(133)
			p.Match(TsvsheetParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(134)
			p.Span()
		}

		p.SetState(139)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(140)
		p.Match(TsvsheetParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ICountCallContext is an interface to support dynamic dispatch.
type ICountCallContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NAME() antlr.TerminalNode
	LPAREN() antlr.TerminalNode
	Offset() IOffsetContext
	RPAREN() antlr.TerminalNode

	// IsCountCallContext differentiates from other interfaces.
	IsCountCallContext()
}

type CountCallContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCountCallContext() *CountCallContext {
	var p = new(CountCallContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_countCall
	return p
}

func InitEmptyCountCallContext(p *CountCallContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_countCall
}

func (*CountCallContext) IsCountCallContext() {}

func NewCountCallContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CountCallContext {
	var p = new(CountCallContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_countCall

	return p
}

func (s *CountCallContext) GetParser() antlr.Parser { return s.parser }

func (s *CountCallContext) NAME() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNAME, 0)
}

func (s *CountCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserLPAREN, 0)
}

func (s *CountCallContext) Offset() IOffsetContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOffsetContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOffsetContext)
}

func (s *CountCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserRPAREN, 0)
}

func (s *CountCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CountCallContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CountCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterCountCall(s)
	}
}

func (s *CountCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitCountCall(s)
	}
}

func (s *CountCallContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitCountCall(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) CountCall() (localctx ICountCallContext) {
	localctx = NewCountCallContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 22, TsvsheetParserRULE_countCall)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(142)
		p.Match(TsvsheetParserNAME)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(143)
		p.Match(TsvsheetParserLPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(144)
		p.Offset()
	}
	{
		p.SetState(145)
		p.Match(TsvsheetParserRPAREN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISpanContext is an interface to support dynamic dispatch.
type ISpanContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllEndpoint() []IEndpointContext
	Endpoint(i int) IEndpointContext
	COLON() antlr.TerminalNode

	// IsSpanContext differentiates from other interfaces.
	IsSpanContext()
}

type SpanContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySpanContext() *SpanContext {
	var p = new(SpanContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_span
	return p
}

func InitEmptySpanContext(p *SpanContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_span
}

func (*SpanContext) IsSpanContext() {}

func NewSpanContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SpanContext {
	var p = new(SpanContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_span

	return p
}

func (s *SpanContext) GetParser() antlr.Parser { return s.parser }

func (s *SpanContext) AllEndpoint() []IEndpointContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IEndpointContext); ok {
			len++
		}
	}

	tst := make([]IEndpointContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IEndpointContext); ok {
			tst[i] = t.(IEndpointContext)
			i++
		}
	}

	return tst
}

func (s *SpanContext) Endpoint(i int) IEndpointContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IEndpointContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IEndpointContext)
}

func (s *SpanContext) COLON() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOLON, 0)
}

func (s *SpanContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SpanContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SpanContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterSpan(s)
	}
}

func (s *SpanContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitSpan(s)
	}
}

func (s *SpanContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitSpan(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) Span() (localctx ISpanContext) {
	localctx = NewSpanContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 24, TsvsheetParserRULE_span)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(147)
		p.Endpoint()
	}
	{
		p.SetState(148)
		p.Match(TsvsheetParserCOLON)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(149)
		p.Endpoint()
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IEndpointContext is an interface to support dynamic dispatch.
type IEndpointContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	COL() antlr.TerminalNode
	Offset() IOffsetContext

	// IsEndpointContext differentiates from other interfaces.
	IsEndpointContext()
}

type EndpointContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyEndpointContext() *EndpointContext {
	var p = new(EndpointContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_endpoint
	return p
}

func InitEmptyEndpointContext(p *EndpointContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_endpoint
}

func (*EndpointContext) IsEndpointContext() {}

func NewEndpointContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *EndpointContext {
	var p = new(EndpointContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_endpoint

	return p
}

func (s *EndpointContext) GetParser() antlr.Parser { return s.parser }

func (s *EndpointContext) COL() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOL, 0)
}

func (s *EndpointContext) Offset() IOffsetContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IOffsetContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IOffsetContext)
}

func (s *EndpointContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *EndpointContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *EndpointContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterEndpoint(s)
	}
}

func (s *EndpointContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitEndpoint(s)
	}
}

func (s *EndpointContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitEndpoint(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) Endpoint() (localctx IEndpointContext) {
	localctx = NewEndpointContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 26, TsvsheetParserRULE_endpoint)
	p.SetState(153)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case TsvsheetParserCOL:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(151)
			p.Match(TsvsheetParserCOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case TsvsheetParserDASH, TsvsheetParserNUMBER:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(152)
			p.Offset()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IOffsetContext is an interface to support dynamic dispatch.
type IOffsetContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMBER() antlr.TerminalNode
	DASH() antlr.TerminalNode

	// IsOffsetContext differentiates from other interfaces.
	IsOffsetContext()
}

type OffsetContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyOffsetContext() *OffsetContext {
	var p = new(OffsetContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_offset
	return p
}

func InitEmptyOffsetContext(p *OffsetContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_offset
}

func (*OffsetContext) IsOffsetContext() {}

func NewOffsetContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *OffsetContext {
	var p = new(OffsetContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_offset

	return p
}

func (s *OffsetContext) GetParser() antlr.Parser { return s.parser }

func (s *OffsetContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNUMBER, 0)
}

func (s *OffsetContext) DASH() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserDASH, 0)
}

func (s *OffsetContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *OffsetContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *OffsetContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterOffset(s)
	}
}

func (s *OffsetContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitOffset(s)
	}
}

func (s *OffsetContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitOffset(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) Offset() (localctx IOffsetContext) {
	localctx = NewOffsetContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 28, TsvsheetParserRULE_offset)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(156)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserDASH {
		{
			p.SetState(155)
			p.Match(TsvsheetParserDASH)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(158)
		p.Match(TsvsheetParserNUMBER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *TsvsheetParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 0:
		var t *ExpressionContext = nil
		if localctx != nil {
			t = localctx.(*ExpressionContext)
		}
		return p.Expression_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *TsvsheetParser) Expression_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 13)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 11)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 10)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 9)

	case 4:
		return p.Precpred(p.GetParserRuleContext(), 8)

	case 5:
		return p.Precpred(p.GetParserRuleContext(), 14)

	case 6:
		return p.Precpred(p.GetParserRuleContext(), 7)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
