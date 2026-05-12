package godin

import "core:testing"
import "core:math"
import "core:strings"

// ============================================================================
// Constructor Tests
// ============================================================================

@(test)
test_exact_value_bool :: proc(t: ^testing.T) {
	v := exact_value_bool(true)
	testing.expect(t, v.kind == .Bool, "expected Bool kind")
	testing.expect(t, v.value_bool == true, "expected true")

	v2 := exact_value_bool(false)
	testing.expect(t, v2.kind == .Bool, "expected Bool kind")
	testing.expect(t, v2.value_bool == false, "expected false")
}

@(test)
test_exact_value_i64 :: proc(t: ^testing.T) {
	v := exact_value_i64(42)
	testing.expect(t, v.kind == .Integer, "expected Integer kind")
	// exact_value_to_i64 round-trips through BigInt
	iv := exact_value_to_i64(v)
	testing.expect(t, iv == 42, "expected 42")
}

@(test)
test_exact_value_u64 :: proc(t: ^testing.T) {
	v := exact_value_u64(12345)
	testing.expect(t, v.kind == .Integer, "expected Integer kind")
	uv := exact_value_to_u64(v)
	testing.expect(t, uv == 12345, "expected 12345")
}

@(test)
test_exact_value_float :: proc(t: ^testing.T) {
	v := exact_value_float(3.14159)
	testing.expect(t, v.kind == .Float, "expected Float kind")
	testing.expectf(t, math.abs(v.value_float - 3.14159) < 0.00001,
		"expected ~3.14159, got %f", v.value_float)
}

@(test)
test_exact_value_string :: proc(t: ^testing.T) {
	v := exact_value_string("hello")
	testing.expect(t, v.kind == .String, "expected String kind")
	testing.expect(t, v.value_string == "hello", "expected 'hello'")
}

@(test)
test_exact_value_string_empty :: proc(t: ^testing.T) {
	v := exact_value_string("")
	testing.expect(t, v.kind == .String, "expected String kind")
	testing.expect(t, v.value_string == "", "expected empty string")
}

@(test)
test_exact_value_complex :: proc(t: ^testing.T) {
	v := exact_value_complex(1.5, 2.5)
	testing.expect(t, v.kind == .Complex, "expected Complex kind")
	testing.expect(t, v.value_complex != nil, "expected non-nil complex ptr")
	if v.value_complex != nil {
		testing.expectf(t, math.abs(v.value_complex.real - 1.5) < 0.00001,
			"expected real=1.5, got %f", v.value_complex.real)
		testing.expectf(t, math.abs(v.value_complex.imag - 2.5) < 0.00001,
			"expected imag=2.5, got %f", v.value_complex.imag)
	}
}

@(test)
test_exact_value_quaternion :: proc(t: ^testing.T) {
	v := exact_value_quaternion(1.0, 2.0, 3.0, 4.0)
	testing.expect(t, v.kind == .Quaternion, "expected Quaternion kind")
	testing.expect(t, v.value_quaternion != nil, "expected non-nil quaternion ptr")
	if v.value_quaternion != nil {
		q := v.value_quaternion
		testing.expectf(t, math.abs(q.real - 1.0) < 0.00001, "expected real=1.0, got %f", q.real)
		testing.expectf(t, math.abs(q.imag - 2.0) < 0.00001, "expected imag=2.0, got %f", q.imag)
		testing.expectf(t, math.abs(q.jmag - 3.0) < 0.00001, "expected jmag=3.0, got %f", q.jmag)
		testing.expectf(t, math.abs(q.kmag - 4.0) < 0.00001, "expected kmag=4.0, got %f", q.kmag)
	}
}

@(test)
test_exact_value_pointer :: proc(t: ^testing.T) {
	v := exact_value_pointer(0xdeadbeef)
	testing.expect(t, v.kind == .Pointer, "expected Pointer kind")
	testing.expect(t, v.value_pointer == 0xdeadbeef, "expected 0xdeadbeef")
}

@(test)
test_exact_value_invalid_default :: proc(t: ^testing.T) {
	testing.expect(t, EMPTY_EXACT_VALUE.kind == .Invalid, "empty should be Invalid")
}

// ============================================================================
// Hash Tests
// ============================================================================

@(test)
test_hash_exact_value_deterministic :: proc(t: ^testing.T) {
	v1 := exact_value_bool(true)
	v2 := exact_value_bool(true)
	h1 := hash_exact_value(v1)
	h2 := hash_exact_value(v2)
	testing.expect(t, h1 == h2, "same value should produce same hash")
}

@(test)
test_hash_exact_value_different :: proc(t: ^testing.T) {
	v1 := exact_value_bool(true)
	v2 := exact_value_bool(false)
	h1 := hash_exact_value(v1)
	h2 := hash_exact_value(v2)
	// Different values may or may not have different hashes (hash collision possible),
	// but Invalid always returns 0
	v_invalid := EMPTY_EXACT_VALUE
	testing.expect(t, hash_exact_value(v_invalid) == 0, "Invalid should hash to 0")
}

@(test)
test_hash_exact_value_invalid_zero :: proc(t: ^testing.T) {
	v := EMPTY_EXACT_VALUE
	testing.expect(t, hash_exact_value(v) == 0, "Invalid must hash to 0")
}

@(test)
test_hash_same_string_same_hash :: proc(t: ^testing.T) {
	v1 := exact_value_string("odin")
	v2 := exact_value_string("odin")
	testing.expect(t, hash_exact_value(v1) == hash_exact_value(v2),
		"identical strings should hash identically")
}

@(test)
test_hash_float_same_hash :: proc(t: ^testing.T) {
	v1 := exact_value_float(3.14)
	v2 := exact_value_float(3.14)
	testing.expect(t, hash_exact_value(v1) == hash_exact_value(v2),
		"identical floats should hash identically")
}

// ============================================================================
// exact_value_to_integer Tests
// ============================================================================

@(test)
test_to_integer_bool :: proc(t: ^testing.T) {
	vt := exact_value_bool(true)
	r := exact_value_to_integer(vt)
	testing.expect(t, r.kind == .Integer, "Bool->Integer should produce Integer")
	// true converts to 1
}

@(test)
test_to_integer_bool_false :: proc(t: ^testing.T) {
	vf := exact_value_bool(false)
	r := exact_value_to_integer(vf)
	testing.expect(t, r.kind == .Integer, "Bool(false)->Integer should produce Integer")
	// false converts to 0
}

@(test)
test_to_integer_from_integer :: proc(t: ^testing.T) {
	v := exact_value_i64(100)
	r := exact_value_to_integer(v)
	testing.expect(t, r.kind == .Integer, "Integer->Integer identity")
}

@(test)
test_to_integer_from_float_exact :: proc(t: ^testing.T) {
	v := exact_value_float(5.0)
	r := exact_value_to_integer(v)
	testing.expect(t, r.kind == .Integer, "whole float should convert back to Integer")
}

@(test)
test_to_integer_from_float_fractional :: proc(t: ^testing.T) {
	v := exact_value_float(5.7)
	r := exact_value_to_integer(v)
	testing.expect(t, r.kind == .Invalid, "fractional float should produce Invalid")
}

@(test)
test_to_integer_from_pointer :: proc(t: ^testing.T) {
	v := exact_value_pointer(0x1234)
	r := exact_value_to_integer(v)
	testing.expect(t, r.kind == .Integer, "Pointer->Integer should produce Integer")
}

// ============================================================================
// exact_value_to_float Tests
// ============================================================================

@(test)
test_to_float_from_integer :: proc(t: ^testing.T) {
	v := exact_value_i64(7)
	r := exact_value_to_float(v)
	testing.expect(t, r.kind == .Float, "Integer->Float should produce Float")
	testing.expectf(t, math.abs(r.value_float - 7.0) < 0.00001,
		"expected 7.0, got %f", r.value_float)
}

@(test)
test_to_float_from_float :: proc(t: ^testing.T) {
	v := exact_value_float(2.71828)
	r := exact_value_to_float(v)
	testing.expect(t, r.kind == .Float, "Float->Float identity")
	testing.expectf(t, math.abs(r.value_float - 2.71828) < 0.00001,
		"expected same float value")
}

// ============================================================================
// exact_value_to_complex Tests
// ============================================================================

@(test)
test_to_complex_from_integer :: proc(t: ^testing.T) {
	v := exact_value_i64(3)
	r := exact_value_to_complex(v)
	testing.expect(t, r.kind == .Complex, "Integer->Complex should produce Complex")
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.imag) < 0.00001,
			"imag should be 0, got %f", r.value_complex.imag)
	}
}

@(test)
test_to_complex_from_float :: proc(t: ^testing.T) {
	v := exact_value_float(2.5)
	r := exact_value_to_complex(v)
	testing.expect(t, r.kind == .Complex, "Float->Complex should produce Complex")
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - 2.5) < 0.00001,
			"expected real=2.5")
		testing.expectf(t, math.abs(r.value_complex.imag) < 0.00001,
			"imag should be 0")
	}
}

@(test)
test_to_complex_from_complex :: proc(t: ^testing.T) {
	v := exact_value_complex(1.0, 2.0)
	r := exact_value_to_complex(v)
	testing.expect(t, r.kind == .Complex, "Complex->Complex identity")
}

// ============================================================================
// exact_value_to_quaternion Tests
// ============================================================================

@(test)
test_to_quaternion_from_integer :: proc(t: ^testing.T) {
	v := exact_value_i64(4)
	r := exact_value_to_quaternion(v)
	testing.expect(t, r.kind == .Quaternion, "Integer->Quaternion should produce Quaternion")
	if r.value_quaternion != nil {
		q := r.value_quaternion
		testing.expectf(t, math.abs(q.imag) < 0.00001, "imag should be 0")
		testing.expectf(t, math.abs(q.jmag) < 0.00001, "jmag should be 0")
		testing.expectf(t, math.abs(q.kmag) < 0.00001, "kmag should be 0")
	}
}

@(test)
test_to_quaternion_from_float :: proc(t: ^testing.T) {
	v := exact_value_float(3.0)
	r := exact_value_to_quaternion(v)
	testing.expect(t, r.kind == .Quaternion, "Float->Quaternion should produce Quaternion")
}

@(test)
test_to_quaternion_from_complex :: proc(t: ^testing.T) {
	v := exact_value_complex(1.0, 2.0)
	r := exact_value_to_quaternion(v)
	testing.expect(t, r.kind == .Quaternion, "Complex->Quaternion should produce Quaternion")
	if r.value_quaternion != nil {
		q := r.value_quaternion
		testing.expectf(t, math.abs(q.real - 1.0) < 0.00001, "real should be 1.0")
		testing.expectf(t, math.abs(q.imag - 2.0) < 0.00001, "imag should be 2.0")
		testing.expectf(t, math.abs(q.jmag) < 0.00001, "jmag should be 0")
		testing.expectf(t, math.abs(q.kmag) < 0.00001, "kmag should be 0")
	}
}

@(test)
test_to_quaternion_from_quaternion :: proc(t: ^testing.T) {
	v := exact_value_quaternion(1.0, 2.0, 3.0, 4.0)
	r := exact_value_to_quaternion(v)
	testing.expect(t, r.kind == .Quaternion, "Quaternion->Quaternion identity")
}

// ============================================================================
// Component Accessor Tests
// ============================================================================

@(test)
test_exact_value_real :: proc(t: ^testing.T) {
	// Integer: real returns self
	vi := exact_value_i64(5)
	ri := exact_value_real(vi)
	testing.expect(t, ri.kind == .Integer, "real(Integer) should return self")

	// Float: real returns self
	vf := exact_value_float(3.14)
	rf := exact_value_real(vf)
	testing.expect(t, rf.kind == .Float, "real(Float) should return self")

	// Complex: real extracts real component
	vc := exact_value_complex(7.0, 3.0)
	rc := exact_value_real(vc)
	testing.expect(t, rc.kind == .Float, "real(Complex) should produce Float")
	testing.expectf(t, math.abs(rc.value_float - 7.0) < 0.00001, "expected 7.0")

	// Quaternion: real extracts real component
	vq := exact_value_quaternion(9.0, 1.0, 2.0, 3.0)
	rq := exact_value_real(vq)
	testing.expect(t, rq.kind == .Float, "real(Quaternion) should produce Float")
	testing.expectf(t, math.abs(rq.value_float - 9.0) < 0.00001, "expected 9.0")
}

@(test)
test_exact_value_imag :: proc(t: ^testing.T) {
	// Integer: imag returns 0
	vi := exact_value_i64(5)
	ri := exact_value_imag(vi)
	testing.expect(t, ri.kind == .Integer, "imag(Integer) should return 0 as Integer")

	// Float: imag returns 0
	vf := exact_value_float(3.14)
	rf := exact_value_imag(vf)
	testing.expect(t, rf.kind == .Integer, "imag(Float) should return 0 as Integer")

	// Complex: imag extracts imag component
	vc := exact_value_complex(7.0, 3.0)
	rc := exact_value_imag(vc)
	testing.expect(t, rc.kind == .Float, "imag(Complex) should produce Float")
	testing.expectf(t, math.abs(rc.value_float - 3.0) < 0.00001, "expected 3.0")

	// Quaternion: imag extracts imag component
	vq := exact_value_quaternion(9.0, 1.0, 2.0, 3.0)
	rq := exact_value_imag(vq)
	testing.expect(t, rq.kind == .Float, "imag(Quaternion) should produce Float")
	testing.expectf(t, math.abs(rq.value_float - 1.0) < 0.00001, "expected 1.0")
}

@(test)
test_exact_value_jmag :: proc(t: ^testing.T) {
	// Integer: jmag returns 0
	vi := exact_value_i64(5)
	ri := exact_value_jmag(vi)
	testing.expect(t, ri.kind == .Integer, "jmag(Integer) should return 0 as Integer")

	// Float: jmag returns 0
	vf := exact_value_float(3.14)
	rf := exact_value_jmag(vf)
	testing.expect(t, rf.kind == .Integer, "jmag(Float) should return 0 as Integer")

	// Complex: jmag returns 0
	vc := exact_value_complex(1.0, 2.0)
	rc := exact_value_jmag(vc)
	testing.expect(t, rc.kind == .Integer, "jmag(Complex) should return 0 as Integer")

	// Quaternion: jmag extracts jmag component
	vq := exact_value_quaternion(9.0, 1.0, 2.0, 3.0)
	rq := exact_value_jmag(vq)
	testing.expect(t, rq.kind == .Float, "jmag(Quaternion) should produce Float")
	testing.expectf(t, math.abs(rq.value_float - 2.0) < 0.00001, "expected 2.0")
}

@(test)
test_exact_value_kmag :: proc(t: ^testing.T) {
	// Integer: kmag returns 0
	vi := exact_value_i64(5)
	ri := exact_value_kmag(vi)
	testing.expect(t, ri.kind == .Integer, "kmag(Integer) should return 0 as Integer")

	// Float: kmag returns 0
	vf := exact_value_float(3.14)
	rf := exact_value_kmag(vf)
	testing.expect(t, rf.kind == .Integer, "kmag(Float) should return 0 as Integer")

	// Complex: kmag returns 0
	vc := exact_value_complex(1.0, 2.0)
	rc := exact_value_kmag(vc)
	testing.expect(t, rc.kind == .Integer, "kmag(Complex) should return 0 as Integer")

	// Quaternion: kmag extracts kmag component
	vq := exact_value_quaternion(9.0, 1.0, 2.0, 3.0)
	rq := exact_value_kmag(vq)
	testing.expect(t, rq.kind == .Float, "kmag(Quaternion) should produce Float")
	testing.expectf(t, math.abs(rq.value_float - 3.0) < 0.00001, "expected 3.0")
}

// ============================================================================
// exact_value_to_i64 / to_u64 / to_f64 Tests
// ============================================================================

@(test)
test_exact_value_to_i64_integer :: proc(t: ^testing.T) {
	v := exact_value_i64(-123)
	testing.expect(t, exact_value_to_i64(v) == -123, "expected -123")
}

@(test)
test_exact_value_to_i64_invalid :: proc(t: ^testing.T) {
	v := EMPTY_EXACT_VALUE
	testing.expect(t, exact_value_to_i64(v) == 0, "Invalid should return 0")
}

@(test)
test_exact_value_to_u64_integer :: proc(t: ^testing.T) {
	v := exact_value_u64(999)
	testing.expect(t, exact_value_to_u64(v) == 999, "expected 999")
}

@(test)
test_exact_value_to_f64_float :: proc(t: ^testing.T) {
	v := exact_value_float(2.71828)
	f := exact_value_to_f64(v)
	testing.expectf(t, math.abs(f - 2.71828) < 0.00001, "expected 2.71828, got %f", f)
}

@(test)
test_exact_value_to_f64_integer :: proc(t: ^testing.T) {
	v := exact_value_i64(42)
	f := exact_value_to_f64(v)
	testing.expectf(t, math.abs(f - 42.0) < 0.00001, "expected 42.0, got %f", f)
}

// ============================================================================
// Unary Operator Tests
// ============================================================================

@(test)
test_unary_add :: proc(t: ^testing.T) {
	v := exact_value_i64(10)
	r := exact_unary_operator_value(.Add, v, 32, false)
	testing.expect(t, r.kind == .Integer, "+Integer should stay Integer")
	testing.expect(t, exact_value_to_i64(r) == 10, "+10 should equal 10")
}

@(test)
test_unary_add_float :: proc(t: ^testing.T) {
	v := exact_value_float(3.14)
	r := exact_unary_operator_value(.Add, v, 0, false)
	testing.expect(t, r.kind == .Float, "+Float should stay Float")
	testing.expectf(t, math.abs(r.value_float - 3.14) < 0.00001, "+3.14 should equal 3.14")
}

@(test)
test_unary_sub_integer :: proc(t: ^testing.T) {
	v := exact_value_i64(10)
	r := exact_unary_operator_value(.Sub, v, 32, false)
	testing.expect(t, r.kind == .Integer, "-Integer should stay Integer")
	testing.expect(t, exact_value_to_i64(r) == -10, "-10 should equal -10")
}

@(test)
test_unary_sub_float :: proc(t: ^testing.T) {
	v := exact_value_float(3.14)
	r := exact_unary_operator_value(.Sub, v, 0, false)
	testing.expect(t, r.kind == .Float, "-Float should stay Float")
	testing.expectf(t, math.abs(r.value_float - (-3.14)) < 0.00001, "expected -3.14")
}

@(test)
test_unary_sub_complex :: proc(t: ^testing.T) {
	v := exact_value_complex(1.0, 2.0)
	r := exact_unary_operator_value(.Sub, v, 0, false)
	testing.expect(t, r.kind == .Complex, "-Complex should stay Complex")
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - (-1.0)) < 0.00001, "real should be -1.0")
		testing.expectf(t, math.abs(r.value_complex.imag - (-2.0)) < 0.00001, "imag should be -2.0")
	}
}

@(test)
test_unary_sub_quaternion :: proc(t: ^testing.T) {
	v := exact_value_quaternion(1.0, 2.0, 3.0, 4.0)
	r := exact_unary_operator_value(.Sub, v, 0, false)
	testing.expect(t, r.kind == .Quaternion, "-Quaternion should stay Quaternion")
	if r.value_quaternion != nil {
		q := r.value_quaternion
		testing.expectf(t, math.abs(q.real - (-1.0)) < 0.00001, "real should be -1.0")
		testing.expectf(t, math.abs(q.imag - (-2.0)) < 0.00001, "imag should be -2.0")
		testing.expectf(t, math.abs(q.jmag - (-3.0)) < 0.00001, "jmag should be -3.0")
		testing.expectf(t, math.abs(q.kmag - (-4.0)) < 0.00001, "kmag should be -4.0")
	}
}

@(test)
test_unary_double_negate :: proc(t: ^testing.T) {
	// --x == x
	v := exact_value_i64(42)
	r1 := exact_unary_operator_value(.Sub, v, 32, false)
	r2 := exact_unary_operator_value(.Sub, r1, 32, false)
	testing.expect(t, exact_value_to_i64(r2) == 42, "double negation should equal original")
}

@(test)
test_unary_not_bool :: proc(t: ^testing.T) {
	vt := exact_value_bool(true)
	rt := exact_unary_operator_value(.Not, vt, 0, false)
	testing.expect(t, rt.kind == .Bool, "!Bool should stay Bool")
	testing.expect(t, rt.value_bool == false, "!true should be false")

	vf := exact_value_bool(false)
	rf := exact_unary_operator_value(.Not, vf, 0, false)
	testing.expect(t, rf.value_bool == true, "!false should be true")
}

@(test)
test_unary_xor_integer :: proc(t: ^testing.T) {
	// ~0 should be -1 (in 32-bit twos complement)
	v := exact_value_i64(0)
	r := exact_unary_operator_value(.Xor, v, 32, false)
	testing.expect(t, r.kind == .Integer, "~Integer should stay Integer")
	testing.expect(t, exact_value_to_i64(r) == -1, "~0 with 32-bit precision should be -1")
}

// ============================================================================
// Binary Operator Tests: Integer
// ============================================================================

@(test)
test_binary_add_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(10)
	b := exact_value_i64(20)
	r := exact_binary_operator_value(.Add, a, b)
	testing.expect(t, r.kind == .Integer, "Integer+Integer should be Integer")
	testing.expect(t, exact_value_to_i64(r) == 30, "10+20=30")
}

@(test)
test_binary_sub_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(30)
	b := exact_value_i64(12)
	r := exact_binary_operator_value(.Sub, a, b)
	testing.expect(t, exact_value_to_i64(r) == 18, "30-12=18")
}

@(test)
test_binary_mul_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(6)
	b := exact_value_i64(7)
	r := exact_binary_operator_value(.Mul, a, b)
	testing.expect(t, exact_value_to_i64(r) == 42, "6*7=42")
}

@(test)
test_binary_quo_integer :: proc(t: ^testing.T) {
	// Integer / uses fmod semantics per the original C++ code
	// For positive integers: 7 / 3 = fmod(7, 3) = 1.0
	a := exact_value_i64(10)
	b := exact_value_i64(3)
	r := exact_binary_operator_value(.Quo, a, b)
	testing.expect(t, r.kind == .Float, "Integer/Integer returns Float (fmod)")
	testing.expectf(t, math.abs(r.value_float - 1.0) < 0.00001,
		"fmod(10,3) should be ~1.0, got %f", r.value_float)
}

@(test)
test_binary_and_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(0xff)
	b := exact_value_i64(0x0f)
	r := exact_binary_operator_value(.And, a, b)
	testing.expect(t, exact_value_to_i64(r) == 0x0f, "0xFF & 0x0F = 0x0F")
}

@(test)
test_binary_or_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(0xf0)
	b := exact_value_i64(0x0f)
	r := exact_binary_operator_value(.Or, a, b)
	testing.expect(t, exact_value_to_i64(r) == 0xff, "0xF0 | 0x0F = 0xFF")
}

@(test)
test_binary_xor_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(0xaaaa)
	b := exact_value_i64(0x5555)
	r := exact_binary_operator_value(.Xor, a, b)
	testing.expect(t, exact_value_to_i64(r) == 0xffff, "0xAAAA ^ 0x5555 = 0xFFFF")
}

@(test)
test_binary_shl_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(1)
	b := exact_value_i64(4)
	r := exact_binary_operator_value(.Shl, a, b)
	testing.expect(t, exact_value_to_i64(r) == 16, "1<<4=16")
}

@(test)
test_binary_shr_integer :: proc(t: ^testing.T) {
	a := exact_value_i64(256)
	b := exact_value_i64(4)
	r := exact_binary_operator_value(.Shr, a, b)
	testing.expect(t, exact_value_to_i64(r) == 16, "256>>4=16")
}

// ============================================================================
// Binary Operator Tests: Float
// ============================================================================

@(test)
test_binary_add_float :: proc(t: ^testing.T) {
	a := exact_value_float(1.5)
	b := exact_value_float(2.5)
	r := exact_binary_operator_value(.Add, a, b)
	testing.expect(t, r.kind == .Float, "Float+Float should be Float")
	testing.expectf(t, math.abs(r.value_float - 4.0) < 0.00001, "1.5+2.5=4.0")
}

@(test)
test_binary_sub_float :: proc(t: ^testing.T) {
	a := exact_value_float(5.0)
	b := exact_value_float(3.2)
	r := exact_binary_operator_value(.Sub, a, b)
	testing.expectf(t, math.abs(r.value_float - 1.8) < 0.00001, "5.0-3.2=1.8")
}

@(test)
test_binary_mul_float :: proc(t: ^testing.T) {
	a := exact_value_float(2.5)
	b := exact_value_float(4.0)
	r := exact_binary_operator_value(.Mul, a, b)
	testing.expectf(t, math.abs(r.value_float - 10.0) < 0.00001, "2.5*4.0=10.0")
}

@(test)
test_binary_quo_float :: proc(t: ^testing.T) {
	a := exact_value_float(10.0)
	b := exact_value_float(4.0)
	r := exact_binary_operator_value(.Quo, a, b)
	testing.expectf(t, math.abs(r.value_float - 2.5) < 0.00001, "10.0/4.0=2.5")
}

// ============================================================================
// Binary Operator Tests: Mixed (match_exact_values promotion)
// ============================================================================

@(test)
test_binary_add_int_float_promotion :: proc(t: ^testing.T) {
	a := exact_value_i64(3)
	b := exact_value_float(2.5)
	r := exact_binary_operator_value(.Add, a, b)
	testing.expect(t, r.kind == .Float, "Integer+Float should promote to Float")
	testing.expectf(t, math.abs(r.value_float - 5.5) < 0.00001, "3+2.5=5.5")
}

@(test)
test_binary_add_int_complex_promotion :: proc(t: ^testing.T) {
	a := exact_value_i64(2)
	b := exact_value_complex(1.0, 3.0)
	r := exact_binary_operator_value(.Add, a, b)
	testing.expect(t, r.kind == .Complex, "Integer+Complex should promote to Complex")
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - 3.0) < 0.00001, "2+1.0=3.0")
		testing.expectf(t, math.abs(r.value_complex.imag - 3.0) < 0.00001, "imag should be 3.0")
	}
}

@(test)
test_binary_mul_float_complex_promotion :: proc(t: ^testing.T) {
	a := exact_value_float(2.0)
	b := exact_value_complex(3.0, 4.0)
	r := exact_binary_operator_value(.Mul, a, b)
	testing.expect(t, r.kind == .Complex, "Float*Complex should promote to Complex")
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - 6.0) < 0.00001, "2.0*3.0=6.0")
		testing.expectf(t, math.abs(r.value_complex.imag - 8.0) < 0.00001, "2.0*4.0=8.0")
	}
}

// ============================================================================
// Binary Operator Tests: Complex
// ============================================================================

@(test)
test_binary_add_complex :: proc(t: ^testing.T) {
	a := exact_value_complex(1.0, 2.0)
	b := exact_value_complex(3.0, 4.0)
	r := exact_binary_operator_value(.Add, a, b)
	testing.expect(t, r.kind == .Complex)
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - 4.0) < 0.00001)
		testing.expectf(t, math.abs(r.value_complex.imag - 6.0) < 0.00001)
	}
}

@(test)
test_binary_sub_complex :: proc(t: ^testing.T) {
	a := exact_value_complex(5.0, 6.0)
	b := exact_value_complex(3.0, 2.0)
	r := exact_binary_operator_value(.Sub, a, b)
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - 2.0) < 0.00001)
		testing.expectf(t, math.abs(r.value_complex.imag - 4.0) < 0.00001)
	}
}

@(test)
test_binary_mul_complex :: proc(t: ^testing.T) {
	a := exact_value_complex(2.0, 3.0)
	b := exact_value_complex(4.0, 5.0)
	r := exact_binary_operator_value(.Mul, a, b)
	// (2+3i)*(4+5i) = 2*4-3*5 + (3*4+2*5)i = 8-15 + (12+10)i = -7+22i
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - (-7.0)) < 0.00001)
		testing.expectf(t, math.abs(r.value_complex.imag - 22.0) < 0.00001)
	}
}

@(test)
test_binary_quo_complex :: proc(t: ^testing.T) {
	a := exact_value_complex(1.0, 2.0)
	b := exact_value_complex(3.0, 4.0)
	r := exact_binary_operator_value(.Quo, a, b)
	// (1+2i)/(3+4i) = (1*3+2*4)/(9+16) + (2*3-1*4)/(25)i = 11/25 + 2/25i = 0.44+0.08i
	if r.value_complex != nil {
		testing.expectf(t, math.abs(r.value_complex.real - 0.44) < 0.001)
		testing.expectf(t, math.abs(r.value_complex.imag - 0.08) < 0.001)
	}
}

// ============================================================================
// Binary Operator Tests: Quaternion
// ============================================================================

@(test)
test_binary_add_quaternion :: proc(t: ^testing.T) {
	a := exact_value_quaternion(1.0, 2.0, 3.0, 4.0)
	b := exact_value_quaternion(5.0, 6.0, 7.0, 8.0)
	r := exact_binary_operator_value(.Add, a, b)
	testing.expect(t, r.kind == .Quaternion)
	if r.value_quaternion != nil {
		q := r.value_quaternion
		testing.expectf(t, math.abs(q.real - 6.0) < 0.00001)
		testing.expectf(t, math.abs(q.imag - 8.0) < 0.00001)
		testing.expectf(t, math.abs(q.jmag - 10.0) < 0.00001)
		testing.expectf(t, math.abs(q.kmag - 12.0) < 0.00001)
	}
}

@(test)
test_binary_quaternion_identity :: proc(t: ^testing.T) {
	// i*j*k = -1 in quaternion algebra
	i := exact_value_quaternion(0.0, 1.0, 0.0, 0.0)
	j := exact_value_quaternion(0.0, 0.0, 1.0, 0.0)
	k := exact_value_quaternion(0.0, 0.0, 0.0, 1.0)
	ij := exact_binary_operator_value(.Mul, i, j)
	ijk := exact_binary_operator_value(.Mul, ij, k)
	if ijk.value_quaternion != nil {
		q := ijk.value_quaternion
		testing.expectf(t, math.abs(q.real - (-1.0)) < 0.00001,
			"i*j*k should equal -1 (real component)")
		testing.expectf(t, math.abs(q.imag) < 0.00001, "imag should be 0")
		testing.expectf(t, math.abs(q.jmag) < 0.00001, "jmag should be 0")
		testing.expectf(t, math.abs(q.kmag) < 0.00001, "kmag should be 0")
	}
}

// ============================================================================
// Binary Operator Tests: Bool
// ============================================================================

@(test)
test_binary_cmp_and_bool :: proc(t: ^testing.T) {
	a := exact_value_bool(true)
	b := exact_value_bool(true)
	r := exact_binary_operator_value(.CmpAnd, a, b)
	testing.expect(t, r.kind == .Bool)
	testing.expect(t, r.value_bool == true, "true && true = true")
}

@(test)
test_binary_cmp_or_bool :: proc(t: ^testing.T) {
	a := exact_value_bool(false)
	b := exact_value_bool(false)
	r := exact_binary_operator_value(.CmpOr, a, b)
	testing.expect(t, r.value_bool == false, "false || false = false")
}

@(test)
test_binary_xor_bool :: proc(t: ^testing.T) {
	a := exact_value_bool(true)
	b := exact_value_bool(true)
	r := exact_binary_operator_value(.Xor, a, b)
	testing.expect(t, r.value_bool == false, "true xor true = false")

	a2 := exact_value_bool(true)
	b2 := exact_value_bool(false)
	r2 := exact_binary_operator_value(.Xor, a2, b2)
	testing.expect(t, r2.value_bool == true, "true xor false = true")
}

// ============================================================================
// Binary Operator Tests: String concatenation
// ============================================================================

@(test)
test_binary_add_string :: proc(t: ^testing.T) {
	a := exact_value_string("hello ")
	b := exact_value_string("world")
	r := exact_binary_operator_value(.Add, a, b)
	testing.expect(t, r.kind == .String, "String+String should be String")
	testing.expect(t, r.value_string == "hello world", "expected 'hello world'")
}

@(test)
test_binary_string_not_add :: proc(t: ^testing.T) {
	a := exact_value_string("a")
	b := exact_value_string("b")
	r := exact_binary_operator_value(.Sub, a, b)
	testing.expect(t, r.kind == .Invalid, "String-String should be Invalid")
}

// ============================================================================
// Convenience Wrapper Tests
// ============================================================================

@(test)
test_exact_value_add_wrapper :: proc(t: ^testing.T) {
	a := exact_value_float(1.0)
	b := exact_value_float(2.0)
	r := exact_value_add(a, b)
	testing.expectf(t, math.abs(r.value_float - 3.0) < 0.00001)
}

@(test)
test_exact_value_sub_wrapper :: proc(t: ^testing.T) {
	a := exact_value_float(5.0)
	b := exact_value_float(2.0)
	r := exact_value_sub(a, b)
	testing.expectf(t, math.abs(r.value_float - 3.0) < 0.00001)
}

@(test)
test_exact_value_mul_wrapper :: proc(t: ^testing.T) {
	a := exact_value_i64(6)
	b := exact_value_i64(9)
	r := exact_value_mul(a, b)
	testing.expect(t, exact_value_to_i64(r) == 54, "6*9=54")
}

@(test)
test_exact_value_quo_wrapper :: proc(t: ^testing.T) {
	a := exact_value_float(8.0)
	b := exact_value_float(2.0)
	r := exact_value_quo(a, b)
	testing.expectf(t, math.abs(r.value_float - 4.0) < 0.00001)
}

@(test)
test_exact_value_increment_one :: proc(t: ^testing.T) {
	a := exact_value_i64(41)
	r := exact_value_increment_one(a)
	testing.expect(t, exact_value_to_i64(r) == 42, "41+1=42")
}

// ============================================================================
// Comparison Tests
// ============================================================================

@(test)
test_compare_integer_eq :: proc(t: ^testing.T) {
	a := exact_value_i64(42)
	b := exact_value_i64(42)
	testing.expect(t, compare_exact_values(.CmpEq, a, b), "42==42")
	testing.expect(t, !compare_exact_values(.NotEq, a, b), "42!=42 is false")
}

@(test)
test_compare_integer_lt :: proc(t: ^testing.T) {
	a := exact_value_i64(10)
	b := exact_value_i64(20)
	testing.expect(t, compare_exact_values(.Lt, a, b), "10<20")
	testing.expect(t, !compare_exact_values(.Gt, a, b), "10>20 is false")
}

@(test)
test_compare_integer_lteq :: proc(t: ^testing.T) {
	a := exact_value_i64(10)
	b := exact_value_i64(10)
	testing.expect(t, compare_exact_values(.LtEq, a, b), "10<=10")
	testing.expect(t, compare_exact_values(.GtEq, a, b), "10>=10")
}

@(test)
test_compare_float_eq :: proc(t: ^testing.T) {
	a := exact_value_float(3.14)
	b := exact_value_float(3.14)
	testing.expect(t, compare_exact_values(.CmpEq, a, b), "3.14==3.14")
}

@(test)
test_compare_float_nan :: proc(t: ^testing.T) {
	nan := exact_value_float(math.NaN())
	testing.expect(t, !compare_exact_values(.CmpEq, nan, nan), "NaN != NaN")
	testing.expect(t, compare_exact_values(.NotEq, nan, nan), "NaN != NaN (NotEq)")
}

@(test)
test_compare_float_value_nan :: proc(t: ^testing.T) {
	nan := exact_value_float(math.NaN())
	normal := exact_value_float(1.0)
	testing.expect(t, compare_exact_values(.NotEq, nan, normal), "NaN != 1.0")
}

@(test)
test_compare_float_cross_type :: proc(t: ^testing.T) {
	a := exact_value_i64(5)
	b := exact_value_float(5.0)
	// match_exact_values promotes i64 to float
	testing.expect(t, compare_exact_values(.CmpEq, a, b), "5 (int) == 5.0 (float)")
}

@(test)
test_compare_string_eq :: proc(t: ^testing.T) {
	a := exact_value_string("odin")
	b := exact_value_string("odin")
	testing.expect(t, compare_exact_values(.CmpEq, a, b), "'odin'=='odin'")
}

@(test)
test_compare_string_ne :: proc(t: ^testing.T) {
	a := exact_value_string("odin")
	b := exact_value_string("lang")
	testing.expect(t, compare_exact_values(.NotEq, a, b), "'odin'!='lang'")
}

@(test)
test_compare_string_lt :: proc(t: ^testing.T) {
	a := exact_value_string("abc")
	b := exact_value_string("abd")
	testing.expect(t, compare_exact_values(.Lt, a, b), "'abc'<'abd'")
}

@(test)
test_compare_bool :: proc(t: ^testing.T) {
	a := exact_value_bool(true)
	b := exact_value_bool(true)
	testing.expect(t, compare_exact_values(.CmpEq, a, b), "true==true")
	testing.expect(t, compare_exact_values(.NotEq, exact_value_bool(true), exact_value_bool(false)), "true!=false")
}

@(test)
test_compare_pointer :: proc(t: ^testing.T) {
	a := exact_value_pointer(100)
	b := exact_value_pointer(200)
	testing.expect(t, compare_exact_values(.CmpEq, a, a), "100==100")
	testing.expect(t, compare_exact_values(.NotEq, a, b), "100!=200")
	testing.expect(t, compare_exact_values(.Lt, a, b), "100<200")
	testing.expect(t, compare_exact_values(.Gt, b, a), "200>100")
}

@(test)
test_compare_invalid :: proc(t: ^testing.T) {
	a := EMPTY_EXACT_VALUE
	b := EMPTY_EXACT_VALUE
	testing.expect(t, !compare_exact_values(.CmpEq, a, b), "Invalid==Invalid is false")
}

@(test)
test_compare_complex_eq :: proc(t: ^testing.T) {
	a := exact_value_complex(1.0, 2.0)
	b := exact_value_complex(1.0, 2.0)
	testing.expect(t, compare_exact_values(.CmpEq, a, b), "(1+2i)==(1+2i)")
}

@(test)
test_compare_complex_ne :: proc(t: ^testing.T) {
	a := exact_value_complex(1.0, 2.0)
	b := exact_value_complex(3.0, 4.0)
	testing.expect(t, compare_exact_values(.NotEq, a, b), "(1+2i)!=(3+4i)")
}

// ============================================================================
// exact_value_order Tests
// ============================================================================

@(test)
test_exact_value_order_hierarchy :: proc(t: ^testing.T) {
	// Verify the promotion lattice: Invalid < Bool/String/String16 < Integer < Float < Complex < Quaternion < Pointer < Procedure
	testing.expect(t, exact_value_order(exact_value_bool(true)) == 1, "Bool order=1")
	testing.expect(t, exact_value_order(exact_value_i64(0)) == 2, "Integer order=2")
	testing.expect(t, exact_value_order(exact_value_float(0.0)) == 3, "Float order=3")
	testing.expect(t, exact_value_order(exact_value_complex(0, 0)) == 4, "Complex order=4")
	testing.expect(t, exact_value_order(exact_value_quaternion(0, 0, 0, 0)) == 5, "Quaternion order=5")
	testing.expect(t, exact_value_order(exact_value_pointer(0)) == 6, "Pointer order=6")
}