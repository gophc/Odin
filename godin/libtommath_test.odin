// libtommath_test.odin – 固定后的完整单元测试
package godin

import "core:math"
import "core:testing"

// 辅助：用字符串初始化（十进制）
@(private)
mp_from_string :: proc(a: ^mp_int, str: string) -> mp_err {
    return mp_read_radix(a, str, 10)
}

// 辅助：转为十进制字符串（用于调试）
@(private)
mp_to_string :: proc(a: ^mp_int) -> string {
    s, _ := mp_to_radix(a, 10)
    return s
}

// 辅助：用十六进制字符串初始化
@(private)
mp_from_hex :: proc(a: ^mp_int, str: string) -> mp_err {
    return mp_read_radix(a, str, 16)
}

// 辅助：断言两个 mp_int 相等
@(private)
expect_eq :: proc(t: ^testing.T, a, b: ^mp_int, msg: string) {
    if mp_cmp(a, b) != .MP_EQ {
        testing.expectf(t, false, "%s: expected %v, got %v", msg, mp_to_string(b), mp_to_string(a))
    }
}

// 辅助：断言 mp_int 等于 i64
@(private)
expect_i64 :: proc(t: ^testing.T, a: ^mp_int, expected: i64, msg: string) {
    val := mp_get_i64(a)
    testing.expectf(t, val == expected, "%s: expected %v, got %v", msg, expected, val)
}

// 简化错误检查：若 err != OK 则记录失败并返回
@(private)
expect_ok :: proc(t: ^testing.T, err: mp_err, msg: string) -> bool {
    if err != .MP_OKAY {
        testing.expectf(t, false, "%s: got error %v", msg, err)
        return false
    }
    return true
}

// ==================== 初始化 & 内存管理 ====================
@(test)
test_mp_init_clear :: proc(t: ^testing.T) {
    a: mp_int
    err := mp_init(&a)
    if !expect_ok(t, err, "mp_init") do return
    testing.expectf(t, a.alloc == MP_MIN_ALLOC, "alloc should be %v, got %v", MP_MIN_ALLOC, a.alloc)
    testing.expect(t, a.used == 0)
    testing.expect(t, a.sign == .MP_ZPOS)

    mp_clear(&a)
    testing.expect(t, a.dp == nil)
    testing.expect(t, a.alloc == 0)
}

@(test)
test_mp_init_size :: proc(t: ^testing.T) {
    a: mp_int
    err := mp_init_size(&a, 64)
    if !expect_ok(t, err, "mp_init_size 64")  do return
    testing.expect(t, a.alloc >= 64)
    mp_clear(&a)

    // 超限测试
    err = mp_init_size(&a, max(int) / 2 + 1)
    testing.expect(t, err == .MP_OVF)
}

@(test)
test_mp_grow_shrink :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a);   if !expect_ok(t, err, "mp_init")  do return
    err = mp_grow(&a, 10); if !expect_ok(t, err, "mp_grow")  do return
    testing.expect(t, a.alloc >= 10)

    err = mp_shrink(&a);  if !expect_ok(t, err, "mp_shrink")  do return
    testing.expect(t, a.alloc == 32) // 回到最小分配

    // 设定非零值后收缩
    mp_set(&a, 12345)
    err = mp_shrink(&a);  if !expect_ok(t, err, "mp_shrink after set")  do return
    testing.expect(t, a.alloc >= 1)
    testing.expect(t, a.used == 1)
}

@(test)
test_mp_set_and_zero :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a);   if !expect_ok(t, err, "mp_init")  do return
    mp_set(&a, 777)
    testing.expect(t, a.used == 1)
    testing.expect(t, a.dp[0] == 777)
    testing.expect(t, a.sign == .MP_ZPOS)

    mp_zero(&a)
    testing.expect(t, a.used == 0)
    testing.expect(t, a.sign == .MP_ZPOS)
}

// ==================== 复制与交换 ====================
@(test)
test_mp_copy :: proc(t: ^testing.T) {
    a, b: mp_int
    defer mp_clear(&a)
    defer mp_clear(&b)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return

    mp_set(&a, 65535)
    err = mp_copy(&a, &b); if !expect_ok(t, err, "copy positive")  do return
    testing.expect(t, b.used == a.used)
    testing.expect(t, b.dp[0] == 65535)
    testing.expect(t, b.sign == .MP_ZPOS)

    // 复制负数
    a.sign = .MP_NEG
    err = mp_copy(&a, &b); if !expect_ok(t, err, "copy negative")  do return
    testing.expect(t, b.sign == .MP_NEG)
}

@(test)
test_mp_exch :: proc(t: ^testing.T) {
    a, b: mp_int
    defer mp_clear(&a)
    defer mp_clear(&b)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return

    mp_set(&a, 100)
    mp_set(&b, 200)
    mp_exch(&a, &b)
    testing.expect(t, a.dp[0] == 200)
    testing.expect(t, b.dp[0] == 100)
}

// ==================== 比较运算 ====================
@(test)
test_comparisons :: proc(t: ^testing.T) {
    a, b: mp_int
    defer mp_clear(&a)
    defer mp_clear(&b)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return

    // 零相等
    testing.expect(t, mp_cmp(&a, &b) == .MP_EQ)
    testing.expect(t, mp_cmp_mag(&a, &b) == .MP_EQ)

    // 正 vs 零
    mp_set(&a, 5)
    testing.expect(t, mp_cmp(&a, &b) == .MP_GT)
    testing.expect(t, mp_cmp_mag(&a, &b) == .MP_GT)

    // 负 vs 零
    a.sign = .MP_NEG
    testing.expect(t, mp_cmp(&a, &b) == .MP_LT)

    // 两个负数绝对值相等
    mp_set(&b, 5); b.sign = .MP_NEG
    testing.expect(t, mp_cmp(&a, &b) == .MP_EQ)

    // 绝对值大小比较
    mp_set(&a, 1); mp_set(&b, 2)
    testing.expect(t, mp_cmp_mag(&a, &b) == .MP_LT)

    // mp_cmp_d
    mp_set(&a, 1)
    testing.expect(t, mp_cmp_d(&a, 0) == .MP_GT)
    testing.expect(t, mp_cmp_d(&a, 1) == .MP_EQ)
    a.sign = .MP_NEG
    testing.expect(t, mp_cmp_d(&a, 1) == .MP_LT)
}

// ==================== 加减法 ====================
@(test)
test_add_sub_basics :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // 2 + 3 = 5
    mp_set(&a, 2); mp_set(&b, 3)
    err = mp_add(&a, &b, &c); if !expect_ok(t, err, "2+3")  do return
    testing.expectf(t, mp_cmp_d(&c, 5) == .MP_EQ, "2+3: expected 5, got %v", mp_to_string(&c))

    // 5 - 3 = 2
    err = mp_sub(&c, &b, &c); if !expect_ok(t, err, "5-3")  do return
    testing.expectf(t, mp_cmp_d(&c, 2) == .MP_EQ, "5-3: expected 2, got %v", mp_to_string(&c))

    // -2 + 3 = 1
    mp_set(&a, 2); a.sign = .MP_NEG
    mp_set(&b, 3)
    err = mp_add(&a, &b, &c); if !expect_ok(t, err, "-2+3")  do return
    testing.expectf(t, mp_cmp_d(&c, 1) == .MP_EQ, "-2+3: expected 1")

    // 3 - (-2) = 5
    err = mp_sub(&b, &a, &c); if !expect_ok(t, err, "3-(-2)")  do return
    testing.expectf(t, mp_cmp_d(&c, 5) == .MP_EQ, "3-(-2): expected 5")
}

@(test)
test_add_d_sub_d :: proc(t: ^testing.T) {
    a, c: mp_int
    defer mp_clear(&a); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // 正数加单数字：100 + 55 = 155
    mp_set(&a, 100)
    err = mp_add_d(&a, 55, &c); if !expect_ok(t, err, "100+55")  do return
    testing.expectf(t, mp_cmp_d(&c, 155) == .MP_EQ, "100+55: expected 155, got %v", mp_to_string(&c))

    // 负数加单数字：-100 + 55 结果应为 -45
    a.sign = .MP_NEG
    err = mp_add_d(&a, 55, &c); if !expect_ok(t, err, "-100+55")  do return
    // 预期：绝对值 45，符号负
    testing.expectf(t, mp_get_i64(&c) == -45 && c.sign == .MP_NEG,
    "-100+55: expected -45, got sign=%v value=%s", c.sign, mp_to_string(&c))

    // 正数减单数字：100 - 30 = 70
    mp_set(&a, 100)
    err = mp_sub_d(&a, 30, &c); if !expect_ok(t, err, "100-30")  do return
    testing.expectf(t, mp_cmp_d(&c, 70) == .MP_EQ, "100-30: expected 70, got %v", mp_to_string(&c))

    // 负数减单数字：-100 - 30 = -130
    a.sign = .MP_NEG
    err = mp_sub_d(&a, 30, &c); if !expect_ok(t, err, "-100-30")  do return
    testing.expectf(t, mp_get_i64(&c) == -130 && c.sign == .MP_NEG,
    "-100-30: expected -130, got %v", mp_to_string(&c))
}

// ==================== 乘法 ====================
@(test)
test_mul_small :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    mp_set(&a, 123); mp_set(&b, 456)
    err = mp_mul(&a, &b, &c); if !expect_ok(t, err, "123*456")  do return
    testing.expectf(t, mp_cmp_d(&c, 123 * 456) == .MP_EQ,
    "123*456: expected %v, got %v", 123 * 456, mp_to_string(&c))

    // 平方
    err = mp_mul(&a, &a, &c); if !expect_ok(t, err, "123^2")  do return
    testing.expectf(t, mp_cmp_d(&c, 123 * 123) == .MP_EQ,
    "123^2: expected %v, got %v", 123 * 123, mp_to_string(&c))
}

@(test)
test_mul_d :: proc(t: ^testing.T) {
    a, c: mp_int
    defer mp_clear(&a); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    mp_set(&a, 65537)
    err = mp_mul_d(&a, 2, &c); if !expect_ok(t, err, "65537*2")  do return
    testing.expectf(t, mp_cmp_d(&c, 65537 * 2) == .MP_EQ,
    "65537*2: expected %v, got %v", 65537 * 2, mp_to_string(&c))

    // 乘以 2 的幂：65537 * 8 = 524296
    err = mp_mul_d(&a, 8, &c); if !expect_ok(t, err, "65537*8")  do return
    testing.expectf(t, mp_cmp_d(&c, 65537 * 8) == .MP_EQ,
    "65537*8: expected %v, got %v", 65537 * 8, mp_to_string(&c))
}

@(test)
test_mul_2_mul_2d :: proc(t: ^testing.T) {
    a, c: mp_int
    defer mp_clear(&a); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    mp_set(&a, 1)
    err = mp_mul_2d(&a, 62, &c); if !expect_ok(t, err, "2^62")  do return
    val := mp_get_mag_u64(&c)
    testing.expectf(t, val == (u64(1) << 62), "2^62: expected %v, got %v", u64(1) << 62, val)

    // 再乘 2
    err = mp_mul_2(&c, &c); if !expect_ok(t, err, "mul 2")  do return
    val = mp_get_mag_u64(&c)
    testing.expect(t, val == (u64(1) << 62) * 2)
}

// ==================== 除法 ====================
@(test)
test_div_2_div_2d :: proc(t: ^testing.T) {
    a, c, d: mp_int
    defer mp_clear(&a); defer mp_clear(&c); defer mp_clear(&d)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return
    err = mp_init(&d);  if !expect_ok(t, err, "init d")  do return

    mp_set(&a, 1024)
    err = mp_div_2(&a, &c); if !expect_ok(t, err, "1024/2")  do return
    testing.expectf(t, mp_cmp_d(&c, 512) == .MP_EQ,
    "1024/2: expected 512, got %v", mp_to_string(&c))

    // 右移 3 位，同时取余数
    err = mp_div_2d(&a, 3, &c, &d); if !expect_ok(t, err, "1024/8 rem")  do return
    testing.expectf(t, mp_cmp_d(&c, 128) == .MP_EQ, "1024/8: expected 128")
    testing.expectf(t, mp_cmp_d(&d, 0) == .MP_EQ, "1024%8: expected 0")

    mp_set(&a, 1025)
    err = mp_div_2d(&a, 3, &c, &d); if !expect_ok(t, err, "1025/8 rem")  do return
    testing.expectf(t, mp_cmp_d(&c, 128) == .MP_EQ, "1025/8: expected 128")
    testing.expectf(t, mp_cmp_d(&d, 1) == .MP_EQ, "1025%8: expected 1")
}

@(test)
test_mod_2d :: proc(t: ^testing.T) {
    a, c: mp_int
    defer mp_clear(&a); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    mp_set(&a, 1000)
    err = mp_mod_2d(&a, 4, &c); if !expect_ok(t, err, "1000 mod 16")  do return
    testing.expectf(t, mp_cmp_d(&c, 8) == .MP_EQ, "1000%%16: expected 8, got %v", mp_to_string(&c))

    // 模 1
    err = mp_mod_2d(&a, 0, &c); if !expect_ok(t, err, "mod 1")  do return
    testing.expectf(t, mp_cmp_d(&c, 0) == .MP_EQ, "mod 1: expected 0")

    // 模大于位数时复制原数
    err = mp_mod_2d(&a, a.used * 28 + 1, &c); if !expect_ok(t, err, "mod large")  do return
    testing.expect(t, mp_cmp(&a, &c) == .MP_EQ)
}

@(test)
test_div_standard :: proc(t: ^testing.T) {
    a, b, q, r: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&q); defer mp_clear(&r)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&q);  if !expect_ok(t, err, "init q")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return

    // 1000 / 3 = 333 余 1
    mp_set(&a, 1000); mp_set(&b, 3)
    err = mp_div(&a, &b, &q, &r); if !expect_ok(t, err, "1000/3")  do return
    testing.expectf(t, mp_cmp_d(&q, 333) == .MP_EQ,
    "1000//3: expected 333, got %v", mp_to_string(&q))
    testing.expectf(t, mp_cmp_d(&r, 1) == .MP_EQ,
    "1000%%3: expected 1, got %v", mp_to_string(&r))

    // 除数为零应报错
    mp_zero(&b)
    err = mp_div(&a, &b, &q, &r)
    testing.expect(t, err == .MP_VAL)
}

@(test)
test_div_d :: proc(t: ^testing.T) {
    a, q: mp_int
    rem: mp_digit
    defer mp_clear(&a); defer mp_clear(&q)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&q);  if !expect_ok(t, err, "init q")  do return

    mp_set(&a, 1000)
    err = mp_div_d(&a, 7, &q, &rem); if !expect_ok(t, err, "1000/7")  do return
    testing.expectf(t, mp_cmp_d(&q, 1000 / 7) == .MP_EQ,
    "quot: expected %v, got %v", 1000 / 7, mp_to_string(&q))
    testing.expectf(t, rem == 1000 % 7, "rem: expected %v, got %v", 1000 % 7, rem)

    // 除 1
    mp_set(&a, 12345)
    err = mp_div_d(&a, 1, &q, &rem); if !expect_ok(t, err, "div by 1")  do return
    testing.expect(t, mp_cmp(&a, &q) == .MP_EQ)
    testing.expect(t, rem == 0)
}

// ==================== GCD, LCM, 模逆 ====================
@(test)
test_gcd :: proc(t: ^testing.T) {
    a, b, g: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&g)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&g);  if !expect_ok(t, err, "init g")  do return

    mp_set(&a, 48); mp_set(&b, 180)
    err = mp_gcd(&a, &b, &g); if !expect_ok(t, err, "gcd(48,180)")  do return
    testing.expectf(t, mp_cmp_d(&g, 12) == .MP_EQ,
    "gcd(48,180): expected 12, got %v", mp_to_string(&g))

    // gcd(a,0) = |a|
    mp_zero(&b)
    err = mp_gcd(&a, &b, &g); if !expect_ok(t, err, "gcd(a,0)")  do return
    testing.expect(t, mp_cmp(&a, &g) == .MP_EQ)
}

@(test)
test_lcm :: proc(t: ^testing.T) {
    a, b, l: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&l)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&l);  if !expect_ok(t, err, "init l")  do return

    mp_set(&a, 12); mp_set(&b, 18)
    err = mp_lcm(&a, &b, &l); if !expect_ok(t, err, "lcm(12,18)")  do return
    testing.expectf(t, mp_cmp_d(&l, 36) == .MP_EQ,
    "lcm(12,18): expected 36, got %v", mp_to_string(&l))
}

@(test)
test_invmod_odd :: proc(t: ^testing.T) {
    a, m, inv: mp_int
    defer mp_clear(&a); defer mp_clear(&m); defer mp_clear(&inv)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&m);  if !expect_ok(t, err, "init m")  do return
    err = mp_init(&inv);if !expect_ok(t, err, "init inv")  do return

    // 3^{-1} mod 11 = 4
    mp_set(&a, 3); mp_set(&m, 11)
    err = mp_invmod(&a, &m, &inv); if !expect_ok(t, err, "invmod(3,11)")  do return
    testing.expectf(t, mp_cmp_d(&inv, 4) == .MP_EQ,
    "3^-1 mod 11: expected 4, got %v", mp_to_string(&inv))

    // 模数为零
    mp_zero(&m)
    err = mp_invmod(&a, &m, &inv)
    testing.expect(t, err == .MP_VAL)
}

// ==================== 位运算 ====================
@(test)
test_bitwise :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    mp_set(&a, 0xFF00); mp_set(&b, 0x0FF0)

    err = mp_and(&a, &b, &c); if !expect_ok(t, err, "and")  do return
    testing.expectf(t, mp_cmp_d(&c, 0x0F00) == .MP_EQ,
    "0xFF00 & 0x0FF0: expected 0x0F00, got 0x%X", mp_get_mag_u32(&c))

    err = mp_or(&a, &b, &c); if !expect_ok(t, err, "or")  do return
    testing.expectf(t, mp_cmp_d(&c, 0xFFF0) == .MP_EQ,
    "0xFF00 | 0x0FF0: expected 0xFFF0, got 0x%X", mp_get_mag_u32(&c))

    err = mp_xor(&a, &b, &c); if !expect_ok(t, err, "xor")  do return
    testing.expectf(t, mp_cmp_d(&c, 0xF0F0) == .MP_EQ,
    "0xFF00 ^ 0x0FF0: expected 0xF0F0, got 0x%X", mp_get_mag_u32(&c))
}

@(test)
test_complement_signed_rsh :: proc(t: ^testing.T) {
    a, c: mp_int
    defer mp_clear(&a); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    mp_set(&a, 0x1234)
    err = mp_complement(&a, &c); if !expect_ok(t, err, "complement")  do return
    // 对正数补码即 -(a+1)
    mp_add_d(&a, 1, &a); if !expect_ok(t, err, "a+1")  do return // 忽略前一个 err，重新捕获
    a.sign = .MP_NEG
    testing.expectf(t, mp_cmp(&a, &c) == .MP_EQ, "complement positive: expected %v, got %v", mp_to_string(&a), mp_to_string(&c))

    // 有符号右移 -1 仍为 -1
    mp_set(&a, 1); a.sign = .MP_NEG
    err = mp_signed_rsh(&a, 5, &c); if !expect_ok(t, err, "signed rsh -1")  do return
    testing.expectf(t, mp_get_i64(&c) == -1,
    "-1>>5: expected -1, got %v", mp_to_string(&c))
}

// ==================== 平方根 & 幂 ====================
@(test)
test_sqrt :: proc(t: ^testing.T) {
    x, r: mp_int
    defer mp_clear(&x); defer mp_clear(&r)

    err := mp_init(&x); if !expect_ok(t, err, "init x")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return

    mp_set(&x, 144)
    err = mp_sqrt(&x, &r); if !expect_ok(t, err, "sqrt 144")  do return
    testing.expectf(t, mp_cmp_d(&r, 12) == .MP_EQ,
    "sqrt(144): expected 12, got %v", mp_to_string(&r))

    // sqrt(0)
    mp_zero(&x)
    err = mp_sqrt(&x, &r); if !expect_ok(t, err, "sqrt 0")  do return
    testing.expectf(t, mp_cmp_d(&r, 0) == .MP_EQ, "sqrt(0): expected 0")

    // 负数报错
    x.sign = .MP_NEG
    err = mp_sqrt(&x, &r)
    testing.expect(t, err == .MP_VAL)
}

@(test)
test_expt_n :: proc(t: ^testing.T) {
    base, exp, res: mp_int
    defer mp_clear(&base); defer mp_clear(&exp); defer mp_clear(&res)

    err := mp_init(&base); if !expect_ok(t, err, "init base")  do return
    err = mp_init(&exp);   if !expect_ok(t, err, "init exp")  do return
    err = mp_init(&res);   if !expect_ok(t, err, "init res")  do return

    mp_set(&base, 2)
    err = mp_expt_n(&base, 10, &res); if !expect_ok(t, err, "2^10")  do return
    testing.expectf(t, mp_cmp_d(&res, 1024) == .MP_EQ,
    "2^10: expected 1024, got %v", mp_to_string(&res))

    err = mp_expt_n(&base, 0, &res); if !expect_ok(t, err, "2^0")  do return
    testing.expectf(t, mp_cmp_d(&res, 1) == .MP_EQ, "2^0: expected 1")

// 0^0? 库可能返回 1
}

// ==================== n 次方根 ====================
@(test)
test_root_n :: proc(t: ^testing.T) {
    a, r: mp_int
    defer mp_clear(&a); defer mp_clear(&r)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return

    mp_set(&a, 125)
    err = mp_root_n(&a, 3, &r); if !expect_ok(t, err, "cbrt 125")  do return
    testing.expectf(t, mp_cmp_d(&r, 5) == .MP_EQ,
    "cbrt(125): expected 5, got %v", mp_to_string(&r))

    // 较大的数
    mp_set(&a, 1000000)
    err = mp_root_n(&a, 6, &r); if !expect_ok(t, err, "root6 1e6")  do return
    approx := mp_get_i32(&r)
    testing.expectf(t, approx >= 9 && approx <= 10,
    "root6(1e6): expected ~10, got %v", approx)
}

// ==================== 类型转换 ====================
@(test)
test_conversions :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    // u32
    mp_set_u32(&a, 0xABCDEF01)
    testing.expect(t, mp_get_mag_u32(&a) == 0xABCDEF01)

    // i32 负数
    mp_set_i32(&a, -1234)
    testing.expect(t, mp_get_i32(&a) == -1234)

    // u64
    big := u64(0xFFFFFFFFFFFFFFF0)
    mp_set_u64(&a, big)
    testing.expect(t, mp_get_mag_u64(&a) == big)

    // double 近似
    err = mp_set_double(&a, 3.14159); if !expect_ok(t, err, "set double")  do return
    d := mp_get_double(&a)
    testing.expect_value(t, d, 3.0)
}

// ==================== 打包 / 解包 ====================
@(test)
test_pack_unpack :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    mp_set(&a, 0x12345678)

    buf := make([]u8, 8)
    defer delete(buf)

    written, err2 := mp_pack(buf, .MP_MSB_FIRST, 4, .MP_BIG_ENDIAN, 0, &a)
    if !expect_ok(t, err2, "pack")  do return
    testing.expect(t, written == 1)

    // 解包
    b: mp_int
    defer mp_clear(&b)
    err = mp_init(&b); if !expect_ok(t, err, "init b")  do return
    err = mp_unpack(&b, 1, .MP_MSB_FIRST, 4, .MP_BIG_ENDIAN, 0, buf[:4])
    if !expect_ok(t, err, "unpack")  do return
    testing.expectf(t, mp_cmp_d(&b, 0x12345678) == .MP_EQ,
    "unpack: expected 0x12345678, got %v", mp_to_string(&b))
}

// ==================== 进制转换 ====================
@(test)
test_radix :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    err = mp_from_string(&a, "12345"); if !expect_ok(t, err, "from dec")  do return
    str, err2 := mp_to_radix(&a, 10)

    testing.expectf(t, err2 == .MP_OKAY && str == "12345",
    "dec roundtrip: expected 12345, got %v (err=%v)", str, err2)

    // 十六进制
    err = mp_read_radix(&a, "1A2B3C", 16); if !expect_ok(t, err, "read hex")  do return
    hex, _ := mp_to_radix(&a, 16)

    // 转换结果应全大写
    testing.expectf(t, hex == "1A2B3C", "hex roundtrip: expected 1A2B3C, got %v", hex)
}

// ==================== 平方检测 ====================
@(test)
test_is_square :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    mp_set(&a, 1234321)  // 1111^2
    res: bool
    err = mp_is_square(&a, &res); if !expect_ok(t, err, "is square 1234321")  do return
    testing.expect(t, res)

    mp_set(&a, 1234322)
    err = mp_is_square(&a, &res); if !expect_ok(t, err, "is square 1234322")  do return
    testing.expect(t, !res)
}

// ==================== 位计数 ====================
@(test)
test_bit_counts :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    mp_set(&a, 0)
    testing.expect(t, mp_count_bits(&a) == 0)

    mp_set(&a, 1)
    testing.expect(t, mp_count_bits(&a) == 1)

    mp_set(&a, 255)
    testing.expect(t, mp_count_bits(&a) == 8)

    // cnt_lsb: 32 = 2^5, LSB 连续 0 的个数为 5
    mp_set(&a, 32)
    testing.expect(t, mp_cnt_lsb(&a) == 5)

    mp_set(&a, 17)
    testing.expect(t, mp_cnt_lsb(&a) == 0)

    // cnt_lsb on zero
    mp_zero(&a)
    testing.expect(t, mp_cnt_lsb(&a) == 0)

    // count_bits on larger value
    mp_set(&a, 0xFFFFFFFF)
    n := mp_count_bits(&a)
    testing.expect(t, n == 32)
}

// ==================== 否决与绝对值 ====================
@(test)
test_neg_abs :: proc(t: ^testing.T) {
    a, b: mp_int
    defer mp_clear(&a); defer mp_clear(&b)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return

    // 正数取否
    mp_set(&a, 123)
    err = mp_neg(&a, &b); if !expect_ok(t, err, "neg 123")  do return
    testing.expect(t, b.sign == .MP_NEG && mp_cmp_mag(&a, &b) == .MP_EQ)

    // 负数取否
    err = mp_neg(&b, &a); if !expect_ok(t, err, "neg -123")  do return
    testing.expect(t, a.sign == .MP_ZPOS && mp_cmp_d(&a, 123) == .MP_EQ)

    // 零取否
    mp_zero(&a)
    err = mp_neg(&a, &b); if !expect_ok(t, err, "neg 0")  do return
    testing.expect(t, b.sign == .MP_ZPOS && b.used == 0)

    // abs of negative
    mp_set(&a, 456); a.sign = .MP_NEG
    err = mp_abs(&a, &b); if !expect_ok(t, err, "abs -456")  do return
    testing.expect(t, b.sign == .MP_ZPOS && mp_cmp_d(&b, 456) == .MP_EQ)

    // abs of positive
    mp_set(&a, 789)
    err = mp_abs(&a, &b); if !expect_ok(t, err, "abs 789")  do return
    testing.expect(t, mp_cmp(&a, &b) == .MP_EQ)
}

// ==================== 2expt ====================
@(test)
test_2expt :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    err = mp_2expt(&a, 0); if !expect_ok(t, err, "2^0")  do return
    testing.expectf(t, mp_cmp_d(&a, 1) == .MP_EQ, "2^0: expected 1, got %v", mp_to_string(&a))

    err = mp_2expt(&a, 62); if !expect_ok(t, err, "2^62")  do return
    val := mp_get_mag_u64(&a)
    testing.expectf(t, val == (u64(1) << 62), "2^62: expected %v, got %v", u64(1) << 62, val)

    err = mp_2expt(&a, 100); if !expect_ok(t, err, "2^100")  do return
    testing.expect(t, mp_count_bits(&a) == 101)
    testing.expect(t, a.sign == .MP_ZPOS && a.used > 1)
}

// ==================== 自引用运算 ====================
@(test)
test_self_reference :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    // add self
    mp_set(&a, 5)
    err = mp_add(&a, &a, &a); if !expect_ok(t, err, "add self")  do return
    testing.expectf(t, mp_cmp_d(&a, 10) == .MP_EQ, "5+5: expected 10, got %v", mp_to_string(&a))

    // sub self (result zero)
    mp_set(&a, 7)
    err = mp_sub(&a, &a, &a); if !expect_ok(t, err, "sub self")  do return
    testing.expect(t, mp_cmp_d(&a, 0) == .MP_EQ)

    // mul self (square)
    mp_set(&a, 6)
    err = mp_mul(&a, &a, &a); if !expect_ok(t, err, "mul self")  do return
    testing.expectf(t, mp_cmp_d(&a, 36) == .MP_EQ, "6^2: expected 36, got %v", mp_to_string(&a))

    // copy self
    mp_set(&a, 99)
    err = mp_copy(&a, &a); if !expect_ok(t, err, "copy self")  do return
    testing.expect(t, mp_cmp_d(&a, 99) == .MP_EQ)
}

// ==================== 取模运算 ====================
@(test)
test_mod_and_modular_arith :: proc(t: ^testing.T) {
    a, b, m, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&m); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&m);  if !expect_ok(t, err, "init m")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // mp_mod: 10 mod 3 = 1
    mp_set(&a, 10); mp_set(&m, 3)
    err = mp_mod(&a, &m, &c); if !expect_ok(t, err, "10 mod 3")  do return
    testing.expectf(t, mp_cmp_d(&c, 1) == .MP_EQ, "10 mod 3: expected 1, got %v", mp_to_string(&c))

    // mp_mod: negative dividend
    mp_set(&a, 10); a.sign = .MP_NEG
    mp_set(&m, 3)
    err = mp_mod(&a, &m, &c); if !expect_ok(t, err, "-10 mod 3")  do return
    testing.expectf(t, mp_cmp_d(&c, 2) == .MP_EQ, "-10 mod 3: expected 2, got %v", mp_to_string(&c))

    // addmod: (7 + 8) mod 5 = 0
    mp_set(&a, 7); mp_set(&b, 8); mp_set(&m, 5)
    err = mp_addmod(&a, &b, &m, &c); if !expect_ok(t, err, "(7+8) mod 5")  do return
    testing.expectf(t, mp_cmp_d(&c, 0) == .MP_EQ, "(7+8) mod 5: expected 0, got %v", mp_to_string(&c))

    // submod: (3 - 7) mod 10 = 6
    mp_set(&a, 3); mp_set(&b, 7); mp_set(&m, 10)
    err = mp_submod(&a, &b, &m, &c); if !expect_ok(t, err, "(3-7) mod 10")  do return
    testing.expectf(t, mp_cmp_d(&c, 6) == .MP_EQ, "(3-7) mod 10: expected 6, got %v", mp_to_string(&c))

    // mulmod: (7 * 8) mod 11 = 1
    mp_set(&a, 7); mp_set(&b, 8); mp_set(&m, 11)
    err = mp_mulmod(&a, &b, &m, &c); if !expect_ok(t, err, "(7*8) mod 11")  do return
    testing.expectf(t, mp_cmp_d(&c, 1) == .MP_EQ, "(7*8) mod 11: expected 1, got %v", mp_to_string(&c))

    // sqrmod: 4^2 mod 7 = 2
    mp_set(&a, 4); mp_set(&m, 7)
    err = mp_sqrmod(&a, &m, &c); if !expect_ok(t, err, "4^2 mod 7")  do return
    testing.expectf(t, mp_cmp_d(&c, 2) == .MP_EQ, "4^2 mod 7: expected 2, got %v", mp_to_string(&c))
}

// ==================== invmod 偶数模 ====================
@(test)
test_invmod_even :: proc(t: ^testing.T) {
    a, m, inv: mp_int
    defer mp_clear(&a); defer mp_clear(&m); defer mp_clear(&inv)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&m);  if !expect_ok(t, err, "init m")  do return
    err = mp_init(&inv); if !expect_ok(t, err, "init inv")  do return

    // 对偶数模且 gcd(a,m)=1 的情况使用 s_mp_invmod
    // 3^{-1} mod 10 = 7
    mp_set(&a, 3); mp_set(&m, 10)
    err = mp_invmod(&a, &m, &inv); if !expect_ok(t, err, "invmod(3,10)")  do return
    // Verify: 3*7=21 mod 10 = 1
    mp_mulmod(&a, &inv, &m, &a); if !expect_ok(t, err, "verify")  do return
    testing.expectf(t, mp_cmp_d(&a, 1) == .MP_EQ, "3*inv mod 10: expected 1, got %v", mp_to_string(&a))

    // invmod with a modulus of 1
    mp_set(&a, 0); mp_set(&m, 1)
    err = mp_invmod(&a, &m, &inv); if !expect_ok(t, err, "invmod(0,1)")  do return
    testing.expectf(t, mp_cmp_d(&inv, 0) == .MP_EQ, "invmod(0,1): expected 0")
}

// ==================== DR 取约简 ====================
@(test)
test_dr_reduction :: proc(t: ^testing.T) {
    a, m: mp_int
    rho: mp_digit
    defer mp_clear(&a); defer mp_clear(&m)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&m);  if !expect_ok(t, err, "init m")  do return

    // n = 2^28 - 1 = MP_MASK 是 DR moduli 的典型例子
    mp_set(&m, MP_MASK)
    testing.expect(t, mp_dr_is_modulus(&m) == false) // only one digit

    // 设置 DR moduli: 低digit任意，高digit全1
    mp_set(&a, 12345)
    // 构建大于1位数的DR moduli
    mp_set_u64(&m, u64(0xFFFFFFF) | (u64(MP_MASK) << 28))
    if mp_dr_is_modulus(&m) {
        mp_dr_setup(&m, &rho)
        testing.expect(t, rho > 0)
        // 计算 a mod m via DR
        mp_mod(&a, &m, &a) // direct mod for reference
        b: mp_int; defer mp_clear(&b)
        err = mp_init_copy(&b, &a);  if !expect_ok(t, err, "init_copy")  do return
        // DR reduce should give same result (if |a| < m)
        err = mp_dr_reduce(&b, &m, rho); if !expect_ok(t, err, "dr_reduce")  do return
        testing.expect(t, mp_cmp(&a, &b) == .MP_EQ)
    }
}

// ==================== 取模与大数的除法与乘法 ====================
@(test)
test_div_mod_large :: proc(t: ^testing.T) {
    a, b, q, r: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&q); defer mp_clear(&r)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&q);  if !expect_ok(t, err, "init q")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return

    // 大数据除法 触发递归除法（或学校除法）
    mp_set_u64(&a, u64(1234567890123456))
    mp_set_u64(&b, u64(9876543210))
    err = mp_div(&a, &b, &q, &r); if !expect_ok(t, err, "div large")  do return
    // 验证: a = q*b + r
    t1: mp_int; defer mp_clear(&t1)
    err = mp_init(&t1); if !expect_ok(t, err, "init t1")  do return
    err = mp_mul(&q, &b, &t1); if !expect_ok(t, err, "q*b")  do return
    err = mp_add(&t1, &r, &t1); if !expect_ok(t, err, "q*b+r")  do return
    expect_eq(t, &t1, &a, "q*b+r == a")

    // 负数被除数
    a.sign = .MP_NEG
    err = mp_div(&a, &b, &q, &r); if !expect_ok(t, err, "div -a/b")  do return
    testing.expect(t, q.sign == .MP_NEG)

    // 负数除数
    mp_set_u64(&a, u64(1234567890123456))
    b.sign = .MP_NEG
    err = mp_div(&a, &b, &q, &r); if !expect_ok(t, err, "div a/-b")  do return
    testing.expect(t, q.sign == .MP_NEG)
}

// ==================== 乘法：Karatsuba 路径 ====================
@(test)
test_mul_karatsuba :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // 超过 80 位数的乘法会触发 Karatsuba
    mp_2expt(&a, 84) // 2^84
    mp_set(&b, 123456789)
    err = mp_mul(&a, &b, &c); if !expect_ok(t, err, "2^84 * n")  do return
    // 验证: c / a = b
    q, r: mp_int
    defer mp_clear(&q); defer mp_clear(&r)
    err = mp_init(&q);  if !expect_ok(t, err, "init q")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return
    err = mp_div(&c, &a, &q, &r);  if !expect_ok(t, err, "mp_div")  do return
    testing.expectf(t, mp_cmp(&b, &q) == .MP_EQ, "(2^84 * n) / 2^84 == n")
}

// ==================== 乘法：负数 ====================
@(test)
test_mul_negative :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // (-7) * 3 = -21
    mp_set_i32(&a, -7); mp_set(&b, 3)
    err = mp_mul(&a, &b, &c); if !expect_ok(t, err, "-7*3")  do return
    vv := mp_get_i64(&c)
    testing.expect(t, vv == -21, "-7*3 == -21")

    // (-7) * (-3) = 21
    mp_set_i32(&a, -7); mp_set_i32(&b, -3)
    err = mp_mul(&a, &b, &c); if !expect_ok(t, err, "-7*-3")  do return
    testing.expect(t, mp_get_i64(&c) == 21)

    // 0 * any = 0
    mp_zero(&a); mp_set(&b, 12345)
    err = mp_mul(&a, &b, &c); if !expect_ok(t, err, "0*n")  do return
    testing.expect(t, mp_cmp_d(&c, 0) == .MP_EQ)
}

// ==================== 除法：div_d 边界 ====================
@(test)
test_div_d_edge :: proc(t: ^testing.T) {
    a, q: mp_int
    rem: mp_digit
    defer mp_clear(&a); defer mp_clear(&q)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&q);  if !expect_ok(t, err, "init q")  do return

    // a=0
    mp_zero(&a)
    err = mp_div_d(&a, 5, &q, &rem); if !expect_ok(t, err, "0/5")  do return
    testing.expect(t, mp_cmp_d(&q, 0) == .MP_EQ && rem == 0)

    // 除以 0 报错
    mp_set(&a, 10)
    err = mp_div_d(&a, 0, &q, &rem)
    testing.expect(t, err == .MP_VAL)

    // 除以 2 的幂 (触发快速路径)
    mp_set(&a, 64)
    err = mp_div_d(&a, 4, &q, &rem); if !expect_ok(t, err, "64/4")  do return
    testing.expect(t, mp_cmp_d(&q, 16) == .MP_EQ && rem == 0)

    // 除以 3 (触发特殊路径)
    mp_set(&a, 100)
    err = mp_div_d(&a, 3, &q, &rem); if !expect_ok(t, err, "100/3")  do return
    testing.expect(t, mp_cmp_d(&q, 33) == .MP_EQ && rem == 1)

    // nil c
    mp_set(&a, 1000)
    err = mp_div_d(&a, 7, nil, &rem); if !expect_ok(t, err, "1000/7 nil q")  do return
    testing.expect(t, rem == 1000 % 7)

    // nil d
    err = mp_div_d(&a, 9, &q, nil); if !expect_ok(t, err, "1000/9 nil rem")  do return
    testing.expect(t, mp_cmp_d(&q, 1000 / 9) == .MP_EQ)
}

// ==================== mul_d 边界 ====================
@(test)
test_mul_d_edge :: proc(t: ^testing.T) {
    a, c: mp_int
    defer mp_clear(&a); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // mul by 1
    mp_set(&a, 12345)
    err = mp_mul_d(&a, 1, &c); if !expect_ok(t, err, "*1")  do return
    testing.expect(t, mp_cmp(&a, &c) == .MP_EQ)

    // mul by 2 (trigger mp_mul_2)
    err = mp_mul_d(&a, 2, &c); if !expect_ok(t, err, "*2")  do return
    testing.expect(t, mp_cmp_d(&c, 24690) == .MP_EQ)

    // mul by 4 (two's-power shortcut)
    err = mp_mul_d(&a, 4, &c); if !expect_ok(t, err, "*4")  do return
    testing.expect(t, mp_cmp_d(&c, 49380) == .MP_EQ)

    // mul by 0
    err = mp_mul_d(&a, 0, &c); if !expect_ok(t, err, "*0")  do return
    testing.expect(t, mp_cmp_d(&c, 0) == .MP_EQ)

    // negative * d
    a.sign = .MP_NEG
    err = mp_mul_d(&a, 3, &c); if !expect_ok(t, err, "-n*3")  do return
    testing.expect(t, mp_get_i64(&c) == -(12345 * 3))
}

// ==================== add_d / sub_d 自引用 ====================
@(test)
test_add_sub_d_self :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    // add_d self: positive small
    mp_set(&a, 5)
    err = mp_add_d(&a, 3, &a); if !expect_ok(t, err, "5+3 self")  do return
    testing.expectf(t, mp_cmp_d(&a, 8) == .MP_EQ, "5+3: expected 8, got %v", mp_to_string(&a))

    // add_d self: negative small
    mp_set_i32(&a, -5)
    err = mp_add_d(&a, 3, &a); if !expect_ok(t, err, "-5+3 self")  do return
    testing.expectf(t, mp_get_i64(&a) == -2, "-5+3: expected -2, got %v", mp_to_string(&a))

    // sub_d self: positive large
    mp_set(&a, 500)
    err = mp_sub_d(&a, 100, &a); if !expect_ok(t, err, "500-100 self")  do return
    testing.expectf(t, mp_cmp_d(&a, 400) == .MP_EQ, "500-100: expected 400, got %v", mp_to_string(&a))

    // sub_d self: negative
    mp_set_i32(&a, -10)
    err = mp_sub_d(&a, 5, &a); if !expect_ok(t, err, "-10-5 self")  do return
    testing.expectf(t, mp_get_i64(&a) == -15, "-10-5: expected -15, got %v", mp_to_string(&a))

    // sub_d: result negative
    mp_set(&a, 3)
    err = mp_sub_d(&a, 10, &a); if !expect_ok(t, err, "3-10 self")  do return
    testing.expectf(t, mp_get_i64(&a) == -7, "3-10: expected -7, got %v", mp_to_string(&a))
}

// ==================== 位运算扩展 ====================
@(test)
test_bitwise_extended :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // AND with zeros
    mp_set(&a, 0xFF)
    mp_zero(&b)
    err = mp_and(&a, &b, &c); if !expect_ok(t, err, "and 0")  do return
    testing.expect(t, mp_cmp_d(&c, 0) == .MP_EQ, "mp_and mp_and eq 0")

    // OR with zero
    mp_set(&a, 0xFF); mp_zero(&b)
    err = mp_or(&a, &b, &c); if !expect_ok(t, err, "or 0")  do return
    testing.expect(t, mp_cmp(&a, &c) == .MP_EQ, "mp_or eq 0")

    // XOR self (should be zero)
    mp_set(&a, 0xABCD)
    err = mp_xor(&a, &a, &c); if !expect_ok(t, err, "xor self")  do return
    testing.expect(t, mp_cmp_d(&c, 0) == .MP_EQ, "mp_xor eq 0")

    // complement large
    mp_set(&a, 0)  /* b = ~a */
    err = mp_complement(&a, &c); if !expect_ok(t, err, "complement 0")  do return
    testing.expect(t, mp_get_i64(&c) == -1, "mp_complement eq -1")

    mp_set(&a, 255)  /* b = ~a */
    err = mp_complement(&a, &c); if !expect_ok(t, err, "complement 255")  do return
    testing.expect(t, mp_get_i64(&c) == -256, "mp_complement eq -256")

    // signed_rsh positive
    mp_set(&a, 128)
    err = mp_signed_rsh(&a, 3, &c); if !expect_ok(t, err, "128>>3")  do return
    testing.expect(t, mp_cmp_d(&c, 16) == .MP_EQ, "mp_signed_rsh eq 16")

    // s_mp_get_bit
    mp_set(&a, 0x80) // 1000_0000
    testing.expect(t, s_mp_get_bit(&a, 7) == true)
    testing.expect(t, s_mp_get_bit(&a, 6) == false)
    testing.expect(t, s_mp_get_bit(&a, 100) == false)
}

// ==================== 加法扩展 ====================
@(test)
test_add_extended :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // a+b where a.used < b.used (trip the swap in s_mp_add)
    mp_set(&a, 3)
    mp_set(&b, 1000000000)
    err = mp_add(&a, &b, &c); if !expect_ok(t, err, "3 + 10亿")  do return
    expect_i64(t, &c, 1000000003, "3+1e9")

    // -a + (-b) both negative
    mp_set_i32(&a, -5); mp_set_i32(&b, -7)
    err = mp_add(&a, &b, &c); if !expect_ok(t, err, "-5+-7")  do return
    expect_i64(t, &c, -12, "-5+-7")

    // a + (-b) where |a| < |b|
    mp_set(&a, 3); mp_set_i32(&b, -7)
    err = mp_add(&a, &b, &c); if !expect_ok(t, err, "3+(-7)")  do return
    expect_i64(t, &c, -4, "3+(-7)")

    // a + (-b) where |a| > |b|
    mp_set(&a, 7); mp_set_i32(&b, -3)
    err = mp_add(&a, &b, &c); if !expect_ok(t, err, "7+(-3)")  do return
    expect_i64(t, &c, 4, "7+(-3)")

    // addition that produces 进位
    mp_2expt(&a, 62); mp_2expt(&b, 62)
    testing.expect(t, mp_count_bits(&a) == 63) // 2^62
    err = mp_add(&a, &b, &c); if !expect_ok(t, err, "2^62+2^62")  do return
    mp_2expt(&a, 6)
    testing.expect(t, mp_count_bits(&a) == 7, "a 2^6 expect") // 2^6
    mp_2expt(&b, 63)
    testing.expect(t, mp_count_bits(&b) == 64, "b 2^63 expect") // 2^63
    expect_eq(t, &c, &b, "2^62+2^62 != 2^63 mp_2expt")
    n := mp_count_bits(&c)
    testing.expect(t, n == 64, "c 2^63 expect") // 2^63
}

// ==================== 减法扩展 ====================
@(test)
test_sub_extended :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // a-b where |a| < |b|, both positive
    mp_set(&a, 3); mp_set(&b, 7)
    err = mp_sub(&a, &b, &c); if !expect_ok(t, err, "3-7")  do return
    expect_i64(t, &c, -4, "3-7")

    // -a - b where both negative? 不对: a - (-b) = a + b
    mp_set(&a, 5); mp_set_i32(&b, -3)
    err = mp_sub(&a, &b, &c); if !expect_ok(t, err, "5-(-3)")  do return
    expect_i64(t, &c, 8, "5-(-3)")

    // -a - b
    mp_set_i32(&a, -5); mp_set(&b, 3)
    err = mp_sub(&a, &b, &c); if !expect_ok(t, err, "-5-3")  do return
    expect_i64(t, &c, -8, "-5-3")

    // 结果为0
    mp_set(&a, 100); mp_set(&b, 100)
    err = mp_sub(&a, &b, &c); if !expect_ok(t, err, "100-100")  do return
    testing.expect(t, mp_cmp_d(&c, 0) == .MP_EQ && c.sign == .MP_ZPOS)
}

// ==================== 大数字自指 ====================
@(test)
test_convert_edge :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    // set_double edge: 0.0
    err = mp_set_double(&a, 0.0); if !expect_ok(t, err, "set_double 0")  do return
    testing.expect(t, mp_cmp_d(&a, 0) == .MP_EQ)

    // set_double: negative
    err = mp_set_double(&a, -42.0); if !expect_ok(t, err, "set_double -42")  do return
    testing.expect(t, mp_get_double(&a) == -42.0)
    testing.expect(t, a.sign == .MP_NEG)

    // NaN / Inf should error
    err = mp_set_double(&a, math.inf_f64(0)) // infinity
    testing.expect(t, err == .MP_VAL)

    // get_double 0
    mp_zero(&a)
    testing.expect(t, mp_get_double(&a) == 0.0)

    // get_double negative
    mp_set_i32(&a, -123)
    testing.expect(t, mp_get_double(&a) == -123.0)
}

// ==================== 便捷初始化 ====================
@(test)
test_convenience_inits :: proc(t: ^testing.T) {
    a: mp_int

    // init_u32
    err := mp_init_u32(&a, 42); if !expect_ok(t, err, "init_u32")  do return
    defer mp_clear(&a)
    testing.expect(t, mp_get_mag_u32(&a) == 42)
    mp_clear(&a)

    // init_i32 negative
    err = mp_init_i32(&a, -99); if !expect_ok(t, err, "init_i32 -99")  do return
    testing.expect(t, mp_get_i32(&a) == -99)
    mp_clear(&a)

    // init_u64 large
    err = mp_init_u64(&a, u64(1) << 50); if !expect_ok(t, err, "init_u64 2^50")  do return
    testing.expect(t, mp_get_mag_u64(&a) == (u64(1) << 50))
    mp_clear(&a)

    // init_i64 negative large
    err = mp_init_i64(&a, -(1 << 40)); if !expect_ok(t, err, "init_i64 -2^40")  do return
    testing.expect(t, mp_get_i64(&a) == -(1 << 40))
    mp_clear(&a)

    // init_l
    err = mp_init_l(&a, 777); if !expect_ok(t, err, "init_l")  do return
    testing.expect(t, mp_get_l(&a) == 777)
    mp_clear(&a)

    // init_ul
    err = mp_init_ul(&a, 888); if !expect_ok(t, err, "init_ul")  do return
    testing.expect(t, int(mp_get_mag_ul(&a)) == 888)
    mp_clear(&a)

    // init_set
    err = mp_init_set(&a, 255); if !expect_ok(t, err, "init_set")  do return
    testing.expect(t, mp_cmp_d(&a, 255) == .MP_EQ)
    mp_clear(&a)

    // init_copy
    b: mp_int; defer mp_clear(&b)
    err = mp_init(&b); if !expect_ok(t, err, "init b")  do return
    mp_set(&b, 12345)
    err = mp_init_copy(&a, &b); if !expect_ok(t, err, "init_copy")  do return
    testing.expect(t, mp_cmp(&a, &b) == .MP_EQ)
    mp_clear(&a)

    // init_multi and clear_multi
    x, y, z: mp_int
    err = mp_init_multi(&x, &y, &z, nil); if !expect_ok(t, err, "init_multi")  do return
    defer mp_clear_multi(&x, &y, &z, nil)
    mp_set(&x, 1); mp_set(&y, 2); mp_set(&z, 3)
    testing.expect(t, mp_cmp_d(&x, 1) == .MP_EQ && mp_cmp_d(&y, 2) == .MP_EQ && mp_cmp_d(&z, 3) == .MP_EQ)
}

// ==================== 错误字符串 ====================
@(test)
test_error_to_string :: proc(t: ^testing.T) {
    testing.expect(t, mp_error_to_string(.MP_OKAY) == "Successful")
    testing.expect(t, mp_error_to_string(.MP_MEM) == "Out of heap")
    testing.expect(t, mp_error_to_string(.MP_VAL) == "Value out of range")
    testing.expect(t, mp_error_to_string(.MP_BUF) == "Buffer overflow")
    testing.expect(t, mp_error_to_string(.MP_OVF) == "Integer overflow")
    testing.expect(t, mp_error_to_string(.MP_ITER) == "Max. iterations reached")
    testing.expect(t, mp_error_to_string(.MP_ERR) == "Unknown error")
}

// ==================== get_l / get_mag_ul ====================
@(test)
test_get_set_l_ul :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    mp_set_l(&a, -500)
    testing.expect(t, mp_get_l(&a) == -500)

    mp_set_ul(&a, 1000)
    testing.expect(t, uint(1000) == mp_get_mag_ul(&a))

    mp_set_i64(&a, -9999)
    testing.expect(t, mp_get_l(&a) == -9999)
}

// ==================== ubin / sbin 转换 ====================
@(test)
test_ubin_sbin :: proc(t: ^testing.T) {
    a, b: mp_int
    defer mp_clear(&a); defer mp_clear(&b)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return

    // ubin roundtrip
    mp_set(&a, 0xDEADBEEF)
    size := mp_ubin_size(&a)
    testing.expect(t, size == 4)
    ub, err2 := mp_to_ubin(&a)
    if !expect_ok(t, err2, "to_ubin")  do return
    defer delete(ub)
    err = mp_from_ubin(&b, ub); if !expect_ok(t, err, "from_ubin")  do return
    expect_eq(t, &a, &b, "ubin roundtrip")

    // sbin roundtrip
    a.sign = .MP_NEG
    sb, err3 := mp_to_sbin(&a)
    if !expect_ok(t, err3, "to_sbin")  do return
    defer delete(sb)
    mp_zero(&b)
    err = mp_from_sbin(&b, sb); if !expect_ok(t, err, "from_sbin")  do return
    expect_eq(t, &a, &b, "sbin roundtrip")
    testing.expect(t, mp_sbin_size(&a) == 1 + mp_ubin_size(&a))

    // sbin zero
    mp_zero(&a)
    sz, err4 := mp_to_sbin(&a)
    if !expect_ok(t, err4, "to_sbin 0")  do return
    defer delete(sz)
    mp_zero(&b)
    err = mp_from_sbin(&b, sz); if !expect_ok(t, err, "from_sbin 0")  do return
    testing.expect(t, mp_cmp_d(&b, 0) == .MP_EQ)
}

// ==================== radix_size ====================
@(test)
test_radix_size :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    // zero
    sz, err2 := mp_radix_size(&a, 10); if !expect_ok(t, err2, "radix_size 0")  do return
    testing.expect(t, sz == 2)

    // positive
    mp_set(&a, 1000)
    sz, err2 = mp_radix_size(&a, 10); if !expect_ok(t, err2, "radix_size 1000")  do return
    testing.expect(t, sz >= 4) // "1000" + null

    // negative
    a.sign = .MP_NEG
    sz, err2 = mp_radix_size(&a, 10); if !expect_ok(t, err2, "radix_size -1000")  do return
    testing.expect(t, sz >= 5) // "-1000" + null

    // overestimate
    osz, err3 := mp_radix_size_overestimate(&a, 10); if !expect_ok(t, err3, "overestimate")  do return
    testing.expect(t, osz >= sz)

    // power-of-two radix shortcut
    osz, err3 = mp_radix_size_overestimate(&a, 16); if !expect_ok(t, err3, "overestimate 16")  do return
    testing.expect(t, osz > 0)
}

// ==================== log_n ====================
@(test)
test_log_n :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    // log_2(256) = 8
    mp_set(&a, 256)
    log: int
    err = mp_log_n(&a, 2, &log); if !expect_ok(t, err, "log2(256)")  do return
    testing.expect(t, log == 8)

    // log_10(1000) = 3
    mp_set(&a, 1000)
    err = mp_log_n(&a, 10, &log); if !expect_ok(t, err, "log10(1000)")  do return
    testing.expect(t, log == 3)

    // log_10(1) = 0
    mp_set(&a, 1)
    err = mp_log_n(&a, 10, &log); if !expect_ok(t, err, "log10(1)")  do return
    testing.expect(t, log == 0)

    // single-digit above base
    mp_set(&a, 127)
    err = mp_log_n(&a, 10, &log); if !expect_ok(t, err, "log10(127)")  do return
    testing.expect(t, log == 2) // 10^2=100 <= 127 < 10^3=1000

    // 非法: 负数的 log
    a.sign = .MP_NEG
    err = mp_log_n(&a, 2, &log)
    testing.expect(t, err == .MP_VAL)

    // 非法: base < 2
    mp_set(&a, 100)
    err = mp_log_n(&a, 1, &log)
    testing.expect(t, err == .MP_VAL)
}

// ==================== sqrt 大数 ====================
@(test)
test_sqrt_large :: proc(t: ^testing.T) {
    x, r: mp_int
    defer mp_clear(&x); defer mp_clear(&r)

    err := mp_init(&x); if !expect_ok(t, err, "init x")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return

    // sqrt(2^30) must be 2^15
    mp_2expt(&x, 60)
    err = mp_sqrt(&x, &r); if !expect_ok(t, err, "sqrt 2^60")  do return
    testing.expect(t, mp_get_i64(&r) == 1 << 30, "sqrt 2^60")

    // sqrt of 1
    mp_set(&x, 1)
    err = mp_sqrt(&x, &r); if !expect_ok(t, err, "sqrt 1")  do return
    testing.expect(t, mp_cmp_d(&r, 1) == .MP_EQ)

    // sqrt of 0
    mp_zero(&x)
    err = mp_sqrt(&x, &r); if !expect_ok(t, err, "sqrt 0")  do return
    testing.expect(t, mp_cmp_d(&r, 0) == .MP_EQ)
}

// ==================== n 次方根边界 ====================
@(test)
test_root_n_edge :: proc(t: ^testing.T) {
    a, r: mp_int
    defer mp_clear(&a); defer mp_clear(&r)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return

    // 1^{anything} = 1
    mp_set(&a, 1)
    err = mp_root_n(&a, 7, &r); if !expect_ok(t, err, "root7 1")  do return
    testing.expect(t, mp_cmp_d(&r, 1) == .MP_EQ, "root7 1")

    // 0^{anything} = 0
    mp_zero(&a)
    err = mp_root_n(&a, 3, &r); if !expect_ok(t, err, "root3 0")  do return
    testing.expect_value(t, mp_get_i64(&r), 0)

    // 小数的根
    mp_set(&a, 27)
    err = mp_root_n(&a, 3, &r); if !expect_ok(t, err, "cbrt 27")  do return
    testing.expect(t, mp_cmp_d(&r, 3) == .MP_EQ, "cbrt 27")

    // 偶数次根检查: root_n(x, 2, ...) 对正数等于 sqrt
    mp_set(&a, 81)
    err = mp_root_n(&a, 2, &r); if !expect_ok(t, err, "root2 81")  do return
    testing.expect(t, mp_cmp_d(&r, 9) == .MP_EQ, "root2 81")

    // 偶数根负数报错
    a.sign = .MP_NEG
    err = mp_root_n(&a, 2, &r)
    testing.expect(t, err == .MP_VAL)

    // b < 0
    mp_set(&a, 8)
    err = mp_root_n(&a, -1, &r)
    testing.expect(t, err == .MP_VAL)
}

// ==================== 比较边界 ====================
@(test)
test_comparisons_ext :: proc(t: ^testing.T) {
    a, b: mp_int
    defer mp_clear(&a); defer mp_clear(&b)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return

    // cmp_d on larger-than-single-digit
    mp_set_u64(&a, u64(1) << 32)
    testing.expect(t, mp_cmp_d(&a, 0) == .MP_GT)
    testing.expect(t, mp_cmp_d(&a, 1) == .MP_GT)

    // cmp_d on zero
    mp_zero(&a)
    testing.expect(t, mp_cmp_d(&a, 0) == .MP_EQ)
    testing.expect(t, mp_cmp_d(&a, 1) == .MP_LT)

    // cmp with different lengths
    mp_set(&a, 1)
    mp_set_u64(&b, u64(1) << 32)
    testing.expect(t, mp_cmp(&a, &b) == .MP_LT)
    testing.expect(t, mp_cmp_mag(&a, &b) == .MP_LT)

    // both negative, different lengths
    mp_set_i32(&a, -1)
    mp_set_i64(&b, -(1 << 40))
    testing.expect(t, mp_cmp(&a, &b) == .MP_GT) // -1 > -(1<<40)
}

// ==================== 除法自引用 ====================
@(test)
test_div_self_reference :: proc(t: ^testing.T) {
    a, b, q, r: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&q); defer mp_clear(&r)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&q);  if !expect_ok(t, err, "init q")  do return
    err = mp_init(&r);  if !expect_ok(t, err, "init r")  do return

    // 被除数小于除数
    mp_set(&a, 5); mp_set(&b, 100)
    err = mp_div(&a, &b, &q, &r); if !expect_ok(t, err, "5/100")  do return
    testing.expect(t, mp_cmp_d(&q, 0) == .MP_EQ)
    testing.expect(t, mp_cmp_d(&r, 5) == .MP_EQ)

    // 被除数等于除数
    mp_set(&a, 42); mp_set(&b, 42)
    err = mp_div(&a, &b, &q, &r); if !expect_ok(t, err, "42/42")  do return
    testing.expect(t, mp_cmp_d(&q, 1) == .MP_EQ)
    testing.expect(t, mp_cmp_d(&r, 0) == .MP_EQ)

    // nil q, only remainder
    mp_set(&a, 1000); mp_set(&b, 17)
    err = mp_div(&a, &b, nil, &r); if !expect_ok(t, err, "1000/17 nil q")  do return
    testing.expect(t, mp_cmp_d(&r, 1000 % 17) == .MP_EQ)

    // nil r
    err = mp_div(&a, &b, &q, nil); if !expect_ok(t, err, "1000/17 nil r")  do return
    testing.expect(t, mp_cmp_d(&q, 1000 / 17) == .MP_EQ)
}

// ==================== Pack/Unpack 变种 ====================
@(test)
test_pack_unpack_variants :: proc(t: ^testing.T) {
    a, b: mp_int
    defer mp_clear(&a); defer mp_clear(&b)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return

    // LSB_FIRST + LITTLE_ENDIAN
    mp_set(&a, 0x12345678)
    buf := make([]u8, 8)
    defer delete(buf)
    written, err2 := mp_pack(buf, .MP_LSB_FIRST, 4, .MP_LITTLE_ENDIAN, 0, &a)
    if !expect_ok(t, err2, "pack lsb+le")  do return
    err = mp_unpack(&b, written, .MP_LSB_FIRST, 4, .MP_LITTLE_ENDIAN, 0, buf[:written * 4])
    if !expect_ok(t, err, "unpack lsb+le")  do return
    testing.expectf(t, mp_cmp_d(&b, 0x12345678) == .MP_EQ,
    "lsb+le roundtrip: expected 0x12345678, got %v", mp_to_string(&b))

    // MSB_FIRST + BIG_ENDIAN
    mp_set(&a, 0xDEAD)
    written, err2 = mp_pack(buf, .MP_MSB_FIRST, 2, .MP_BIG_ENDIAN, 0, &a)
    if !expect_ok(t, err2, "pack msb+be")  do return
    mp_zero(&b)
    err = mp_unpack(&b, written, .MP_MSB_FIRST, 2, .MP_BIG_ENDIAN, 0, buf[:written * 2])
    if !expect_ok(t, err, "unpack msb+be")  do return
    testing.expectf(t, mp_cmp_d(&b, 0xDEAD) == .MP_EQ,
    "msb+be roundtrip: expected 0xDEAD, got %v", mp_to_string(&b))

    // pack with nails
    mp_set(&a, 0x55)
    written, err2 = mp_pack(buf, .MP_MSB_FIRST, 1, .MP_BIG_ENDIAN, 2, &a)
    if !expect_ok(t, err2, "pack nails=2")  do return

    // pack_buffer overflow
    small := make([]u8, 1)
    defer delete(small)
    mp_set(&a, 0xFFFFF)
    _, err2 = mp_pack(small, .MP_MSB_FIRST, 4, .MP_BIG_ENDIAN, 0, &a)
    testing.expect(t, err2 == .MP_BUF)
}

// ==================== 模反向扩展 ====================
@(test)
test_invmod_extended :: proc(t: ^testing.T) {
    a, m, inv: mp_int
    defer mp_clear(&a); defer mp_clear(&m); defer mp_clear(&inv)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&m);  if !expect_ok(t, err, "init m")  do return
    err = mp_init(&inv); if !expect_ok(t, err, "init inv")  do return

    // invmod with non-coprime: gcd(2,4)=2
    mp_set(&a, 2); mp_set(&m, 4)
    err = mp_invmod(&a, &m, &inv)
    testing.expect(t, err == .MP_VAL)

    // invmod negative modulus
    mp_set(&a, 3); mp_set_i32(&m, -11)
    err = mp_invmod(&a, &m, &inv)
    testing.expect(t, err == .MP_VAL)
    mp_zero(&inv)

    // invmod where a > m
    mp_set(&a, 14); mp_set(&m, 11)
    err = mp_invmod(&a, &m, &inv); if !expect_ok(t, err, "invmod(14,11)")  do return
    // 3 mod 11, inverse = 4
    testing.expect(t, mp_cmp_d(&inv, 4) == .MP_EQ, "invmod(14,11)")

    // invmod negative a
    mp_set_i32(&a, -3); mp_set(&m, 11)
    err = mp_invmod(&a, &m, &inv); if !expect_ok(t, err, "invmod(-3,11)")  do return
    vv := mp_get_i64(&inv)
    testing.expect(t, vv == -7, "invmod(-3,11) = -7") // -3 mod 11 = 8, inv of 8 is 7
}

// ==================== 比较自引用 ====================
@(test)
test_cmp_self :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    mp_set(&a, 12345)
    testing.expect(t, mp_cmp(&a, &a) == .MP_EQ)
    testing.expect(t, mp_cmp_mag(&a, &a) == .MP_EQ)
}

// ==================== 大数乘法: Toom-Cook ====================
@(test)
test_mul_toom :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // 超过 MP_MUL_TOOM_CUTOFF (350) 位的乘法会触发 Toom-Cook-3
    mp_2expt(&a, 352)
    mp_2expt(&b, 355)
    err = mp_mul(&a, &b, &c); if !expect_ok(t, err, "2^352 * 2^355")  do return
    testing.expect(t, mp_count_bits(&c) == 352 + 355 + 1) // or 708

    // 验证: c / a = b
    q: mp_int; defer mp_clear(&q)
    err = mp_init(&q); if !expect_ok(t, err, "init q")  do return
    err = mp_div(&c, &a, &q, nil); if !expect_ok(t, err, "mp_div")  do return
    expect_eq(t, &b, &q, "c/a == b")
}

// ==================== 大数乘法: Balance ====================
@(test)
test_mul_balance :: proc(t: ^testing.T) {
    a, b, c: mp_int
    defer mp_clear(&a); defer mp_clear(&b); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&b);  if !expect_ok(t, err, "init b")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    // Balance: a 大很多位数超过 b, min >= karatsuba cutoff, max / 2 >= karatsuba cutoff
    mp_2expt(&a, 400) // ~15 digits
    mp_set_u64(&b, u64(1234567890123456)) // fits in 64 bits
    err = mp_mul(&a, &b, &c); if !expect_ok(t, err, "balance mul")  do return

    q: mp_int; defer mp_clear(&q)
    err = mp_init(&q); if !expect_ok(t, err, "init q")  do return
    err = mp_div(&c, &a, &q, nil); if !expect_ok(t, err, "mp_div")  do return
    expect_eq(t, &b, &q, "c/a == b")
}

// ==================== Pack 计数 ====================
@(test)
test_pack_count :: proc(t: ^testing.T) {
    a: mp_int
    defer mp_clear(&a)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return

    mp_set(&a, 0)
    testing.expect_value(t, mp_pack_count(&a, 0, 1), 0)

    mp_set(&a, 255)
    cnt := mp_pack_count(&a, 0, 1)
    testing.expect_value(t, cnt, 1)

    mp_set(&a, 0xFFFF)
    cnt = mp_pack_count(&a, 0, 1)
    testing.expect_value(t, cnt, 2)
}