package quanticscript

// Node is the base interface for all AST nodes
type Node interface {
	Loc() SourceLocation
	Accept(visitor Visitor) interface{}
}

// Visitor pattern for AST traversal
type Visitor interface {
	VisitProgram(*Program) interface{}
	VisitConstDecl(*ConstDecl) interface{}
	VisitFunctionDecl(*FunctionDecl) interface{}
	VisitParameter(*Parameter) interface{}
	VisitBlockStmt(*BlockStmt) interface{}
	VisitReturnStmt(*ReturnStmt) interface{}
	VisitLetStmt(*LetStmt) interface{}
	VisitExprStmt(*ExprStmt) interface{}
	VisitIfStmt(*IfStmt) interface{}
	VisitWhileStmt(*WhileStmt) interface{}
	VisitForStmt(*ForStmt) interface{}
	VisitAssignmentStmt(*AssignmentStmt) interface{}
	VisitAsmBlock(*AsmBlock) interface{}
	VisitBinaryExpr(*BinaryExpr) interface{}
	VisitUnaryExpr(*UnaryExpr) interface{}
	VisitCallExpr(*CallExpr) interface{}
	VisitIdentExpr(*IdentExpr) interface{}
	VisitIntLiteral(*IntLiteral) interface{}
	VisitStringLiteral(*StringLiteral) interface{}
	VisitBoolLiteral(*BoolLiteral) interface{}
	VisitNullLiteral(*NullLiteral) interface{}
	VisitArrayExpr(*ArrayExpr) interface{}
	VisitIndexExpr(*IndexExpr) interface{}
	VisitMemberExpr(*MemberExpr) interface{}
	VisitTypeAnnotation(*TypeAnnotation) interface{}
}

// Program represents the entire source file
type Program struct {
	Location     SourceLocation
	Imports      []*ImportDecl
	Declarations []Node
}

func (p *Program) Loc() SourceLocation          { return p.Location }
func (p *Program) Accept(v Visitor) interface{} { return v.VisitProgram(p) }

// ImportDecl represents an import statement
type ImportDecl struct {
	Location SourceLocation
	Names    []string
	Path     string
}

func (i *ImportDecl) Loc() SourceLocation { return i.Location }

// ConstDecl represents a constant declaration
type ConstDecl struct {
	Location SourceLocation
	Name     string
	Type     *TypeAnnotation
	Value    Expression
}

func (c *ConstDecl) Loc() SourceLocation          { return c.Location }
func (c *ConstDecl) Accept(v Visitor) interface{} { return v.VisitConstDecl(c) }

// FunctionDecl represents a function declaration
type FunctionDecl struct {
	Location   SourceLocation
	Name       string
	Parameters []*Parameter
	ReturnType *TypeAnnotation
	Body       *BlockStmt
	IsExport   bool
}

func (f *FunctionDecl) Loc() SourceLocation          { return f.Location }
func (f *FunctionDecl) Accept(v Visitor) interface{} { return v.VisitFunctionDecl(f) }

// Parameter represents a function parameter
type Parameter struct {
	Location SourceLocation
	Name     string
	Type     *TypeAnnotation
}

func (p *Parameter) Loc() SourceLocation          { return p.Location }
func (p *Parameter) Accept(v Visitor) interface{} { return v.VisitParameter(p) }

// TypeAnnotation represents a type annotation
type TypeAnnotation struct {
	Location SourceLocation
	Name     string
	IsArray  bool
	Nullable bool
}

func (t *TypeAnnotation) Loc() SourceLocation          { return t.Location }
func (t *TypeAnnotation) Accept(v Visitor) interface{} { return v.VisitTypeAnnotation(t) }

// Statement interface
type Statement interface {
	Node
	statementNode()
}

// Expression interface
type Expression interface {
	Node
	expressionNode()
}

// BlockStmt represents a block of statements
type BlockStmt struct {
	Location   SourceLocation
	Statements []Statement
}

func (b *BlockStmt) Loc() SourceLocation          { return b.Location }
func (b *BlockStmt) Accept(v Visitor) interface{} { return v.VisitBlockStmt(b) }
func (b *BlockStmt) statementNode()               {}

// ReturnStmt represents a return statement
type ReturnStmt struct {
	Location SourceLocation
	Value    Expression
}

func (r *ReturnStmt) Loc() SourceLocation          { return r.Location }
func (r *ReturnStmt) Accept(v Visitor) interface{} { return v.VisitReturnStmt(r) }
func (r *ReturnStmt) statementNode()               {}

// LetStmt represents a variable declaration
type LetStmt struct {
	Location SourceLocation
	Name     string
	Type     *TypeAnnotation
	Value    Expression
	IsConst  bool
}

func (l *LetStmt) Loc() SourceLocation          { return l.Location }
func (l *LetStmt) Accept(v Visitor) interface{} { return v.VisitLetStmt(l) }
func (l *LetStmt) statementNode()               {}

// ExprStmt represents an expression statement
type ExprStmt struct {
	Location   SourceLocation
	Expression Expression
}

func (e *ExprStmt) Loc() SourceLocation          { return e.Location }
func (e *ExprStmt) Accept(v Visitor) interface{} { return v.VisitExprStmt(e) }
func (e *ExprStmt) statementNode()               {}

// IfStmt represents an if statement
type IfStmt struct {
	Location  SourceLocation
	Condition Expression
	ThenBlock *BlockStmt
	ElseBlock *BlockStmt
}

func (i *IfStmt) Loc() SourceLocation          { return i.Location }
func (i *IfStmt) Accept(v Visitor) interface{} { return v.VisitIfStmt(i) }
func (i *IfStmt) statementNode()               {}

// WhileStmt represents a while loop
type WhileStmt struct {
	Location  SourceLocation
	Condition Expression
	Body      *BlockStmt
}

func (w *WhileStmt) Loc() SourceLocation          { return w.Location }
func (w *WhileStmt) Accept(v Visitor) interface{} { return v.VisitWhileStmt(w) }
func (w *WhileStmt) statementNode()               {}

// ForStmt represents a for loop
type ForStmt struct {
	Location  SourceLocation
	Init      Statement
	Condition Expression
	Update    Statement
	Body      *BlockStmt
}

func (f *ForStmt) Loc() SourceLocation          { return f.Location }
func (f *ForStmt) Accept(v Visitor) interface{} { return v.VisitForStmt(f) }
func (f *ForStmt) statementNode()               {}

// AssignmentStmt represents an assignment statement
type AssignmentStmt struct {
	Location SourceLocation
	Target   Expression
	Operator TokenType // TOKEN_ASSIGN, TOKEN_PLUS_ASSIGN, etc.
	Value    Expression
}

func (a *AssignmentStmt) Loc() SourceLocation          { return a.Location }
func (a *AssignmentStmt) Accept(v Visitor) interface{} { return v.VisitAssignmentStmt(a) }
func (a *AssignmentStmt) statementNode()               {}

// AsmBlock represents an inline assembly block
type AsmBlock struct {
	Location     SourceLocation
	Instructions []AsmInstruction
}

func (a *AsmBlock) Loc() SourceLocation          { return a.Location }
func (a *AsmBlock) Accept(v Visitor) interface{} { return v.VisitAsmBlock(a) }
func (a *AsmBlock) statementNode()               {}

// AsmInstruction represents a single assembly instruction
type AsmInstruction struct {
	Location SourceLocation
	Mnemonic string
	Operands []string
}

// BinaryExpr represents a binary expression
type BinaryExpr struct {
	Location SourceLocation
	Left     Expression
	Operator TokenType
	Right    Expression
}

func (b *BinaryExpr) Loc() SourceLocation          { return b.Location }
func (b *BinaryExpr) Accept(v Visitor) interface{} { return v.VisitBinaryExpr(b) }
func (b *BinaryExpr) expressionNode()              {}

// UnaryExpr represents a unary expression
type UnaryExpr struct {
	Location SourceLocation
	Operator TokenType
	Operand  Expression
}

func (u *UnaryExpr) Loc() SourceLocation          { return u.Location }
func (u *UnaryExpr) Accept(v Visitor) interface{} { return v.VisitUnaryExpr(u) }
func (u *UnaryExpr) expressionNode()              {}

// CallExpr represents a function call
type CallExpr struct {
	Location  SourceLocation
	Function  Expression
	Arguments []Expression
}

func (c *CallExpr) Loc() SourceLocation          { return c.Location }
func (c *CallExpr) Accept(v Visitor) interface{} { return v.VisitCallExpr(c) }
func (c *CallExpr) expressionNode()              {}

// IdentExpr represents an identifier
type IdentExpr struct {
	Location SourceLocation
	Name     string
}

func (i *IdentExpr) Loc() SourceLocation          { return i.Location }
func (i *IdentExpr) Accept(v Visitor) interface{} { return v.VisitIdentExpr(i) }
func (i *IdentExpr) expressionNode()              {}

// IntLiteral represents an integer literal
type IntLiteral struct {
	Location SourceLocation
	Value    string
}

func (i *IntLiteral) Loc() SourceLocation          { return i.Location }
func (i *IntLiteral) Accept(v Visitor) interface{} { return v.VisitIntLiteral(i) }
func (i *IntLiteral) expressionNode()              {}

// StringLiteral represents a string literal
type StringLiteral struct {
	Location SourceLocation
	Value    string
}

func (s *StringLiteral) Loc() SourceLocation          { return s.Location }
func (s *StringLiteral) Accept(v Visitor) interface{} { return v.VisitStringLiteral(s) }
func (s *StringLiteral) expressionNode()              {}

// BoolLiteral represents a boolean literal
type BoolLiteral struct {
	Location SourceLocation
	Value    bool
}

func (b *BoolLiteral) Loc() SourceLocation          { return b.Location }
func (b *BoolLiteral) Accept(v Visitor) interface{} { return v.VisitBoolLiteral(b) }
func (b *BoolLiteral) expressionNode()              {}

// NullLiteral represents a null literal
type NullLiteral struct {
	Location SourceLocation
}

func (n *NullLiteral) Loc() SourceLocation          { return n.Location }
func (n *NullLiteral) Accept(v Visitor) interface{} { return v.VisitNullLiteral(n) }
func (n *NullLiteral) expressionNode()              {}

// ArrayExpr represents an array literal
type ArrayExpr struct {
	Location SourceLocation
	Elements []Expression
}

func (a *ArrayExpr) Loc() SourceLocation          { return a.Location }
func (a *ArrayExpr) Accept(v Visitor) interface{} { return v.VisitArrayExpr(a) }
func (a *ArrayExpr) expressionNode()              {}

// IndexExpr represents array indexing
type IndexExpr struct {
	Location SourceLocation
	Array    Expression
	Index    Expression
}

func (i *IndexExpr) Loc() SourceLocation          { return i.Location }
func (i *IndexExpr) Accept(v Visitor) interface{} { return v.VisitIndexExpr(i) }
func (i *IndexExpr) expressionNode()              {}

// MemberExpr represents member access (dot notation)
type MemberExpr struct {
	Location SourceLocation
	Object   Expression
	Member   string
}

func (m *MemberExpr) Loc() SourceLocation          { return m.Location }
func (m *MemberExpr) Accept(v Visitor) interface{} { return v.VisitMemberExpr(m) }
func (m *MemberExpr) expressionNode()              {}
