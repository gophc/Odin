// exact_value.odin - Pure Odin rewrite of src/cipp/exact_value.i.cpp
// Package: godin
// compile-time exact value representation

package godin

import "core:fmt"
import "core:math"
import "core:mem"
import "core:strings"
import "core:strconv"
import "core:unicode/utf8"

// ============================================================================
// External type placeholders — real definitions are in sibling package files.
// These minimal declarations let this file be reviewed standalone.
// ============================================================================

Ast	:: rawptr
Type   :: rawptr
Entity :: rawptr

// BigInt is an arbitrary-precision integer defined elsewhere in the package.
// The actual struct must support zero-initialization (BigInt{}) and & address-of.
BigInt :: struct { _: [8]u64 }

// TokenKind mirrors the token type enum in the tokenizer file.
TokenKind :: enum i32 {
	Invalid,
	String,
	Integer,
	Float,
	Imag,
	Rune,
	Add,
	Sub,
	Mul,
	Quo,
	QuoEq,
	Mod,
	ModMod,
	And,
	Or,
	Xor,
	AndNot,
	Shl,
	Shr,
	Not,
	CmpAnd,
	CmpOr,
	CmpEq,
	NotEq,
	Lt,
	LtEq,
	Gt,
	GtEq,
}

// Token_* aliases — shorthand for enum values used throughout the code.
Token_Add	 :: TokenKind.Add
Token_Sub	 :: TokenKind.Sub
Token_Mul	 :: TokenKind.Mul
Token_Quo	 :: TokenKind.Quo
Token_QuoEq   :: TokenKind.QuoEq
Token_Mod	 :: TokenKind.Mod
Token_ModMod  :: TokenKind.ModMod
Token_And	 :: TokenKind.And
Token_Or	  :: TokenKind.Or
Token_Xor	 :: TokenKind.Xor
Token_AndNot  :: TokenKind.AndNot
Token_Shl	 :: TokenKind.Shl
Token_Shr	 :: TokenKind.Shr
Token_Not	 :: TokenKind.Not
Token_CmpAnd  :: TokenKind.CmpAnd
Token_CmpOr   :: TokenKind.CmpOr
Token_CmpEq   :: TokenKind.CmpEq
Token_NotEq   :: TokenKind.NotEq
Token_Lt	  :: TokenKind.Lt
Token_LtEq	:: TokenKind.LtEq
Token_Gt	  :: TokenKind.Gt
Token_GtEq	:: TokenKind.GtEq

// ============================================================================
// External procedures — forward declarations. Bodies live in sibling files.
// ============================================================================

// BigInt constructors
big_int_from_i64	:: proc(b: ^BigInt, i: i64)
big_int_from_u64	:: proc(b: ^BigInt, i: u64)
big_int_from_string :: proc(b: ^BigInt, s: string, success: ^bool)

// BigInt conversions
big_int_to_f64	:: proc(b: ^BigInt) -> f64
big_int_to_i64	:: proc(b: ^BigInt) -> i64
big_int_to_u64	:: proc(b: ^BigInt) -> u64
big_int_to_string :: proc(allocator: mem.Allocator, b: ^BigInt) -> string

// BigInt arithmetic — result written to first parameter
big_int_neg		   :: proc(c, a: ^BigInt)
big_int_not		   :: proc(c, a: ^BigInt, precision: i32, is_signed: bool)
big_int_add		   :: proc(c, a, b: ^BigInt)
big_int_sub		   :: proc(c, a, b: ^BigInt)
big_int_mul		   :: proc(c, a, b: ^BigInt)
big_int_quo		   :: proc(c, a, b: ^BigInt)
big_int_rem		   :: proc(c, a, b: ^BigInt)
big_int_euclidean_mod :: proc(c, a, b: ^BigInt)
big_int_and		   :: proc(c, a, b: ^BigInt)
big_int_or			:: proc(c, a, b: ^BigInt)
big_int_xor		   :: proc(c, a, b: ^BigInt)
big_int_and_not	   :: proc(c, a, b: ^BigInt)
big_int_shl		   :: proc(c, a, b: ^BigInt)
big_int_shr		   :: proc(c, a, b: ^BigInt)

// BigInt comparison — returns <0 / 0 / >0
big_int_cmp :: proc(a, b: ^BigInt) -> i32

// Allocator accessors
permanent_allocator :: proc() -> mem.Allocator
heap_allocator	  :: proc() -> mem.Allocator
temporary_allocator :: proc() -> mem.Allocator

// String quoting — returns a double-quoted printable representation.
quote_to_ascii	:: proc(allocator: mem.Allocator, s: string) -> string
quote_to_ascii_16 :: proc(allocator: mem.Allocator, s: []u16) -> string

// AST expression printer (defined in the AST printing file)
write_expr_to_string :: proc(b: ^strings.Builder, node: Ast, shorthand: bool)

// Entity unwrapping stubs (defined in entity/checker files)
are_types_identical		 :: proc(x, y: Type) -> bool
strip_entity_wrapping_expr  :: proc(expr: Ast) -> Entity
strip_entity_wrapping_entity :: proc(e: Entity) -> Entity

// Compiler error (varargs)
compiler_error :: proc(fmt_str: string, args: ..any)

// ============================================================================
// Internal helpers
// ============================================================================

// fnv32a computes the 32-bit FNV-1a hash of byte data.
@(private)
fnv32a :: proc(data: []byte) -> u32 {
	h: u32 = 0x811c9dc5
	for b in data {
		h = (h ~ u32(b)) * 0x01000193
	}
	return h
}

// ptr_hash hashes a pointer-sized integer as a 32-bit value.
@(private)
ptr_hash :: proc(p: i64) -> u32 {
	return fnv32a(transmute([]byte)mem.Raw_String{&p, size_of(p)})
}

// permanent_alloc_item allocates a single item of type T from the permanent arena.
@(private)
permanent_alloc_item :: proc($T: typeid) -> ^T {
	ptr, _ := mem.alloc(size_of(T), align_of(T), permanent_allocator())
	return (^T)(ptr)
}

// raw_string_from_ptr creates an Odin string from raw pointer + length.
@(private)
raw_string_from_ptr :: proc(data: ^u8, len: int) -> string {
	if len <= 0 do return ""
	return transmute(string)mem.Raw_String{data, len}
}

// raw_string16_from_ptr creates a []u16 from raw pointer + length.
@(private)
raw_string16_from_ptr :: proc(data: ^u16, len: int) -> []u16 {
	if len <= 0 do return nil
	return mem.slice_ptr(data, len)
}

// strip_underscores returns a copy of s with all '_' characters removed.
@(private)
strip_underscores :: proc(s: string, allocator := context.allocator) -> string {
	buf := make([dynamic]byte, 0, len(s), allocator)
	for c in s {
		if c != '_' do append(&buf, byte(c))
	}
	return string(buf[:])
}

// f16_to_f32 converts an IEEE 754 half-precision float (stored as u16 bits) to f32.
@(private)
f16_to_f32 :: proc(h: u16) -> f32 {
	h32 := u32(h)
	s: u32 = (h32 & 0x8000) << 16
	e: u32 = (h32 >> 10) & 0x1f
	m: u32 = h32 & 0x3ff
	result: u32
	if e == 0 {
		if m == 0 {
			result = s  // signed zero
		} else {
			// subnormal — normalize
			e2: u32 = 113
			m2 := m << 1
			for m2 & 0x400 == 0 {
				m2 <<= 1
				e2 -= 1
			}
			m2 &= 0x3ff
			result = s | (e2 << 23) | (m2 << 13)
		}
	} else if e == 31 {
		result = s | 0x7f800000 | (m << 13)
	} else {
		result = s | ((e + 112) << 23) | (m << 13)
	}
	return transmute(f32)result
}

// u64_from_hex_str parses a u64 from a hex string (handles 0h / 0x prefix).
@(private)
u64_from_hex_str :: proc(s: string) -> u64 {
	str := s
	if len(s) >= 2 && (s[:2] == "0h" || s[:2] == "0x") {
		str = s[2:]
	}
	u, _ := strconv.parse_u64(str, 16)
	return u
}

// float_from_string parses a float string, stripping underscores and normalising 'E' → 'e'.
@(private)
float_from_string :: proc(s: string, allocator := context.allocator) -> (f64, bool) {
	buf := make([dynamic]byte, 0, len(s), allocator)
	for c in s {
		if c == '_' do continue
		ch := c
		if c == 'E' do ch = 'e'
		append(&buf, byte(ch))
	}
	return strconv.atof(string(buf[:]))
}

// cmp_f64 returns -1, 0, or 1 comparing two f64 values (NaN-safe).
@(private)
cmp_f64 :: proc(a, b: f64) -> i32 {
	if a > b do return 1
	if a < b do return -1
	return 0
}

// ============================================================================
// Value kind string constants (used for debug / panic messages)
// ============================================================================

EXACT_VALUE_INVALID_MSG	:: "How'd you get here? Invalid Value.kind %d"
MATCH_EXACT_VALUES_MSG	 :: "match_exact_values: How'd you get here? Invalid ExactValueKind %d"
COMPARE_EXACT_VALUES_MSG   :: "Invalid comparison: %d"
INVALID_IMAG_LITERAL_MSG   :: "Invalid imaginary basic literal"
INVALID_HEX_FLOAT_MSG	  :: "Invalid hexadecimal float, expected 4, 8, or 16 digits"

// ============================================================================
// Types
// ============================================================================

Complex128 :: struct {
	real, imag: f64,
}

Quaternion256 :: struct {
	imag, jmag, kmag, real: f64,
}

ExactValueKind :: enum i32 {
	Invalid	= 0,
	Bool	   = 1,
	String	 = 2,
	Integer	= 3,
	Float	  = 4,
	Complex	= 5,
	Quaternion = 6,
	Pointer	= 7,
	Compound   = 8,
	Procedure  = 9,
	Typeid	 = 10,
	String16   = 11,
	Count,
}

// ExactValue holds a compile-time-evaluated constant value.
// Uses a flat struct layout (instead of a union) for Odin compatibility,
// preserving the original field-access patterns.
ExactValue :: struct {
	kind:			 ExactValueKind,
	value_bool:	   bool,
	value_string:	 string,
	value_integer:	BigInt,
	value_float:	  f64,
	value_pointer:	i64,
	value_complex:	^Complex128,
	value_quaternion: ^Quaternion256,
	value_compound:   rawptr,
	value_procedure:  rawptr,
	value_typeid:	 rawptr,
	value_string16:   []u16,
}

// ============================================================================
// Constants
// ============================================================================

EMPTY_EXACT_VALUE :: ExactValue{}

// ============================================================================
// Hash
// ============================================================================

// hash_exact_value computes a 31-bit hash for deduplication / map keys.
hash_exact_value :: proc(v: ExactValue) -> uintptr {
	res: u32 = 0
	#partial switch v.kind {
	case .Invalid:
		return 0
	case .Bool:
		res = fnv32a(transmute([]byte)mem.Raw_String{&v.value_bool, size_of(v.value_bool)})
	case .String:
		res = fnv32a(transmute([]byte)mem.Raw_String{raw_data(v.value_string), len(v.value_string)})
	case .String16:
		res = fnv32a(mem.slice_to_bytes(v.value_string16))
	case .Integer:
		// Hash the entire BigInt struct (approximates original .dp + .sign logic).
		res = fnv32a(transmute([]byte)mem.Raw_String{&v.value_integer, size_of(v.value_integer)})
	case .Float:
		res = fnv32a(transmute([]byte)mem.Raw_String{&v.value_float, size_of(v.value_float)})
	case .Pointer:
		res = ptr_hash(v.value_pointer)
	case .Complex:
		if v.value_complex != nil {
			res = fnv32a(transmute([]byte)mem.Raw_String{v.value_complex, size_of(Complex128)})
		}
	case .Quaternion:
		if v.value_quaternion != nil {
			res = fnv32a(transmute([]byte)mem.Raw_String{v.value_quaternion, size_of(Quaternion256)})
		}
	case .Compound:
		res = ptr_hash(i64(uintptr(v.value_compound)))
	case .Procedure:
		res = ptr_hash(i64(uintptr(v.value_procedure)))
	case .Typeid:
		res = ptr_hash(i64(uintptr(v.value_typeid)))
	case:
		res = fnv32a(transmute([]byte)mem.Raw_String{&v, size_of(v)})
	}
	return uintptr(res & 0x7fffffff)
}

// ============================================================================
// Constructors
// ============================================================================

exact_value_compound :: proc(node: Ast) -> ExactValue {
	return ExactValue{kind = .Compound, value_compound = node}
}

exact_value_bool :: proc(b: bool) -> ExactValue {
	return ExactValue{kind = .Bool, value_bool = b}
}

exact_value_string :: proc(s: string) -> ExactValue {
	return ExactValue{kind = .String, value_string = s}
}

exact_value_string16 :: proc(s: []u16) -> ExactValue {
	return ExactValue{kind = .String16, value_string16 = s}
}

exact_value_i64 :: proc(i: i64) -> ExactValue {
	result := ExactValue{kind = .Integer}
	big_int_from_i64(&result.value_integer, i)
	return result
}

exact_value_u64 :: proc(i: u64) -> ExactValue {
	result := ExactValue{kind = .Integer}
	big_int_from_u64(&result.value_integer, i)
	return result
}

exact_value_float :: proc(f: f64) -> ExactValue {
	return ExactValue{kind = .Float, value_float = f}
}

exact_value_complex :: proc(real, imag: f64) -> ExactValue {
	result := ExactValue{kind = .Complex}
	result.value_complex = permanent_alloc_item(Complex128)
	result.value_complex.real = real
	result.value_complex.imag = imag
	return result
}

exact_value_quaternion :: proc(real, imag, jmag, kmag: f64) -> ExactValue {
	result := ExactValue{kind = .Quaternion}
	result.value_quaternion = permanent_alloc_item(Quaternion256)
	result.value_quaternion.real = real
	result.value_quaternion.imag = imag
	result.value_quaternion.jmag = jmag
	result.value_quaternion.kmag = kmag
	return result
}

exact_value_pointer :: proc(ptr: i64) -> ExactValue {
	return ExactValue{kind = .Pointer, value_pointer = ptr}
}

exact_value_procedure :: proc(node: Ast) -> ExactValue {
	return ExactValue{kind = .Procedure, value_procedure = node}
}

exact_value_typeid :: proc(type_: Type) -> ExactValue {
	return ExactValue{kind = .Typeid, value_typeid = type_}
}

exact_value_integer_from_string :: proc(s: string) -> ExactValue {
	result := ExactValue{kind = .Integer}
	success: bool
	big_int_from_string(&result.value_integer, s, &success)
	if !success {
		return EMPTY_EXACT_VALUE
	}
	return result
}

// ============================================================================
// Float / hex-float parsing
// ============================================================================

// exact_value_float_from_string parses a float literal string into an ExactValue.
// Handles: hex floats (0h...), integers disguised as floats, and regular floats.
exact_value_float_from_string :: proc(s: string, allocator := context.allocator) -> ExactValue {
	// Hex float: 0hXXXX
	if len(s) > 2 && s[:2] == "0h" {
		digit_count := 0
		for c in s[2:] {
			if c != '_' do digit_count += 1
		}
		u := u64_from_hex_str(s)
		switch digit_count {
		case 4:
			f := f16_to_f32(u16(u))
			return exact_value_float(f64(f))
		case 8:
			f := transmute(f32)u32(u)
			return exact_value_float(f64(f))
		case 16:
			return exact_value_float(transmute(f64)u)
		case:
			panic(INVALID_HEX_FLOAT_MSG)
		}
	}

	// If the string looks like an integer (no decimal point or exponent minus),
	// try integer parsing first.
	if !strings.contains_rune(s, '.') && !strings.contains_rune(s, '-') {
		return exact_value_integer_from_string(s)
	}

	f, ok := float_from_string(s, allocator)
	if !ok {
		return EMPTY_EXACT_VALUE
	}
	return exact_value_float(f)
}

// ============================================================================
// exact_value_from_basic_literal
// ============================================================================

// exact_value_from_basic_literal constructs an ExactValue from a token kind and string.
exact_value_from_basic_literal :: proc(kind: TokenKind, s: string, allocator := context.allocator) -> ExactValue {
	switch kind {
	case .String:
		return exact_value_string(s)
	case .Integer:
		return exact_value_integer_from_string(s)
	case .Float:
		return exact_value_float_from_string(s, allocator)
	case .Imag:
		if len(s) == 0 {
			return EMPTY_EXACT_VALUE
		}
		last_rune := rune(s[len(s) - 1])
		trimmed := s[:len(s) - 1]
		imag, _ := float_from_string(trimmed, allocator)
		switch last_rune {
		case 'i': return exact_value_complex(0, imag)
		case 'j': return exact_value_quaternion(0, 0, imag, 0)
		case 'k': return exact_value_quaternion(0, 0, 0, imag)
		case:	 panic(INVALID_IMAG_LITERAL_MSG)
		}
	case .Rune:
		r: rune = 0xfffd
		if len(s) == 1 {
			r = rune(s[0])
		} else {
			r, _ = utf8.decode_rune_in_string(s)
		}
		return exact_value_i64(i64(r))
	}
	return EMPTY_EXACT_VALUE
}

// ============================================================================
// Type conversion: to_integer / to_float / to_complex / to_quaternion
// ============================================================================

exact_value_to_integer :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Bool:
		i: i64 = 0
		if v.value_bool do i = 1
		return exact_value_i64(i)
	case .Integer:
		return v
	case .Float:
		i := i64(v.value_float)
		if f64(i) == v.value_float {
			return exact_value_i64(i)
		}
	case .Pointer:
		return exact_value_i64(i64(v.value_pointer))
	}
	return EMPTY_EXACT_VALUE
}

exact_value_to_float :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Integer:
		return exact_value_float(big_int_to_f64(&v.value_integer))
	case .Float:
		return v
	}
	return EMPTY_EXACT_VALUE
}

exact_value_to_complex :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Integer:
		return exact_value_complex(big_int_to_f64(&v.value_integer), 0)
	case .Float:
		return exact_value_complex(v.value_float, 0)
	case .Complex:
		return v
	}
	return EMPTY_EXACT_VALUE
}

exact_value_to_quaternion :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Integer:
		return exact_value_quaternion(big_int_to_f64(&v.value_integer), 0, 0, 0)
	case .Float:
		return exact_value_quaternion(v.value_float, 0, 0, 0)
	case .Complex:
		if v.value_complex != nil {
			return exact_value_quaternion(v.value_complex.real, v.value_complex.imag, 0, 0)
		}
	case .Quaternion:
		return v
	}
	return EMPTY_EXACT_VALUE
}

// ============================================================================
// Component accessors
// ============================================================================

exact_value_real :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Integer, .Float:
		return v
	case .Complex:
		if v.value_complex != nil {
			return exact_value_float(v.value_complex.real)
		}
	case .Quaternion:
		if v.value_quaternion != nil {
			return exact_value_float(v.value_quaternion.real)
		}
	}
	return EMPTY_EXACT_VALUE
}

exact_value_imag :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Integer, .Float:
		return exact_value_i64(0)
	case .Complex:
		if v.value_complex != nil {
			return exact_value_float(v.value_complex.imag)
		}
	case .Quaternion:
		if v.value_quaternion != nil {
			return exact_value_float(v.value_quaternion.imag)
		}
	}
	return EMPTY_EXACT_VALUE
}

exact_value_jmag :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Integer, .Float, .Complex:
		return exact_value_i64(0)
	case .Quaternion:
		if v.value_quaternion != nil {
			return exact_value_float(v.value_quaternion.jmag)
		}
	}
	return EMPTY_EXACT_VALUE
}

exact_value_kmag :: proc(v: ExactValue) -> ExactValue {
	switch v.kind {
	case .Integer, .Float, .Complex:
		return exact_value_i64(0)
	case .Quaternion:
		if v.value_quaternion != nil {
			return exact_value_float(v.value_quaternion.kmag)
		}
	}
	return EMPTY_EXACT_VALUE
}

// ============================================================================
// Scalar extraction
// ============================================================================

exact_value_to_i64 :: proc(v: ExactValue) -> i64 {
	iv := exact_value_to_integer(v)
	if iv.kind == .Integer {
		return big_int_to_i64(&iv.value_integer)
	}
	return 0
}

exact_value_to_u64 :: proc(v: ExactValue) -> u64 {
	iv := exact_value_to_integer(v)
	if iv.kind == .Integer {
		return big_int_to_u64(&iv.value_integer)
	}
	return 0
}

exact_value_to_f64 :: proc(v: ExactValue) -> f64 {
	fv := exact_value_to_float(v)
	if fv.kind == .Float {
		return fv.value_float
	}
	return 0.0
}

// ============================================================================
// Unary operators
// ============================================================================

// exact_unary_operator_value evaluates a unary operator on an exact value.
// `precision` and `is_unsigned` are only meaningful for bitwise-NOT on integers.
exact_unary_operator_value :: proc(op: TokenKind, v: ExactValue, precision: i32, is_unsigned: bool) -> ExactValue {
	switch op {
	case .Add:
		switch v.kind {
		case .Invalid, .Integer, .Float, .Complex, .Quaternion:
			return v
		}
	case .Sub:
		switch v.kind {
		case .Invalid:
			return v
		case .Integer:
			i := ExactValue{kind = .Integer}
			big_int_neg(&i.value_integer, &v.value_integer)
			return i
		case .Float:
			i := v
			i.value_float = -i.value_float
			return i
		case .Complex:
			if v.value_complex != nil {
				return exact_value_complex(-v.value_complex.real, -v.value_complex.imag)
			}
		case .Quaternion:
			if v.value_quaternion != nil {
				q := v.value_quaternion
				return exact_value_quaternion(-q.real, -q.imag, -q.jmag, -q.kmag)
			}
		}
	case .Xor:
		// Unary ~ (bitwise NOT) — Token_Xor represents the ~ token.
		switch v.kind {
		case .Invalid:
			return v
		case .Integer:
			assert(precision != 0, "precision must be non-zero for bitwise NOT")
			i := ExactValue{kind = .Integer}
			big_int_not(&i.value_integer, &v.value_integer, precision, !is_unsigned)
			return i
		case:
			return EMPTY_EXACT_VALUE
		}
	case .Not:
		switch v.kind {
		case .Invalid:
			return v
		case .Bool:
			return exact_value_bool(!v.value_bool)
		}
	}
	return EMPTY_EXACT_VALUE
}

// ============================================================================
// Value ordering (for match_exact_values promotion lattice)
// ============================================================================

// exact_value_order returns a rank for promotion: higher rank wins.
exact_value_order :: proc(v: ExactValue) -> i32 {
	switch v.kind {
	case .Invalid, .Compound:
		return 0
	case .Bool, .String, .String16:
		return 1
	case .Integer:
		return 2
	case .Float:
		return 3
	case .Complex:
		return 4
	case .Quaternion:
		return 5
	case .Pointer:
		return 6
	case .Procedure:
		return 7
	case:
		panic(EXACT_VALUE_INVALID_MSG, v.kind)
	}
	return -1
}

// ============================================================================
// match_exact_values — promote x to y's type level for binary ops
// ============================================================================

match_exact_values :: proc(x, y: ^ExactValue) {
	// Ensure y has the higher-ranked kind.
	if exact_value_order(y^) < exact_value_order(x^) {
		match_exact_values(y, x)
		return
	}
	switch x.kind {
	case .Invalid:
		y^ = x^
		return
	case .Bool, .String, .String16, .Quaternion, .Pointer, .Compound, .Procedure, .Typeid:
		return
	case .Integer:
		switch y.kind {
		case .Integer:
			return
		case .Float:
			x^ = exact_value_float(big_int_to_f64(&x.value_integer))
			return
		case .Complex:
			x^ = exact_value_complex(big_int_to_f64(&x.value_integer), 0)
			return
		case .Quaternion:
			x^ = exact_value_quaternion(big_int_to_f64(&x.value_integer), 0, 0, 0)
			return
		}
	case .Float:
		switch y.kind {
		case .Float:
			return
		case .Complex:
			x^ = exact_value_to_complex(x^)
			return
		case .Quaternion:
			x^ = exact_value_to_quaternion(x^)
			return
		}
	case .Complex:
		switch y.kind {
		case .Complex:
			return
		case .Quaternion:
			x^ = exact_value_to_quaternion(x^)
			return
		}
	}
	compiler_error(MATCH_EXACT_VALUES_MSG, x.kind)
}

// ============================================================================
// Binary operators
// ============================================================================

// exact_binary_operator_value evaluates a binary operator on two exact values.
exact_binary_operator_value :: proc(op: TokenKind, x_in, y_in: ExactValue) -> ExactValue {
	x, y := x_in, y_in
	match_exact_values(&x, &y)

	switch x.kind {
	case .Invalid:
		return x

	case .Bool:
		switch op {
		case .CmpAnd: return exact_value_bool(x.value_bool && y.value_bool)
		case .CmpOr:  return exact_value_bool(x.value_bool || y.value_bool)
		case .And:	return exact_value_bool(x.value_bool && y.value_bool)  // bitwise & → logical
		case .Or:	 return exact_value_bool(x.value_bool || y.value_bool)
		case .AndNot: return exact_value_bool(x.value_bool && !y.value_bool)
		case .Xor:	return exact_value_bool(x.value_bool != y.value_bool)
		case:		 return EMPTY_EXACT_VALUE
		}

	case .Integer:
		a := &x.value_integer
		b := &y.value_integer
		c: BigInt
		switch op {
		case .Add:
			big_int_add(&c, a, b)
		case .Sub:
			big_int_sub(&c, a, b)
		case .Mul:
			big_int_mul(&c, a, b)
		case .Quo:
			// Integer / → fmod of floats (preserving original behaviour)
			return exact_value_float(math.mod(big_int_to_f64(a), big_int_to_f64(b)))
		case .QuoEq:
			big_int_quo(&c, a, b)
		case .Mod:
			big_int_rem(&c, a, b)
		case .ModMod:
			big_int_euclidean_mod(&c, a, b)
		case .And:
			big_int_and(&c, a, b)
		case .Or:
			big_int_or(&c, a, b)
		case .Xor:
			big_int_xor(&c, a, b)
		case .AndNot:
			big_int_and_not(&c, a, b)
		case .Shl:
			big_int_shl(&c, a, b)
		case .Shr:
			big_int_shr(&c, a, b)
		case:
			return EMPTY_EXACT_VALUE
		}
		return ExactValue{kind = .Integer, value_integer = c}

	case .Float:
		a := x.value_float
		b := y.value_float
		switch op {
		case .Add: return exact_value_float(a + b)
		case .Sub: return exact_value_float(a - b)
		case .Mul: return exact_value_float(a * b)
		case .Quo: return exact_value_float(a / b)
		case:	  return EMPTY_EXACT_VALUE
		}

	case .Complex:
		y = exact_value_to_complex(y)
		a, b := x.value_complex.real, x.value_complex.imag
		c, d := y.value_complex.real, y.value_complex.imag
		switch op {
		case .Add:
			return exact_value_complex(a + c, b + d)
		case .Sub:
			return exact_value_complex(a - c, b - d)
		case .Mul:
			return exact_value_complex(a*c - b*d, b*c + a*d)
		case .Quo:
			s := c*c + d*d
			return exact_value_complex((a*c + b*d) / s, (b*c - a*d) / s)
		case:
			return EMPTY_EXACT_VALUE
		}

	case .Quaternion:
		y = exact_value_to_quaternion(y)
		xr, xi, xj, xk := x.value_quaternion.real, x.value_quaternion.imag, x.value_quaternion.jmag, x.value_quaternion.kmag
		yr, yi, yj, yk := y.value_quaternion.real, y.value_quaternion.imag, y.value_quaternion.jmag, y.value_quaternion.kmag
		switch op {
		case .Add:
			return exact_value_quaternion(xr + yr, xi + yi, xj + yj, xk + yk)
		case .Sub:
			return exact_value_quaternion(xr - yr, xi - yi, xj - yj, xk - yk)
		case .Mul:
			return exact_value_quaternion(
				xr*yr - xi*yi - xj*yj - xk*yk,
				xr*yi + xi*yr + xj*yk - xk*yj,
				xr*yj - xi*yk + xj*yr + xk*yi,
				xr*yk + xi*yj - xj*yi + xk*yr,
			)
		case .Quo:
			invmag2 := 1.0 / (yr*yr + yi*yi + yj*yj + yk*yk)
			return exact_value_quaternion(
				( xr*yr + xi*yi + xj*yj + xk*yk) * invmag2,
				(-xr*yi + xi*yr - xj*yk + xk*yj) * invmag2,
				(-xr*yj + xi*yk + xj*yr - xk*yi) * invmag2,
				(-xr*yk - xi*yj + xj*yi + xk*yr) * invmag2,
			)
		case:
			return EMPTY_EXACT_VALUE
		}

	case .String:
		if op == .Add {
			combined := strings.concatenate({x.value_string, y.value_string}, permanent_allocator())
			return exact_value_string(combined)
		}
		return EMPTY_EXACT_VALUE

	case .String16:
		if op == .Add {
			// Concatenate two []u16 slices
			sx, sy := x.value_string16, y.value_string16
			total := len(sx) + len(sy)
			data := make([]u16, total, permanent_allocator())
			copy(data, sx)
			copy(data[len(sx):], sy)
			return exact_value_string16(data)
		}
		return EMPTY_EXACT_VALUE
	}

	return EMPTY_EXACT_VALUE
}

// ============================================================================
// Convenience wrappers around exact_binary_operator_value
// ============================================================================

exact_value_add :: proc(x, y: ExactValue) -> ExactValue {
	return exact_binary_operator_value(.Add, x, y)
}

exact_value_sub :: proc(x, y: ExactValue) -> ExactValue {
	return exact_binary_operator_value(.Sub, x, y)
}

exact_value_mul :: proc(x, y: ExactValue) -> ExactValue {
	return exact_binary_operator_value(.Mul, x, y)
}

exact_value_quo :: proc(x, y: ExactValue) -> ExactValue {
	return exact_binary_operator_value(.Quo, x, y)
}

exact_value_shift :: proc(op: TokenKind, x, y: ExactValue) -> ExactValue {
	return exact_binary_operator_value(op, x, y)
}

exact_value_increment_one :: proc(x: ExactValue) -> ExactValue {
	return exact_binary_operator_value(.Add, x, exact_value_i64(1))
}

// ============================================================================
// Comparison
// ============================================================================

// compare_exact_values_compound_lit — declared elsewhere; stub with comment.
// Body is defined in the AST / checker module. This file provides no
// implementation for it.
//
// compare_exact_values_compound_lit :: proc(op: TokenKind, x, y: ExactValue) -> bool {
//	 // Defined in another file — compares compound literal trees.
// }

// compare_exact_values compares two exact values using the given relational operator.
compare_exact_values :: proc(op: TokenKind, x_in, y_in: ExactValue) -> bool {
	x, y := x_in, y_in
	match_exact_values(&x, &y)

	switch x.kind {
	case .Invalid:
		return false

	case .Bool:
		switch op {
		case .CmpEq: return x.value_bool == y.value_bool
		case .NotEq: return x.value_bool != y.value_bool
		}

	case .Integer:
		cmp := big_int_cmp(&x.value_integer, &y.value_integer)
		switch op {
		case .CmpEq: return cmp == 0
		case .NotEq: return cmp != 0
		case .Lt:	return cmp <  0
		case .LtEq:  return cmp <= 0
		case .Gt:	return cmp >  0
		case .GtEq:  return cmp >= 0
		}

	case .Float:
		a, b := x.value_float, y.value_float
		if math.is_nan(a) || math.is_nan(b) {
			return op == .NotEq
		}
		c := cmp_f64(a, b)
		switch op {
		case .CmpEq: return c == 0
		case .NotEq: return c != 0
		case .Lt:	return c <  0
		case .LtEq:  return c <= 0
		case .Gt:	return c >  0
		case .GtEq:  return c >= 0
		}

	case .Complex:
		if x.value_complex == nil || y.value_complex == nil do return false
		a, b := x.value_complex.real, x.value_complex.imag
		c, d := y.value_complex.real, y.value_complex.imag
		switch op {
		case .CmpEq: return cmp_f64(a, c) == 0 && cmp_f64(b, d) == 0
		case .NotEq: return cmp_f64(a, c) != 0 || cmp_f64(b, d) != 0
		}

	case .String:
		a, b := x.value_string, y.value_string
		switch op {
		case .CmpEq: return a == b
		case .NotEq: return a != b
		case .Lt:	return strings.compare(a, b) <  0
		case .LtEq:  return strings.compare(a, b) <= 0
		case .Gt:	return strings.compare(a, b) >  0
		case .GtEq:  return strings.compare(a, b) >= 0
		}

	case .String16:
		a, b := x.value_string16, y.value_string16
		switch op {
		case .CmpEq: return string(a) == string(b)  // fallback: compare raw bytes
		case .NotEq: return string(a) != string(b)
		// Full []u16 comparison would need a custom comparator;
		// for now only equality / inequality is supported.
		}

	case .Pointer:
		switch op {
		case .CmpEq: return x.value_pointer == y.value_pointer
		case .NotEq: return x.value_pointer != y.value_pointer
		case .Lt:	return x.value_pointer <  y.value_pointer
		case .LtEq:  return x.value_pointer <= y.value_pointer
		case .Gt:	return x.value_pointer >  y.value_pointer
		case .GtEq:  return x.value_pointer >= y.value_pointer
		}

	case .Typeid:
		switch op {
		case .CmpEq: return x.value_typeid == y.value_typeid
		case .NotEq: return x.value_typeid != y.value_typeid
		}

	case .Procedure:
		switch op {
		case .CmpEq: return x.value_procedure == y.value_procedure
		case .NotEq: return x.value_procedure != y.value_procedure
		}

	case .Compound:
		if op != .CmpEq && op != .NotEq do return false
		if x.kind != y.kind do return false
		// compare_exact_values_compound_lit is defined in the AST module.
		// For now, fall through to the assertion.
		panic(COMPARE_EXACT_VALUES_MSG, x.kind)
	}

	panic(COMPARE_EXACT_VALUES_MSG, x.kind)
	return false
}

// ============================================================================
// String conversion
// ============================================================================

// write_exact_value_to_string appends a human-readable representation of v to builder b.
write_exact_value_to_string :: proc(b: ^strings.Builder, v: ExactValue, string_limit: int = 36, allocator := context.allocator) {
	limit := max(string_limit, 36)

	switch v.kind {
	case .Invalid:
		return

	case .Bool:
		if v.value_bool {
			strings.write_string(b, "true")
		} else {
			strings.write_string(b, "false")
		}

	case .String:
		quoted := strconv.quote(v.value_string, '"')
		defer delete(quoted, allocator)
		if len(quoted) <= limit {
			strings.write_string(b, quoted)
		} else {
			n := limit / 5
			if n < 1 do n = 1
			strings.write_string(b, quoted[:n])
			fmt.sbprintf(b, "\"..%d chars..\"", len(quoted) - 2*n)
			strings.write_string(b, quoted[len(quoted)-n:])
		}

	case .String16:
		// Convert []u16 to a quoted ASCII representation via external helper.
		// The original calls quote_to_ascii(heap_allocator(), v.value_string16).
		// We assume quote_to_ascii_16 exists in the package.
		// NOTE: If not available, falls back to "<string16>".
		// quote_to_ascii_16 is defined in the string utilities file.
		// For now, produce a placeholder.
		quoted := quote_to_ascii_16(allocator, v.value_string16)
		defer delete(quoted, allocator)
		if len(quoted) <= limit {
			strings.write_string(b, quoted)
		} else {
			n := limit / 5
			if n < 1 do n = 1
			strings.write_string(b, quoted[:n])
			fmt.sbprintf(b, "\"..%d chars..\"", len(quoted) - 2*n)
			strings.write_string(b, quoted[len(quoted)-n:])
		}

	case .Integer:
		str := big_int_to_string(allocator, &v.value_integer)
		defer delete(str, allocator)
		strings.write_string(b, str)

	case .Float:
		fmt.sbprintf(b, "%f", v.value_float)

	case .Complex:
		if v.value_complex != nil {
			fmt.sbprintf(b, "%f+%fi", v.value_complex.real, v.value_complex.imag)
		}

	case .Quaternion:
		if v.value_quaternion != nil {
			q := v.value_quaternion
			fmt.sbprintf(b, "%f+%fi+%fj+%fk", q.real, q.imag, q.jmag, q.kmag)
		}

	case .Pointer:
		return

	case .Compound:
		write_expr_to_string(b, v.value_compound, false)

	case .Procedure:
		write_expr_to_string(b, v.value_procedure, false)
	}
}

// exact_value_to_string returns a string representation of the exact value.
exact_value_to_string :: proc(v: ExactValue, string_limit: int = 36, allocator := context.allocator) -> string {
	b := strings.builder_make(allocator)
	write_exact_value_to_string(&b, v, string_limit, allocator)
	return strings.to_string(b)
}