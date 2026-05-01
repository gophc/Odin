package big_int

import (
	"math/big"
	"strings"
)

type BigInt = big.Int

func New() *BigInt {
	return new(big.Int)
}

func NewFromU64(x uint64) *BigInt {
	return new(big.Int).SetUint64(x)
}

func NewFromI64(x int64) *BigInt {
	return big.NewInt(x)
}

func NewFromString(s string, base int) *BigInt {
	z := new(big.Int)
	ok := parseBigInt(z, s, base)
	if !ok {
		return nil
	}
	return z
}

func Copy(x *BigInt) *BigInt {
	if x == nil {
		return nil
	}
	return new(big.Int).Set(x)
}

func Clear(z *BigInt) {
	// math/big.Int requires no explicit cleanup; let GC handle it.
	// Keeping function for API compatibility.
}

func ToU64(x *BigInt) uint64 {
	if x == nil || x.Sign() < 0 {
		panic("big_int: value is negative or nil")
	}
	return x.Uint64()
}

func ToI64(x *BigInt) int64 {
	if x == nil {
		return 0
	}
	return x.Int64()
}

func ToFloat64(x *BigInt) float64 {
	if x == nil {
		return 0.0
	}
	f, _ := new(big.Float).SetInt(x).Float64()
	return f
}

func String(x *BigInt) string {
	if x == nil {
		return "0"
	}
	return x.String()
}

func StringBase(x *BigInt, base int) string {
	if x == nil {
		return "0"
	}
	if base < 2 || base > 16 {
		panic("base must be between 2 and 16")
	}
	return x.Text(base)
}

func Sign(x *BigInt) int {
	if x == nil {
		return 0
	}
	return x.Sign()
}

func IsZero(x *BigInt) bool {
	return x == nil || x.Sign() == 0
}

func Cmp(a, b *BigInt) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -b.Sign()
	}
	if b == nil {
		return a.Sign()
	}
	return a.Cmp(b)
}

func Abs(x *BigInt) *BigInt {
	if x == nil {
		return new(big.Int)
	}
	return new(big.Int).Abs(x)
}

func Neg(x *BigInt) *BigInt {
	if x == nil {
		return new(big.Int)
	}
	return new(big.Int).Neg(x)
}

func Add(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	return new(big.Int).Add(x, y)
}

func Sub(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	return new(big.Int).Sub(x, y)
}

func Mul(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	return new(big.Int).Mul(x, y)
}

func MulU64(x *BigInt, y uint64) *BigInt {
	if x == nil {
		return nil
	}
	return new(big.Int).Mul(x, new(big.Int).SetUint64(y))
}

func ExpU64(x *BigInt, y uint64) *BigInt {
	if x == nil {
		return nil
	}
	if y > 1<<31 {
		return nil
	}
	return new(big.Int).Exp(x, new(big.Int).SetUint64(y), nil)
}

func QuoRem(x, y *BigInt) (q, r *BigInt) {
	if x == nil || y == nil {
		return nil, nil
	}
	q = new(big.Int)
	r = new(big.Int)
	q.QuoRem(x, y, r)
	return q, r
}

func Quo(x, y *BigInt) *BigInt {
	q, _ := QuoRem(x, y)
	return q
}

func Rem(x, y *BigInt) *BigInt {
	_, r := QuoRem(x, y)
	return r
}

func EuclidMod(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	r := new(big.Int).Rem(x, y)
	if r.Sign() < 0 {
		r.Add(r, y)
	}
	return r
}

func Shl(x *BigInt, y uint32) *BigInt {
	if x == nil {
		return nil
	}
	return new(big.Int).Lsh(x, uint(y))
}

func Shr(x *BigInt, y uint32) *BigInt {
	if x == nil {
		return nil
	}
	return new(big.Int).Rsh(x, uint(y))
}

func And(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	return new(big.Int).And(x, y)
}

func AndNot(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	return new(big.Int).AndNot(x, y)
}

func Xor(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	return new(big.Int).Xor(x, y)
}

func Or(x, y *BigInt) *BigInt {
	if x == nil || y == nil {
		return nil
	}
	return new(big.Int).Or(x, y)
}

func Not(x *BigInt, bitCount int, signed bool) *BigInt {
	if x == nil || bitCount <= 0 {
		return new(big.Int)
	}
	if signed {
		// Return Go's bitwise NOT (infinite precision two's complement)
		return new(big.Int).Not(x)
	}
	// Unsigned NOT: mask to bitCount bits
	mask := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bitCount)), big.NewInt(1))
	return new(big.Int).Xor(x, mask)
}

func Log2(x *BigInt) int {
	if x == nil || x.Sign() <= 0 {
		return 0
	}
	return x.BitLen() - 1
}

func FromString(z *BigInt, s string, base int) bool {
	if z == nil {
		return false
	}
	return parseBigInt(z, s, base)
}

// parseBigInt is the internal parser (used by NewFromString and FromString).
func parseBigInt(z *BigInt, s string, base int) bool {
	if base < 2 || base > 16 {
		base = 0 // auto-detect
	}
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
	}
	// Handle prefixes
	if len(s) > 1 && s[0] == '0' {
		switch s[1] {
		case 'b', 'B':
			base = 2
			s = s[2:]
		case 'o', 'O':
			base = 8
			s = s[2:]
		case 'd', 'D':
			base = 10
			s = s[2:]
		case 'x', 'X', 'h', 'H':
			base = 16
			s = s[2:]
		}
	}
	if len(s) == 0 {
		return false
	}
	// Remove underscores
	s = strings.ReplaceAll(s, "_", "")
	if len(s) == 0 {
		return false
	}
	// Handle decimal exponent (base 10 only)
	if base == 10 || base == 0 {
		lower := strings.ToLower(s)
		if idx := strings.IndexAny(lower, "e"); idx != -1 {
			mantissa := s[:idx]
			expStr := s[idx+1:]
			expStr = strings.ReplaceAll(expStr, "_", "")
			if len(expStr) == 0 {
				return false
			}
			if expStr[0] == '+' {
				expStr = expStr[1:]
			} else if expStr[0] == '-' {
				return false // negative exponent not supported for integer
			}
			exp := uint64(0)
			for _, ch := range expStr {
				if ch == '_' {
					continue
				}
				if ch < '0' || ch > '9' {
					return false
				}
				exp = exp*10 + uint64(ch-'0')
			}
			if exp > 308 {
				return false
			}
			// Parse mantissa
			ok := parseBigInt(z, mantissa, 10)
			if !ok {
				return false
			}
			// Multiply by 10^exp
			pow10 := new(big.Int).Exp(big.NewInt(10), new(big.Int).SetUint64(exp), nil)
			z.Mul(z, pow10)
			if neg {
				z.Neg(z)
			}
			return true
		}
	}
	// If base is 0, auto-detect from prefix, but we already handled prefix; fallback to base 10.
	if base == 0 {
		base = 10
	}
	_, ok := z.SetString(s, base)
	if !ok {
		return false
	}
	if neg {
		z.Neg(z)
	}
	return true
}
