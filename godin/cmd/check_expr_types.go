// Depends on: common.odin (Checker, Operand, Ast, Type, Entity, Scope, etc.)
package cmd

type CallArgumentError int

const (
	CallArgumentErrorNone                     CallArgumentError = iota
	CallArgumentErrorNoneProcedureType        CallArgumentError = iota
	CallArgumentErrorWrongTypes               CallArgumentError = iota
	CallArgumentErrorNonVariadicExpand        CallArgumentError = iota
	CallArgumentErrorVariadicTuple            CallArgumentError = iota
	CallArgumentErrorMultipleVariadicExpand   CallArgumentError = iota
	CallArgumentErrorAmbiguousPolymorphicVariadic CallArgumentError = iota
	CallArgumentErrorArgumentCount            CallArgumentError = iota
	CallArgumentErrorTooFewArguments          CallArgumentError = iota
	CallArgumentErrorTooManyArguments         CallArgumentError = iota
	CallArgumentErrorInvalidFieldValue        CallArgumentError = iota
	CallArgumentErrorParameterNotFound        CallArgumentError = iota
	CallArgumentErrorParameterMissing         CallArgumentError = iota
	CallArgumentErrorDuplicateParameter       CallArgumentError = iota
	CallArgumentErrorNoneConstantParameter    CallArgumentError = iota
	CallArgumentErrorOutOfOrderParameters     CallArgumentError = iota
	CallArgumentErrorMAX                      CallArgumentError = iota
)

var CallArgumentErrorStrings = [...]string{
	"None",
	"NoneProcedureType",
	"WrongTypes",
	"NonVariadicExpand",
	"VariadicTuple",
	"MultipleVariadicExpand",
	"AmbiguousPolymorphicVariadic",
	"ArgumentCount",
	"TooFewArguments",
	"TooManyArguments",
	"InvalidFieldValue",
	"ParameterNotFound",
	"ParameterMissing",
	"DuplicateParameter",
	"NoneConstantParameter",
	"OutOfOrderParameters",
}

type CallArgumentErrorMode int

const (
	CallArgumentErrorModeNoErrors   CallArgumentErrorMode = iota
	CallArgumentErrorModeShowErrors CallArgumentErrorMode = iota
)

type CallArgumentData struct {
	GenEntity   Entity
	Score       int64
	ResultType  Type
}

type PolyProcData struct {
	GenEntity Entity
	ProcInfo  ProcInfo
}

type ValidIndexAndScore struct {
	Index isize
	Score int64
}

type CIdentSuggestion struct {
	Name string
	Msg  string
}

var CIdentSuggestions = []CIdentSuggestion{
	{"while", "'for'? Odin only has one loop construct: 'for'"},
	{"sizeof", "'size_of'?"},
	{"alignof", "'align_of'?"},
	{"offsetof", "'offset_of'?"},
	{"_Bool", "'bool'?"},
	{"char", "'u8', 'i8', or 'c.char' (which is part of 'core:c')?"},
	{"short", "'i16' or 'c.short' (which is part of 'core:c')?"},
	{"long", "'c.long' (which is part of 'core:c')?"},
	{"float", "'f32'?"},
	{"double", "'f64'?"},
	{"unsigned", "'c.uint' (which is part of 'core:c')?"},
	{"signed", "'c.int' (which is part of 'core:c')?"},
	{"size_t", "'uint', or 'c.size_t' (which is part of 'core:c')?"},
	{"ssize_t", "'int', or 'c.ssize_t' (which is part of 'core:c')?"},
	{"uintptr_t", "'uintptr'?"},
	{"intptr_t", "'uintptr' or `int` or something else?"},
	{"ptrdiff_t", "'int' or 'c.ptrdiff_t' (which is part of 'core:c')?"},
	{"intmax_t", "'c.intmax_t' (which is part of 'core:c')?"},
	{"uintmax_t", "'c.uintmax_t' (which is part of 'core:c')?"},
	{"uint8_t", "'u8'?"},
	{"int8_t", "'i8'?"},
	{"uint16_t", "'u16'?"},
	{"int16_t", "'i16'?"},
	{"uint32_t", "'u32'?"},
	{"int32_t", "'i32'?"},
	{"uint64_t", "'u64'?"},
	{"int64_t", "'i64'?"},
	{"uint128_t", "'u128'?"},
	{"int128_t", "'i128'?"},
	{"float32", "'f32'?"},
	{"float64", "'f64'?"},
	{"float32_t", "'f32'?"},
	{"float64_t", "'f64'?"},
}

type LoadDirectiveResult int

const (
	LoadDirectiveSuccess LoadDirectiveResult = iota
	LoadDirectiveError   LoadDirectiveResult = iota
	LoadDirectiveNotFound LoadDirectiveResult = iota
)

type TypeAndToken struct {
	Type  Type
	Token Token
}

type SeenMap map[uintptr][]TypeAndToken

type UnpackFlags uint32

const (
	UnpackFlagNone       UnpackFlags = 0
	UnpackFlagAllowOk    UnpackFlags = 1 << 0
	UnpackFlagAllowUndef UnpackFlags = 1 << 1
)
