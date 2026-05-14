// Depends on: common.odin (CheckerContext, Ast, AstKind enum, Slice, String, Operand, Entity, Type, etc.)
package cmd

type TypeSwitchKind int

const (
	TypeSwitchInvalid TypeSwitchKind = iota
	TypeSwitchUnion
	TypeSwitchAny
)

// Forward declaration - defined in other file
// func check_stmt(ctx *CheckerContext, node *Ast, flags u32)
