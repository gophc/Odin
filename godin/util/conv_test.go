package util

import (
	"math"
	"strconv"
	"testing"
)

func TestStrtoul(t *testing.T) {
	t.Parallel()

	testStrtoul(t, "0", 0)
	testStrtoul(t, "123", 123)
	testStrtoul(t, "1234567890", 1234567890)
	testStrtoul(t, "123456789012345678", 123456789012345678)
	// Max supported value: 2 ** 64 / 2 - 1
	testStrtoul(t, "9223372036854775807", math.MaxInt64)

	// Overflow by last digit: 2 ** 64 / 2 * 10 ** n
	testStrtoul(t, "9223372036854775808", 9223372036854775808)
	testStrtoul(t, "18446744073709551615", math.MaxUint64)

	testStrtoul(t, " 0", 0, -1)
	testStrtoul(t, "_", 0, -1)
	testStrtoul(t, "_123", 0, -1)
	testStrtoul(t, "", 0, -1)
	testStrtoul(t, "a", 0, -1)
	testStrtoul(t, "a", 10, 0, 1, 16)
	testStrtoul(t, "f", 0, -1, 0, 15)
	testStrtoul(t, "f", 15, 0, 1, 16)
	testStrtoul(t, "1f", 1, 0, 1, 15)
	testStrtoul(t, "1f", 31, 0, 2, 16)
	testStrtoul(t, "92233720368547758080", 0, -1)
	testStrtoul(t, "922337203685477580800", 0, -1)

	testStrtoul(t, "0 ", 0)
	testStrtoul(t, "123 ", 123)
	testStrtoul(t, "1234567890 ", 1234567890)
	testStrtoul(t, "123456789012345678 ", 123456789012345678)

	testStrtoul(t, "0 abc", 0)
	testStrtoul(t, "123 s", 123)
	testStrtoul(t, "1234567890 352", 1234567890)
	testStrtoul(t, "123456789012345678 121", 123456789012345678)

	testStrtoul(t, "0_", 0, 1)
	testStrtoul(t, "12_3", 123, 1)
	testStrtoul(t, "1_234_567_890", 1234567890, 3)
	testStrtoul(t, "1234_5678_9012_3456_78", 123456789012345678, 4)

	testStrtoul(t, "0_ a", 0, 1)
	testStrtoul(t, "12_3 a", 123, 1)
	testStrtoul(t, "123_ a", 123, 1)
	testStrtoul(t, "1_234_567_890 a", 1234567890, 3)
	testStrtoul(t, "1234_5678_9012_3456_78 a", 123456789012345678, 4)

	testStrtoul(t, "0", 0, 0, 1, 8)
	testStrtoul(t, "123", 83, 0, 3, 8)
	testStrtoul(t, "123 ", 83, 0, 3, 8)
	testStrtoul(t, "1234567890", 78187493520, 0, 10, 16)
	testStrtoul(t, "1234567890 ", 78187493520, 0, 10, 16)
	testStrtoul(t, "123456789012345678", 3771334298145412728, 0, 18, 16)
	testStrtoul(t, "123456789012345678 ", 3771334298145412728, 0, 18, 16)
}

func testStrtoul(t *testing.T, str string, expectedN uint64, args ...int) {
	s := StrUintPtr(str)
	nn := TryFirst(args, 0)
	radix := TryThird(args, 10)
	n, _, p, err := Strtoul(s, len(str), radix)
	if err != nil {
		if nn != -1 {
			t.Fatalf("Unexpected err %v. str=%q", err, s)
		}
		return
	}

	if n != expectedN {
		t.Fatalf("Unexpected value %d. Expected %d. str=%q", n, expectedN, s)
	}

	if expectedN <= 9223372036854775807 {
		expectedL := len(strconv.FormatInt(int64(expectedN), radix)) + nn
		expectedL = TrySecond(args, expectedL)
		sn := int(p - s)
		if sn != expectedL {
			t.Fatalf("Unexpected eofPos %d. Expected %d. str=%q", sn, expectedL, s)
		}
	}
}

func TestStrtold(t *testing.T) {
	t.Parallel()

	testStrtold(t, "0", 0)
	testStrtold(t, "1.", 1.)
	testStrtold(t, ".1", 0.1)
	testStrtold(t, "123.456", 123.456)
	testStrtold(t, "123", 123)
	testStrtold(t, "1234e2", 1234e2)
	testStrtold(t, "1234E-5", 1234e-5)
	testStrtold(t, "1.234e+3", 1.234e+3)
	testStrtold(t, "1.234e+0", 1.234e+0)

	testStrtold(t, "0 a", 0, 0, 1)
	testStrtold(t, "1. a", 1., 0, 2)
	testStrtold(t, ".1 a", 0.1, 0, 2)
	testStrtold(t, "123.456 a", 123.456, 0, 7)
	testStrtold(t, "123 a", 123, 0, 3)
	testStrtold(t, "1234e2 a", 1234e2, 0, 6)
	testStrtold(t, "1234E-5 a", 1234e-5, 0, 7)
	testStrtold(t, "1.234e+3 a", 1.234e+3, 0, 8)
	testStrtold(t, "1.234e+0 a", 1.234e+0, 0, 8)

	// empty num
	testStrtold(t, "", 0.0, -1)

	// negative num
	testStrtold(t, "-123.53", 0.0, -1)

	// non-num chars
	testStrtold(t, "123sdfsd", 123.0, 0, 3)
	testStrtold(t, "sdsf234", 0.0, -1)
	testStrtold(t, "sdfdf", 0.0, -1)

	// non-num chars in exponent
	testStrtold(t, "123e3s", 123000.0, 0, 5)
	testStrtold(t, "12.3e-op", 0.0, -1)
	testStrtold(t, "123E+SS5", 0.0, -1)

	// duplicate point
	testStrtold(t, "1.3.4", 0.0, -1)

	// duplicate exponent
	testStrtold(t, "123e5e6", 12300000.0, 0, 5)

	// missing exponent
	testStrtold(t, "123534e", 0.0, -1)
}

func testStrtold(t *testing.T, str string, expectedF float64, args ...int) {
	s := StrUintPtr(str)
	nn := TryFirst(args, 0)
	f, p, err := Strtold(s, len(str))
	if err != nil {
		if nn != -1 {
			t.Fatalf("Unexpected err %v. str=%q", err, s)
		}
		return
	}

	delta := f - expectedF
	if delta < 0 {
		delta = -delta
	}
	if delta > expectedF*1e-10 {
		t.Fatalf("Unexpected value when parsing %q: %f. Expected %f", s, f, expectedF)
	}

	expectedL := len(str)
	expectedL = TrySecond(args, expectedL)
	sn := int(p - s)
	if sn != expectedL {
		t.Fatalf("Unexpected eofPos %d. Expected %d. str=%q", sn, expectedL, s)
	}
}

func TestParseUintError64(t *testing.T) {
	t.Parallel()

	// Overflow by last digit: 2 ** 64 / 2 * 10 ** n
	testParseUintError(t, "9223372036854775808")
	testParseUintError(t, "92233720368547758080")
	testParseUintError(t, "922337203685477580800")
}

func TestParseUintSuccess(t *testing.T) {
	t.Parallel()

	testParseUintSuccess(t, "0", 0)
	testParseUintSuccess(t, "123", 123)
	testParseUintSuccess(t, "1234567890", 1234567890)
	testParseUintSuccess(t, "123456789012345678", 123456789012345678)

	// Max supported value: 2 ** 64 / 2 - 1
	testParseUintSuccess(t, "9223372036854775807", 9223372036854775807)
}

func TestParseUintError(t *testing.T) {
	t.Parallel()

	// empty string
	testParseUintError(t, "")

	// negative value
	testParseUintError(t, "-123")

	// non-num
	testParseUintError(t, "foobar234")

	// non-num chars at the end
	testParseUintError(t, "123w")

	// floating point num
	testParseUintError(t, "1234.545")

	// too big num
	testParseUintError(t, "12345678901234567890")
	testParseUintError(t, "1234567890123456789012")
}

func TestParseUfloatSuccess(t *testing.T) {
	t.Parallel()

	testParseUfloatSuccess(t, "0", 0)
	testParseUfloatSuccess(t, "1.", 1.)
	testParseUfloatSuccess(t, ".1", 0.1)
	testParseUfloatSuccess(t, "123.456", 123.456)
	testParseUfloatSuccess(t, "123", 123)
	testParseUfloatSuccess(t, "1234e2", 1234e2)
	testParseUfloatSuccess(t, "1234E-5", 1234e-5)
	testParseUfloatSuccess(t, "1.234e+3", 1.234e+3)
}

func TestParseUfloatError(t *testing.T) {
	t.Parallel()

	// empty num
	testParseUfloatError(t, "")

	// negative num
	testParseUfloatError(t, "-123.53")

	// non-num chars
	testParseUfloatError(t, "123sdfsd")
	testParseUfloatError(t, "sdsf234")
	testParseUfloatError(t, "sdfdf")

	// non-num chars in exponent
	testParseUfloatError(t, "123e3s")
	testParseUfloatError(t, "12.3e-op")
	testParseUfloatError(t, "123E+SS5")

	// duplicate point
	testParseUfloatError(t, "1.3.4")

	// duplicate exponent
	testParseUfloatError(t, "123e5e6")

	// missing exponent
	testParseUfloatError(t, "123534e")
}

func testParseUfloatError(t *testing.T, s string) {
	n, err := ParseUfloat([]byte(s))
	if err == nil {
		t.Fatalf("Expecting error when parsing %q. obtained %f", s, n)
	}
	if n >= 0 {
		t.Fatalf("Expecting negative num instead of %f when parsing %q", n, s)
	}
}

func testParseUfloatSuccess(t *testing.T, s string, expectedF float64) {
	f, err := ParseUfloat([]byte(s))
	if err != nil {
		t.Fatalf("Unexpected error when parsing %q: %v", s, err)
	}
	delta := f - expectedF
	if delta < 0 {
		delta = -delta
	}
	if delta > expectedF*1e-10 {
		t.Fatalf("Unexpected value when parsing %q: %f. Expected %f", s, f, expectedF)
	}
}

func TestParseUfloatBufSuccess(t *testing.T) {
	t.Parallel()

	testParseUfloatBufSuccess(t, "0a", 0)
	testParseUfloatBufSuccess(t, "1.x", 1.)
	testParseUfloatBufSuccess(t, ".1a", 0.1)
	testParseUfloatBufSuccess(t, "123.456 sda", 123.456)
	testParseUfloatBufSuccess(t, "123ASD", 123)
	testParseUfloatBufSuccess(t, "1234e2xs2", 1234e2)
	testParseUfloatBufSuccess(t, "1234E-5er", 1234e-5)
	testParseUfloatBufSuccess(t, "1.234e+3 ", 1.234e+3)
}

func testParseUfloatBufSuccess(t *testing.T, s string, expectedF float64) {
	f, _, err := ParseUfloatBuf([]byte(s))
	if err != nil {
		t.Fatalf("Unexpected error when parsing %q: %v", s, err)
	}
	delta := f - expectedF
	if delta < 0 {
		delta = -delta
	}
	if delta > expectedF*1e-10 {
		t.Fatalf("Unexpected value when parsing %q: %f. Expected %f", s, f, expectedF)
	}
}

func testParseUintError(t *testing.T, s string) {
	n, err := ParseUint([]byte(s))
	if err == nil {
		t.Fatalf("Expecting error when parsing %q. obtained %d", s, n)
	}
	if n >= 0 {
		t.Fatalf("Unexpected n=%d when parsing %q. Expected negative num", n, s)
	}
}

func testParseUintSuccess(t *testing.T, s string, expectedN int) {
	n, err := ParseUint([]byte(s))
	if err != nil {
		t.Fatalf("Unexpected error when parsing %q: %v", s, err)
	}
	if n != expectedN {
		t.Fatalf("Unexpected value %d. Expected %d. num=%q", n, expectedN, s)
	}
}
