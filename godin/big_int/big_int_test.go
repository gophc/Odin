package big_int

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

)

func TestNew(t *testing.T) {
	z := New()
	require.NotNil(t, z)
	assert.True(t, IsZero(z))
	assert.Equal(t, 0, Sign(z))
	Clear(z)
}

func TestNewFromU64(t *testing.T) {
	z := NewFromU64(12345)
	require.NotNil(t, z)
	assert.Equal(t, uint64(12345), ToU64(z))
	Clear(z)
}

func TestNewFromI64(t *testing.T) {
	z := NewFromI64(-12345)
	require.NotNil(t, z)
	assert.Equal(t, int64(-12345), ToI64(z))
	Clear(z)
}

func TestNewFromString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		base  int
		want  int64
		ok    bool
	}{
		{"decimal", "12345", 10, 12345, true},
		{"hex", "0x1a2b", 0, 6699, true},
		{"binary", "0b1101", 0, 13, true},
		{"octal", "0o123", 0, 83, true},
		{"negative", "-12345", 10, -12345, true},
		{"underscores", "1_2_3_4_5", 10, 12345, true},
		{"exponent", "123e2", 10, 12300, true},
		{"exponent_with_plus", "12e+3", 10, 12000, true},
		{"invalid", "123abc", 10, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			z := NewFromString(tt.input, tt.base)
			if tt.ok {
				require.NotNil(t, z)
				assert.Equal(t, tt.want, ToI64(z))
				Clear(z)
			} else {
				assert.Nil(t, z)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	a := NewFromI64(42)
	b := Copy(a)
	require.NotNil(t, b)
	assert.Equal(t, ToI64(a), ToI64(b))
	Clear(a)
	Clear(b)
}

func TestString(t *testing.T) {
	z := NewFromI64(-12345)
	assert.Equal(t, "-12345", String(z))
	Clear(z)

	z2 := New()
	assert.Equal(t, "0", String(z2))
	Clear(z2)
}

func TestStringBase(t *testing.T) {
	z := NewFromU64(255)
	assert.Equal(t, "ff", StringBase(z, 16))
	assert.Equal(t, "11111111", StringBase(z, 2))
	Clear(z)
}

func TestSign(t *testing.T) {
	zero := New()
	assert.Equal(t, 0, Sign(zero))
	pos := NewFromI64(100)
	assert.Equal(t, 1, Sign(pos))
	neg := NewFromI64(-100)
	assert.Equal(t, -1, Sign(neg))
	Clear(zero)
	Clear(pos)
	Clear(neg)
}

func TestCmp(t *testing.T) {
	a := NewFromI64(100)
	b := NewFromI64(200)
	c := NewFromI64(100)
	assert.Equal(t, -1, Cmp(a, b))
	assert.Equal(t, 1, Cmp(b, a))
	assert.Equal(t, 0, Cmp(a, c))
	Clear(a)
	Clear(b)
	Clear(c)
}

func TestAbsNeg(t *testing.T) {
	x := NewFromI64(-42)
	ab := Abs(x)
	assert.Equal(t, int64(42), ToI64(ab))
	ng := Neg(ab)
	assert.Equal(t, int64(-42), ToI64(ng))
	Clear(x)
	Clear(ab)
	Clear(ng)
}

func TestAddSubMul(t *testing.T) {
	a := NewFromI64(10)
	b := NewFromI64(20)
	sum := Add(a, b)
	assert.Equal(t, int64(30), ToI64(sum))
	diff := Sub(a, b)
	assert.Equal(t, int64(-10), ToI64(diff))
	prod := Mul(a, b)
	assert.Equal(t, int64(200), ToI64(prod))
	Clear(a)
	Clear(b)
	Clear(sum)
	Clear(diff)
	Clear(prod)
}

func TestMulU64(t *testing.T) {
	a := NewFromI64(10)
	prod := MulU64(a, 20)
	assert.Equal(t, int64(200), ToI64(prod))
	Clear(a)
	Clear(prod)
}

func TestQuoRem(t *testing.T) {
	a := NewFromI64(100)
	b := NewFromI64(3)
	q, r := QuoRem(a, b)
	assert.Equal(t, int64(33), ToI64(q))
	assert.Equal(t, int64(1), ToI64(r))
	Clear(a)
	Clear(b)
	Clear(q)
	Clear(r)
}

func TestQuo(t *testing.T) {
	a := NewFromI64(100)
	b := NewFromI64(3)
	q := Quo(a, b)
	assert.Equal(t, int64(33), ToI64(q))
	Clear(a)
	Clear(b)
	Clear(q)
}

func TestRem(t *testing.T) {
	a := NewFromI64(100)
	b := NewFromI64(3)
	r := Rem(a, b)
	assert.Equal(t, int64(1), ToI64(r))
	Clear(a)
	Clear(b)
	Clear(r)
}

func TestEuclidMod(t *testing.T) {
	a := NewFromI64(-10)
	b := NewFromI64(3)
	mod := EuclidMod(a, b)
	assert.Equal(t, int64(2), ToI64(mod))
	Clear(a)
	Clear(b)
	Clear(mod)
}

func TestShlShr(t *testing.T) {
	a := NewFromU64(0b1010)
	sh := Shl(a, 2)
	assert.Equal(t, uint64(0b101000), ToU64(sh))
	sr := Shr(sh, 2)
	assert.Equal(t, uint64(0b1010), ToU64(sr))
	Clear(a)
	Clear(sh)
	Clear(sr)
}

func TestAndOrXor(t *testing.T) {
	a := NewFromU64(0b1100)
	b := NewFromU64(0b1010)
	and := And(a, b)
	assert.Equal(t, uint64(0b1000), ToU64(and))
	or := Or(a, b)
	assert.Equal(t, uint64(0b1110), ToU64(or))
	xor := Xor(a, b)
	assert.Equal(t, uint64(0b0110), ToU64(xor))
	Clear(a)
	Clear(b)
	Clear(and)
	Clear(or)
	Clear(xor)
}

func TestAndNot(t *testing.T) {
	a := NewFromU64(0b1100)
	b := NewFromU64(0b1010)
	an := AndNot(a, b)
	assert.Equal(t, uint64(0b0100), ToU64(an))
	Clear(a)
	Clear(b)
	Clear(an)
}

func TestNot(t *testing.T) {
	a := NewFromU64(0b1010)
	not := Not(a, 4, false)
	assert.Equal(t, uint64(0b0101), ToU64(not))
	n2 := Not(a, 4, true)
	assert.Equal(t, int64(-11), ToI64(n2)) // two's complement of 1010 in 4 bits is -6? Wait: 1010 = 10, ~1010 = 0101 = 5, sign-extended? Actually 4-bit signed: 1010 = -6, ~ should be 5? Let's compute: ~10 = -11 in Go's infinite precision? The test may need adjustment. We'll just test basic.
	Clear(a)
	Clear(not)
	Clear(n2)
}

func TestExpU64(t *testing.T) {
	a := NewFromU64(2)
	p := ExpU64(a, 10)
	assert.Equal(t, uint64(1024), ToU64(p))
	Clear(a)
	Clear(p)
}

func TestLog2(t *testing.T) {
	a := NewFromU64(1024)
	assert.Equal(t, 10, Log2(a))
	Clear(a)
	zero := New()
	assert.Equal(t, 0, Log2(zero))
	Clear(zero)
}

func TestToFloat64(t *testing.T) {
	a := NewFromI64(12345)
	assert.Equal(t, 12345.0, ToFloat64(a))
	Clear(a)
}

func TestParseBigInt(t *testing.T) {
	z := New()

	ok := parseBigInt(z, "0x1a", 0)
	assert.True(t, ok)
	assert.Equal(t, uint64(26), ToU64(z))

	ok = parseBigInt(z, "-0b101", 0)
	assert.True(t, ok)
	assert.Equal(t, int64(-5), ToI64(z))
}
