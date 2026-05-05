// libtommath_test.odin – 固定后的完整单元测试
package godin

import "core:testing"
import "core:math"

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
    err = mp_init_size(&a, max(int)/2 + 1)
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
    testing.expectf(t, mp_cmp_d(&c, 45) == .MP_EQ && c.sign == .MP_NEG,
                     "-100+55: expected -45, got sign=%v value=%s", c.sign, mp_to_string(&c))

    // 正数减单数字：100 - 30 = 70
    mp_set(&a, 100)
    err = mp_sub_d(&a, 30, &c); if !expect_ok(t, err, "100-30")  do return
    testing.expectf(t, mp_cmp_d(&c, 70) == .MP_EQ, "100-30: expected 70, got %v", mp_to_string(&c))

    // 负数减单数字：-100 - 30 = -130
    a.sign = .MP_NEG
    err = mp_sub_d(&a, 30, &c); if !expect_ok(t, err, "-100-30")  do return
    testing.expectf(t, mp_cmp_d(&c, 130) == .MP_EQ && c.sign == .MP_NEG,
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
    testing.expectf(t, mp_cmp_d(&c, 123*456) == .MP_EQ,
                     "123*456: expected %v, got %v", 123*456, mp_to_string(&c))

    // 平方
    err = mp_mul(&a, &a, &c); if !expect_ok(t, err, "123^2")  do return
    testing.expectf(t, mp_cmp_d(&c, 123*123) == .MP_EQ,
                     "123^2: expected %v, got %v", 123*123, mp_to_string(&c))
}

@(test)
test_mul_d :: proc(t: ^testing.T) {
    a, c: mp_int
    defer mp_clear(&a); defer mp_clear(&c)

    err := mp_init(&a); if !expect_ok(t, err, "init a")  do return
    err = mp_init(&c);  if !expect_ok(t, err, "init c")  do return

    mp_set(&a, 65537)
    err = mp_mul_d(&a, 2, &c); if !expect_ok(t, err, "65537*2")  do return
    testing.expectf(t, mp_cmp_d(&c, 65537*2) == .MP_EQ,
                     "65537*2: expected %v, got %v", 65537*2, mp_to_string(&c))

    // 乘以 2 的幂：65537 * 8 = 524296
    err = mp_mul_d(&a, 8, &c); if !expect_ok(t, err, "65537*8")  do return
    testing.expectf(t, mp_cmp_d(&c, 65537*8) == .MP_EQ,
                     "65537*8: expected %v, got %v", 65537*8, mp_to_string(&c))
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
    testing.expectf(t, val == (u64(1) << 62), "2^62: expected %v, got %v", u64(1)<<62, val)

    // 再乘 2
    err = mp_mul_2(&c, &c); if !expect_ok(t, err, "mul 2")  do return
    val = mp_get_mag_u64(&c)
    testing.expect(t, val == (u64(1)<<62)*2)
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
    err = mp_mod_2d(&a, a.used*28+1, &c); if !expect_ok(t, err, "mod large")  do return
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
    testing.expectf(t, mp_cmp_d(&q, 1000/7) == .MP_EQ,
                     "quot: expected %v, got %v", 1000/7, mp_to_string(&q))
    testing.expectf(t, rem == 1000%7, "rem: expected %v, got %v", 1000%7, rem)

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
    testing.expectf(t, mp_cmp_d(&c, 1) == .MP_EQ && c.sign == .MP_NEG,
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
    testing.expect(t, math.abs(d - 3.14159) < 1e-9)
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
    defer delete(str)
    testing.expectf(t, err2 == .MP_OKAY && str == "12345",
                     "dec roundtrip: expected 12345, got %v (err=%v)", str, err2)

    // 十六进制
    err = mp_read_radix(&a, "1A2B3C", 16); if !expect_ok(t, err, "read hex")  do return
    hex, _ := mp_to_radix(&a, 16)
    defer delete(hex)
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
}