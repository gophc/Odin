package gochibicc

//
// tokenize.c
//

type TokenKind int

const (
	TK_NONE    TokenKind = iota
	TK_IDENT             // Identifiers
	TK_PUNCT             // Punctuators
	TK_KEYWORD           // Keywords
	TK_STR               // String literals
	TK_NUM               // Numeric literals
	TK_PP_NUM            // Preprocessing numbers
	TK_EOF               // End-of-file markers
)

type File struct {
	name     string
	file_no  int
	contents string

	// For #line directive
	display_name string
	line_delta   int
}

type Token struct {
	kind TokenKind // Token kind
	next *Token    // Next token
	val  int64     // If kind is TK_NUM, its value
	fval float64   // If kind is TK_NUM, its value
	loc  int       // Token location
	len  int       // Token length
	ty   *Type     // Used if TK_NUM or TK_STR
	str  string    // String literal contents including terminating '\0'

	file       *File    // Source location
	filename   string   // Filename
	line_no    int      // Line number
	line_delta int      // Line number
	at_bol     bool     // True if this token is at beginning of line
	has_space  bool     // True if this token follows a space character
	hideset    *Hideset // For macro expansion
	origin     *Token   // If this is expanded from a macro, the original token
}

/*
void convert_pp_tokens(Token *tok);
File **get_input_files(void);
File *new_file(char *name, int file_no, char *contents);
Token *tokenize_string_literal(Token *tok, Type *basety);
Token *tokenize(File *file);
Token *tokenize_file(char *filename);
*/

//
// preprocess.c
//

type Hideset struct {
	next *Hideset
	name string
}

/*
char *search_include_paths(char *filename);
void init_macros(void);
void define_macro(char *name, char *buf);
void undef_macro(char *name);
Token *preprocess(Token *tok);
*/

//
// parse.c
//

// Obj Variable or function
type Obj struct {
	next     *Obj
	name     string // Variable name
	ty       *Type  // Type
	tok      *Token // representative token
	is_local bool   // local or global/function
	align    int    // alignment

	// Local variable
	offset int

	// Global variable or function
	is_function   bool
	is_definition bool
	is_static     bool

	// Global variable
	is_tentative bool
	is_tls       bool
	init_data    string
	rel          Relocation

	// Function
	is_inline     bool
	params        *Obj
	body          *Node
	locals        *Obj
	va_area       *Obj
	alloca_bottom *Obj
	stack_size    int

	// Static inline function
	is_live bool
	is_root bool
	refs    []string
}

// Relocation Global variable can be initialized either by a constant expression
// or a pointer to another global variable. This struct represents the
// latter.
type Relocation struct {
	next   *Relocation
	offset int
	label  string
	addend int64
}

// AST node

type NodeKind int

const (
	ND_NONE      NodeKind = iota
	ND_NULL_EXPR          // Do nothing
	ND_ADD                // +
	ND_SUB                // -
	ND_MUL                // *
	ND_DIV                // /
	ND_NEG                // unary -
	ND_MOD                // %
	ND_BITAND             // &
	ND_BITOR              // |
	ND_BITXOR             // ^
	ND_SHL                // <<
	ND_SHR                // >>
	ND_EQ                 // ==
	ND_NE                 // !=
	ND_LT                 // <
	ND_LE                 // <=
	ND_ASSIGN             // =
	ND_COND               // ?:
	ND_COMMA              // ,
	ND_MEMBER             // . (struct member access)
	ND_ADDR               // unary &
	ND_DEREF              // unary *
	ND_NOT                // !
	ND_BITNOT             // ~
	ND_LOGAND             // &&
	ND_LOGOR              // ||
	ND_RETURN             // "return"
	ND_IF                 // "if"
	ND_FOR                // "for" or "while"
	ND_DO                 // "do"
	ND_SWITCH             // "switch"
	ND_CASE               // "case"
	ND_BLOCK              // { ... }
	ND_GOTO               // "goto"
	ND_GOTO_EXPR          // "goto" labels-as-values
	ND_LABEL              // Labeled statement
	ND_LABEL_VAL          // [GNU] Labels-as-values
	ND_FUNCALL            // Function call
	ND_EXPR_STMT          // Expression statement
	ND_STMT_EXPR          // Statement expression
	ND_VAR                // Variable
	ND_VLA_PTR            // VLA designator
	ND_NUM                // Integer
	ND_CAST               // Type cast
	ND_MEMZERO            // Zero-clear a stack variable
	ND_ASM                // "asm"
	ND_CAS                // Atomic compare-and-swap
	ND_EXCH               // Atomic exchange
)

// Node AST node type
type Node struct {
	kind NodeKind // Node kind
	next *Node    // Next node
	ty   *Type    // Type, e.g. int or pointer to int
	tok  *Token   // Representative token

	lhs *Node // Left-hand side
	rhs *Node // Right-hand side

	// "if" or "for" statement
	cond *Node
	then *Node
	els  *Node
	init *Node
	inc  *Node

	// "break" and "continue" labels
	brk_label  string
	cont_label string

	// Block or statement expression
	body *Node

	// Struct member access
	member *Member

	// Function call
	func_ty       *Type
	args          *Node
	pass_by_stack bool
	ret_buffer    *Obj

	// Goto or labeled statement, or labels-as-values
	label        string
	unique_label string
	goto_next    *Node

	// Switch
	case_next    *Node
	default_case *Node

	// Case
	begin int64
	end   int64

	// "asm" string literal
	asm_str string

	// Atomic compare-and-swap
	cas_addr *Node
	cas_old  *Node
	cas_new  *Node

	// Atomic op= operators
	atomic_addr *Obj
	atomic_expr *Node

	// Variable
	var_ *Obj

	// Numeric literal
	val  int64
	fval float64
}

/*
Node *new_cast(Node *expr, Type *ty);
int64_t const_expr(Token **rest, Token *tok);
Obj *parse(Token *tok);
*/

//
// type.c
//

type TypeKind int

const (
	TY_NONE TypeKind = iota
	TY_VOID
	TY_BOOL
	TY_CHAR
	TY_SHORT
	TY_INT
	TY_LONG
	TY_FLOAT
	TY_DOUBLE
	TY_LDOUBLE
	TY_ENUM
	TY_PTR
	TY_FUNC
	TY_ARRAY
	TY_VLA // variable-length array
	TY_STRUCT
	TY_UNION
)

type Type struct {
	kind        TypeKind
	size        int   // sizeof() value
	align       int   // alignment
	is_unsigned bool  // unsigned or signed
	is_atomic   bool  // true if _Atomic
	origin      *Type // for type compatibility check

	// Pointer-to or array-of type. We intentionally use the same member
	// to represent pointer/array duality in C.
	//
	// In many contexts in which a pointer is expected, we examine this
	// member instead of "kind" member to determine whether a type is a
	// pointer or not. That means in many contexts "array of T" is
	// naturally handled as if it were "pointer to T", as required by
	// the C spec.
	base *Type

	// Declaration
	name     *Token
	name_pos *Token

	// Array
	array_len int

	// Variable-length array
	vla_len  *Node // # of elements
	vla_size *Obj  // sizeof() value

	// Struct
	members     *Member
	is_flexible bool
	is_packed   bool

	// Function type
	return_ty   *Type
	params      *Type
	is_variadic bool
	next        *Type
}

// Member Struct member
type Member struct {
	next   *Member
	ty     *Type
	tok    *Token // for error message
	name   *Token
	idx    int
	align  int
	offset int

	// Bitfield
	is_bitfield bool
	bit_offset  int
	bit_width   int
}

//
// codegen.c
//

/*
void codegen(Obj *prog, FILE *out);
int align_to(int n, int align);
*/

//
// unicode.c
//

/*
int encode_utf8(char *buf, uint32_t c);
uint32_t decode_utf8(char **new_pos, char *p);
bool is_ident1(uint32_t c);
bool is_ident2(uint32_t c);
int display_width(char *p, int len);
*/

type ChibiccApp struct {
	include_paths []string
	opt_fpic      bool
	opt_fcommon   bool
	base_file     string

	// tokenize.c
	current_file *File
	input_files  *File
	at_bol       bool
	has_space    bool

	fp *FileProvider
}




var ty_void, ty_bool *Type
var ty_char, ty_short, ty_int, ty_long *Type
var ty_uchar, ty_ushort, ty_uint, ty_ulong *Type
var ty_float, ty_double, ty_ldouble *Type

func init() {
	ty_void = &Type{kind: TY_VOID, size: 1, align: 1}
	ty_bool = &Type{kind: TY_BOOL, size: 1, align: 1}

	ty_char = &Type{kind: TY_CHAR, size: 1, align: 1}
	ty_short = &Type{kind: TY_SHORT, size: 2, align: 2}
	ty_int = &Type{kind: TY_INT, size: 4, align: 4}
	ty_long = &Type{kind: TY_LONG, size: 8, align: 8}

	ty_uchar = &Type{kind: TY_CHAR, size: 1, align: 1, is_unsigned: true}
	ty_ushort = &Type{kind: TY_SHORT, size: 2, align: 2, is_unsigned: true}
	ty_uint = &Type{kind: TY_INT, size: 4, align: 4, is_unsigned: true}
	ty_ulong = &Type{kind: TY_LONG, size: 8, align: 8, is_unsigned: true}

	ty_float = &Type{kind: TY_FLOAT, size: 4, align: 4}
	ty_double = &Type{kind: TY_DOUBLE, size: 8, align: 8}
	ty_ldouble = &Type{kind: TY_LDOUBLE, size: 16, align: 16}
}
