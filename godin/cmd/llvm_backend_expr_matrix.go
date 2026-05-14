package cmd

func lb_is_matrix_simdable(t *Type) bool {
	mt := base_type(t)
	gb_assert_handler("Assertion Failure", "mt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 677, 0)
	elem := core_type(mt.Matrix.Elem)
	if is_type_complex(elem) {
		return false
	}
	if is_type_different_to_arch_endianness(elem) {
		return false
	}
	switch buildContext.Metrics.Arch {
	default:
		return false
	case TargetArchAmd64:
	case TargetArchArm64:
	}
	if type_align_of(t) < 16 {
		return false
	}
	if (mt.Matrix.RowCount & 1) ^ (mt.Matrix.ColumnCount & 1) != 0 {
		return false
	}
	if mt.Matrix.IsRowMajor {
		return false
	}
	if elem.Kind == Type_Basic {
		switch elem.Basic.Kind {
		case BasicF16, BasicF16le, BasicF16be:
			switch buildContext.Metrics.Arch {
			case TargetArchAmd64:
				return false
			case TargetArchArm64:
				return true
			case TargetArchI386, TargetArchWasm32, TargetArchWasm64p32:
				return false
			}
		}
	}
	return true
}

func lb_matrix_to_vector(p *lbProcedure, matrix lbValue) LLVMValueRef {
	mt := base_type(matrix.Type)
	gb_assert_handler("Assertion Failure", "mt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 733, 0)
	elem_type := lb_type(p.Module, mt.Matrix.Elem)
	total_count := uint(matrix_type_total_internal_elems(mt))
	total_matrix_type := LLVMVectorType(elem_type, total_count)
	ptr := lb_address_from_load_or_generate_local(p, matrix).Value
	matrix_vector_ptr := LLVMBuildPointerCast(p.Builder, ptr, LLVMPointerType(total_matrix_type, 0), "")
	matrix_vector := OdinLLVMBuildLoadAligned(p, total_matrix_type, matrix_vector_ptr, type_align_of(mt))
	return matrix_vector
}

func lb_matrix_trimmed_vector_mask(p *lbProcedure, mt_ *Type) LLVMValueRef {
	mt := base_type(mt_)
	gb_assert_handler("Assertion Failure", "mt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 752, 0)
	stride := uint(matrix_type_stride_in_elems(mt))
	row_count := uint(mt.Matrix.RowCount)
	column_count := uint(mt.Matrix.ColumnCount)
	mask_elems := make([]LLVMValueRef, row_count*column_count)
	mask_elems_index := uint(0)
	for j := uint(0); j < column_count; j++ {
		for i := uint(0); i < row_count; i++ {
			offset := stride*j + i
			mask_elems[mask_elems_index] = lb_const_int(p.Module, t_u32, u64(offset)).Value
			mask_elems_index++
		}
	}
	mask := LLVMConstVector(mask_elems, uint(len(mask_elems)))
	return mask
}

func lb_matrix_to_trimmed_vector(p *lbProcedure, m lbValue) LLVMValueRef {
	vector := lb_matrix_to_vector(p, m)
	mt := base_type(m.Type)
	gb_assert_handler("Assertion Failure", "mt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 774, 0)
	stride := uint(matrix_type_stride_in_elems(mt))
	row_count := uint(mt.Matrix.RowCount)
	if stride == row_count {
		return vector
	}
	mask := lb_matrix_trimmed_vector_mask(p, mt)
	trimmed_vector := llvm_basic_shuffle(p, vector, mask)
	return trimmed_vector
}

func lb_emit_matrix_transpose(p *lbProcedure, m lbValue, typ *Type) lbValue {
	if is_type_array(m.Type) {
		rank := type_math_rank(m.Type)
		if rank == 2 {
			addr := lb_add_local_generated(p, typ, false)
			dst := addr.Addr
			src := m
			n := i32(get_array_type_count(m.Type))
			m_ := i32(get_array_type_count(typ))
			for j := i32(0); j < m_; j++ {
				dst_col := lb_emit_struct_ep(p, dst, j)
				for i := i32(0); i < n; i++ {
					dst_row := lb_emit_struct_ep(p, dst_col, i)
					src_col := lb_emit_struct_ev(p, src, i)
					src_row := lb_emit_struct_ev(p, src_col, j)
					lb_emit_store(p, dst_row, src_row)
				}
			}
			return lb_addr_load(p, addr)
		}
		m.Type = typ
		return m
	}
	mt := base_type(m.Type)
	gb_assert_handler("Assertion Failure", "mt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 816, 0)
	rt := base_type(typ)
	if rt.Kind == Type_Matrix && rt.Matrix.IsRowMajor != mt.Matrix.IsRowMajor {
		gb_assert_handler("Panic", 0, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 820, "TODO: transpose with changing layout")
	}
	if lb_is_matrix_simdable(mt) && lb_is_matrix_simdable(typ) {
		stride := uint(matrix_type_stride_in_elems(mt))
		row_count := uint(mt.Matrix.RowCount)
		column_count := uint(mt.Matrix.ColumnCount)
		rows := make([]LLVMValueRef, row_count)
		mask_elems := make([]LLVMValueRef, column_count)
		vector := lb_matrix_to_vector(p, m)
		for i := uint(0); i < row_count; i++ {
			for j := uint(0); j < column_count; j++ {
				offset := stride*j + i
				mask_elems[j] = lb_const_int(p.Module, t_u32, u64(offset)).Value
			}
			mask := LLVMConstVector(mask_elems, column_count)
			row := llvm_basic_shuffle(p, vector, mask)
			rows[i] = row
		}
		res := lb_add_local_generated(p, typ, true)
		for i := isize(0); i < isize(len(rows)); i++ {
			row := rows[i]
			dst_row_ptr := lb_emit_matrix_epi(p, res.Addr, 0, i)
			ptr := dst_row_ptr.Value
			ptr = LLVMBuildPointerCast(p.Builder, ptr, LLVMPointerType(LLVMTypeOf(row), 0), "")
			LLVMBuildStore(p.Builder, row, ptr)
		}
		return lb_addr_load(p, res)
	}
	res := lb_add_local_generated(p, typ, true)
	r_count := mt.Matrix.RowCount
	c_count := mt.Matrix.ColumnCount
	for j := i64(0); j < c_count; j++ {
		for i := i64(0); i < r_count; i++ {
			src := lb_emit_matrix_ev(p, m, i, j)
			dst := lb_emit_matrix_epi(p, res.Addr, j, i)
			lb_emit_store(p, dst, src)
		}
	}
	return lb_addr_load(p, res)
}

func lb_matrix_cast_vector_to_type(p *lbProcedure, vector LLVMValueRef, typ *Type) lbValue {
	res := lb_add_local_generated(p, typ, true)
	res_ptr := res.Addr.Value
	align_type := type_align_of(typ)
	align_vector := lb_alignof(LLVMTypeOf(vector))
	alignment := uint(uint64(align_vector))
	if align_type > align_vector {
		alignment = uint(uint64(align_type))
	}
	LLVMSetAlignment(res_ptr, alignment)
	res_ptr = LLVMBuildPointerCast(p.Builder, res_ptr, LLVMPointerType(LLVMTypeOf(vector), 0), "")
	LLVMBuildStore(p.Builder, vector, res_ptr)
	return lb_addr_load(p, res)
}

func lb_emit_matrix_flatten(p *lbProcedure, m lbValue, typ *Type) lbValue {
	if is_type_array(m.Type) {
		m.Type = typ
		return m
	}
	mt := base_type(m.Type)
	gb_assert_handler("Assertion Failure", "mt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 889, 0)
	res := lb_add_local_generated(p, typ, true)
	gb_assert_handler("Assertion Failure", "type_size_of(type) == type_size_of(m.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 893, 0)
	m_ptr := lb_address_from_load_or_generate_local(p, m)
	n := lb_const_int(p.Module, t_int, u64(type_size_of(typ)))
	lb_mem_copy_non_overlapping(p, res.Addr, m_ptr, n)
	return lb_addr_load(p, res)
}

func lb_emit_outer_product(p *lbProcedure, a lbValue, b lbValue, typ *Type) lbValue {
	mt := base_type(typ)
	at := base_type(a.Type)
	bt := base_type(b.Type)
	gb_assert_handler("Assertion Failure", "mt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 932, 0)
	gb_assert_handler("Assertion Failure", "at->kind == Type_Array", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 933, 0)
	gb_assert_handler("Assertion Failure", "bt->kind == Type_Array", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 934, 0)
	row_count := mt.Matrix.RowCount
	column_count := mt.Matrix.ColumnCount
	gb_assert_handler("Assertion Failure", "row_count == at->Array.count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 940, 0)
	gb_assert_handler("Assertion Failure", "column_count == bt->Array.count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 941, 0)
	res := lb_add_local_generated(p, typ, true)
	for j := i64(0); j < column_count; j++ {
		for i := i64(0); i < row_count; i++ {
			x := lb_emit_struct_ev(p, a, i32(i))
			y := lb_emit_struct_ev(p, b, i32(j))
			src := lb_emit_arith(p, TokenMul, x, y, mt.Matrix.Elem)
			dst := lb_emit_matrix_epi(p, res.Addr, i, j)
			lb_emit_store(p, dst, src)
		}
	}
	return lb_addr_load(p, res)
}

func lb_emit_matrix_mul(p *lbProcedure, lhs lbValue, rhs lbValue, typ *Type) lbValue {
	xt := base_type(lhs.Type)
	yt := base_type(rhs.Type)
	gb_assert_handler("Assertion Failure", "is_type_matrix(type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 965, 0)
	gb_assert_handler("Assertion Failure", "is_type_matrix(xt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 966, 0)
	gb_assert_handler("Assertion Failure", "is_type_matrix(yt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 967, 0)
	gb_assert_handler("Assertion Failure", "xt->Matrix.column_count == yt->Matrix.row_count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 968, 0)
	gb_assert_handler("Assertion Failure", "are_types_identical(xt->Matrix.elem, yt->Matrix.elem)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 969, 0)
	gb_assert_handler("Assertion Failure", "xt->Matrix.is_row_major == yt->Matrix.is_row_major", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 970, 0)
	elem := xt.Matrix.Elem
	outer_rows := uint(xt.Matrix.RowCount)
	inner := uint(xt.Matrix.ColumnCount)
	outer_columns := uint(yt.Matrix.ColumnCount)
	if !xt.Matrix.IsRowMajor && lb_is_matrix_simdable(xt) {
		x_stride := uint(matrix_type_stride_in_elems(xt))
		y_stride := uint(matrix_type_stride_in_elems(yt))
		x_rows := make([]LLVMValueRef, outer_rows)
		y_columns := make([]LLVMValueRef, outer_columns)
		x_vector := lb_matrix_to_vector(p, lhs)
		y_vector := lb_matrix_to_vector(p, rhs)
		mask_elems := make([]LLVMValueRef, inner)
		for i := uint(0); i < outer_rows; i++ {
			for j := uint(0); j < inner; j++ {
				offset := x_stride*j + i
				mask_elems[j] = lb_const_int(p.Module, t_u32, u64(offset)).Value
			}
			mask := LLVMConstVector(mask_elems, inner)
			row := llvm_basic_shuffle(p, x_vector, mask)
			x_rows[i] = row
		}
		for i := uint(0); i < outer_columns; i++ {
			mask := llvm_mask_iota(p.Module, y_stride*i, inner)
			column := llvm_basic_shuffle(p, y_vector, mask)
			y_columns[i] = column
		}
		res := lb_add_local_generated(p, typ, true)
		for i := isize(0); i < isize(len(x_rows)); i++ {
			x_row := x_rows[i]
			for j := isize(0); j < isize(len(y_columns)); j++ {
				y_column := y_columns[j]
				elem_val := llvm_vector_dot(p, x_row, y_column)
				dst := lb_emit_matrix_epi(p, res.Addr, i, j)
				LLVMBuildStore(p.Builder, elem_val, dst.Value)
			}
		}
		return lb_addr_load(p, res)
	}
	if !xt.Matrix.IsRowMajor {
		res := lb_add_local_generated(p, typ, true)
		inners := make([][2]lbValue, inner)
		for j := uint(0); j < outer_columns; j++ {
			for i := uint(0); i < outer_rows; i++ {
				dst := lb_emit_matrix_epi(p, res.Addr, isize(i), isize(j))
				for k := uint(0); k < inner; k++ {
					inners[k][0] = lb_emit_matrix_ev(p, lhs, isize(i), isize(k))
					inners[k][1] = lb_emit_matrix_ev(p, rhs, isize(k), isize(j))
				}
				sum := lb_const_nil(p.Module, elem)
				for k := uint(0); k < inner; k++ {
					a := inners[k][0]
					b := inners[k][1]
					sum = lb_emit_mul_add(p, a, b, sum, elem)
				}
				lb_emit_store(p, dst, sum)
			}
		}
		return lb_addr_load(p, res)
	} else {
		res := lb_add_local_generated(p, typ, true)
		inners := make([][2]lbValue, inner)
		for i := uint(0); i < outer_rows; i++ {
			for j := uint(0); j < outer_columns; j++ {
				dst := lb_emit_matrix_epi(p, res.Addr, isize(i), isize(j))
				for k := uint(0); k < inner; k++ {
					inners[k][0] = lb_emit_matrix_ev(p, lhs, isize(i), isize(k))
					inners[k][1] = lb_emit_matrix_ev(p, rhs, isize(k), isize(j))
				}
				sum := lb_const_nil(p.Module, elem)
				for k := uint(0); k < inner; k++ {
					a := inners[k][0]
					b := inners[k][1]
					sum = lb_emit_mul_add(p, a, b, sum, elem)
				}
				lb_emit_store(p, dst, sum)
			}
		}
		return lb_addr_load(p, res)
	}
}

func lb_emit_matrix_mul_vector(p *lbProcedure, lhs lbValue, rhs lbValue, typ *Type) lbValue {
	mt := base_type(lhs.Type)
	vt := base_type(rhs.Type)
	gb_assert_handler("Assertion Failure", "is_type_matrix(mt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1077, 0)
	gb_assert_handler("Assertion Failure", "is_type_array_like(vt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1078, 0)
	vector_count := get_array_type_count(vt)
	gb_assert_handler("Assertion Failure", "mt->Matrix.column_count == vector_count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1082, 0)
	gb_assert_handler("Assertion Failure", "are_types_identical(mt->Matrix.elem, base_array_type(vt))", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1083, 0)
	elem := mt.Matrix.Elem
	if !mt.Matrix.IsRowMajor && lb_is_matrix_simdable(mt) {
		stride := uint(matrix_type_stride_in_elems(mt))
		row_count := uint(mt.Matrix.RowCount)
		column_count := uint(mt.Matrix.ColumnCount)
		m_columns := make([]LLVMValueRef, column_count)
		v_rows := make([]LLVMValueRef, column_count)
		matrix_vector := lb_matrix_to_vector(p, lhs)
		for column_index := uint(0); column_index < column_count; column_index++ {
			mask := llvm_mask_iota(p.Module, stride*column_index, row_count)
			column := llvm_basic_shuffle(p, matrix_vector, mask)
			m_columns[column_index] = column
		}
		for row_index := uint(0); row_index < column_count; row_index++ {
			value := LLVMBuildExtractValue(p.Builder, rhs.Value, row_index, "")
			row := llvm_vector_broadcast(p, value, row_count)
			v_rows[row_index] = row
		}
		gb_assert_handler("Assertion Failure", "column_count > 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1109, 0)
		var vector LLVMValueRef
		for i := uint(0); i < column_count; i++ {
			if i == 0 {
				vector = llvm_vector_mul(p, m_columns[i], v_rows[i])
			} else {
				vector = llvm_vector_mul_add(p, m_columns[i], v_rows[i], vector)
			}
		}
		return lb_matrix_cast_vector_to_type(p, vector, typ)
	}
	res := lb_add_local_generated(p, typ, true)
	vector_elem_type := base_array_type(rhs.Type)
	for i := i64(0); i < mt.Matrix.RowCount; i++ {
		for j := i64(0); j < mt.Matrix.ColumnCount; j++ {
			dst := lb_emit_matrix_epi(p, res.Addr, i, 0)
			d0 := lb_emit_load(p, dst)
			a := lb_emit_matrix_ev(p, lhs, i, j)
			b_value := LLVMBuildExtractValue(p.Builder, rhs.Value, uint(j), "")
			b := lbValue{Value: b_value, Type: vector_elem_type}
			c := lb_emit_mul_add(p, a, b, d0, elem)
			lb_emit_store(p, dst, c)
		}
	}
	return lb_addr_load(p, res)
}

func lb_emit_vector_mul_matrix(p *lbProcedure, lhs lbValue, rhs lbValue, typ *Type) lbValue {
	mt := base_type(rhs.Type)
	vt := base_type(lhs.Type)
	gb_assert_handler("Assertion Failure", "is_type_matrix(mt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1152, 0)
	gb_assert_handler("Assertion Failure", "is_type_array_like(vt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1153, 0)
	vector_count := get_array_type_count(vt)
	gb_assert_handler("Assertion Failure", "vector_count == mt->Matrix.row_count", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1157, 0)
	gb_assert_handler("Assertion Failure", "are_types_identical(mt->Matrix.elem, base_array_type(vt))", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1158, 0)
	elem := mt.Matrix.Elem
	if !mt.Matrix.IsRowMajor && lb_is_matrix_simdable(mt) {
		stride := uint(matrix_type_stride_in_elems(mt))
		row_count := uint(mt.Matrix.RowCount)
		column_count := uint(mt.Matrix.ColumnCount)
		_ = column_count
		m_columns := make([]LLVMValueRef, row_count)
		v_rows := make([]LLVMValueRef, row_count)
		matrix_vector := lb_matrix_to_vector(p, rhs)
		mask_elems := make([]LLVMValueRef, column_count)
		for row_index := uint(0); row_index < row_count; row_index++ {
			for column_index := uint(0); column_index < column_count; column_index++ {
				offset := row_index + column_index*stride
				mask_elems[column_index] = lb_const_int(p.Module, t_u32, u64(offset)).Value
			}
			mask := LLVMConstVector(mask_elems, column_count)
			column := llvm_basic_shuffle(p, matrix_vector, mask)
			m_columns[row_index] = column
		}
		for column_index := uint(0); column_index < row_count; column_index++ {
			value := LLVMBuildExtractValue(p.Builder, lhs.Value, column_index, "")
			row := llvm_vector_broadcast(p, value, column_count)
			v_rows[column_index] = row
		}
		gb_assert_handler("Assertion Failure", "row_count > 0", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1191, 0)
		var vector LLVMValueRef
		for i := uint(0); i < row_count; i++ {
			if i == 0 {
				vector = llvm_vector_mul(p, v_rows[i], m_columns[i])
			} else {
				vector = llvm_vector_mul_add(p, v_rows[i], m_columns[i], vector)
			}
		}
		res := lb_add_local_generated(p, typ, true)
		res_ptr := res.Addr.Value
		align_type := type_align_of(typ)
		align_vector := lb_alignof(LLVMTypeOf(vector))
		alignment := uint(uint64(align_vector))
		if align_type > align_vector {
			alignment = uint(uint64(align_type))
		}
		LLVMSetAlignment(res_ptr, alignment)
		res_ptr = LLVMBuildPointerCast(p.Builder, res_ptr, LLVMPointerType(LLVMTypeOf(vector), 0), "")
		LLVMBuildStore(p.Builder, vector, res_ptr)
		return lb_addr_load(p, res)
	}
	res := lb_add_local_generated(p, typ, true)
	vector_elem_type := base_array_type(rhs.Type)
	for j := i64(0); j < mt.Matrix.ColumnCount; j++ {
		for k := i64(0); k < mt.Matrix.RowCount; k++ {
			dst := lb_emit_matrix_epi(p, res.Addr, 0, j)
			d0 := lb_emit_load(p, dst)
			a_value := LLVMBuildExtractValue(p.Builder, lhs.Value, uint(k), "")
			a := lbValue{Value: a_value, Type: vector_elem_type}
			b := lb_emit_matrix_ev(p, rhs, k, j)
			c := lb_emit_mul_add(p, a, b, d0, elem)
			lb_emit_store(p, dst, c)
		}
	}
	return lb_addr_load(p, res)
}

func lb_emit_arith_matrix(p *lbProcedure, op TokenKind, lhs lbValue, rhs lbValue, typ *Type, component_wise bool) lbValue {
	gb_assert_handler("Assertion Failure", "is_type_matrix(lhs.type) || is_type_matrix(rhs.type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1238, 0)
	if op == TokenMul && !component_wise {
		xt := base_type(lhs.Type)
		yt := base_type(rhs.Type)
		if xt.Kind == Type_Matrix {
			if yt.Kind == Type_Matrix {
				return lb_emit_matrix_mul(p, lhs, rhs, typ)
			} else if is_type_array_like(yt) {
				return lb_emit_matrix_mul_vector(p, lhs, rhs, typ)
			}
		} else if is_type_array_like(xt) {
			gb_assert_handler("Assertion Failure", "yt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1251, 0)
			return lb_emit_vector_mul_matrix(p, lhs, rhs, typ)
		} else {
			gb_assert_handler("Assertion Failure", "xt->kind == Type_Basic", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1254, 0)
			gb_assert_handler("Assertion Failure", "yt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1255, 0)
			gb_assert_handler("Assertion Failure", "is_type_matrix(type)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1256, 0)
			array_type := alloc_type_array(yt.Matrix.Elem, matrix_type_total_internal_elems(yt), nil)
			gb_assert_handler("Assertion Failure", "type_size_of(array_type) == type_size_of(yt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1259, 0)
			array_lhs := lb_emit_conv(p, lhs, array_type)
			array_rhs := rhs
			array_rhs.Type = array_type
			array_val := lb_emit_arith(p, op, array_lhs, array_rhs, array_type)
			array_val.Type = typ
			return array_val
		}
	} else {
		if is_type_matrix(lhs.Type) {
			rhs = lb_emit_conv(p, rhs, lhs.Type)
		} else {
			lhs = lb_emit_conv(p, lhs, rhs.Type)
		}
		xt := base_type(lhs.Type)
		yt := base_type(rhs.Type)
		gb_assert_handler("Assertion Failure", "are_types_identical(xt, yt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1279, "%s %.*s %s", type_to_string(lhs.Type), token_strings[op].Len, token_strings[op].Data, type_to_string(rhs.Type))
		gb_assert_handler("Assertion Failure", "xt->kind == Type_Matrix", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1280, 0)
		array_lhs := lhs
		array_rhs := rhs
		array_type := alloc_type_array(xt.Matrix.Elem, matrix_type_total_internal_elems(xt), nil)
		gb_assert_handler("Assertion Failure", "type_size_of(array_type) == type_size_of(xt)", "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1286, 0)
		array_lhs.Type = array_type
		array_rhs.Type = array_type
		if token_is_comparison(op) {
			res := lb_emit_comp(p, op, array_lhs, array_rhs)
			return lb_emit_conv(p, res, typ)
		} else {
			array_val := lb_emit_arith(p, op, array_lhs, array_rhs, array_type)
			array_val.Type = typ
			return array_val
		}
	}
	gb_assert_handler("Panic", 0, "G:\\b0pass-win\\Odin\\src\\llvm_backend_expr.cpp", 1302, "TODO: lb_emit_arith_matrix")
	return lbValue{}
}
