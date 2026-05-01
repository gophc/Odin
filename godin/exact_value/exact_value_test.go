package exact_value

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBool(t *testing.T) {
	v := NewBool(true)
	assert.Equal(t, ExactValue_Bool, v.Kind)
	assert.True(t, v.ValueBool())

	v = NewBool(false)
	assert.False(t, v.ValueBool())
}

func TestNewInteger(t *testing.T) {
	v := NewIntegerFromI64(42)
	assert.Equal(t, ExactValue_Integer, v.Kind)
	assert.Equal(t, int64(42), ToI64(v))

	v = NewIntegerFromU64(100)
	assert.Equal(t, uint64(100), ToU64(v))
}

func TestNewFloat(t *testing.T) {
	v := NewFloat(3.14)
	assert.Equal(t, ExactValue_Float, v.Kind)
	assert.InDelta(t, 3.14, ToF64(v), 1e-9)
}

func TestNewComplex(t *testing.T) {
	v := NewComplex(1.5, 2.5)
	assert.Equal(t, ExactValue_Complex, v.Kind)
	require.NotNil(t, v.ValueComplex())
	assert.InDelta(t, 1.5, v.ValueComplex().Real, 1e-9)
	assert.InDelta(t, 2.5, v.ValueComplex().Imag, 1e-9)
}

func TestNewQuaternion(t *testing.T) {
	v := NewQuaternion(1, 2, 3, 4)
	assert.Equal(t, ExactValue_Quaternion, v.Kind)
	require.NotNil(t, v.ValueQuaternion())
	assert.InDelta(t, 1, v.ValueQuaternion().Real, 1e-9)
	assert.InDelta(t, 2, v.ValueQuaternion().Imag, 1e-9)
	assert.InDelta(t, 3, v.ValueQuaternion().Jmag, 1e-9)
	assert.InDelta(t, 4, v.ValueQuaternion().Kmag, 1e-9)
}

func TestUnaryOperators(t *testing.T) {
	// Negation
	v := NewIntegerFromI64(5)
	neg := UnaryOperator(tokenSub, v, 0, false)
	assert.Equal(t, int64(-5), ToI64(neg))

	// Float negation
	vf := NewFloat(3.14)
	negf := UnaryOperator(tokenSub, vf, 0, false)
	assert.InDelta(t, -3.14, ToF64(negf), 1e-9)

	// Complex negation
	vc := NewComplex(1, 2)
	negc := UnaryOperator(tokenSub, vc, 0, false)
	assert.InDelta(t, -1, negc.ValueComplex().Real, 1e-9)
	assert.InDelta(t, -2, negc.ValueComplex().Imag, 1e-9)

	// Logical NOT
	vb := NewBool(true)
	notb := UnaryOperator(tokenNot, vb, 0, false)
	assert.False(t, notb.ValueBool())
}

func TestBinaryOperators(t *testing.T) {
	// Integer addition
	a := NewIntegerFromI64(10)
	b := NewIntegerFromI64(5)
	sum := Add(a, b)
	assert.Equal(t, int64(15), ToI64(sum))

	// Integer subtraction
	diff := Sub(a, b)
	assert.Equal(t, int64(5), ToI64(diff))

	// Integer multiplication
	prod := Mul(a, b)
	assert.Equal(t, int64(50), ToI64(prod))

	// Float division
	fa := NewFloat(10.0)
	fb := NewFloat(3.0)
	quo := Quo(fa, fb)
	assert.InDelta(t, 10.0/3.0, ToF64(quo), 1e-9)

	// Complex addition
	ca := NewComplex(1, 2)
	cb := NewComplex(3, 4)
	csum := Add(ca, cb)
	assert.InDelta(t, 4, csum.ValueComplex().Real, 1e-9)
	assert.InDelta(t, 6, csum.ValueComplex().Imag, 1e-9)

	// Quaternion multiplication
	qa := NewQuaternion(1, 2, 3, 4)
	qb := NewQuaternion(5, 6, 7, 8)
	qprod := Mul(qa, qb)
	// Expected values from quaternion multiplication formula
	expectedReal := 1*5 - 2*6 - 3*7 - 4*8 // 5 -12 -21 -32 = -60
	expectedImag := 1*6 + 2*5 + 3*8 - 4*7 // 6 +10 +24 -28 = 12
	expectedJmag := 1*7 - 2*8 + 3*5 + 4*6 // 7 -16 +15 +24 = 30
	expectedKmag := 1*8 + 2*7 - 3*6 + 4*5 // 8 +14 -18 +20 = 24
	assert.InDelta(t, float64(expectedReal), qprod.ValueQuaternion().Real, 1e-9)
	assert.InDelta(t, float64(expectedImag), qprod.ValueQuaternion().Imag, 1e-9)
	assert.InDelta(t, float64(expectedJmag), qprod.ValueQuaternion().Jmag, 1e-9)
	assert.InDelta(t, float64(expectedKmag), qprod.ValueQuaternion().Kmag, 1e-9)
}

func TestComparison(t *testing.T) {
	// Integer comparisons
	a := NewIntegerFromI64(10)
	b := NewIntegerFromI64(20)
	assert.True(t, Compare(tokenLt, a, b))
	assert.False(t, Compare(tokenGt, a, b))
	assert.True(t, Compare(tokenCmpEq, a, a))

	// Float comparisons
	fa := NewFloat(3.14)
	fb := NewFloat(2.71)
	assert.True(t, Compare(tokenGt, fa, fb))
	assert.False(t, Compare(tokenLt, fa, fb))

	// String comparisons
	s1 := NewString("hello")
	s2 := NewString("world")
	assert.True(t, Compare(tokenLt, s1, s2))
	assert.True(t, Compare(tokenCmpEq, s1, s1))
}

func TestFromString(t *testing.T) {
	// Integer parsing
	v := ExactValueFromString("12345")
	assert.Equal(t, ExactValue_Integer, v.Kind)
	assert.Equal(t, int64(12345), ToI64(v))

	// Float parsing
	v = ExactValueFromString("3.14159")
	assert.Equal(t, ExactValue_Float, v.Kind)
	assert.InDelta(t, 3.14159, ToF64(v), 1e-9)

	// Hexadecimal float
	v = ExactValueFromString("0h3f800000") // 1.0 as float32
	assert.Equal(t, ExactValue_Float, v.Kind)
	assert.InDelta(t, 1.0, ToF64(v), 1e-9)

	// Invalid
	v = ExactValueFromString("not a number")
	assert.Equal(t, ExactValue_Invalid, v.Kind)
}

func TestFromBasicLiteral(t *testing.T) {
	v := ExactValueFromBasicLiteral(LitString, "test")
	assert.Equal(t, ExactValue_String, v.Kind)
	assert.Equal(t, "test", string(v.ValueString().Text))

	v = ExactValueFromBasicLiteral(LitInteger, "42")
	assert.Equal(t, ExactValue_Integer, v.Kind)
	assert.Equal(t, int64(42), ToI64(v))

	v = ExactValueFromBasicLiteral(LitFloat, "3.14")
	assert.Equal(t, ExactValue_Float, v.Kind)
	assert.InDelta(t, 3.14, ToF64(v), 1e-9)

	v = ExactValueFromBasicLiteral(LitImag, "3i")
	assert.Equal(t, ExactValue_Complex, v.Kind)
	assert.InDelta(t, 0, v.ValueComplex().Real, 1e-9)
	assert.InDelta(t, 3, v.ValueComplex().Imag, 1e-9)

	v = ExactValueFromBasicLiteral(LitRune, "A")
	assert.Equal(t, ExactValue_Integer, v.Kind)
	assert.Equal(t, int64('A'), ToI64(v))
}

func TestHash(t *testing.T) {
	h1 := HashExactValue(NewIntegerFromI64(123))
	h2 := HashExactValue(NewIntegerFromI64(123))
	assert.Equal(t, h1, h2)

	h3 := HashExactValue(NewString("hello"))
	h4 := HashExactValue(NewString("hello"))
	assert.Equal(t, h3, h4)

	h5 := HashExactValue(NewString("world"))
	assert.NotEqual(t, h3, h5)
}

func TestToString(t *testing.T) {
	s := ExactValueToString(NewBool(true), 36)
	assert.Equal(t, "true", s)

	s = ExactValueToString(NewIntegerFromI64(12345), 36)
	assert.Equal(t, "12345", s)

	s = ExactValueToString(NewFloat(3.14), 36)
	assert.Contains(t, s, "3.14")

	s = ExactValueToString(NewComplex(1, 2), 36)
	assert.Contains(t, s, "1.000000+2.000000i")
}

// TestBigIntegerOperations tests big integer arithmetic beyond 64-bit
func TestBigIntegerOperations(t *testing.T) {
	big1 := NewIntegerFromString("12345678901234567890")
	big2 := NewIntegerFromString("9876543210987654321")
	sum := Add(big1, big2)
	expectedSum := NewIntegerFromString("22222222112222222211")
	assert.True(t, Compare(tokenCmpEq, sum, expectedSum))

	prod := Mul(big1, big2)
	expectedProd := NewIntegerFromString("121932631137021795223746380111126352690")
	assert.True(t, Compare(tokenCmpEq, prod, expectedProd))
}

func TestMatchExactValues(t *testing.T) {
	intVal := NewIntegerFromI64(5)
	floatVal := NewFloat(3.14)

	matchExactValues(&intVal, &floatVal)
	// After match, intVal should be promoted to float
	assert.Equal(t, ExactValue_Float, intVal.Kind)
	assert.InDelta(t, 5.0, ToF64(intVal), 1e-9)

	intVal2 := NewIntegerFromI64(10)
	complexVal := NewComplex(1, 2)
	matchExactValues(&intVal2, &complexVal)
	assert.Equal(t, ExactValue_Complex, intVal2.Kind)
	assert.InDelta(t, 10.0, intVal2.ValueComplex().Real, 1e-9)
	assert.InDelta(t, 0.0, intVal2.ValueComplex().Imag, 1e-9)
}
