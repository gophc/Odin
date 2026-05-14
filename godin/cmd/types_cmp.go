package cmd

import "strings"

const (
	FailureSize      int64 = -1
	FailureAlignment int64 = -2
)

func areTypesIdentical(x, y *Type) bool {
	if x == y {
		return true
	}
	if (x == nil && y != nil) || (x != nil && y == nil) {
		return false
	}
	if x.Kind == TypeNamed {
		e := x.Named.TypeName
		if e != nil && e.TypeName.IsTypeAlias {
			x = x.Named.Base
		}
	}
	if y.Kind == TypeNamed {
		e := y.Named.TypeName
		if e != nil && e.TypeName.IsTypeAlias {
			y = y.Named.Base
		}
	}
	if x == nil || y == nil || x.Kind != y.Kind {
		return false
	}
	return areTypesIdenticalInternal(x, y, false)
}

func areTypesIdenticalUniqueTuples(x, y *Type) bool {
	if x == y {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	if x.Kind == TypeNamed {
		e := x.Named.TypeName
		if e != nil && e.TypeName.IsTypeAlias {
			x = x.Named.Base
		}
	}
	if y.Kind == TypeNamed {
		e := y.Named.TypeName
		if e != nil && e.TypeName.IsTypeAlias {
			y = y.Named.Base
		}
	}
	if x.Kind != y.Kind {
		return false
	}
	return areTypesIdenticalInternal(x, y, true)
}

func areProcPropertiesIdentical(x, y *Type) bool {
	return x.Proc.CallingConvention == y.Proc.CallingConvention &&
		x.Proc.CVararg == y.Proc.CVararg &&
		x.Proc.Variadic == y.Proc.Variadic &&
		x.Proc.Diverging == y.Proc.Diverging &&
		x.Proc.OptionalOK == y.Proc.OptionalOK
}

func areTypesIdenticalInternal(x, y *Type, checkTupleNames bool) bool {
	if x == y {
		return true
	}
	if x == nil || y == nil {
		return false
	}

	switch x.Kind {
	case TypeGeneric:
		return areTypesIdentical(x.Generic.Specialized, y.Generic.Specialized)

	case TypeBasic:
		return x.Basic.Kind == y.Basic.Kind

	case TypeEnumeratedArray:
		return areTypesIdentical(x.EnumeratedArray.Index, y.EnumeratedArray.Index) &&
			areTypesIdentical(x.EnumeratedArray.Elem, y.EnumeratedArray.Elem)

	case TypeArray:
		return x.Array.Count == y.Array.Count &&
			areTypesIdentical(x.Array.Elem, y.Array.Elem)

	case TypeMatrix:
		return x.Matrix.RowCount == y.Matrix.RowCount &&
			x.Matrix.ColumnCount == y.Matrix.ColumnCount &&
			x.Matrix.IsRowMajor == y.Matrix.IsRowMajor &&
			areTypesIdentical(x.Matrix.Elem, y.Matrix.Elem)

	case TypeDynamicArray:
		return areTypesIdentical(x.DynamicArray.Elem, y.DynamicArray.Elem)

	case TypeFixedCapacityDynamicArray:
		return x.FixedCapacityDynamicArray.Capacity == y.FixedCapacityDynamicArray.Capacity &&
			areTypesIdentical(x.FixedCapacityDynamicArray.Elem, y.FixedCapacityDynamicArray.Elem)

	case TypeSlice:
		return areTypesIdentical(x.Slice.Elem, y.Slice.Elem)

	case TypeBitSet:
		if areTypesIdentical(x.BitSet.Elem, y.BitSet.Elem) &&
			areTypesIdentical(x.BitSet.Underlying, y.BitSet.Underlying) {
			if isTypeEnum(x.BitSet.Elem) {
				return true
			}
			return x.BitSet.Lower == y.BitSet.Lower && x.BitSet.Upper == y.BitSet.Upper
		}
		return false

	case TypeEnum:
		if x == y {
			return true
		}
		if len(x.Enum.Fields) != len(y.Enum.Fields) {
			return false
		}
		if !areTypesIdentical(x.Enum.BaseType, y.Enum.BaseType) {
			return false
		}
		if x.Enum.MinValueIndex != y.Enum.MinValueIndex {
			return false
		}
		if x.Enum.MaxValueIndex != y.Enum.MaxValueIndex {
			return false
		}
		for i := range x.Enum.Fields {
			a := x.Enum.Fields[i]
			b := y.Enum.Fields[i]
			if a.Token.String != b.Token.String {
				return false
			}
			if a.Kind != b.Kind || a.Kind != EntityConstant {
				gbAssertHandler("unexpected entity kind in enum comparison")
			}
			if !compareExactValues(TokenCmpEq, a.Constant.Value, b.Constant.Value) {
				return false
			}
		}
		return true

	case TypeUnion:
		if len(x.Union.Variants) == len(y.Union.Variants) &&
			x.Union.Kind == y.Union.Kind {
			if x.Union.CustomAlign != y.Union.CustomAlign {
				if typeAlignOf(x) != typeAlignOf(y) {
					return false
				}
			}
			for i := range x.Union.Variants {
				if !areTypesIdentical(x.Union.Variants[i], y.Union.Variants[i]) {
					return false
				}
			}
			return true
		}

	case TypeStruct:
		if x.Struct.IsRawUnion == y.Struct.IsRawUnion &&
			len(x.Struct.Fields) == len(y.Struct.Fields) &&
			x.Struct.IsPacked == y.Struct.IsPacked &&
			x.Struct.IsAllOrNone == y.Struct.IsAllOrNone &&
			x.Struct.SoaKind == y.Struct.SoaKind &&
			x.Struct.SoaCount == y.Struct.SoaCount &&
			areTypesIdentical(x.Struct.SoaElem, y.Struct.SoaElem) {
			if x.Struct.CustomAlign != y.Struct.CustomAlign {
				if typeAlignOf(x) != typeAlignOf(y) {
					return false
				}
			}
			for i := range x.Struct.Fields {
				xf := x.Struct.Fields[i]
				yf := y.Struct.Fields[i]
				if xf.Kind != yf.Kind {
					return false
				}
				if !areTypesIdentical(xf.Type, yf.Type) {
					return false
				}
				if xf.Token.String != yf.Token.String {
					return false
				}
				if x.Struct.Tags[i] != y.Struct.Tags[i] {
					return false
				}
				xfFlags := xf.Flags & EntityFlagsIsSubtype
				yfFlags := yf.Flags & EntityFlagsIsSubtype
				if xfFlags != yfFlags {
					return false
				}
			}
			return true
		}

	case TypePointer:
		return areTypesIdentical(x.Pointer.Elem, y.Pointer.Elem)

	case TypeMultiPointer:
		return areTypesIdentical(x.MultiPointer.Elem, y.MultiPointer.Elem)

	case TypeSoaPointer:
		return areTypesIdentical(x.SoaPointer.Elem, y.SoaPointer.Elem)

	case TypeNamed:
		return x.Named.TypeName == y.Named.TypeName

	case TypeTuple:
		if len(x.Tuple.Variables) == len(y.Tuple.Variables) &&
			x.Tuple.IsPacked == y.Tuple.IsPacked {
			for i := range x.Tuple.Variables {
				xe := x.Tuple.Variables[i]
				ye := y.Tuple.Variables[i]
				if xe.Kind != ye.Kind || !areTypesIdentical(xe.Type, ye.Type) {
					return false
				}
				if checkTupleNames {
					if xe.Token.String != ye.Token.String {
						return false
					}
				}
				if xe.Kind == EntityConstant && !compareExactValues(TokenCmpEq, xe.Constant.Value, ye.Constant.Value) {
					return false
				}
			}
			return true
		}

	case TypeProc:
		return areProcPropertiesIdentical(x, y) &&
			areTypesIdenticalInternal(x.Proc.Params, y.Proc.Params, checkTupleNames) &&
			areTypesIdenticalInternal(x.Proc.Results, y.Proc.Results, checkTupleNames)

	case TypeMap:
		return areTypesIdentical(x.Map.Key, y.Map.Key) &&
			areTypesIdentical(x.Map.Value, y.Map.Value)

	case TypeSimdVector:
		if x.SimdVector.Count == y.SimdVector.Count {
			return areTypesIdentical(x.SimdVector.Elem, y.SimdVector.Elem)
		}

	case TypeBitField:
		if areTypesIdentical(x.BitField.BackingType, y.BitField.BackingType) &&
			len(x.BitField.Fields) == len(y.BitField.Fields) {
			for i := range x.BitField.Fields {
				a := x.BitField.Fields[i]
				b := y.BitField.Fields[i]
				if !areTypesIdentical(a.Type, b.Type) {
					return false
				}
				if a.Token.String != b.Token.String {
					return false
				}
				if x.BitField.BitSizes[i] != y.BitField.BitSizes[i] {
					return false
				}
				if x.BitField.BitOffsets[i] != y.BitField.BitOffsets[i] {
					return false
				}
			}
			return true
		}
	}

	return false
}

func defaultType(type_ *Type) *Type {
	if type_ == nil {
		return tInvalid
	}
	if type_.Kind == TypeBasic {
		switch type_.Basic.Kind {
		case BasicUntypedBool:
			return tBool
		case BasicUntypedInteger:
			return tInt
		case BasicUntypedFloat:
			return tF64
		case BasicUntypedComplex:
			return tComplex128
		case BasicUntypedQuaternion:
			return tQuaternion256
		case BasicUntypedString:
			return tString
		case BasicUntypedRune:
			return tRune
		}
	} else if type_.Kind == TypeGeneric {
		if type_.Generic.Specialized != nil {
			return defaultType(type_.Generic.Specialized)
		}
	}
	return type_
}

func cVarargPromoteType(type_ *Type) *Type {
	if type_ == nil {
		return nil
	}
	core := coreType(type_)
	if core.Kind == TypeBitSet {
		gbAssertHandler("c_vararg_promote_type called with BitSet type")
	}
	if core.Kind == TypeBasic {
		switch core.Basic.Kind {
		case BasicF16, BasicF32, BasicUntypedFloat:
			return tF64
		case BasicF16le, BasicF32le:
			return tF64le
		case BasicF16be, BasicF32be:
			return tF64be
		case BasicUntypedBool, BasicBool, BasicB8, BasicB16,
			BasicI8, BasicI16, BasicU8, BasicU16:
			return tI32
		case BasicI16le, BasicU16le:
			return tI32le
		case BasicI16be, BasicU16be:
			return tI32be
		}
	}
	return type_
}

func unionVariantIndexTypesEqual(v, vt *Type) bool {
	if areTypesIdentical(v, vt) {
		return true
	}
	if isTypeProc(v) && isTypeProc(vt) {
		return areTypesIdentical(baseType(v), baseType(vt))
	}
	return false
}

func unionVariantIndexChecked(u, v *Type) int64 {
	u = baseType(u)
	if u.Kind != TypeUnion {
		gbAssertHandler("union_variant_index_checked: expected union type")
	}
	for i, vt := range u.Union.Variants {
		if unionVariantIndexTypesEqual(v, vt) {
			if u.Union.Kind == UnionTypeNoNil {
				return int64(i)
			} else {
				return int64(i + 1)
			}
		}
	}
	return -1
}

func unionIsVariantOf(u, v *Type) bool {
	u = baseType(u)
	if u.Kind != TypeUnion {
		gbAssertHandler("union_is_variant_of: expected union type")
	}
	for _, vt := range u.Union.Variants {
		if unionVariantIndexTypesEqual(v, vt) {
			return true
		}
	}
	return false
}

func unionTagSize(u *Type) int64 {
	u = baseType(u)
	if u.Kind != TypeUnion {
		gbAssertHandler("union_tag_size: expected union type")
	}
	if u.Union.TagSize > 0 {
		return int64(u.Union.TagSize)
	}
	n := uint64(len(u.Union.Variants))
	if n == 0 {
		return 0
	}
	maxAlign := int64(1)
	switch {
	case n < 1<<8:
		maxAlign = 1
	case n < 1<<16:
		maxAlign = 2
	case n < 1<<32:
		maxAlign = 4
	default:
		compilerError("how many variants do you have?! %d", n)
	}
	if u.Union.CustomAlign > 0 {
		if maxAlign < u.Union.CustomAlign {
			maxAlign = u.Union.CustomAlign
		}
	} else {
		for _, variantType := range u.Union.Variants {
			align := typeAlignOf(variantType)
			if maxAlign < align {
				maxAlign = align
			}
		}
	}
	tagSize := maxAlign
	if buildContext.MaxAlign < tagSize {
		tagSize = buildContext.MaxAlign
	}
	if tagSize > 8 {
		tagSize = 8
	}
	u.Union.TagSize = int16(tagSize)
	return tagSize
}

func unionTagType(u *Type) *Type {
	s := unionTagSize(u)
	switch s {
	case 0, 1:
		return tU8
	case 2:
		return tU16
	case 4:
		return tU32
	case 8:
		return tU64
	}
	return tUint
}

func matchedTargetFeatures(t TypeProc) int {
	if len(t.RequireTargetFeature) == 0 {
		return 0
	}
	matches := 0
	for _, part := range strings.Split(t.RequireTargetFeature, ",") {
		if part == "" {
			continue
		}
		if checkTargetFeatureIsValidForTargetArch(part, "") {
			matches++
		}
	}
	return matches
}

func areProcTypesOverloadSafe(x, y *Type) ProcTypeOverloadKind {
	if x == nil || y == nil {
		return ProcOverloadNotProcedure
	}
	if !isTypeProc(x) {
		return ProcOverloadNotProcedure
	}
	if !isTypeProc(y) {
		return ProcOverloadNotProcedure
	}

	px := baseType(x).Proc
	py := baseType(y).Proc

	if px.ParamCount != py.ParamCount {
		return ProcOverloadParamCount
	}
	for i := int32(0); i < px.ParamCount; i++ {
		ex := px.Params.Tuple.Variables[i]
		ey := py.Params.Tuple.Variables[i]
		if !areTypesIdentical(ex.Type, ey.Type) {
			return ProcOverloadParamTypes
		}
	}
	if px.Variadic != py.Variadic {
		return ProcOverloadParamVariadic
	}
	if px.IsPolymorphic != py.IsPolymorphic {
		return ProcOverloadPolymorphic
	}
	if px.ResultCount != py.ResultCount {
		return ProcOverloadResultCount
	}
	for i := int32(0); i < px.ResultCount; i++ {
		ex := px.Results.Tuple.Variables[i]
		ey := py.Results.Tuple.Variables[i]
		if !areTypesIdentical(ex.Type, ey.Type) {
			return ProcOverloadResultTypes
		}
	}
	if matchedTargetFeatures(px) != matchedTargetFeatures(py) {
		return ProcOverloadTargetFeatures
	}
	return ProcOverloadIdentical
}

func matrixAlignOf(t *Type, tp *TypePath) int64 {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("matrix_align_of: expected matrix type")
	}
	elem := t.Matrix.Elem
	rowCount := t.Matrix.RowCount
	if rowCount < 1 {
		rowCount = 1
	}
	columnCount := t.Matrix.ColumnCount
	if columnCount < 1 {
		columnCount = 1
	}

	pop := typePathPush(tp, elem)
	if tp.Failure {
		return FailureAlignment
	}
	elemAlign := typeAlignOfInternal(elem, tp)
	if pop {
		typePathPop(tp)
	}
	elemSize := typeSizeOf(elem)

	totalExpectedSize := rowCount * columnCount * elemSize
	minAlignment := prevPow2(totalExpectedSize)
	for totalExpectedSize != 0 && (totalExpectedSize%minAlignment) != 0 {
		minAlignment >>= 1
	}
	if minAlignment < elemAlign {
		minAlignment = elemAlign
	}
	align := minAlignment
	if buildContext.MaxSimdAlign < align {
		align = buildContext.MaxSimdAlign
	}
	return align
}

func matrixTypeStrideInBytes(t *Type, tp *TypePath) int64 {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("matrix_type_stride_in_bytes: expected matrix type")
	}
	if t.Matrix.StrideInBytes != 0 {
		return t.Matrix.StrideInBytes
	}
	if t.Matrix.RowCount == 0 {
		return 0
	}
	var elemSize int64
	if tp != nil {
		elemSize = typeSizeOfInternal(t.Matrix.Elem, tp)
	} else {
		elemSize = typeSizeOf(t.Matrix.Elem)
	}
	var strideInBytes int64
	if t.Matrix.IsRowMajor {
		strideInBytes = elemSize * t.Matrix.ColumnCount
	} else {
		strideInBytes = elemSize * t.Matrix.RowCount
	}
	t.Matrix.StrideInBytes = strideInBytes
	return strideInBytes
}

func matrixTypeStrideInElems(t *Type) int64 {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("matrix_type_stride_in_elems: expected matrix type")
	}
	stride := matrixTypeStrideInBytes(t, nil)
	elemSize := typeSizeOf(t.Matrix.Elem)
	if elemSize < 1 {
		elemSize = 1
	}
	return stride / elemSize
}

func matrixTypeTotalInternalElems(t *Type) int64 {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("matrix_type_total_internal_elems: expected matrix type")
	}
	size := typeSizeOf(t)
	elemSize := typeSizeOf(t.Matrix.Elem)
	if elemSize < 1 {
		elemSize = 1
	}
	return size / elemSize
}

func matrixIndicesToOffset(t *Type, rowIndex, columnIndex int64) int64 {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("matrix_indices_to_offset: expected matrix type")
	}
	if rowIndex < 0 || rowIndex >= t.Matrix.RowCount {
		gbAssertHandler("matrix_indices_to_offset: row index out of range")
	}
	if columnIndex < 0 || columnIndex >= t.Matrix.ColumnCount {
		gbAssertHandler("matrix_indices_to_offset: column index out of range")
	}
	strideElems := matrixTypeStrideInElems(t)
	if t.Matrix.IsRowMajor {
		return columnIndex + strideElems*rowIndex
	} else {
		return rowIndex + strideElems*columnIndex
	}
}

func matrixRowMajorIndexToOffset(t *Type, index int64) int64 {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("matrix_row_major_index_to_offset: expected matrix type")
	}
	rowIndex := index / t.Matrix.ColumnCount
	columnIndex := index % t.Matrix.ColumnCount
	return matrixIndicesToOffset(t, rowIndex, columnIndex)
}

func matrixColumnMajorIndexToOffset(t *Type, index int64) int64 {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("matrix_column_major_index_to_offset: expected matrix type")
	}
	rowIndex := index % t.Matrix.RowCount
	columnIndex := index / t.Matrix.RowCount
	return matrixIndicesToOffset(t, rowIndex, columnIndex)
}

func isMatrixSquare(t *Type) bool {
	t = baseType(t)
	if t.Kind != TypeMatrix {
		gbAssertHandler("is_matrix_square: expected matrix type")
	}
	return t.Matrix.RowCount == t.Matrix.ColumnCount
}

func isTypeValidForMatrixElems(t *Type) bool {
	t = baseType(t)
	if t == nil {
		return false
	}
	if isTypeInteger(t) || isTypeFloat(t) || isTypeComplex(t) {
		return true
	}
	if t.Kind == TypeGeneric {
		return true
	}
	return false
}
