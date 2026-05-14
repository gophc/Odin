package cmd

func checkBuiltinProc_overflow_add(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var x, y Operand
	check_expr(c, &x, ce.Args[0])
	check_expr(c, &y, ce.Args[1])
	if x.Mode == Addressing_Invalid || y.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, &y, x.Type)
	if y.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, &x, y.Type)
	if is_type_untyped(x.Type) {
		xts := type_to_string(x.Type)
		error(x.Expr, "Expected a typed integer for '%s', got %s", builtin_name, xts)
		gb_string_free(xts)
		return false
	}
	if !is_type_integer(x.Type) {
		xts := type_to_string(x.Type)
		error(x.Expr, "Expected an integer for '%s', got %s", builtin_name, xts)
		gb_string_free(xts)
		return false
	}
	ct := core_type(x.Type)
	if is_type_different_to_arch_endianness(ct) {
		if ct.Basic.Flags&(BasicFlag_EndianLittle|BasicFlag_EndianBig) != 0 {
			xts := type_to_string(x.Type)
			error(x.Expr, "Expected an integer which does not specify the explicit endianness for '%s', got %s", builtin_name, xts)
			gb_string_free(xts)
			return false
		}
	}
	if !are_types_identical(x.Type, y.Type) {
		xts := type_to_string(x.Type)
		yts := type_to_string(y.Type)
		error(x.Expr, "Mismatched types for '%s', got %s vs %s", builtin_name, xts, yts)
		gb_string_free(yts)
		gb_string_free(xts)
		return false
	}
	operand.Mode = Addressing_Value
	operand.Type = make_optional_ok_type(default_type(x.Type))
	return true
}

func checkBuiltinProc_saturating_add(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var x, y Operand
	check_expr(c, &x, ce.Args[0])
	check_expr(c, &y, ce.Args[1])
	if x.Mode == Addressing_Invalid || y.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, &y, x.Type)
	if y.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, &x, y.Type)
	if is_type_untyped(x.Type) {
		xts := type_to_string(x.Type)
		error(x.Expr, "Expected a typed integer for '%s', got %s", builtin_name, xts)
		gb_string_free(xts)
		return false
	}
	if !is_type_integer(x.Type) {
		xts := type_to_string(x.Type)
		error(x.Expr, "Expected an integer for '%s', got %s", builtin_name, xts)
		gb_string_free(xts)
		return false
	}
	ct := core_type(x.Type)
	if is_type_different_to_arch_endianness(ct) {
		if ct.Basic.Flags&(BasicFlag_EndianLittle|BasicFlag_EndianBig) != 0 {
			xts := type_to_string(x.Type)
			error(x.Expr, "Expected an integer which does not specify the explicit endianness for '%s', got %s", builtin_name, xts)
			gb_string_free(xts)
			return false
		}
	}
	if !are_types_identical(x.Type, y.Type) {
		xts := type_to_string(x.Type)
		yts := type_to_string(y.Type)
		error(x.Expr, "Mismatched types for '%s', got %s vs %s", builtin_name, xts, yts)
		gb_string_free(yts)
		gb_string_free(xts)
		return false
	}
	operand.Mode = Addressing_Value
	operand.Type = default_type(x.Type)
	return true
}

func checkBuiltinProc_mem_copy(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	operand.Mode = Addressing_NoValue
	operand.Type = t_invalid
	var dst, src, ln Operand
	check_expr(c, &dst, ce.Args[0])
	check_expr(c, &src, ce.Args[1])
	check_expr(c, &ln, ce.Args[2])
	if dst.Mode == Addressing_Invalid || src.Mode == Addressing_Invalid || ln.Mode == Addressing_Invalid {
		return false
	}
	if !is_type_pointer(dst.Type) && !is_type_multi_pointer(dst.Type) {
		str := type_to_string(dst.Type)
		error(dst.Expr, "Expected a pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if !is_type_pointer(src.Type) && !is_type_multi_pointer(src.Type) {
		str := type_to_string(src.Type)
		error(src.Expr, "Expected a pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if !is_type_integer(ln.Type) {
		str := type_to_string(ln.Type)
		error(ln.Expr, "Expected an integer value for the number of bytes for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if ln.Mode == Addressing_Constant {
		n := exact_value_to_i64(ln.Value)
		if n < 0 {
			str := expr_to_string(ln.Expr)
			error(ln.Expr, "Expected a non-negative integer value for the number of bytes for '%s', got %s", builtin_name, str)
			gb_string_free(str)
		}
	}
	return true
}

func checkBuiltinProc_mem_zero(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	operand.Mode = Addressing_NoValue
	operand.Type = t_invalid
	var ptr, ln Operand
	check_expr(c, &ptr, ce.Args[0])
	check_expr(c, &ln, ce.Args[1])
	if ptr.Mode == Addressing_Invalid || ln.Mode == Addressing_Invalid {
		return false
	}
	if !is_type_pointer(ptr.Type) && !is_type_multi_pointer(ptr.Type) {
		str := type_to_string(ptr.Type)
		error(ptr.Expr, "Expected a pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if !is_type_integer(ln.Type) {
		str := type_to_string(ln.Type)
		error(ln.Expr, "Expected an integer value for the number of bytes for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if ln.Mode == Addressing_Constant {
		n := exact_value_to_i64(ln.Value)
		if n < 0 {
			str := expr_to_string(ln.Expr)
			error(ln.Expr, "Expected a non-negative integer value for the number of bytes for '%s', got %s", builtin_name, str)
			gb_string_free(str)
		}
	}
	return true
}

func checkBuiltinProc_ptr_offset(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var ptr, offset Operand
	check_expr(c, &ptr, ce.Args[0])
	check_expr(c, &offset, ce.Args[1])
	if ptr.Mode == Addressing_Invalid {
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	if offset.Mode == Addressing_Invalid {
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	operand.Mode = Addressing_Value
	operand.Type = ptr.Type
	if !is_type_pointer(ptr.Type) && !is_type_multi_pointer(ptr.Type) {
		str := type_to_string(ptr.Type)
		error(ptr.Expr, "Expected a pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if are_types_identical(core_type(ptr.Type), t_rawptr) {
		str := type_to_string(ptr.Type)
		error(ptr.Expr, "Expected a dereferenceable pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if !is_type_integer(offset.Type) {
		str := type_to_string(offset.Type)
		error(offset.Expr, "Expected an integer value for the offset parameter for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	return true
}

func checkBuiltinProc_ptr_sub(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	operand.Mode = Addressing_NoValue
	operand.Type = t_invalid
	var ptr0, ptr1 Operand
	check_expr(c, &ptr0, ce.Args[0])
	check_expr(c, &ptr1, ce.Args[1])
	if ptr0.Mode == Addressing_Invalid {
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	if ptr1.Mode == Addressing_Invalid {
		operand.Mode = Addressing_Invalid
		operand.Type = t_invalid
		return false
	}
	operand.Mode = Addressing_Value
	operand.Type = t_int
	if !is_type_pointer(ptr0.Type) && !is_type_multi_pointer(ptr0.Type) {
		str := type_to_string(ptr0.Type)
		error(ptr0.Expr, "Expected a pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if are_types_identical(core_type(ptr0.Type), t_rawptr) {
		str := type_to_string(ptr0.Type)
		error(ptr0.Expr, "Expected a dereferenceable pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if !is_type_pointer(ptr1.Type) && !is_type_multi_pointer(ptr1.Type) {
		str := type_to_string(ptr1.Type)
		error(ptr1.Expr, "Expected a pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if are_types_identical(core_type(ptr1.Type), t_rawptr) {
		str := type_to_string(ptr1.Type)
		error(ptr1.Expr, "Expected a dereferenceable pointer value for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	if !are_types_identical(ptr0.Type, ptr1.Type) {
		xts := type_to_string(ptr0.Type)
		yts := type_to_string(ptr1.Type)
		error(ptr0.Expr, "Mismatched types for '%s', %s vs %s", builtin_name, xts, yts)
		gb_string_free(yts)
		gb_string_free(xts)
		return false
	}
	elem := type_deref(ptr0.Type, false)
	if type_size_of(elem) == 0 {
		str := type_to_string(ptr0.Type)
		error(ptr0.Expr, "Expected a pointer to a non-zero sized element for '%s', got %s", builtin_name, str)
		gb_string_free(str)
		return false
	}
	return true
}

func checkBuiltinProc_atomic_type_is_lock_free(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	expr := ce.Args[0]
	var o Operand
	check_expr_or_type(c, &o, expr)
	if o.Mode == Addressing_Invalid || o.Mode == Addressing_Builtin {
		return false
	}
	if o.Type == nil || o.Type == t_invalid || is_type_asm_proc(o.Type) {
		error(o.Expr, "Invalid argument to '%s'", builtin_name)
		return false
	}
	if is_type_polymorphic(o.Type) {
		error(o.Expr, "'%s' of polymorphic type cannot be determined", builtin_name)
		return false
	}
	if is_type_untyped(o.Type) {
		error(o.Expr, "'%s' of untyped type is not allowed", builtin_name)
		return false
	}
	t := o.Type
	is_lock_free := is_type_lock_free(t)
	operand.Mode = Addressing_Constant
	operand.Type = t_untyped_bool
	operand.Value = exact_value_bool(is_lock_free)
	return true
}

func checkBuiltinProc_atomic_thread_fence(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var memory_order OdinAtomicMemoryOrder
	if !check_atomic_memory_order_argument(c, ce.Args[0], builtin_name, &memory_order) {
		return false
	}
	switch memory_order {
	case OdinAtomicMemoryOrderAcquire, OdinAtomicMemoryOrderRelease,
		OdinAtomicMemoryOrderAcqRel, OdinAtomicMemoryOrderSeqCst:
		break
	default:
		error(ce.Args[0], "Illegal memory ordering for '%s', got .%s", builtin_name, OdinAtomicMemoryOrderStrings[memory_order])
	}
	operand.Mode = Addressing_NoValue
	return true
}

func checkBuiltinProc_volatile_store(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if BuiltinProcId(id) == BuiltinProc_atomic_store && !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	var x Operand
	check_expr_with_type_hint(c, &x, ce.Args[1], elem)
	check_assignment(c, &x, elem, builtin_name)
	operand.Type = nil
	operand.Mode = Addressing_NoValue
	return true
}

func checkBuiltinProc_atomic_store_explicit(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	var x Operand
	check_expr_with_type_hint(c, &x, ce.Args[1], elem)
	check_assignment(c, &x, elem, builtin_name)
	var memory_order OdinAtomicMemoryOrder
	if !check_atomic_memory_order_argument(c, ce.Args[2], builtin_name, &memory_order) {
		return false
	}
	switch memory_order {
	case OdinAtomicMemoryOrderConsume, OdinAtomicMemoryOrderAcquire, OdinAtomicMemoryOrderAcqRel:
		error(ce.Args[2], "Illegal memory order .%s for '%s'", OdinAtomicMemoryOrderStrings[memory_order], builtin_name)
	}
	operand.Type = nil
	operand.Mode = Addressing_NoValue
	return true
}

func checkBuiltinProc_volatile_load(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if BuiltinProcId(id) == BuiltinProc_atomic_load && !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	operand.Type = elem
	operand.Mode = Addressing_Value
	return true
}

func checkBuiltinProc_atomic_load_explicit(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	var memory_order OdinAtomicMemoryOrder
	if !check_atomic_memory_order_argument(c, ce.Args[1], builtin_name, &memory_order) {
		return false
	}
	switch memory_order {
	case OdinAtomicMemoryOrderRelease, OdinAtomicMemoryOrderAcqRel:
		error(ce.Args[1], "Illegal memory order .%s for '%s'", OdinAtomicMemoryOrderStrings[memory_order], builtin_name)
	}
	operand.Type = elem
	operand.Mode = Addressing_Value
	return true
}

func checkBuiltinProc_atomic_rmw(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	var x Operand
	check_expr_with_type_hint(c, &x, ce.Args[1], elem)
	check_assignment(c, &x, elem, builtin_name)
	t := type_deref(operand.Type, false)
	if BuiltinProcId(id) != BuiltinProc_atomic_exchange {
		if !is_type_integer_like(t) {
			str := type_to_string(t)
			error(operand.Expr, "Expected an integer type for '%s', got %s", builtin_name, str)
			gb_string_free(str)
		} else if is_type_different_to_arch_endianness(t) {
			str := type_to_string(t)
			error(operand.Expr, "Expected an integer type of the same platform endianness for '%s', got %s", builtin_name, str)
			gb_string_free(str)
		}
	}
	operand.Type = elem
	operand.Mode = Addressing_Value
	return true
}

func checkBuiltinProc_atomic_rmw_explicit(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	var x Operand
	check_expr_with_type_hint(c, &x, ce.Args[1], elem)
	check_assignment(c, &x, elem, builtin_name)
	if !check_atomic_memory_order_argument(c, ce.Args[2], builtin_name, nil) {
		return false
	}
	t := type_deref(operand.Type, false)
	if BuiltinProcId(id) != BuiltinProc_atomic_exchange_explicit {
		if !is_type_integer_like(t) {
			str := type_to_string(t)
			error(operand.Expr, "Expected an integer type for '%s', got %s", builtin_name, str)
			gb_string_free(str)
		} else if is_type_different_to_arch_endianness(t) {
			str := type_to_string(t)
			error(operand.Expr, "Expected an integer type of the same platform endianness for '%s', got %s", builtin_name, str)
			gb_string_free(str)
		}
	}
	operand.Type = elem
	operand.Mode = Addressing_Value
	return true
}

func checkBuiltinProc_atomic_compare_exchange_strong(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	var x, y Operand
	check_expr_with_type_hint(c, &x, ce.Args[1], elem)
	check_expr_with_type_hint(c, &y, ce.Args[2], elem)
	check_assignment(c, &x, elem, builtin_name)
	check_assignment(c, &y, elem, builtin_name)
	t := type_deref(operand.Type, false)
	if !is_type_comparable(t) {
		str := type_to_string(t)
		error(operand.Expr, "Expected a comparable type for '%s', got %s", builtin_name, str)
		gb_string_free(str)
	}
	operand.Mode = Addressing_OptionalOk
	operand.Type = elem
	return true
}

func checkBuiltinProc_atomic_compare_exchange_explicit(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var elem *Type
	if !is_type_normal_pointer(operand.Type, &elem) {
		error(operand.Expr, "Expected a pointer for '%s'", builtin_name)
		return false
	}
	if !check_atomic_ptr_argument(operand, builtin_name, elem) {
		return false
	}
	var x, y Operand
	check_expr_with_type_hint(c, &x, ce.Args[1], elem)
	check_expr_with_type_hint(c, &y, ce.Args[2], elem)
	check_assignment(c, &x, elem, builtin_name)
	check_assignment(c, &y, elem, builtin_name)
	var success_memory_order, failure_memory_order OdinAtomicMemoryOrder
	if !check_atomic_memory_order_argument(c, ce.Args[3], builtin_name, &success_memory_order, "success ordering") {
		return false
	}
	if !check_atomic_memory_order_argument(c, ce.Args[4], builtin_name, &failure_memory_order, "failure ordering") {
		return false
	}
	t := type_deref(operand.Type, false)
	if !is_type_comparable(t) {
		str := type_to_string(t)
		error(operand.Expr, "Expected a comparable type for '%s', got %s", builtin_name, str)
		gb_string_free(str)
	}
	invalid_combination := false
	switch success_memory_order {
	case OdinAtomicMemoryOrderRelaxed, OdinAtomicMemoryOrderRelease:
		if failure_memory_order != OdinAtomicMemoryOrderRelaxed {
			invalid_combination = true
		}
	case OdinAtomicMemoryOrderConsume:
		switch failure_memory_order {
		case OdinAtomicMemoryOrderRelaxed, OdinAtomicMemoryOrderConsume:
		default:
			invalid_combination = true
		}
	case OdinAtomicMemoryOrderAcquire, OdinAtomicMemoryOrderAcqRel:
		switch failure_memory_order {
		case OdinAtomicMemoryOrderRelaxed, OdinAtomicMemoryOrderConsume, OdinAtomicMemoryOrderAcquire:
		default:
			invalid_combination = true
		}
	case OdinAtomicMemoryOrderSeqCst:
		switch failure_memory_order {
		case OdinAtomicMemoryOrderRelaxed, OdinAtomicMemoryOrderConsume, OdinAtomicMemoryOrderAcquire, OdinAtomicMemoryOrderSeqCst:
		default:
			invalid_combination = true
		}
	default:
		invalid_combination = true
	}
	if invalid_combination {
		error(ce.Args[3], "Illegal memory order pairing for '%s', success = .%s, failure = .%s",
			builtin_name,
			OdinAtomicMemoryOrderStrings[success_memory_order],
			OdinAtomicMemoryOrderStrings[failure_memory_order],
		)
	}
	operand.Mode = Addressing_OptionalOk
	operand.Type = elem
	return true
}

func checkBuiltinProc_fixed_point_mul(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type, builtin_name string) bool {
	ce := &call.CallExpr
	var x, y, z Operand
	check_expr(c, &x, ce.Args[0])
	if x.Mode == Addressing_Invalid {
		return false
	}
	check_expr(c, &y, ce.Args[1])
	if y.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, &x, y.Type)
	if x.Mode == Addressing_Invalid {
		return false
	}
	convert_to_typed(c, &y, x.Type)
	if y.Mode == Addressing_Invalid {
		return false
	}
	if !are_types_identical(x.Type, y.Type) {
		xts := type_to_string(x.Type)
		yts := type_to_string(y.Type)
		error(x.Expr, "Mismatched types for '%s', %s vs %s", builtin_name, xts, yts)
		gb_string_free(yts)
		gb_string_free(xts)
		return false
	}
	if !is_type_integer(x.Type) || is_type_untyped(x.Type) {
		xts := type_to_string(x.Type)
		error(x.Expr, "Expected an integer type for '%s', got %s", builtin_name, xts)
		gb_string_free(xts)
		return false
	}
	check_expr(c, &z, ce.Args[2])
	if z.Mode == Addressing_Invalid {
		return false
	}
	if z.Mode != Addressing_Constant || !is_type_integer(z.Type) {
		error(z.Expr, "Expected a constant integer for the scale in '%s'", builtin_name)
		return false
	}
	n := exact_value_to_i64(z.Value)
	if n <= 0 {
		error(z.Expr, "Scale parameter in '%s' must be positive, got %d", builtin_name, n)
		return false
	}
	sz := 8 * type_size_of(x.Type)
	if n > sz {
		error(z.Expr, "Scale parameter in '%s' is larger than the base integer bit width, got %d, expected a maximum of %d", builtin_name, n, sz)
		return false
	}
	if sz >= 64 {
		if is_type_unsigned(x.Type) || is_type_unsigned(y.Type) {
			add_package_dependency(c, "runtime", "umodti3", true)
			add_package_dependency(c, "runtime", "udivti3", true)
		} else {
			add_package_dependency(c, "runtime", "modti3", true)
			add_package_dependency(c, "runtime", "divti3", true)
		}
	}
	operand.Type = x.Type
	operand.Mode = Addressing_Value
	return true
}
