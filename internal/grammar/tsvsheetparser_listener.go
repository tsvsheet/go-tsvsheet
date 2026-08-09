// Code generated from TsvsheetParser.g4 by ANTLR 4.13.2. DO NOT EDIT.

package tsvsheetgrammar // TsvsheetParser
import "github.com/antlr4-go/antlr/v4"

// TsvsheetParserListener is a complete listener for a parse tree produced by TsvsheetParser.
type TsvsheetParserListener interface {
	antlr.ParseTreeListener

	// EnterFormula is called when entering the formula production.
	EnterFormula(c *FormulaContext)

	// EnterMetaClause is called when entering the metaClause production.
	EnterMetaClause(c *MetaClauseContext)

	// EnterMetaArgs is called when entering the metaArgs production.
	EnterMetaArgs(c *MetaArgsContext)

	// EnterErrorExpr is called when entering the errorExpr production.
	EnterErrorExpr(c *ErrorExprContext)

	// EnterPipeExpr is called when entering the pipeExpr production.
	EnterPipeExpr(c *PipeExprContext)

	// EnterNumberExpr is called when entering the numberExpr production.
	EnterNumberExpr(c *NumberExprContext)

	// EnterParenExpr is called when entering the parenExpr production.
	EnterParenExpr(c *ParenExprContext)

	// EnterConcatExpr is called when entering the concatExpr production.
	EnterConcatExpr(c *ConcatExprContext)

	// EnterStringExpr is called when entering the stringExpr production.
	EnterStringExpr(c *StringExprContext)

	// EnterUnaryExpr is called when entering the unaryExpr production.
	EnterUnaryExpr(c *UnaryExprContext)

	// EnterNameRefExpr is called when entering the nameRefExpr production.
	EnterNameRefExpr(c *NameRefExprContext)

	// EnterAddExpr is called when entering the addExpr production.
	EnterAddExpr(c *AddExprContext)

	// EnterRefExpr is called when entering the refExpr production.
	EnterRefExpr(c *RefExprContext)

	// EnterMulExpr is called when entering the mulExpr production.
	EnterMulExpr(c *MulExprContext)

	// EnterPercentExpr is called when entering the percentExpr production.
	EnterPercentExpr(c *PercentExprContext)

	// EnterCallExpr is called when entering the callExpr production.
	EnterCallExpr(c *CallExprContext)

	// EnterBoolExpr is called when entering the boolExpr production.
	EnterBoolExpr(c *BoolExprContext)

	// EnterPowExpr is called when entering the powExpr production.
	EnterPowExpr(c *PowExprContext)

	// EnterCompareExpr is called when entering the compareExpr production.
	EnterCompareExpr(c *CompareExprContext)

	// EnterFunctionCall is called when entering the functionCall production.
	EnterFunctionCall(c *FunctionCallContext)

	// EnterArgList is called when entering the argList production.
	EnterArgList(c *ArgListContext)

	// EnterReference is called when entering the reference production.
	EnterReference(c *ReferenceContext)

	// EnterSheetQualifier is called when entering the sheetQualifier production.
	EnterSheetQualifier(c *SheetQualifierContext)

	// EnterCellRef is called when entering the cellRef production.
	EnterCellRef(c *CellRefContext)

	// EnterDirectiveValue is called when entering the directiveValue production.
	EnterDirectiveValue(c *DirectiveValueContext)

	// EnterAxisCall is called when entering the axisCall production.
	EnterAxisCall(c *AxisCallContext)

	// EnterItemList is called when entering the itemList production.
	EnterItemList(c *ItemListContext)

	// EnterItem is called when entering the item production.
	EnterItem(c *ItemContext)

	// EnterRangeCall is called when entering the rangeCall production.
	EnterRangeCall(c *RangeCallContext)

	// EnterCountCall is called when entering the countCall production.
	EnterCountCall(c *CountCallContext)

	// EnterSpan is called when entering the span production.
	EnterSpan(c *SpanContext)

	// EnterEndpoint is called when entering the endpoint production.
	EnterEndpoint(c *EndpointContext)

	// EnterOffset is called when entering the offset production.
	EnterOffset(c *OffsetContext)

	// ExitFormula is called when exiting the formula production.
	ExitFormula(c *FormulaContext)

	// ExitMetaClause is called when exiting the metaClause production.
	ExitMetaClause(c *MetaClauseContext)

	// ExitMetaArgs is called when exiting the metaArgs production.
	ExitMetaArgs(c *MetaArgsContext)

	// ExitErrorExpr is called when exiting the errorExpr production.
	ExitErrorExpr(c *ErrorExprContext)

	// ExitPipeExpr is called when exiting the pipeExpr production.
	ExitPipeExpr(c *PipeExprContext)

	// ExitNumberExpr is called when exiting the numberExpr production.
	ExitNumberExpr(c *NumberExprContext)

	// ExitParenExpr is called when exiting the parenExpr production.
	ExitParenExpr(c *ParenExprContext)

	// ExitConcatExpr is called when exiting the concatExpr production.
	ExitConcatExpr(c *ConcatExprContext)

	// ExitStringExpr is called when exiting the stringExpr production.
	ExitStringExpr(c *StringExprContext)

	// ExitUnaryExpr is called when exiting the unaryExpr production.
	ExitUnaryExpr(c *UnaryExprContext)

	// ExitNameRefExpr is called when exiting the nameRefExpr production.
	ExitNameRefExpr(c *NameRefExprContext)

	// ExitAddExpr is called when exiting the addExpr production.
	ExitAddExpr(c *AddExprContext)

	// ExitRefExpr is called when exiting the refExpr production.
	ExitRefExpr(c *RefExprContext)

	// ExitMulExpr is called when exiting the mulExpr production.
	ExitMulExpr(c *MulExprContext)

	// ExitPercentExpr is called when exiting the percentExpr production.
	ExitPercentExpr(c *PercentExprContext)

	// ExitCallExpr is called when exiting the callExpr production.
	ExitCallExpr(c *CallExprContext)

	// ExitBoolExpr is called when exiting the boolExpr production.
	ExitBoolExpr(c *BoolExprContext)

	// ExitPowExpr is called when exiting the powExpr production.
	ExitPowExpr(c *PowExprContext)

	// ExitCompareExpr is called when exiting the compareExpr production.
	ExitCompareExpr(c *CompareExprContext)

	// ExitFunctionCall is called when exiting the functionCall production.
	ExitFunctionCall(c *FunctionCallContext)

	// ExitArgList is called when exiting the argList production.
	ExitArgList(c *ArgListContext)

	// ExitReference is called when exiting the reference production.
	ExitReference(c *ReferenceContext)

	// ExitSheetQualifier is called when exiting the sheetQualifier production.
	ExitSheetQualifier(c *SheetQualifierContext)

	// ExitCellRef is called when exiting the cellRef production.
	ExitCellRef(c *CellRefContext)

	// ExitDirectiveValue is called when exiting the directiveValue production.
	ExitDirectiveValue(c *DirectiveValueContext)

	// ExitAxisCall is called when exiting the axisCall production.
	ExitAxisCall(c *AxisCallContext)

	// ExitItemList is called when exiting the itemList production.
	ExitItemList(c *ItemListContext)

	// ExitItem is called when exiting the item production.
	ExitItem(c *ItemContext)

	// ExitRangeCall is called when exiting the rangeCall production.
	ExitRangeCall(c *RangeCallContext)

	// ExitCountCall is called when exiting the countCall production.
	ExitCountCall(c *CountCallContext)

	// ExitSpan is called when exiting the span production.
	ExitSpan(c *SpanContext)

	// ExitEndpoint is called when exiting the endpoint production.
	ExitEndpoint(c *EndpointContext)

	// ExitOffset is called when exiting the offset production.
	ExitOffset(c *OffsetContext)
}
