package main

type Type struct {
	kind TypeKind

	// Pointer
	base *Type

	// Declaration
	name *Token
}

type TypeKind int

const (
	TY_INT TypeKind = iota
	TY_PTR
)

var (
	tyInt *Type = &Type{kind: TY_INT}
)

func (t *Type) isInt() bool {
	return t.kind == TY_INT
}

func (t *Type) isPtr() bool {
	return t.kind == TY_PTR
}

func pointerTo(base *Type) *Type {
	ty := new(Type)
	ty.kind = TY_PTR
	ty.base = base
	return ty
}

// recursively adds type information to node and children
func addType(node *Node) {
	if node == nil || node.ty != nil {
		return
	}

	addType(node.lhs)
	addType(node.rhs)
	addType(node.cond)
	addType(node.then)
	addType(node.els)
	addType(node.init)
	addType(node.inc)

	// traverse linked list
	for n := node.body; n != nil; n = n.next {
		addType(n)
	}

	switch node.kind {
	case ND_ADD,
		ND_SUB,
		ND_MUL,
		ND_DIV,
		ND_NEG,
		ND_ASSIGN:
		node.ty = node.lhs.ty
		return
	case ND_EQ,
		ND_NE,
		ND_LT,
		ND_LE,
		ND_NUM,
		ND_FUNCALL:
		node.ty = tyInt
		return
	case ND_VAR:
		node.ty = node.variable.ty
		return
	case ND_ADDR:
		node.ty = pointerTo(node.lhs.ty)
		return
	case ND_DEREF:
		if node.lhs.ty.kind != TY_PTR {
			errorTok(node.tok, "invalid pointer dereference")
		}
		node.ty = node.lhs.ty.base
		return
	}
}
