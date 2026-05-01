package exact_value

import (
	"hash/fnv"
	"unsafe"

	"github.com/gophc/Odin/godin/big_int"
)

// hashUintptr hashes a uintptr value using FNV-1a
func hashUintptr(ptr uintptr) uint32 {
	h := fnv.New32a()
	// Write the bytes of the pointer
	ptrBytes := (*[unsafe.Sizeof(ptr)]byte)(unsafe.Pointer(ptr))
	h.Write(ptrBytes[:])
	return h.Sum32()
}

// hashString hashes a string using FNV-1a
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// hashString16 hashes a UTF-16 string using FNV-1a
func hashString16(s String16) uint32 {
	h := fnv.New32a()
	// Write each uint16 as bytes (little-endian)
	for _, ch := range s.Text {
		// Convert uint16 to 2 bytes
		b := [2]byte{byte(ch), byte(ch >> 8)}
		h.Write(b[:])
	}
	return h.Sum32()
}

// hashBigInt hashes a big.Int using FNV-1a
func hashBigInt(i *big_int.BigInt) uint32 {
	if i == nil || i.Sign() == 0 {
		return 0
	}
	h := fnv.New32a()
	// Write the bytes of the absolute value
	bytes := i.Bytes()
	h.Write(bytes)
	// Include the sign in the hash
	if i.Sign() < 0 {
		h.Write([]byte{0xFF})
	} else {
		h.Write([]byte{0x00})
	}
	return h.Sum32()
}

// HashExactValue computes a hash of an ExactValue
func HashExactValue(v ExactValue) uint32 {
	var res uint32 = 0
	switch v.Kind {
	case ExactValue_Invalid:
		return 0
	case ExactValue_Bool:
		h := fnv.New32a()
		if v.valueBool {
			h.Write([]byte{1})
		} else {
			h.Write([]byte{0})
		}
		res = h.Sum32()
	case ExactValue_String:
		res = hashString(string(v.valueString.Text))
	case ExactValue_String16:
		res = hashString16(v.valueString16)
	case ExactValue_Integer:
		res = hashBigInt(&v.valueInteger)
	case ExactValue_Float:
		h := fnv.New32a()
		// Write the float64 bits
		bits := (*[8]byte)(unsafe.Pointer(&v.valueFloat))
		h.Write(bits[:])
		res = h.Sum32()
	case ExactValue_Pointer:
		res = hashUintptr(uintptr(v.valuePointer))
	case ExactValue_Complex:
		if v.valueComplex != nil {
			h := fnv.New32a()
			bitsReal := (*[8]byte)(unsafe.Pointer(&v.valueComplex.Real))
			bitsImag := (*[8]byte)(unsafe.Pointer(&v.valueComplex.Imag))
			h.Write(bitsReal[:])
			h.Write(bitsImag[:])
			res = h.Sum32()
		}
	case ExactValue_Quaternion:
		if v.valueQuaternion != nil {
			h := fnv.New32a()
			bitsReal := (*[8]byte)(unsafe.Pointer(&v.valueQuaternion.Real))
			bitsImag := (*[8]byte)(unsafe.Pointer(&v.valueQuaternion.Imag))
			bitsJmag := (*[8]byte)(unsafe.Pointer(&v.valueQuaternion.Jmag))
			bitsKmag := (*[8]byte)(unsafe.Pointer(&v.valueQuaternion.Kmag))
			h.Write(bitsReal[:])
			h.Write(bitsImag[:])
			h.Write(bitsJmag[:])
			h.Write(bitsKmag[:])
			res = h.Sum32()
		}
	case ExactValue_Compound:
		res = hashUintptr(v.valueCompound)
	case ExactValue_Procedure:
		res = hashUintptr(v.valueProcedure)
	case ExactValue_Typeid:
		res = hashUintptr(v.valueTypeid)
	default:
		h := fnv.New32a()
		// Hash the struct itself as a fallback
		// This is not ideal but matches the C++ behavior
		vBytes := (*[unsafe.Sizeof(v)]byte)(unsafe.Pointer(&v))
		h.Write(vBytes[:])
		res = h.Sum32()
	}
	return res & 0x7fffffff
}
