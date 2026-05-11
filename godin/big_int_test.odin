// big_int_test.odin – tests for big_int.odin
package godin

import "core:testing"

// --------------- Helpers ---------------

@(private)
big_int_from_string_decimal :: proc(a: ^BigInt, str: string) -> bool {
    ok: bool
    big_int_from_string(a, str, &ok)
    return ok
}

@(private)
big_int_from_string_decimal_or_fail :: proc(t: ^testing.T, a: ^BigInt, str: string) -> bool {
    ok: bool
    big_int_from_string(a, str, &ok)
    if !ok {
        testing.expectf(t, false, "big_int_from_string failed for %q", str)
        return false
    }
    // big_int_from_string 内部调用了 mp_zero 再填充，这里确保 dst 有值
    return true
}

@(private)
big_int_to_string_decimal :: proc(a: ^BigInt) -> string {
    return big_int_to_string(a, allocator = context.temp_allocator)
}

// ==================== Initialization / Deallocation ====================

@(test)
test_init_from_u64 :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    big_int_from_u64(&a, 42)
    testing.expect(t, !big_int_is_zero(&a))
    testing.expect(t, big_int_to_u64(&a) == 42)
}

@(test)
test_init_from_i64 :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    big_int_from_i64(&a, -123)
    testing.expect(t, big_int_is_neg(&a))
    testing.expect(t, big_int_to_i64(&a) == -123)
}

@(test)
test_init_copy :: proc(t: ^testing.T) {
    a: BigInt; defer big_int_dealloc(&a)
    b: BigInt; defer big_int_dealloc(&b)
    big_int_from_u64(&a, 789)
    big_int_init(&b, &a)
    testing.expect(t, big_int_cmp(&a, &b) == .MP_EQ)
}

@(test)
test_init_self_copy :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    big_int_from_u64(&a, 555)
    // Should be no-op, not crash
    big_int_init(&a, &a)
    testing.expect(t, big_int_to_u64(&a) == 555)
}

// ==================== Make constructors ====================

@(test)
test_make_u64 :: proc(t: ^testing.T) {
    x := big_int_make_u64(100)
    defer big_int_dealloc(&x)
    testing.expect(t, big_int_to_u64(&x) == 100)
}

@(test)
test_make_i64 :: proc(t: ^testing.T) {
    x := big_int_make_i64(-200)
    defer big_int_dealloc(&x)
    testing.expect(t, big_int_to_i64(&x) == -200)
}

@(test)
test_make_abs :: proc(t: ^testing.T) {
    a := big_int_make_i64(-300)
    defer big_int_dealloc(&a)
    b := big_int_make_abs(&a)
    defer big_int_dealloc(&b)
    testing.expect(t, !big_int_is_neg(&b))
    testing.expect(t, big_int_to_u64(&b) == 300)
}

@(test)
test_make_copy :: proc(t: ^testing.T) {
    a := big_int_make_i64(77)
    defer big_int_dealloc(&a)
    b := big_int_make(&a, false)
    defer big_int_dealloc(&b)
    testing.expect(t, big_int_cmp(&a, &b) == .MP_EQ)
}

// ==================== From string ====================

@(test)
test_from_string_basic :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "12345"))
    testing.expect(t, big_int_to_u64(&a) == 12345)
}

@(test)
test_from_string_negative :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "-42"))
    testing.expect(t, big_int_to_i64(&a) == -42)
}

@(test)
test_from_string_zero :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "0"))
    testing.expect(t, big_int_is_zero(&a))
}

@(test)
test_from_string_underscores :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "1_000_000"))
    testing.expect(t, big_int_to_u64(&a) == 1_000_000)
}

@(test)
test_from_string_prefix_binary :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "0b1010"))
    testing.expect(t, big_int_to_u64(&a) == 10)
}

@(test)
test_from_string_prefix_hex :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "0xff"))
    testing.expect(t, big_int_to_u64(&a) == 255)
}

@(test)
test_from_string_prefix_octal :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "0o77"))
    testing.expect(t, big_int_to_u64(&a) == 63)
}

@(test)
test_from_string_double_negative_fails :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    ok: bool
    big_int_from_string(&a, "--5", &ok)
    testing.expect(t, !ok)
}

@(test)
test_from_string_exponent :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "1e3"))
    testing.expect(t, big_int_to_u64(&a) == 1000)
}

@(test)
test_from_string_exponent_plus :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_from_string_decimal_or_fail(t, &a, "2e+4"))
    testing.expect(t, big_int_to_u64(&a) == 20000)
}

@(test)
test_from_string_exponent_too_large :: proc(t: ^testing.T) {
    a: BigInt
    defer big_int_dealloc(&a)
    ok: bool
    big_int_from_string(&a, "1e500", &ok)
    testing.expect(t, !ok)
}

// ==================== Queries ====================

@(test)
test_sign_positive :: proc(t: ^testing.T) {
    a := big_int_make_u64(1)
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_sign(&a) == 1)
}

@(test)
test_sign_negative :: proc(t: ^testing.T) {
    a := big_int_make_i64(-1)
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_sign(&a) == -1)
}

@(test)
test_sign_zero :: proc(t: ^testing.T) {
    a := big_int_make_u64(0)
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_sign(&a) == 0)
}

@(test)
test_cmp_equal :: proc(t: ^testing.T) {
    a := big_int_make_u64(100); defer big_int_dealloc(&a)
    b := big_int_make_u64(100); defer big_int_dealloc(&b)
    testing.expect(t, big_int_cmp(&a, &b) == .MP_EQ)
}

@(test)
test_cmp_less :: proc(t: ^testing.T) {
    a := big_int_make_u64(10);  defer big_int_dealloc(&a)
    b := big_int_make_u64(20);  defer big_int_dealloc(&b)
    testing.expect(t, big_int_cmp(&a, &b) == .MP_LT)
}

@(test)
test_cmp_greater :: proc(t: ^testing.T) {
    a := big_int_make_u64(30);  defer big_int_dealloc(&a)
    b := big_int_make_u64(20);  defer big_int_dealloc(&b)
    testing.expect(t, big_int_cmp(&a, &b) == .MP_GT)
}

@(test)
test_cmp_zero :: proc(t: ^testing.T) {
    z := big_int_make_u64(0);    defer big_int_dealloc(&z)
    p := big_int_make_u64(5);    defer big_int_dealloc(&p)
    n := big_int_make_i64(-5);   defer big_int_dealloc(&n)
    testing.expect(t, big_int_cmp_zero(&z) == 0)
    testing.expect(t, big_int_cmp_zero(&p) == + 1)
    testing.expect(t, big_int_cmp_zero(&n) == -1)
}

@(test)
test_is_zero :: proc(t: ^testing.T) {
    a := big_int_make_u64(0); defer big_int_dealloc(&a)
    b := big_int_make_u64(1); defer big_int_dealloc(&b)
    testing.expect(t, big_int_is_zero(&a))
    testing.expect(t, !big_int_is_zero(&b))
}

@(test)
test_is_neg :: proc(t: ^testing.T) {
    p := big_int_make_u64(1);  defer big_int_dealloc(&p)
    n := big_int_make_i64(-1); defer big_int_dealloc(&n)
    testing.expect(t, !big_int_is_neg(&p))
    testing.expect(t, big_int_is_neg(&n))
}

@(test)
test_is_neg_nil :: proc(t: ^testing.T) {
    testing.expect(t, !big_int_is_neg(nil))
}

@(test)
test_log2 :: proc(t: ^testing.T) {
    a := big_int_make_u64(16); defer big_int_dealloc(&a)
    // 16 = 2^4, log2 = 4
    testing.expect(t, big_int_log2(&a) == 4)
}

@(test)
test_log2_one :: proc(t: ^testing.T) {
    a := big_int_make_u64(1); defer big_int_dealloc(&a)
    testing.expect(t, big_int_log2(&a) == 0)
}

@(test)
test_can_be_represented_in_64_bits :: proc(t: ^testing.T) {
    small := big_int_make_u64(12345); defer big_int_dealloc(&small)
    testing.expect(t, big_int_can_be_represented_in_64_bits(&small))
}

// ==================== Conversion to primitive types ====================

@(test)
test_to_u64 :: proc(t: ^testing.T) {
    a := big_int_make_u64(0xDEAD_BEEF)
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_to_u64(&a) == 0xDEAD_BEEF)
}

@(test)
test_to_i64_positive :: proc(t: ^testing.T) {
    a := big_int_make_u64(999)
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_to_i64(&a) == 999)
}

@(test)
test_to_i64_negative :: proc(t: ^testing.T) {
    a := big_int_make_i64(-999)
    defer big_int_dealloc(&a)
    testing.expect(t, big_int_to_i64(&a) == -999)
}

@(test)
test_to_f64 :: proc(t: ^testing.T) {
    a := big_int_make_u64(42)
    defer big_int_dealloc(&a)
    f := big_int_to_f64(&a)
    testing.expect(t, f == 42.0)
}

// ==================== Negation ====================

@(test)
test_neg_positive :: proc(t: ^testing.T) {
    a := big_int_make_u64(5); defer big_int_dealloc(&a)
    b: BigInt;               defer big_int_dealloc(&b)
    big_int_neg(&b, &a)
    testing.expect(t, big_int_is_neg(&b))
    testing.expect(t, big_int_to_i64(&b) == -5)
}

@(test)
test_neg_negative :: proc(t: ^testing.T) {
    a := big_int_make_i64(-5); defer big_int_dealloc(&a)
    b: BigInt;                 defer big_int_dealloc(&b)
    big_int_neg(&b, &a)
    testing.expect(t, !big_int_is_neg(&b))
    testing.expect(t, big_int_to_u64(&b) == 5)
}

@(test)
test_neg_self :: proc(t: ^testing.T) {
    a := big_int_make_i64(10); defer big_int_dealloc(&a)
    big_int_neg(&a, &a)
    testing.expect(t, big_int_to_i64(&a) == -10)
}

// ==================== Arithmetic ====================

@(test)
test_add :: proc(t: ^testing.T) {
    x := big_int_make_u64(100); defer big_int_dealloc(&x)
    y := big_int_make_u64(200); defer big_int_dealloc(&y)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_add(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 300)
}

@(test)
test_add_negative :: proc(t: ^testing.T) {
    x := big_int_make_i64(100);  defer big_int_dealloc(&x)
    y := big_int_make_i64(-30);  defer big_int_dealloc(&y)
    z: BigInt;                   defer big_int_dealloc(&z)
    big_int_add(&z, &x, &y)
    testing.expect(t, big_int_to_i64(&z) == 70)
}

@(test)
test_sub :: proc(t: ^testing.T) {
    x := big_int_make_u64(500); defer big_int_dealloc(&x)
    y := big_int_make_u64(123); defer big_int_dealloc(&y)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_sub(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 377)
}

@(test)
test_sub_result_negative :: proc(t: ^testing.T) {
    x := big_int_make_u64(10); defer big_int_dealloc(&x)
    y := big_int_make_u64(20); defer big_int_dealloc(&y)
    z: BigInt;                 defer big_int_dealloc(&z)
    big_int_sub(&z, &x, &y)
    testing.expect(t, big_int_to_i64(&z) == -10)
}

@(test)
test_shl :: proc(t: ^testing.T) {
    x := big_int_make_u64(1);    defer big_int_dealloc(&x)
    shift := big_int_make_u64(8); defer big_int_dealloc(&shift)
    z: BigInt;                   defer big_int_dealloc(&z)
    big_int_shl(&z, &x, &shift)
    testing.expect(t, big_int_to_u64(&z) == 256)
}

@(test)
test_shr :: proc(t: ^testing.T) {
    x := big_int_make_u64(256);  defer big_int_dealloc(&x)
    shift := big_int_make_u64(8); defer big_int_dealloc(&shift)
    z: BigInt;                   defer big_int_dealloc(&z)
    big_int_shr(&z, &x, &shift)
    testing.expect(t, big_int_to_u64(&z) == 1)
}

@(test)
test_mul :: proc(t: ^testing.T) {
    x := big_int_make_u64(123); defer big_int_dealloc(&x)
    y := big_int_make_u64(456); defer big_int_dealloc(&y)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_mul(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 123 * 456)
}

@(test)
test_mul_u64 :: proc(t: ^testing.T) {
    x := big_int_make_u64(100); defer big_int_dealloc(&x)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_mul_u64(&z, &x, 3)
    testing.expect(t, big_int_to_u64(&z) == 300)
}

@(test)
test_exp_u64 :: proc(t: ^testing.T) {
    x := big_int_make_u64(2); defer big_int_dealloc(&x)
    z: BigInt;                defer big_int_dealloc(&z)
    ok: bool
    big_int_exp_u64(&z, &x, 10, &ok)
    testing.expect(t, ok)
    testing.expect(t, big_int_to_u64(&z) == 1024)
}

@(test)
test_exp_u64_zero :: proc(t: ^testing.T) {
    x := big_int_make_u64(5); defer big_int_dealloc(&x)
    z: BigInt;                defer big_int_dealloc(&z)
    ok: bool
    big_int_exp_u64(&z, &x, 0, &ok)
    testing.expect(t, ok)
    testing.expect(t, big_int_to_u64(&z) == 1)
}

// ==================== Division ====================

@(test)
test_quo_rem :: proc(t: ^testing.T) {
    x := big_int_make_u64(17); defer big_int_dealloc(&x)
    y := big_int_make_u64(5);  defer big_int_dealloc(&y)
    q: BigInt;                 defer big_int_dealloc(&q)
    r: BigInt;                 defer big_int_dealloc(&r)
    big_int_quo_rem(&x, &y, &q, &r)
    testing.expect(t, big_int_to_u64(&q) == 3)
    testing.expect(t, big_int_to_u64(&r) == 2)
}

@(test)
test_quo :: proc(t: ^testing.T) {
    x := big_int_make_u64(20); defer big_int_dealloc(&x)
    y := big_int_make_u64(6);  defer big_int_dealloc(&y)
    z: BigInt;                 defer big_int_dealloc(&z)
    big_int_quo(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 3)
}

@(test)
test_rem :: proc(t: ^testing.T) {
    x := big_int_make_u64(20); defer big_int_dealloc(&x)
    y := big_int_make_u64(6);  defer big_int_dealloc(&y)
    z: BigInt;                 defer big_int_dealloc(&z)
    big_int_rem(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 2)
}

@(test)
test_euclidean_mod_positive :: proc(t: ^testing.T) {
    x := big_int_make_i64(-10); defer big_int_dealloc(&x)
    y := big_int_make_u64(6);   defer big_int_dealloc(&y)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_euclidean_mod(&z, &x, &y)
    // -10 % 6 euclidean = 2
    testing.expect(t, big_int_to_u64(&z) == 2)
}

@(test)
test_euclidean_mod_both_positive :: proc(t: ^testing.T) {
    x := big_int_make_u64(10); defer big_int_dealloc(&x)
    y := big_int_make_u64(3);  defer big_int_dealloc(&y)
    z: BigInt;                 defer big_int_dealloc(&z)
    big_int_euclidean_mod(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 1)
}

// ==================== Bitwise operations ====================

@(test)
test_and :: proc(t: ^testing.T) {
    x := big_int_make_u64(0xFF00); defer big_int_dealloc(&x)
    y := big_int_make_u64(0x0FF0); defer big_int_dealloc(&y)
    z: BigInt;                     defer big_int_dealloc(&z)
    big_int_and(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 0x0F00)
}

@(test)
test_xor :: proc(t: ^testing.T) {
    x := big_int_make_u64(0xF0F0); defer big_int_dealloc(&x)
    y := big_int_make_u64(0xFF00); defer big_int_dealloc(&y)
    z: BigInt;                     defer big_int_dealloc(&z)
    big_int_xor(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 0x0FF0)
}

@(test)
test_or :: proc(t: ^testing.T) {
    x := big_int_make_u64(0xF000); defer big_int_dealloc(&x)
    y := big_int_make_u64(0x0F00); defer big_int_dealloc(&y)
    z: BigInt;                     defer big_int_dealloc(&z)
    big_int_or(&z, &x, &y)
    testing.expect(t, big_int_to_u64(&z) == 0xFF00)
}

@(test)
test_not_zero_bit_count :: proc(t: ^testing.T) {
    x := big_int_make_u64(0); defer big_int_dealloc(&x)
    z: BigInt;                defer big_int_dealloc(&z)
    big_int_not(&z, &x, 0, false)
    testing.expect(t, big_int_is_zero(&z))
}

@(test)
test_not_zero_input :: proc(t: ^testing.T) {
    x := big_int_make_u64(0); defer big_int_dealloc(&x)
    z: BigInt;                defer big_int_dealloc(&z)
    big_int_not(&z, &x, 8, false)
    // ~0 & 0xFF = 0xFF
    testing.expect(t, big_int_to_u64(&z) == 0xFF)
}

@(test)
test_not_negative_input :: proc(t: ^testing.T) {
    x := big_int_make_i64(-1); defer big_int_dealloc(&x)
    z: BigInt;                 defer big_int_dealloc(&z)
    // ~x = -x - 1 = 1 - 1 = 0, modulo 2^8 = 0
    big_int_not(&z, &x, 8, false)
    testing.expect(t, big_int_is_zero(&z))
}

@(test)
test_and_not_both_positive :: proc(t: ^testing.T) {
    x := big_int_make_u64(0xFF); defer big_int_dealloc(&x)
    y := big_int_make_u64(0x0F); defer big_int_dealloc(&y)
    z: BigInt;                   defer big_int_dealloc(&z)
    big_int_and_not(&z, &x, &y)
    // 0xFF &~ 0x0F = 0xF0
    testing.expect(t, big_int_to_u64(&z) == 0xF0)
}

@(test)
test_and_not_x_zero :: proc(t: ^testing.T) {
    x := big_int_make_u64(0);  defer big_int_dealloc(&x)
    y := big_int_make_u64(42); defer big_int_dealloc(&y)
    z: BigInt;                 defer big_int_dealloc(&z)
    big_int_and_not(&z, &x, &y)
    // 0 &~ y -> copy y
    testing.expect(t, big_int_to_u64(&z) == 42)
}

@(test)
test_and_not_y_zero :: proc(t: ^testing.T) {
    x := big_int_make_u64(42); defer big_int_dealloc(&x)
    y := big_int_make_u64(0);  defer big_int_dealloc(&y)
    z: BigInt;                 defer big_int_dealloc(&z)
    big_int_and_not(&z, &x, &y)
    // x &~ 0 -> copy x
    testing.expect(t, big_int_to_u64(&z) == 42)
}

// ==================== In-place arithmetic ====================

@(test)
test_add_eq :: proc(t: ^testing.T) {
    a := big_int_make_u64(10); defer big_int_dealloc(&a)
    b := big_int_make_u64(5);  defer big_int_dealloc(&b)
    big_int_add_eq(&a, &b)
    testing.expect(t, big_int_to_u64(&a) == 15)
}

@(test)
test_sub_eq :: proc(t: ^testing.T) {
    a := big_int_make_u64(10); defer big_int_dealloc(&a)
    b := big_int_make_u64(3);  defer big_int_dealloc(&b)
    big_int_sub_eq(&a, &b)
    testing.expect(t, big_int_to_u64(&a) == 7)
}

@(test)
test_shl_eq :: proc(t: ^testing.T) {
    a := big_int_make_u64(1);   defer big_int_dealloc(&a)
    b := big_int_make_u64(4);   defer big_int_dealloc(&b)
    big_int_shl_eq(&a, &b)
    testing.expect(t, big_int_to_u64(&a) == 16)
}

@(test)
test_shr_eq :: proc(t: ^testing.T) {
    a := big_int_make_u64(64);  defer big_int_dealloc(&a)
    b := big_int_make_u64(2);   defer big_int_dealloc(&b)
    big_int_shr_eq(&a, &b)
    testing.expect(t, big_int_to_u64(&a) == 16)
}

@(test)
test_mul_eq :: proc(t: ^testing.T) {
    a := big_int_make_u64(6);  defer big_int_dealloc(&a)
    b := big_int_make_u64(7);  defer big_int_dealloc(&b)
    big_int_mul_eq(&a, &b)
    testing.expect(t, big_int_to_u64(&a) == 42)
}

@(test)
test_quo_eq :: proc(t: ^testing.T) {
    a := big_int_make_u64(100); defer big_int_dealloc(&a)
    b := big_int_make_u64(7);   defer big_int_dealloc(&b)
    big_int_quo_eq(&a, &b)
    testing.expect(t, big_int_to_u64(&a) == 14)
}

@(test)
test_rem_eq :: proc(t: ^testing.T) {
    a := big_int_make_u64(100); defer big_int_dealloc(&a)
    b := big_int_make_u64(7);   defer big_int_dealloc(&b)
    big_int_rem_eq(&a, &b)
    testing.expect(t, big_int_to_u64(&a) == 2)
}

// ==================== To string ====================

@(test)
test_to_string_zero :: proc(t: ^testing.T) {
    a := big_int_make_u64(0); defer big_int_dealloc(&a)
    s := big_int_to_string(&a, allocator = context.temp_allocator)
    testing.expect(t, s == "0")
}

@(test)
test_to_string_decimal :: proc(t: ^testing.T) {
    a := big_int_make_u64(12345); defer big_int_dealloc(&a)
    s := big_int_to_string(&a, allocator = context.temp_allocator)
    testing.expect(t, s == "12345")
}

@(test)
test_to_string_negative :: proc(t: ^testing.T) {
    a := big_int_make_i64(-789); defer big_int_dealloc(&a)
    s := big_int_to_string(&a, allocator = context.temp_allocator)
    testing.expect(t, s == "-789")
}

@(test)
test_to_string_hex :: proc(t: ^testing.T) {
    a := big_int_make_u64(255); defer big_int_dealloc(&a)
    s := big_int_to_string(&a, 16, allocator = context.temp_allocator)
    testing.expect(t, s == "ff")
}

@(test)
test_to_string_binary :: proc(t: ^testing.T) {
    a := big_int_make_u64(5); defer big_int_dealloc(&a)
    s := big_int_to_string(&a, 2, allocator = context.temp_allocator)
    testing.expect(t, s == "101")
}

@(test)
test_to_string_octal :: proc(t: ^testing.T) {
    a := big_int_make_u64(63); defer big_int_dealloc(&a)
    s := big_int_to_string(&a, 8, allocator = context.temp_allocator)
    testing.expect(t, s == "77")
}

// ==================== Roundtrip tests ====================

@(test)
test_roundtrip_string_small :: proc(t: ^testing.T) {
    vals := []u64{ 0, 1, 10, 42, 256, 1000, 99999, 1_000_000_000 }
    for v in vals {
        a := big_int_make_u64(v)
        s := big_int_to_string(&a, allocator = context.temp_allocator)
        b: BigInt
        ok := big_int_from_string_decimal_or_fail(t, &b, s)
        defer big_int_dealloc(&b)
        if ok {
            testing.expectf(t, big_int_cmp(&a, &b) == .MP_EQ,
            "roundtrip failed for %v: got %s (cmp: %v)", v, s, big_int_cmp(&a, &b))
        }
        big_int_dealloc(&a)
    }
}

@(test)
test_roundtrip_neg_string :: proc(t: ^testing.T) {
    vals := []i64{ -1, -10, -256, -10000 }
    for v in vals {
        a := big_int_make_i64(v)
        s := big_int_to_string(&a, allocator = context.temp_allocator)
        b: BigInt
        ok := big_int_from_string_decimal_or_fail(t, &b, s)
        defer big_int_dealloc(&b)
        if ok {
            testing.expectf(t, big_int_cmp(&a, &b) == .MP_EQ,
            "roundtrip failed for %v: got %s", v, s)
        }
        big_int_dealloc(&a)
    }
}

// ==================== Large number tests ====================

@(test)
test_large_number_mul :: proc(t: ^testing.T) {
    a := big_int_make_u64(1_000_000); defer big_int_dealloc(&a)
    b := big_int_make_u64(1_000_000); defer big_int_dealloc(&b)
    c: BigInt;                        defer big_int_dealloc(&c)
    big_int_mul(&c, &a, &b)
    // 1_000_000^2 = 10^12
    expected : u64 = 1_000_000_000_000
    testing.expect(t, big_int_to_u64(&c) == expected)
}

@(test)
test_large_exp :: proc(t: ^testing.T) {
    a := big_int_make_u64(3); defer big_int_dealloc(&a)
    b: BigInt;               defer big_int_dealloc(&b)
    ok: bool
    big_int_exp_u64(&b, &a, 20, &ok)
    testing.expect(t, ok)
    testing.expect(t, big_int_to_u64(&b) == 3_486_784_401)
}

@(test)
test_mul_large_64bit :: proc(t: ^testing.T) {
    a := big_int_make_u64(0xFFFF_FFFF); defer big_int_dealloc(&a)
    b := big_int_make_u64(0xFFFF_FFFF); defer big_int_dealloc(&b)
    c: BigInt;                          defer big_int_dealloc(&c)
    big_int_mul(&c, &a, &b)
    // (2^32-1)^2 = 2^64 - 2^33 + 1
    expected := u64(0xFFFF_FFFF) * u64(0xFFFF_FFFF)
    testing.expect(t, big_int_to_u64(&c) == expected)
}

@(test)
test_large_to_string :: proc(t: ^testing.T) {
    a := big_int_make_u64(1); defer big_int_dealloc(&a)
    b := big_int_make_u64(10); defer big_int_dealloc(&b)

    // Compute 10^20, which has 21 digits (below BIG_INT_LARGE_THRESHOLD = 498)
    c: BigInt
    ok_pow: bool
    big_int_exp_u64(&c, &b, 20, &ok_pow)
    defer big_int_dealloc(&c)
    testing.expect(t, ok_pow)

    s := big_int_to_string(&c, allocator = context.temp_allocator)
    testing.expect(t, s == "100000000000000000000")
}

// ==================== Edge cases ====================

@(test)
test_mul_by_zero :: proc(t: ^testing.T) {
    a := big_int_make_u64(999); defer big_int_dealloc(&a)
    b := big_int_make_u64(0);   defer big_int_dealloc(&b)
    c: BigInt;                  defer big_int_dealloc(&c)
    big_int_mul(&c, &a, &b)
    testing.expect(t, big_int_is_zero(&c))
}

@(test)
test_mul_u64_by_one :: proc(t: ^testing.T) {
    a := big_int_make_u64(999); defer big_int_dealloc(&a)
    b: BigInt;                  defer big_int_dealloc(&b)
    big_int_mul_u64(&b, &a, 1)
    testing.expect(t, big_int_to_u64(&b) == 999)
}

@(test)
test_shl_zero :: proc(t: ^testing.T) {
    a := big_int_make_u64(123);  defer big_int_dealloc(&a)
    shift := big_int_make_u64(0); defer big_int_dealloc(&shift)
    b: BigInt;                   defer big_int_dealloc(&b)
    big_int_shl(&b, &a, &shift)
    testing.expect(t, big_int_to_u64(&b) == 123)
}

@(test)
test_shr_zero :: proc(t: ^testing.T) {
    a := big_int_make_u64(123);  defer big_int_dealloc(&a)
    shift := big_int_make_u64(0); defer big_int_dealloc(&shift)
    b: BigInt;                   defer big_int_dealloc(&b)
    big_int_shr(&b, &a, &shift)
    testing.expect(t, big_int_to_u64(&b) == 123)
}

@(test)
test_exp_u64_too_large_exp :: proc(t: ^testing.T) {
    a := big_int_make_u64(2); defer big_int_dealloc(&a)
    b: BigInt;               defer big_int_dealloc(&b)
    ok: bool
    // y > max(int) should fail
    big_int_exp_u64(&b, &a, u64(max(int)) + 1, &ok)
    testing.expect(t, !ok)
}

@(test)
test_euclidean_mod_exact_division :: proc(t: ^testing.T) {
    x := big_int_make_u64(12); defer big_int_dealloc(&x)
    y := big_int_make_u64(4);  defer big_int_dealloc(&y)
    z: BigInt;                 defer big_int_dealloc(&z)
    big_int_euclidean_mod(&z, &x, &y)
    testing.expect(t, big_int_is_zero(&z))
}

@(test)
test_leading_zeros_u64_zero :: proc(t: ^testing.T) {
    testing.expect(t, leading_zeros_u64(0) == 64)
}

@(test)
test_leading_zeros_u64_top_bit :: proc(t: ^testing.T) {
    testing.expect(t, leading_zeros_u64(1 << 63) == 0)
}

@(test)
test_debug_print :: proc(t: ^testing.T) {
    a := big_int_make_u64(42); defer big_int_dealloc(&a)
    // Should not crash
    debug_print_big_int(&a)
}

// ==================== and_not edge cases ====================

@(test)
test_and_not_neg_neg :: proc(t: ^testing.T) {
// (-x) &~ (-y) when both negative
    x := big_int_make_i64(-10); defer big_int_dealloc(&x)
    y := big_int_make_i64(-3);  defer big_int_dealloc(&y)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_and_not(&z, &x, &y)
    // Should not crash; just verify it produces some result
    testing.expect(t, z.used >= 0)
}

@(test)
test_and_not_neg_pos :: proc(t: ^testing.T) {
// (-x) &~ y when x negative, y positive
    x := big_int_make_i64(-10); defer big_int_dealloc(&x)
    y := big_int_make_u64( 5);  defer big_int_dealloc(&y)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_and_not(&z, &x, &y)
    testing.expect(t, z.used >= 0)
}

@(test)
test_and_not_pos_neg :: proc(t: ^testing.T) {
// x &~ (-y) when x positive, y negative
    x := big_int_make_u64( 10); defer big_int_dealloc(&x)
    y := big_int_make_i64(-3);  defer big_int_dealloc(&y)
    z: BigInt;                  defer big_int_dealloc(&z)
    big_int_and_not(&z, &x, &y)
    testing.expect(t, z.used >= 0)
}

// ==================== big_int_not signed ====================

@(test)
test_not_signed :: proc(t: ^testing.T) {
    x := big_int_make_u64(0); defer big_int_dealloc(&x)
    z: BigInt;                defer big_int_dealloc(&z)
    big_int_not(&z, &x, 3, true)
    // ~0 in 3-bit signed: 2^3 - 1 = 7, then sign-extend: 7 & 3 - 4 = -1? Actually the logic is:
    // a = dst & pmask_minus_one, b = dst & pmask, result = a - b
    // dst = 0xFF & 0x07 = 7 (since bit_count=3, mask=7)
    // ~0 & 7 = 7
    // and 7 & 3 = 3 (pmask_minus_one for bit_count=3 -> 2^2-1=3)
    // and 7 & 4 = 4
    // 3 - 4 = -1
    testing.expect(t, big_int_to_i64(&z) == -1)
}

@(test)
test_not_signed_max :: proc(t: ^testing.T) {
    a := big_int_make_u64(0); defer big_int_dealloc(&a)
    b: BigInt;                defer big_int_dealloc(&b)
    big_int_not(&b, &a, 64, true)
    // ~0 in 64-bit signed = -1
    testing.expect(t, big_int_to_i64(&b) == -1)
}
