// Package big_int provides a high-level wrapper around libtommath's MpInt.
// This file is kept for backward compatibility but all parsing logic has been moved to big_int.go.
// New code should use NewFromString or FromString directly.
package big_int

// digitValue is kept for compatibility but not used in the new math/big-based implementation.
func digitValue(ch rune) uint64 {
	switch {
	case ch >= '0' && ch <= '9':
		return uint64(ch - '0')
	case ch >= 'a' && ch <= 'f':
		return uint64(ch - 'a' + 10)
	case ch >= 'A' && ch <= 'F':
		return uint64(ch - 'A' + 10)
	default:
		return 16 // invalid
	}
}

// parseBigIntDecimalExponent is a stub for compatibility; the real implementation is in big_int.go.
func parseBigIntDecimalExponent(z *BigInt, baseDigits string, exp uint64) bool {
	return FromString(z, baseDigits+"e"+formatUint64(exp), 10)
}

func formatUint64(n uint64) string {
	if n == 0 {
		return "0"
	}
	var digits [20]byte
	idx := len(digits) - 1
	for n > 0 {
		digits[idx] = byte('0' + n%10)
		n /= 10
		idx--
	}
	return string(digits[idx+1:])
}

// parseBigIntInternal is a stub for compatibility; use parseBigInt from big_int.go instead.
func parseBigIntInternal(z *BigInt, s string, base int) bool {
	return parseBigInt(z, s, base)
}
