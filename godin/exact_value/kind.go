package exact_value

// ExactValueKind represents the type of exact value
type ExactValueKind int

const (
	ExactValue_Invalid ExactValueKind = iota
	ExactValue_Bool
	ExactValue_String
	ExactValue_Integer
	ExactValue_Float
	ExactValue_Complex
	ExactValue_Quaternion
	ExactValue_Pointer
	ExactValue_Compound
	ExactValue_Procedure
	ExactValue_Typeid
	ExactValue_String16
	ExactValue_Count
)
