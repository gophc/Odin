package main

// This file contains a recursive descent parser for C.
//
// Most functions in this file are named after the symbols they are
// supposed to read from an input token list. For example, stmt() is
// responsible for reading a statement from a token list. The function
// then construct an AST node representing a statement.
//
// Each function conceptually returns two values, an AST node and
// remaining part of the input tokens.
//
// Input tokens are represented by a linked list. Unlike many recursive
// descent parsers, we don't have the notion of the "input token stream".
// Most parsing functions don't change the global state of the parser.
// So it is very easy to lookahead arbitrary number of tokens in this
// parser.

import (
	"fmt"
)

// local variable
type Obj struct {
	next   *Obj
	name   string // variable name
	ty     *Type  // Type
	offset int    // offset from rbp
}

// Find a local variable by name.
// walks the token list
func findVar(tok *Token) *Obj {
	for v := locals; v != nil; v = v.next {
		if v.name == string(tok.loc) {
			return v
		}
	}
	return nil
}

// function
type Function struct {
	body      *Node
	locals    *Obj
	stackSize int
}

func (f *Function) String() string {
	str := `{ "function body": [`
	for v := f.body; v != nil; v = v.next {
		str += v.String() + ","
	}
	return str + "]}"
}

// AST node

type NodeKind int

const (
	ND_ADD       NodeKind = iota // +
	ND_SUB                       // -
	ND_MUL                       // *
	ND_DIV                       // /
	ND_NEG                       // unary -
	ND_EQ                        // ==
	ND_NE                        // !=
	ND_LT                        // <
	ND_LE                        // <=
	ND_ASSIGN                    // =
	ND_ADDR                      // unary &
	ND_DEREF                     // unary *
	ND_RETURN                    // "return"
	ND_IF                        // "if"
	ND_FOR                       // "for" or "while"
	ND_BLOCK                     // {...}
	ND_FUNCALL                   // function call
	ND_EXPR_STMT                 // Expression statement
	ND_VAR                       // Variable
	ND_NUM                       // Integer
)

func (nd NodeKind) String() string {
	switch nd {
	case ND_ADD:
		return "+"
	case ND_SUB:
		return " - "
	case ND_MUL:
		return " * "
	case ND_DIV:
		return " / "
	case ND_NEG:
		return " unary - "
	case ND_EQ:
		return "  =="
	case ND_NE:
		return "  !="
	case ND_LT:
		return "  < "
	case ND_LE:
		return "  <= "
	case ND_ASSIGN:
		return " = "
	case ND_EXPR_STMT:
		return " Expression statement"
	case ND_VAR:
		return " Variable "
	case ND_NUM:
		return " Integer "
	case ND_RETURN:
		return "return"
	case ND_IF:
		return "if"
	case ND_FOR:
		return "for"
	case ND_BLOCK:
		return "Block"
	default:
		return "unknown node kind"
	}
}

// AST node type
type Node struct {
	kind NodeKind // node kind
	next *Node    // next node, (nodes are stored in a linked list)
	ty   *Type    // Type, e.g. int or pointer to int
	tok  *Token   // Representative token

	lhs *Node // left hand side
	rhs *Node // right hand side

	// Block, used if kind == ND_BLOCK
	body *Node

	// function call, if kind == ND_FUNCALL
	funcname string
	args     *Node

	// "if" or "for" statement, used if kind == ND_IF || ND_FOR
	cond *Node
	then *Node
	els  *Node
	// "for" statement only
	init *Node
	inc  *Node

	variable *Obj // used if kind == ND_VAR
	val      int  // used if kind == ND_NUM
}

func (n *Node) String() string {
	// print recursively
	if n == nil {
		return "\"nil\""
	}

	lhs := n.lhs.String()
	rhs := n.rhs.String()

	str := fmt.Sprintln(`{ "kind" : "`, n.kind,
		`", "var": "`, n.variable,
		`", "val": "`, n.val, "\",",
		`"body": `, n.body)
	return str + ",\"left\": " + lhs + ", \"right\": " + rhs + "}"
}

func NewNode(kind NodeKind, tok *Token) *Node {
	node := new(Node)
	node.kind = kind
	node.tok = tok

	return node
}

func NewBinary(kind NodeKind, lhs *Node, rhs *Node, tok *Token) *Node {
	node := NewNode(kind, tok)
	node.lhs = lhs
	node.rhs = rhs

	return node
}

func NewUnary(kind NodeKind, expr *Node, tok *Token) *Node {
	node := NewNode(kind, tok)
	node.lhs = expr
	return node
}

func NewNum(val int, tok *Token) *Node {
	node := NewNode(ND_NUM, tok)
	node.val = val

	return node
}

func NewVarNode(v *Obj, tok *Token) *Node {
	node := NewNode(ND_VAR, tok)
	node.variable = v
	return node
}

// adds to the start of the locals (a linked list)
// LVar = Local Variable
func newLVar(name string, ty *Type) *Obj {
	v := new(Obj)
	v.name = name
	v.ty = ty
	v.next = locals
	locals = v
	return v
}

func getIdent(tok *Token) string {
	if tok.Kind != IDENT {
		errorTok(tok, "expected an identifier")
	}
	return string(tok.loc[0:len(tok.loc)])
}

// declspec = "int"
func declspec(rest **Token, tok *Token) *Type {
	*rest = skip(tok, "int")
	return tyInt
}

// declarator = "*"* ident
// pointer of a type
func declarator(rest **Token, tok *Token, ty *Type) *Type {
	for tok.consume(&tok, "*") {
		ty = pointerTo(ty)
	}

	if tok.Kind != IDENT {
		errorTok(tok, "expected a variable name")
	}

	ty.name = tok
	*rest = tok.Next
	return ty
}

// declaration = declspec (declarator ("=" expr)? ("," declarator ("=" expr)?)*)? ";"
func declaration(rest **Token, tok *Token) *Node {
	baseTy := declspec(&tok, tok)

	head := Node{}
	cur := &head
	i := 0

	for !tok.equal(";") {
		if i > 0 {
			tok = skip(tok, ",")
		}
		i++

		ty := declarator(&tok, tok, baseTy)
		v := newLVar(getIdent(ty.name), ty) // var

		if !tok.equal("=") {
			continue
		}

		lhs := NewVarNode(v, ty.name)
		rhs := assign(&tok, tok.Next)
		node := NewBinary(ND_ASSIGN, lhs, rhs, tok)
		cur.next = NewUnary(ND_EXPR_STMT, node, tok)
		cur = cur.next
	}

	node := NewNode(ND_BLOCK, tok)
	node.body = head.next
	*rest = tok.Next
	return node
}

// stmt = "return" expr ";"
//      | "if" "(" expr ")" stmt ("else" stmt)?
//      | "for" "(" expr-stmt expr? ";" expr? ")" stmt
//      | "while" "(" expr ")" stmt
//      | "{" compound-stmt
//      | expr-stmt
func stmt(rest **Token, tok *Token) *Node {
	if tok.equal("return") {
		node := NewNode(ND_RETURN, tok)
		node.lhs = expr(&tok, tok.Next)
		*rest = skip(tok, ";")
		return node
	}

	if tok.equal("if") {
		node := NewNode(ND_IF, tok)
		tok = skip(tok.Next, "(")
		node.cond = expr(&tok, tok)
		tok = skip(tok, ")")
		node.then = stmt(&tok, tok)
		if tok.equal("else") {
			node.els = stmt(&tok, tok.Next)
		}
		*rest = tok
		return node
	}

	if tok.equal("for") {
		node := NewNode(ND_FOR, tok)
		tok = skip(tok.Next, "(")

		// expr-stmt, must finish with a comma
		node.init = exprStmt(&tok, tok)

		// expr?, look if comma exists, if so it means no expr present
		if !tok.equal(";") {
			node.cond = expr(&tok, tok)
		}
		tok = skip(tok, ";")

		if !tok.equal(")") {
			node.inc = expr(&tok, tok)
		}
		tok = skip(tok, ")")

		node.then = stmt(rest, tok)
		return node
	}

	if tok.equal("while") {
		node := NewNode(ND_FOR, tok)
		tok = skip(tok.Next, "(")
		node.cond = expr(&tok, tok)
		tok = skip(tok, ")")

		node.then = stmt(rest, tok)
		return node
	}

	if tok.equal("{") {
		return compoundStmt(rest, tok.Next)
	}

	return exprStmt(rest, tok)
}

// compound-stmt = (declaration | stmt)* "}"
func compoundStmt(rest **Token, tok *Token) *Node {
	head := new(Node)
	cur := head
	for !tok.equal("}") {
		if tok.equal("int") {
			cur.next = declaration(&tok, tok)
			cur = cur.next
		} else {
			cur.next = stmt(&tok, tok)
			cur = cur.next
		}
		addType(cur)
	}

	node := NewNode(ND_BLOCK, tok)
	node.body = head.next
	*rest = tok.Next

	return node
}

// expr-stmt = expr? ";"
func exprStmt(rest **Token, tok *Token) *Node {
	if tok.equal(";") {
		*rest = tok.Next
		return NewNode(ND_BLOCK, tok)
	}

	node := NewNode(ND_EXPR_STMT, tok)
	node.lhs = expr(&tok, tok)
	*rest = skip(tok, ";")
	return node
}

// expr = assign
func expr(rest **Token, tok *Token) *Node {
	return assign(rest, tok)
}

// assign = equality ("=" assign)?
func assign(rest **Token, tok *Token) *Node {
	node := equality(&tok, tok)
	if tok.equal("=") {
		node = NewBinary(ND_ASSIGN, node, assign(&tok, tok.Next), tok)
	}
	*rest = tok
	return node
}

// equality = relational ("==" relational | "!=" relational)*
func equality(rest **Token, tok *Token) *Node {
	node := relational(&tok, tok)

	for {

		if tok.equal("==") {
			node = NewBinary(ND_EQ, node, relational(&tok, tok.Next), tok)
			continue
		}

		if tok.equal("!=") {
			node = NewBinary(ND_NE, node, relational(&tok, tok.Next), tok)
			continue
		}

		*rest = tok
		return node
	}
}

// relational = add ("<" add | "<=" add | ">" add | ">=" add)*
func relational(rest **Token, tok *Token) *Node {
	node := add(&tok, tok)

	for {
		start := tok

		if tok.equal("<") {
			node = NewBinary(ND_LT, node, add(&tok, tok.Next), start)
			continue
		}

		if tok.equal("<=") {
			node = NewBinary(ND_LE, node, add(&tok, tok.Next), start)
			continue
		}

		if tok.equal(">") {
			node = NewBinary(ND_LT, add(&tok, tok.Next), node, start)
			continue
		}

		if tok.equal(">=") {
			node = NewBinary(ND_LE, add(&tok, tok.Next), node, start)
			continue
		}

		*rest = tok
		return node
	}

}

// In C, `+` operator is overloaded to perform the pointer arithmetic.
// If p is a pointer, p+n adds not n but sizeof(*p)*n to the value of p,
// so that p+n points to the location n elements (not bytes) ahead of p.
// In other words, we need to scale an integer value before adding to a
// pointer value. This function takes care of the scaling.
func NewAdd(lhs, rhs *Node, tok *Token) *Node {
	addType(lhs)
	addType(rhs)

	// num + num
	if lhs.ty.isInt() && rhs.ty.isInt() {
		return NewBinary(ND_ADD, lhs, rhs, tok)
	}

	// if both sides are pointer, invalid
	if lhs.ty.isPtr() && rhs.ty.isPtr() {
		errorTok(tok, "invalid operands")
	}

	// cannonicalize `num+ptr` to `ptr + num`
	if lhs.ty.isInt() && rhs.ty.isPtr() {
		lhs, rhs = rhs, lhs
	}

	// ptr+num
	rhs = NewBinary(ND_MUL, rhs, NewNum(8, tok), tok)
	return NewBinary(ND_ADD, lhs, rhs, tok)
}

// Like `+`, `-` is overloaded for the pointer type.
func NewSub(lhs, rhs *Node, tok *Token) *Node {
	addType(lhs)
	addType(rhs)

	// num - num
	if lhs.ty.isInt() && rhs.ty.isInt() {
		return NewBinary(ND_SUB, lhs, rhs, tok)
	}

	// ptr - num
	if lhs.ty.isPtr() && rhs.ty.isInt() {
		// scale
		rhs = NewBinary(ND_MUL, rhs, NewNum(8, tok), tok)
		addType(rhs)
		node := NewBinary(ND_SUB, lhs, rhs, tok)
		node.ty = lhs.ty
		return node
	}

	// ptr - ptr, which returns how many elements are between the two
	if lhs.ty.isPtr() && rhs.ty.isPtr() {
		node := NewBinary(ND_SUB, lhs, rhs, tok)
		node.ty = tyInt
		return NewBinary(ND_DIV, node, NewNum(8, tok), tok)
	}

	errorTok(tok, "invalid operands")
	return nil
}

// add = mul ("+" mul | "-" mul)*
func add(rest **Token, tok *Token) *Node {
	node := mul(&tok, tok)

	for {
		start := tok

		if tok.equal("+") {
			node = NewAdd(node, mul(&tok, tok.Next), start)
			continue
		}

		if tok.equal("-") {
			node = NewSub(node, mul(&tok, tok.Next), start)
			continue
		}

		*rest = tok
		return node
	}
}

// mul = unary ("*" unary | "/" unary)*
func mul(rest **Token, tok *Token) *Node {
	node := unary(&tok, tok) // left node for the new binary node

	for {
		start := tok

		if tok.equal("*") {
			// rhs is primary(&tok,.)
			node = NewBinary(ND_MUL, node, unary(&tok, tok.Next), start)
			continue
		}

		if tok.equal("/") {
			node = NewBinary(ND_DIV, node, unary(&tok, tok.Next), start)
			continue
		}

		*rest = tok
		return node
	}
}

// unary = ("+" | "-" | "*" | "&") unary
//       | primary
func unary(rest **Token, tok *Token) *Node {

	// doesn't affect the sign
	if tok.equal("+") {
		return unary(rest, tok.Next)
	}

	if tok.equal("-") {
		return NewUnary(ND_NEG, unary(rest, tok.Next), tok)
	}

	if tok.equal("&") {
		return NewUnary(ND_ADDR, unary(rest, tok.Next), tok)
	}

	if tok.equal("*") {
		return NewUnary(ND_DEREF, unary(rest, tok.Next), tok)
	}

	return primary(rest, tok)
}

// funcall = ident "(" (assign ("," assign)*)? ")"
func funcall(rest **Token, tok *Token) *Node {
	start := tok
	tok = tok.Next.Next

	head := Node{}
	cur := &head

	for !tok.equal(")") {
		if cur != &head {
			tok = skip(tok, ",")
		}
		cur.next = assign(&tok, tok)
		cur = cur.next
	}

	*rest = skip(tok, ")")

	node := NewNode(ND_FUNCALL, start)
	node.funcname = string(start.loc)
	node.args = head.next
	return node
}

// primary = "(" expr ")" | ident func-args? | num
func primary(rest **Token, tok *Token) *Node {
	if tok.equal("(") {
		node := expr(&tok, tok.Next)
		*rest = skip(tok, ")")
		return node
	}

	if tok.Kind == IDENT {
		// Function call
		if tok.Next.equal("(") {
			return funcall(rest, tok)
		}

		// Variable
		v := findVar(tok)
		if v == nil {
			errorTok(tok, "undefined variable")
		}
		*rest = tok.Next
		return NewVarNode(v, tok)
	}

	if tok.Kind == NUM {
		node := NewNum(tok.val, tok)
		*rest = tok.Next
		return node
	}

	errorTok(tok, "expected an expression")
	return nil
}

// program = stmt*
// the returned Node is also a linked list of Nodes
func parse(tok *Token) *Function {
	// this commit we expect program to start with '{'
	tok = skip(tok, "{")

	prog := new(Function)
	prog.body = compoundStmt(&tok, tok)
	prog.locals = locals
	return prog
}
