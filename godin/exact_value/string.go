package exact_value

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/gophc/Odin/godin/big_int"
)

// String returns a string representation of the exact value.
func (v ExactValue) String() string {
	return ExactValueToString(v, 36)
}

// ExactValueToString converts an exact value to its string representation.
func ExactValueToString(v ExactValue, stringLimit int) string {
	var sb strings.Builder
	writeExactValueToString(&sb, v, stringLimit)
	return sb.String()
}

func writeExactValueToString(sb *strings.Builder, v ExactValue, stringLimit int) {
	if stringLimit <= 0 {
		stringLimit = 36
	}
	switch v.Kind {
	case ExactValue_Invalid:
		return
	case ExactValue_Bool:
		if v.valueBool {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case ExactValue_String:
		quoteToAscii(sb, v.valueString.Text, stringLimit)
	case ExactValue_String16:
		utf8Bytes := utf16ToUTF8(v.valueString16.Text)
		quoteToAscii(sb, utf8Bytes, stringLimit)
	case ExactValue_Integer:
		sb.WriteString(big_int.String(&v.valueInteger))
	case ExactValue_Float:
		// Use %g for cleaner output, but original used %f
		fmt.Fprintf(sb, "%f", v.valueFloat)
	case ExactValue_Complex:
		if v.valueComplex != nil {
			fmt.Fprintf(sb, "%f+%fi", v.valueComplex.Real, v.valueComplex.Imag)
		}
	case ExactValue_Quaternion:
		if v.valueQuaternion != nil {
			fmt.Fprintf(sb, "%f+%fi+%fj+%fk",
				v.valueQuaternion.Real, v.valueQuaternion.Imag,
				v.valueQuaternion.Jmag, v.valueQuaternion.Kmag)
		}
	case ExactValue_Pointer:
		// No output for pointers (like original)
	case ExactValue_Compound, ExactValue_Procedure:
		// Placeholder
		sb.WriteString("<compound>")
	case ExactValue_Typeid:
		sb.WriteString("<typeid>")
	}
}

func quoteToAscii(sb *strings.Builder, data []byte, limit int) {
	if len(data) <= limit {
		sb.WriteByte('"')
		sb.Write(data)
		sb.WriteByte('"')
	} else {
		n := limit / 5
		if n < 1 {
			n = 1
		}
		sb.WriteByte('"')
		sb.Write(data[:n])
		fmt.Fprintf(sb, "\"..%d chars..\"", len(data)-2*n)
		sb.Write(data[len(data)-n:])
		sb.WriteByte('"')
	}
}

func utf16ToUTF8(data []uint16) []byte {
	buf := bytes.Buffer{}
	for _, ch := range data {
		if ch < 0x80 {
			buf.WriteByte(byte(ch))
		} else if ch < 0x800 {
			buf.WriteByte(0xC0 | byte(ch>>6))
			buf.WriteByte(0x80 | byte(ch&0x3F))
		} else {
			buf.WriteByte(0xE0 | byte(ch>>12))
			buf.WriteByte(0x80 | byte((ch>>6)&0x3F))
			buf.WriteByte(0x80 | byte(ch&0x3F))
		}
	}
	return buf.Bytes()
}
