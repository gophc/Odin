package util

type SliceStruct struct {
	Ptr uintptr
	Len int
	Cap int
}

type StrStruct struct {
	Ptr uintptr
	Len int
}

//region constraints

// Signed is a constraint that permits any signed integer type.
// If future releases of Go add new predeclared signed integer types,
// this constraint will be modified to include them.
type Signed interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// Unsigned is a constraint that permits any unsigned integer type.
// If future releases of Go add new predeclared unsigned integer types,
// this constraint will be modified to include them.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Integer is a constraint that permits any integer type.
// If future releases of Go add new predeclared integer types,
// this constraint will be modified to include them.
type Integer interface {
	Signed | Unsigned
}

// Float is a constraint that permits any floating-point type.
// If future releases of Go add new predeclared floating-point types,
// this constraint will be modified to include them.
type Float interface {
	~float32 | ~float64
}

// Complex is a constraint that permits any complex numeric type.
// If future releases of Go add new predeclared complex numeric types,
// this constraint will be modified to include them.
type Complex interface {
	~complex64 | ~complex128
}

// Ordered is a constraint that permits any ordered type: any type
// that supports the operators < <= >= >.
// If future releases of Go add new ordered types,
// this constraint will be modified to include them.
type Ordered interface {
	Integer | Float | ~string
}

//endregion

//region FN

//goland:noinspection GoUnusedExportedFunction
func BinarySearch[T Ordered](arr []T, target T) (int, bool) {
	end := len(arr) - 1
	if end < 0 {
		return -1, false
	}

	if target >= arr[end] {
		if target == arr[end] {
			return end, true
		}
		return -1, false
	}

	mid, left, right := 0, 0, end+1
	for left < right {
		mid = int(uint(left+right) >> 1)
		if target >= arr[mid] {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return left - 1, left > 0 && target == arr[left-1]
}

//goland:noinspection GoUnusedExportedFunction
func Ifi[V Integer, T any](n V, v1 T, v2 T) T {
	return Ifb(n != 0, v1, v2)
}

//goland:noinspection GoUnusedExportedFunction
func Ifif[V Integer, T any](n V, v1 func() T, v2 func() T) T {
	return Ifbf(n != 0, v1, v2)
}

func Ifb[T any](b bool, v1 T, v2 T) T {
	if b {
		return v1
	}
	return v2
}

func Ifbf[T any](b bool, v1 func() T, v2 func() T) T {
	if b {
		return v1()
	}
	return v2()
}

func Tf[T any](v T) func() T {
	return func() T {
		return v
	}
}

func Errf[T error](err T) func() string {
	return func() string {
		return err.Error()
	}
}

func ErrS(err error, s string) string {
	return Ifbf(err != nil, Errf(err), Tf(s))
}

func Contains[T Ordered](arr []T, target T) bool {
	return IndexOf(arr, target) >= 0
}

func IndexOf[T Ordered](arr []T, target T) int {
	for i, t := range arr {
		if t == target {
			return i
		}
	}
	return -1
}

//goland:noinspection GoUnusedExportedFunction
func MapTo[T any, V any](s []T, f func(T) (V, bool)) []V {
	v := make([]V, 0, len(s))
	for _, t := range s {
		i, ok := f(t)
		if ok {
			v = append(v, i)
		}
	}
	return v
}

func TryFirst[T any](s []T, v T) T {
	return TryN(s, 0, v)
}

func TrySecond[T any](s []T, v T) T {
	return TryN(s, 1, v)
}

func TryThird[T any](s []T, v T) T {
	return TryN(s, 2, v)
}

func TryN[T any](s []T, i int, v T) T {
	if i < 0 {
		i = 0
	}

	if len(s) > i {
		return s[i]
	}
	return v
}

//goland:noinspection GoUnusedExportedFunction
func Max[T Ordered](x, y T) T {
	if x > y {
		return x
	}
	return y
}

//goland:noinspection GoUnusedExportedFunction
func Min[T Ordered](x, y T) T {
	if x < y {
		return x
	}
	return y
}

//endregion
