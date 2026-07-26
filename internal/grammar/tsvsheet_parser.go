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
		"cellRef", "rowSelector", "colSelector", "countValue", "rowSpan", "colSpan",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 28, 136, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2, 10, 7,
		10, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1,
		0, 1, 0, 3, 0, 36, 8, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0,
		1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0, 1, 0,
		5, 0, 58, 8, 0, 10, 0, 12, 0, 61, 9, 0, 1, 1, 1, 1, 3, 1, 65, 8, 1, 1,
		1, 1, 1, 3, 1, 69, 8, 1, 1, 1, 1, 1, 3, 1, 73, 8, 1, 1, 2, 1, 2, 1, 2,
		5, 2, 78, 8, 2, 10, 2, 12, 2, 81, 9, 2, 1, 3, 3, 3, 84, 8, 3, 1, 3, 1,
		3, 1, 3, 3, 3, 89, 8, 3, 1, 4, 1, 4, 1, 4, 1, 5, 3, 5, 95, 8, 5, 1, 5,
		1, 5, 3, 5, 99, 8, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 6, 5, 6, 106, 8, 6, 10,
		6, 12, 6, 109, 9, 6, 1, 6, 1, 6, 1, 7, 1, 7, 1, 7, 5, 7, 116, 8, 7, 10,
		7, 12, 7, 119, 9, 7, 1, 7, 1, 7, 1, 8, 1, 8, 1, 8, 1, 9, 1, 9, 1, 9, 3,
		9, 129, 8, 9, 1, 10, 1, 10, 1, 10, 3, 10, 134, 8, 10, 1, 10, 0, 1, 0, 11,
		0, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 0, 5, 1, 0, 16, 17, 1, 0, 6, 7,
		2, 0, 15, 15, 18, 18, 2, 0, 1, 5, 9, 9, 1, 0, 25, 26, 150, 0, 35, 1, 0,
		0, 0, 2, 72, 1, 0, 0, 0, 4, 74, 1, 0, 0, 0, 6, 83, 1, 0, 0, 0, 8, 90, 1,
		0, 0, 0, 10, 94, 1, 0, 0, 0, 12, 102, 1, 0, 0, 0, 14, 112, 1, 0, 0, 0,
		16, 122, 1, 0, 0, 0, 18, 125, 1, 0, 0, 0, 20, 130, 1, 0, 0, 0, 22, 23,
		6, 0, -1, 0, 23, 24, 5, 10, 0, 0, 24, 25, 3, 0, 0, 0, 25, 26, 5, 11, 0,
		0, 26, 36, 1, 0, 0, 0, 27, 28, 7, 0, 0, 0, 28, 36, 3, 0, 0, 12, 29, 36,
		3, 2, 1, 0, 30, 36, 3, 6, 3, 0, 31, 36, 5, 24, 0, 0, 32, 36, 5, 27, 0,
		0, 33, 36, 7, 1, 0, 0, 34, 36, 5, 8, 0, 0, 35, 22, 1, 0, 0, 0, 35, 27,
		1, 0, 0, 0, 35, 29, 1, 0, 0, 0, 35, 30, 1, 0, 0, 0, 35, 31, 1, 0, 0, 0,
		35, 32, 1, 0, 0, 0, 35, 33, 1, 0, 0, 0, 35, 34, 1, 0, 0, 0, 36, 59, 1,
		0, 0, 0, 37, 38, 10, 13, 0, 0, 38, 39, 5, 20, 0, 0, 39, 58, 3, 0, 0, 13,
		40, 41, 10, 11, 0, 0, 41, 42, 7, 2, 0, 0, 42, 58, 3, 0, 0, 12, 43, 44,
		10, 10, 0, 0, 44, 45, 7, 0, 0, 0, 45, 58, 3, 0, 0, 11, 46, 47, 10, 9, 0,
		0, 47, 48, 5, 21, 0, 0, 48, 58, 3, 0, 0, 10, 49, 50, 10, 8, 0, 0, 50, 51,
		7, 3, 0, 0, 51, 58, 3, 0, 0, 9, 52, 53, 10, 14, 0, 0, 53, 58, 5, 19, 0,
		0, 54, 55, 10, 7, 0, 0, 55, 56, 5, 23, 0, 0, 56, 58, 3, 2, 1, 0, 57, 37,
		1, 0, 0, 0, 57, 40, 1, 0, 0, 0, 57, 43, 1, 0, 0, 0, 57, 46, 1, 0, 0, 0,
		57, 49, 1, 0, 0, 0, 57, 52, 1, 0, 0, 0, 57, 54, 1, 0, 0, 0, 58, 61, 1,
		0, 0, 0, 59, 57, 1, 0, 0, 0, 59, 60, 1, 0, 0, 0, 60, 1, 1, 0, 0, 0, 61,
		59, 1, 0, 0, 0, 62, 64, 7, 4, 0, 0, 63, 65, 5, 24, 0, 0, 64, 63, 1, 0,
		0, 0, 64, 65, 1, 0, 0, 0, 65, 66, 1, 0, 0, 0, 66, 68, 5, 10, 0, 0, 67,
		69, 3, 4, 2, 0, 68, 67, 1, 0, 0, 0, 68, 69, 1, 0, 0, 0, 69, 70, 1, 0, 0,
		0, 70, 73, 5, 11, 0, 0, 71, 73, 7, 4, 0, 0, 72, 62, 1, 0, 0, 0, 72, 71,
		1, 0, 0, 0, 73, 3, 1, 0, 0, 0, 74, 79, 3, 0, 0, 0, 75, 76, 5, 13, 0, 0,
		76, 78, 3, 0, 0, 0, 77, 75, 1, 0, 0, 0, 78, 81, 1, 0, 0, 0, 79, 77, 1,
		0, 0, 0, 79, 80, 1, 0, 0, 0, 80, 5, 1, 0, 0, 0, 81, 79, 1, 0, 0, 0, 82,
		84, 3, 8, 4, 0, 83, 82, 1, 0, 0, 0, 83, 84, 1, 0, 0, 0, 84, 85, 1, 0, 0,
		0, 85, 88, 3, 10, 5, 0, 86, 87, 5, 12, 0, 0, 87, 89, 3, 10, 5, 0, 88, 86,
		1, 0, 0, 0, 88, 89, 1, 0, 0, 0, 89, 7, 1, 0, 0, 0, 90, 91, 5, 27, 0, 0,
		91, 92, 5, 22, 0, 0, 92, 9, 1, 0, 0, 0, 93, 95, 5, 14, 0, 0, 94, 93, 1,
		0, 0, 0, 94, 95, 1, 0, 0, 0, 95, 96, 1, 0, 0, 0, 96, 98, 5, 25, 0, 0, 97,
		99, 5, 14, 0, 0, 98, 97, 1, 0, 0, 0, 98, 99, 1, 0, 0, 0, 99, 100, 1, 0,
		0, 0, 100, 101, 5, 24, 0, 0, 101, 11, 1, 0, 0, 0, 102, 107, 3, 18, 9, 0,
		103, 104, 5, 13, 0, 0, 104, 106, 3, 18, 9, 0, 105, 103, 1, 0, 0, 0, 106,
		109, 1, 0, 0, 0, 107, 105, 1, 0, 0, 0, 107, 108, 1, 0, 0, 0, 108, 110,
		1, 0, 0, 0, 109, 107, 1, 0, 0, 0, 110, 111, 5, 0, 0, 1, 111, 13, 1, 0,
		0, 0, 112, 117, 3, 20, 10, 0, 113, 114, 5, 13, 0, 0, 114, 116, 3, 20, 10,
		0, 115, 113, 1, 0, 0, 0, 116, 119, 1, 0, 0, 0, 117, 115, 1, 0, 0, 0, 117,
		118, 1, 0, 0, 0, 118, 120, 1, 0, 0, 0, 119, 117, 1, 0, 0, 0, 120, 121,
		5, 0, 0, 1, 121, 15, 1, 0, 0, 0, 122, 123, 5, 24, 0, 0, 123, 124, 5, 0,
		0, 1, 124, 17, 1, 0, 0, 0, 125, 128, 5, 24, 0, 0, 126, 127, 5, 17, 0, 0,
		127, 129, 5, 24, 0, 0, 128, 126, 1, 0, 0, 0, 128, 129, 1, 0, 0, 0, 129,
		19, 1, 0, 0, 0, 130, 133, 5, 25, 0, 0, 131, 132, 5, 17, 0, 0, 132, 134,
		5, 25, 0, 0, 133, 131, 1, 0, 0, 0, 133, 134, 1, 0, 0, 0, 134, 21, 1, 0,
		0, 0, 15, 35, 57, 59, 64, 68, 72, 79, 83, 88, 94, 98, 107, 117, 128, 133,
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
	TsvsheetParserRULE_rowSelector    = 6
	TsvsheetParserRULE_colSelector    = 7
	TsvsheetParserRULE_countValue     = 8
	TsvsheetParserRULE_rowSpan        = 9
	TsvsheetParserRULE_colSpan        = 10
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
	p.SetState(35)
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
			p.SetState(23)
			p.Match(TsvsheetParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(24)
			p.expression(0)
		}
		{
			p.SetState(25)
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
			p.SetState(27)

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
			p.SetState(28)
			p.expression(12)
		}

	case 3:
		localctx = NewCallExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(29)
			p.FunctionCall()
		}

	case 4:
		localctx = NewRefExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(30)
			p.Reference()
		}

	case 5:
		localctx = NewNumberExprContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(31)
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
			p.SetState(32)
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
			p.SetState(33)
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
			p.SetState(34)
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
	p.SetState(59)
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
			p.SetState(57)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext()) {
			case 1:
				localctx = NewPowExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(37)

				if !(p.Precpred(p.GetParserRuleContext(), 13)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 13)", ""))
					goto errorExit
				}
				{
					p.SetState(38)
					p.Match(TsvsheetParserCARET)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(39)
					p.expression(13)
				}

			case 2:
				localctx = NewMulExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(40)

				if !(p.Precpred(p.GetParserRuleContext(), 11)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 11)", ""))
					goto errorExit
				}
				{
					p.SetState(41)

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
					p.SetState(42)
					p.expression(12)
				}

			case 3:
				localctx = NewAddExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(43)

				if !(p.Precpred(p.GetParserRuleContext(), 10)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 10)", ""))
					goto errorExit
				}
				{
					p.SetState(44)

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
					p.SetState(45)
					p.expression(11)
				}

			case 4:
				localctx = NewConcatExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(46)

				if !(p.Precpred(p.GetParserRuleContext(), 9)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 9)", ""))
					goto errorExit
				}
				{
					p.SetState(47)
					p.Match(TsvsheetParserAMP)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(48)
					p.expression(10)
				}

			case 5:
				localctx = NewCompareExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(49)

				if !(p.Precpred(p.GetParserRuleContext(), 8)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 8)", ""))
					goto errorExit
				}
				{
					p.SetState(50)

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
					p.SetState(51)
					p.expression(9)
				}

			case 6:
				localctx = NewPercentExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(52)

				if !(p.Precpred(p.GetParserRuleContext(), 14)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 14)", ""))
					goto errorExit
				}
				{
					p.SetState(53)
					p.Match(TsvsheetParserPERCENT)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}

			case 7:
				localctx = NewPipeExprContext(p, NewExpressionContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, TsvsheetParserRULE_expression)
				p.SetState(54)

				if !(p.Precpred(p.GetParserRuleContext(), 7)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 7)", ""))
					goto errorExit
				}
				{
					p.SetState(55)
					p.Match(TsvsheetParserPIPE)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(56)
					p.FunctionCall()
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(61)
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

	p.SetState(72)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		p.EnterOuterAlt(localctx, 1)
		{
			p.SetState(62)
			_la = p.GetTokenStream().LA(1)

			if !(_la == TsvsheetParserCOL || _la == TsvsheetParserNAME) {
				p.GetErrorHandler().RecoverInline(p)
			} else {
				p.GetErrorHandler().ReportMatch(p)
				p.Consume()
			}
		}
		p.SetState(64)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if _la == TsvsheetParserNUMBER {
			{
				p.SetState(63)
				p.Match(TsvsheetParserNUMBER)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		}
		{
			p.SetState(66)
			p.Match(TsvsheetParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(68)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&251872704) != 0 {
			{
				p.SetState(67)
				p.ArgList()
			}

		}
		{
			p.SetState(70)
			p.Match(TsvsheetParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		p.EnterOuterAlt(localctx, 2)
		{
			p.SetState(71)
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
		p.SetState(74)
		p.expression(0)
	}
	p.SetState(79)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == TsvsheetParserCOMMA {
		{
			p.SetState(75)
			p.Match(TsvsheetParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(76)
			p.expression(0)
		}

		p.SetState(81)
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
	p.SetState(83)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserSTRING {
		{
			p.SetState(82)
			p.SheetQualifier()
		}

	}
	{
		p.SetState(85)
		p.CellRef()
	}
	p.SetState(88)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(86)
			p.Match(TsvsheetParserCOLON)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(87)
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
		p.SetState(90)
		p.Match(TsvsheetParserSTRING)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(91)
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
	p.SetState(94)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserDOLLAR {
		{
			p.SetState(93)
			p.Match(TsvsheetParserDOLLAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(96)
		p.Match(TsvsheetParserCOL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(98)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserDOLLAR {
		{
			p.SetState(97)
			p.Match(TsvsheetParserDOLLAR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(100)
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

// IRowSelectorContext is an interface to support dynamic dispatch.
type IRowSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllRowSpan() []IRowSpanContext
	RowSpan(i int) IRowSpanContext
	EOF() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsRowSelectorContext differentiates from other interfaces.
	IsRowSelectorContext()
}

type RowSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRowSelectorContext() *RowSelectorContext {
	var p = new(RowSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_rowSelector
	return p
}

func InitEmptyRowSelectorContext(p *RowSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_rowSelector
}

func (*RowSelectorContext) IsRowSelectorContext() {}

func NewRowSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RowSelectorContext {
	var p = new(RowSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_rowSelector

	return p
}

func (s *RowSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *RowSelectorContext) AllRowSpan() []IRowSpanContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IRowSpanContext); ok {
			len++
		}
	}

	tst := make([]IRowSpanContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IRowSpanContext); ok {
			tst[i] = t.(IRowSpanContext)
			i++
		}
	}

	return tst
}

func (s *RowSelectorContext) RowSpan(i int) IRowSpanContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IRowSpanContext); ok {
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

	return t.(IRowSpanContext)
}

func (s *RowSelectorContext) EOF() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserEOF, 0)
}

func (s *RowSelectorContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserCOMMA)
}

func (s *RowSelectorContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOMMA, i)
}

func (s *RowSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RowSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RowSelectorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterRowSelector(s)
	}
}

func (s *RowSelectorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitRowSelector(s)
	}
}

func (s *RowSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitRowSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) RowSelector() (localctx IRowSelectorContext) {
	localctx = NewRowSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 12, TsvsheetParserRULE_rowSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(102)
		p.RowSpan()
	}
	p.SetState(107)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == TsvsheetParserCOMMA {
		{
			p.SetState(103)
			p.Match(TsvsheetParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(104)
			p.RowSpan()
		}

		p.SetState(109)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(110)
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

// IColSelectorContext is an interface to support dynamic dispatch.
type IColSelectorContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllColSpan() []IColSpanContext
	ColSpan(i int) IColSpanContext
	EOF() antlr.TerminalNode
	AllCOMMA() []antlr.TerminalNode
	COMMA(i int) antlr.TerminalNode

	// IsColSelectorContext differentiates from other interfaces.
	IsColSelectorContext()
}

type ColSelectorContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColSelectorContext() *ColSelectorContext {
	var p = new(ColSelectorContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_colSelector
	return p
}

func InitEmptyColSelectorContext(p *ColSelectorContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_colSelector
}

func (*ColSelectorContext) IsColSelectorContext() {}

func NewColSelectorContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ColSelectorContext {
	var p = new(ColSelectorContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_colSelector

	return p
}

func (s *ColSelectorContext) GetParser() antlr.Parser { return s.parser }

func (s *ColSelectorContext) AllColSpan() []IColSpanContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IColSpanContext); ok {
			len++
		}
	}

	tst := make([]IColSpanContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IColSpanContext); ok {
			tst[i] = t.(IColSpanContext)
			i++
		}
	}

	return tst
}

func (s *ColSelectorContext) ColSpan(i int) IColSpanContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IColSpanContext); ok {
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

	return t.(IColSpanContext)
}

func (s *ColSelectorContext) EOF() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserEOF, 0)
}

func (s *ColSelectorContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserCOMMA)
}

func (s *ColSelectorContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOMMA, i)
}

func (s *ColSelectorContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColSelectorContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ColSelectorContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterColSelector(s)
	}
}

func (s *ColSelectorContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitColSelector(s)
	}
}

func (s *ColSelectorContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitColSelector(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) ColSelector() (localctx IColSelectorContext) {
	localctx = NewColSelectorContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 14, TsvsheetParserRULE_colSelector)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(112)
		p.ColSpan()
	}
	p.SetState(117)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for _la == TsvsheetParserCOMMA {
		{
			p.SetState(113)
			p.Match(TsvsheetParserCOMMA)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(114)
			p.ColSpan()
		}

		p.SetState(119)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(120)
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

// ICountValueContext is an interface to support dynamic dispatch.
type ICountValueContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	NUMBER() antlr.TerminalNode
	EOF() antlr.TerminalNode

	// IsCountValueContext differentiates from other interfaces.
	IsCountValueContext()
}

type CountValueContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyCountValueContext() *CountValueContext {
	var p = new(CountValueContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_countValue
	return p
}

func InitEmptyCountValueContext(p *CountValueContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_countValue
}

func (*CountValueContext) IsCountValueContext() {}

func NewCountValueContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *CountValueContext {
	var p = new(CountValueContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_countValue

	return p
}

func (s *CountValueContext) GetParser() antlr.Parser { return s.parser }

func (s *CountValueContext) NUMBER() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNUMBER, 0)
}

func (s *CountValueContext) EOF() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserEOF, 0)
}

func (s *CountValueContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *CountValueContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *CountValueContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterCountValue(s)
	}
}

func (s *CountValueContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitCountValue(s)
	}
}

func (s *CountValueContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitCountValue(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) CountValue() (localctx ICountValueContext) {
	localctx = NewCountValueContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 16, TsvsheetParserRULE_countValue)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(122)
		p.Match(TsvsheetParserNUMBER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(123)
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

// IRowSpanContext is an interface to support dynamic dispatch.
type IRowSpanContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllNUMBER() []antlr.TerminalNode
	NUMBER(i int) antlr.TerminalNode
	DASH() antlr.TerminalNode

	// IsRowSpanContext differentiates from other interfaces.
	IsRowSpanContext()
}

type RowSpanContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyRowSpanContext() *RowSpanContext {
	var p = new(RowSpanContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_rowSpan
	return p
}

func InitEmptyRowSpanContext(p *RowSpanContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_rowSpan
}

func (*RowSpanContext) IsRowSpanContext() {}

func NewRowSpanContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *RowSpanContext {
	var p = new(RowSpanContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_rowSpan

	return p
}

func (s *RowSpanContext) GetParser() antlr.Parser { return s.parser }

func (s *RowSpanContext) AllNUMBER() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserNUMBER)
}

func (s *RowSpanContext) NUMBER(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserNUMBER, i)
}

func (s *RowSpanContext) DASH() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserDASH, 0)
}

func (s *RowSpanContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *RowSpanContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *RowSpanContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterRowSpan(s)
	}
}

func (s *RowSpanContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitRowSpan(s)
	}
}

func (s *RowSpanContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitRowSpan(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) RowSpan() (localctx IRowSpanContext) {
	localctx = NewRowSpanContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 18, TsvsheetParserRULE_rowSpan)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(125)
		p.Match(TsvsheetParserNUMBER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(128)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserDASH {
		{
			p.SetState(126)
			p.Match(TsvsheetParserDASH)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(127)
			p.Match(TsvsheetParserNUMBER)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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

// IColSpanContext is an interface to support dynamic dispatch.
type IColSpanContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllCOL() []antlr.TerminalNode
	COL(i int) antlr.TerminalNode
	DASH() antlr.TerminalNode

	// IsColSpanContext differentiates from other interfaces.
	IsColSpanContext()
}

type ColSpanContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyColSpanContext() *ColSpanContext {
	var p = new(ColSpanContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_colSpan
	return p
}

func InitEmptyColSpanContext(p *ColSpanContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = TsvsheetParserRULE_colSpan
}

func (*ColSpanContext) IsColSpanContext() {}

func NewColSpanContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ColSpanContext {
	var p = new(ColSpanContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = TsvsheetParserRULE_colSpan

	return p
}

func (s *ColSpanContext) GetParser() antlr.Parser { return s.parser }

func (s *ColSpanContext) AllCOL() []antlr.TerminalNode {
	return s.GetTokens(TsvsheetParserCOL)
}

func (s *ColSpanContext) COL(i int) antlr.TerminalNode {
	return s.GetToken(TsvsheetParserCOL, i)
}

func (s *ColSpanContext) DASH() antlr.TerminalNode {
	return s.GetToken(TsvsheetParserDASH, 0)
}

func (s *ColSpanContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ColSpanContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ColSpanContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.EnterColSpan(s)
	}
}

func (s *ColSpanContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(TsvsheetParserListener); ok {
		listenerT.ExitColSpan(s)
	}
}

func (s *ColSpanContext) Accept(visitor antlr.ParseTreeVisitor) interface{} {
	switch t := visitor.(type) {
	case TsvsheetParserVisitor:
		return t.VisitColSpan(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *TsvsheetParser) ColSpan() (localctx IColSpanContext) {
	localctx = NewColSpanContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 20, TsvsheetParserRULE_colSpan)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(130)
		p.Match(TsvsheetParserCOL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(133)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == TsvsheetParserDASH {
		{
			p.SetState(131)
			p.Match(TsvsheetParserDASH)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(132)
			p.Match(TsvsheetParserCOL)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
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
