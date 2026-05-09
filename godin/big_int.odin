package godin

import "core:fmt"
import "core:math/bits"

// --------------- Type alias ---------------
BigInt :: mp_int

// --------------- Constants ---------------
MAX_BIG_INT_SHIFT :: 1024
BIG_INT_STRING_BUF_SIZE :: 256
BIG_INT_DEFAULT_BASE :: u64(10)
BIG_INT_MAX_EXPONENT :: u64(308)
BIG_INT_LARGE_THRESHOLD :: 498  // 2^498 ~ 10^150, switch to scientific notation

// --------------- Helpers ---------------

mp_iszero :: proc(a: ^mp_int) -> bool {
    return a.used == 0
}

mp_get_u64 :: proc(a: ^mp_int) -> u64 {
    return mp_get_mag_u64(a)
}

mp_get_u32 :: proc(a: ^mp_int) -> u32 {
    return mp_get_mag_u32(a)
}

@(private="file")
mp_decr :: proc(a: ^mp_int) -> mp_err {
    return mp_sub_d(a, 1, a)
}

@(private="file")
u64_digit_value :: proc(r: rune) -> u64 {
    switch r {
    case '0' ..= '9': return u64(r - '0')
    case 'a' ..= 'z': return u64(r - 'a' + 10)
    case 'A' ..= 'Z': return u64(r - 'A' + 10)
    }
    return 0xff
}

digit_to_char :: proc(digit: u8) -> u8 {
    if digit <= 9 {
        return digit + '0'
    }
    else if digit <= 15 {
        return digit + 'a' - 10
    }
    return '0'
}

leading_zeros_u64 :: proc(x: u64) -> u64 {
    if x == 0 do return 64
    return u64(bits.count_leading_zeros(x))
}

// --------------- Initialization / Deallocation ---------------

big_int_from_u64 :: proc(dst: ^BigInt, x: u64) {
    mp_init_u64(dst, x)
}

big_int_from_i64 :: proc(dst: ^BigInt, x: i64) {
    mp_init_i64(dst, x)
}

big_int_init :: proc(dst: ^BigInt, src: ^BigInt) {
    if dst == src {
        return
    }
    mp_init_copy(dst, src)
}

big_int_dealloc :: proc(dst: ^BigInt) {
    mp_clear(dst)
}

// --------------- Make (constructors returning BigInt) ---------------

big_int_make :: proc(b: ^BigInt, abs: bool) -> BigInt {
    i: BigInt
    big_int_init(&i, b)
    if abs {
        mp_abs(&i, &i)
    }
    return i
}

big_int_make_abs :: proc(b: ^BigInt) -> BigInt {
    return big_int_make(b, true)
}

big_int_make_u64 :: proc(x: u64) -> BigInt {
    i: BigInt
    big_int_from_u64(&i, x)
    return i
}

big_int_make_i64 :: proc(x: i64) -> BigInt {
    i: BigInt
    big_int_from_i64(&i, x)
    return i
}

// --------------- From string ---------------

big_int_from_string :: proc(dst: ^BigInt, s: string, success: ^bool) {
    success^ = true

    is_negative := false
    base := u64(10)
    has_prefix := false

    if len(s) > 2 && s[0] == '0' {
        switch s[1] {
        case 'b': base = 2;  has_prefix = true
        case 'o': base = 8;  has_prefix = true
        case 'd': base = 10; has_prefix = true
        case 'z': base = 12; has_prefix = true
        case 'x': base = 16; has_prefix = true
        case 'h': base = 16; has_prefix = true
        }
    }

    text := s
    if has_prefix {
        text = text[2:]
    }

    b: BigInt
    big_int_from_u64(&b, base)
    defer big_int_dealloc(&b)

    mp_zero(dst)

    digit: BigInt
    defer big_int_dealloc(&digit)

    i := 0
    for i < len(text) {
        r := rune(text[i])

        if r == '-' {
            if is_negative {
                success^ = false
                return
            }
            is_negative = true
            i += 1
            continue
        }

        if r == '_' {
            i += 1
            continue
        }
        v := u64_digit_value(r)
        if v >= base {
            if r != 'e' && r != 'E' {
                success^ = false
            }
            break
        }

        big_int_from_u64(&digit, v)
        big_int_mul_eq(dst, &b)
        big_int_add_eq(dst, &digit)
        i += 1
    }

    if i < len(text) && (text[i] == 'e' || text[i] == 'E') {
        i += 1
        assert(base == 10)
        assert(text[i] != '-')
        if text[i] == '+' {
            i += 1
        }

        exp: u64
        for i < len(text) {
            r := rune(text[i])
            if r == '_' {
                i += 1; continue
            }
            v: u64
            if r >= '0' && r <= '9' {
                v = u64_digit_value(r)
            } else {
                success^ = false
                break
            }
            exp *= 10
            exp += v
            i += 1
        }

        if exp > BIG_INT_MAX_EXPONENT {
            success^ = false
            return
        }

        tmp: BigInt
        mp_init(&tmp)
        defer big_int_dealloc(&tmp)
        big_int_exp_u64(&tmp, &b, exp, success)
        big_int_mul_eq(dst, &tmp)
    }

    if is_negative {
        big_int_neg(dst, dst)
    }
}

// --------------- Queries ---------------

big_int_can_be_represented_in_64_bits :: proc(x: ^BigInt) -> bool {
    bits_used := (x.used - 1) * DIGIT_BIT
    return bits_used <= 64
}

big_int_sign :: proc(x: ^BigInt) -> i64 {
    if mp_iszero(x) {
        return 0
    }
    return x.sign == .MP_ZPOS ? + 1 : -1
}

big_int_cmp :: proc(x: ^BigInt, y: ^BigInt) -> mp_ord {
    return mp_cmp(x, y)
}

big_int_cmp_zero :: proc(x: ^BigInt) -> int {
    if mp_iszero(x) {
        return 0
    }
    return x.sign == .MP_ZPOS ? + 1 : -1
}

big_int_is_zero :: proc(x: ^BigInt) -> bool {
    return mp_iszero(x)
}

big_int_is_neg :: proc(x: ^BigInt) -> bool {
    if x == nil {
        return false
    }
    return x.sign != .MP_ZPOS
}

big_int_log2 :: proc(x: ^BigInt) -> int {
    return mp_count_bits(x) - 1
}

// --------------- Conversion to primitive types ---------------

big_int_to_u64 :: proc(x: ^BigInt) -> u64 {
    assert(x.sign == .MP_ZPOS)
    return mp_get_mag_u64(x)
}

big_int_to_i64 :: proc(x: ^BigInt) -> i64 {
    return mp_get_i64(x)
}

big_int_to_f64 :: proc(x: ^BigInt) -> f64 {
    return mp_get_double(x)
}

// --------------- Negation ---------------

big_int_neg :: proc(dst: ^BigInt, x: ^BigInt) {
    mp_neg(x, dst)
}

// --------------- Arithmetic ---------------

big_int_add :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    mp_add(x, y, dst)
}

big_int_sub :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    mp_sub(x, y, dst)
}

big_int_shl :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    yy := mp_get_mag_u32(y)
    mp_mul_2d(x, int(yy), dst)
}

big_int_shr :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    yy := mp_get_mag_u32(y)
    d: BigInt
    mp_div_2d(x, int(yy), dst, &d)
    big_int_dealloc(&d)
}

big_int_mul_u64 :: proc(dst: ^BigInt, x: ^BigInt, y: u64) {
    d: BigInt
    big_int_from_u64(&d, y)
    mp_mul(x, &d, dst)
    big_int_dealloc(&d)
}

big_int_exp_u64 :: proc(dst: ^BigInt, x: ^BigInt, y: u64, success: ^bool) {
    if y > u64(max(int)) {
        success^ = false
        return
    }
    mp_init(dst)
    err := mp_expt_n(x, int(y), dst)
    success^ = err == .MP_OKAY
}

big_int_mul :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    mp_mul(x, y, dst)
}

// --------------- Division ---------------

big_int_quo_rem :: proc(x: ^BigInt, y: ^BigInt, q: ^BigInt, r: ^BigInt) {
    mp_div(x, y, q, r)
}

big_int_quo :: proc(z: ^BigInt, x: ^BigInt, y: ^BigInt) {
    r: BigInt
    big_int_quo_rem(x, y, z, &r)
    big_int_dealloc(&r)
}

big_int_rem :: proc(z: ^BigInt, x: ^BigInt, y: ^BigInt) {
    q: BigInt
    big_int_quo_rem(x, y, &q, z)
    big_int_dealloc(&q)
}

big_int_euclidean_mod :: proc(z: ^BigInt, x: ^BigInt, y: ^BigInt) {
    y0: BigInt
    big_int_init(&y0, y)
    defer big_int_dealloc(&y0)

    q: BigInt
    big_int_quo_rem(x, y, &q, z)
    defer big_int_dealloc(&q)

    if z.sign == .MP_NEG {
        if y0.sign == .MP_NEG {
            big_int_sub(z, z, &y0)
        } else {
            big_int_add(z, z, &y0)
        }
    }
}

// --------------- Bitwise operations ---------------

big_int_and :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    mp_and(x, y, dst)
}

big_int_and_not :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    if mp_iszero(x) {
        big_int_init(dst, y)
        return
    }
    if mp_iszero(y) {
        big_int_init(dst, x)
        return
    }

    if x.sign == y.sign {
        if x.sign == .MP_NEG {
        // (-x) &~ (-y) == ~(x-1) &~ ~(y-1) == ~(x-1) & (y-1) == (y-1) &~ (x-1)
            x1 := big_int_make_abs(x); defer big_int_dealloc(&x1)
            y1 := big_int_make_abs(y); defer big_int_dealloc(&y1)
            mp_decr(&x1)
            mp_decr(&y1)

            ny1: BigInt; defer big_int_dealloc(&ny1)
            mp_complement(&y1, &ny1)
            mp_and(&x1, &ny1, dst)
            return
        }

        ny: BigInt; defer big_int_dealloc(&ny)
        mp_complement(y, &ny)
        mp_and(x, &ny, dst)
        return
    }

    if x.sign == .MP_NEG {
    // (-x) &~ y == ~(x-1) &~ y == ~(x-1) & ~y == ~((x-1) | y) == -(((x-1) | y) + 1)
        x1 := big_int_make_abs(x); defer big_int_dealloc(&x1)
        y1 := big_int_make_abs(y); defer big_int_dealloc(&y1)
        mp_decr(&x1)

        z1: BigInt; defer big_int_dealloc(&z1)
        big_int_or(&z1, &x1, &y1)
        mp_add_d(&z1, 1, dst)
        return
    }

    // x &~ (-y) == x &~ ~(y-1) == x & (y-1)
    x1 := big_int_make_abs(x); defer big_int_dealloc(&x1)
    y1 := big_int_make_abs(y); defer big_int_dealloc(&y1)
    mp_decr(&y1)
    big_int_and(dst, &x1, &y1)
}

big_int_xor :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    mp_xor(x, y, dst)
}

big_int_or :: proc(dst: ^BigInt, x: ^BigInt, y: ^BigInt) {
    mp_or(x, y, dst)
}

big_int_not :: proc(dst: ^BigInt, x: ^BigInt, bit_count: i32, is_signed: bool) {
    assert(bit_count >= 0)
    if bit_count == 0 {
        big_int_from_u64(dst, 0)
        return
    }
    if big_int_is_neg(x) {
    // ~x == -x - 1
        big_int_neg(dst, x)
        mp_decr(dst)
        mp_mod_2d(dst, int(bit_count), dst)
        return
    }

    pow2b: BigInt; defer big_int_dealloc(&pow2b)
    mp_2expt(&pow2b, int(bit_count))

    mask: BigInt; defer big_int_dealloc(&mask)
    mp_2expt(&mask, int(bit_count))
    mp_decr(&mask)

    v: BigInt; defer big_int_dealloc(&v)
    mp_init_copy(&v, x)
    mp_mod_2d(&v, int(bit_count), &v)

    mp_xor(&v, &mask, dst)

    if is_signed {
        pmask: BigInt; defer big_int_dealloc(&pmask)
        pmask_minus_one: BigInt; defer big_int_dealloc(&pmask_minus_one)
        mp_2expt(&pmask, int(bit_count - 1))
        mp_sub_d(&pmask, 1, &pmask_minus_one)

        a: BigInt; defer big_int_dealloc(&a)
        b: BigInt; defer big_int_dealloc(&b)
        big_int_and(&a, dst, &pmask_minus_one)
        big_int_and(&b, dst, &pmask)
        big_int_sub(dst, &a, &b)
    }
}

// --------------- In-place arithmetic (_eq variants) ---------------

big_int_add_eq :: proc(dst: ^BigInt, x: ^BigInt) {
    res: BigInt
    big_int_init(&res, dst)
    defer big_int_dealloc(&res)
    big_int_add(dst, &res, x)
}

big_int_sub_eq :: proc(dst: ^BigInt, x: ^BigInt) {
    res: BigInt
    big_int_init(&res, dst)
    defer big_int_dealloc(&res)
    big_int_sub(dst, &res, x)
}

big_int_shl_eq :: proc(dst: ^BigInt, x: ^BigInt) {
    res: BigInt
    big_int_init(&res, dst)
    defer big_int_dealloc(&res)
    big_int_shl(dst, &res, x)
}

big_int_shr_eq :: proc(dst: ^BigInt, x: ^BigInt) {
    res: BigInt
    big_int_init(&res, dst)
    defer big_int_dealloc(&res)
    big_int_shr(dst, &res, x)
}

big_int_mul_eq :: proc(dst: ^BigInt, x: ^BigInt) {
    res: BigInt
    big_int_init(&res, dst)
    defer big_int_dealloc(&res)
    big_int_mul(dst, &res, x)
}

big_int_quo_eq :: proc(dst: ^BigInt, x: ^BigInt) {
    res: BigInt
    big_int_init(&res, dst)
    defer big_int_dealloc(&res)
    big_int_quo(dst, &res, x)
}

big_int_rem_eq :: proc(dst: ^BigInt, x: ^BigInt) {
    res: BigInt
    big_int_init(&res, dst)
    defer big_int_dealloc(&res)
    big_int_rem(dst, &res, x)
}

// --------------- To string ---------------

big_int_to_string :: proc(x: ^BigInt, base: u64 = BIG_INT_DEFAULT_BASE, allocator := context.allocator) -> string {
    assert(base <= 16)

    if mp_iszero(x) {
        buf := make([]u8, 1, allocator)
        buf[0] = '0'
        return string(buf)
    }

    buf: [dynamic]u8
    // Use the provided allocator for the dynamic array backing store.
    // [dynamic] uses context.allocator by default; we set it explicitly.
    context.allocator = allocator

    if x.used >= BIG_INT_LARGE_THRESHOLD {
        val: BigInt; defer big_int_dealloc(&val)
        mp_abs(x, &val)
        exp: int
        err := mp_log_n(&val, 10, &exp)
        assert(err == .MP_OKAY)
        assert(exp >= 100)

        thousand_below: BigInt; defer big_int_dealloc(&thousand_below)
        thousand_above: BigInt; defer big_int_dealloc(&thousand_above)
        mp_init_i32(&thousand_below, 10)

        mp_expt_n(&thousand_below, exp - 3, &thousand_below)
        mp_div(&val, &thousand_below, &thousand_above, nil)

        mant := 1.0e-3 * mp_get_double(&thousand_above)

        sign_str := "-" if x.sign == .MP_NEG else ""
        s := fmt.aprintf("~ %s%.0fe%d", sign_str, mant, exp, allocator = allocator)
        return s
    } else {
        v: BigInt; defer big_int_dealloc(&v)
        mp_init_copy(&v, x)

        if v.sign == .MP_NEG {
            append(&buf, '-')
            mp_abs(&v, &v)
        }

        first_word_idx := len(buf)

        r: BigInt; defer big_int_dealloc(&r)
        b: BigInt; defer big_int_dealloc(&b)
        big_int_from_u64(&b, base)

        for big_int_cmp(&v, &b) != .MP_LT {
            big_int_quo_rem(&v, &b, &v, &r)
            digit := u8(mp_get_mag_u64(&r))
            append(&buf, digit_to_char(digit))
        }

        big_int_rem(&r, &v, &b)
        digit := u8(mp_get_mag_u64(&r))
        append(&buf, digit_to_char(digit))

        // Reverse the digit portion
        for i, j := first_word_idx, len(buf) - 1; i < j; i, j = i + 1, j - 1 {
            buf[i], buf[j] = buf[j], buf[i]
        }

        return string(buf[:])
    }
}

// --------------- Debug ---------------

debug_print_big_int :: proc(x: ^BigInt) {
    s := big_int_to_string(x, 10, context.temp_allocator)
    fmt.eprintf("[DEBUG] %s\n", s)
}
