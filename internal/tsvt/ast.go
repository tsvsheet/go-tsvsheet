// Package tsvt is the covered seam over the ANTLR-generated formula parser: it
// turns a cell's formula source (the text after its leading `=`) into an
// immutable typed AST — an Expr over A1 references and literals — or a sentinel
// syntax error, and hides every ANTLR type from the rest of the program.
package tsvt

// The two AST interfaces below are sealed: each has an unexported marker method,
// carried by a zero-size embedded struct, so only the node types in this package
// can satisfy it. Consumers walk the AST by type switch; the markers bound each
// switch's variant set at compile time.
type (
	exprMarker      struct{}
	referenceMarker struct{}
)

func (exprMarker) isExpr()           {}
func (referenceMarker) isReference() {}

// Expr is a formula expression (SPECIFICATION §5). The set is sealed.
type Expr interface{ isExpr() }

// BinaryOp is a binary operator.
type BinaryOp string

// The binary operators.
const (
	OpMul BinaryOp = "*"
	OpDiv BinaryOp = "/"
	OpAdd BinaryOp = "+"
	OpSub BinaryOp = "-"
	OpPow BinaryOp = "^"
	OpCat BinaryOp = "&"
	OpEq  BinaryOp = "="
	OpNe  BinaryOp = "<>"
	OpLt  BinaryOp = "<"
	OpLe  BinaryOp = "<="
	OpGt  BinaryOp = ">"
	OpGe  BinaryOp = ">="
)

// UnaryOp is a unary sign operator.
type UnaryOp string

// The unary operators.
const (
	OpNeg UnaryOp = "-"
	OpPos UnaryOp = "+"
)

// Binary is a binary operation. IsGrouped records that the author wrapped this
// operation in parentheses — redundant ones included. Grouping is authored
// information, not noise: it is how a reader was told which reading was meant,
// so it survives parsing and rewriting, and it tells the Excel-divergence
// checker that this precedence was chosen rather than tripped over.
//
// A renderer re-emits the parentheses wherever the expression appears as an
// operand. Parentheses around a formula's whole expression are not re-emitted:
// with no surrounding operator they cannot change a reading, and there is no
// operand context in which to place them.
type Binary struct {
	exprMarker
	Left      Expr
	Right     Expr
	Op        BinaryOp
	IsGrouped bool
}

// Unary is a unary sign operation. IsGrouped carries the same authored
// parenthesization Binary does.
type Unary struct {
	exprMarker
	X         Expr
	Op        UnaryOp
	IsGrouped bool
}

// Percent is a postfix-percent operation: `50%` is `Percent{50}` = 0.5.
type Percent struct {
	exprMarker
	X Expr
}

// Call is a function call; Name is case-preserved (identity resolves
// case-insensitively in the evaluator). IsPiped records that the source
// spelled this call with the pipe operator (SPECIFICATION §5.4): `x | f(a)`
// desugars at build to `f(x, a)` with IsPiped set, so evaluation only ever
// sees function application and the flag exists solely so rendering can
// preserve the author's spelling. The desugar always prepends the piped value,
// so a piped call has at least one argument.
type Call struct {
	exprMarker
	Name    string
	Args    []Expr
	IsPiped bool
}

// RefOperand is an A1 reference used as an expression operand.
type RefOperand struct {
	exprMarker
	Ref Reference
}

// Number is a numeric literal; Text preserves the source form.
type Number struct {
	exprMarker
	Text string
}

// StringLit is a double-quoted string literal.
type StringLit struct {
	exprMarker
	Value string
}

// BoolLit is a TRUE/FALSE literal.
type BoolLit struct {
	exprMarker
	IsTrue bool
}

// ErrorLit is an error-value literal (`#N/A`, `#REF!`, …); Code is its text.
type ErrorLit struct {
	exprMarker
	Code string
}

// Reference is an A1 reference (SPECIFICATION §4). The set is sealed.
type Reference interface{ isReference() }

// RangeRef is a single A1 cell (To nil) or a rectangular range of two cells.
// File is a `"path"!` sheet qualifier — empty for the current sheet, otherwise
// the path (relative, bare, or absolute) of the sheet the cells are read from.
type RangeRef struct {
	referenceMarker
	To   *CellRef
	File string
	From CellRef
}

// CellRef is an A1 cell: a column label and a 1-based row. The Is*Absolute
// flags record the `$` markers (`$B$2`, `$B2`, `B$2`), which SPECIFICATION §4
// retains for familiarity and Excel round-tripping. They carry no positional
// difference — evaluation and structural edits treat a pinned and an unpinned
// coordinate identically (Excel-faithful: `$` pins only under copy/fill) — so
// their sole effect is surviving to the re-rendered formula.
type CellRef struct {
	Col           string
	Row           int
	IsColAbsolute bool
	IsRowAbsolute bool
}
