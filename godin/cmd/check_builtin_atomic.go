package cmd

func check_atomic_memory_order_argument(c *CheckerContext, expr *Ast, builtin_name String, memory_order_ *OdinAtomicMemoryOrder, extra_message string) bool {
	var x Operand
	check_expr_with_type_hint(c, &x, expr, t_atomic_memory_order)
	if x.Mode == Addressing_Invalid {
		return false
	}
	if !are_types_identical(x.Type, t_atomic_memory_order) || x.Mode != Addressing_Constant {
		str := type_to_string(x.Type)
		if extra_message != "" {
			error(x.Expr, "Expected a constant Atomic_Memory_Order value for the %s of '%.*s', got %s", extra_message, builtin_name.Len, builtin_name.Data, str)
		} else {
			error(x.Expr, "Expected a constant Atomic_Memory_Order value for '%.*s', got %s", builtin_name.Len, builtin_name.Data, str)
		}
		gb_string_free(str)
		return false
	}
	value := exact_value_to_i64(x.Value)
	if value < 0 || value >= int64(OdinAtomicMemoryOrderCOUNT) {
		error(x.Expr, "Illegal Atomic_Memory_Order value, got %d", value)
		return false
	}
	if memory_order_ != nil {
		*memory_order_ = OdinAtomicMemoryOrder(value)
	}
	return true
}

func check_atomic_ptr_argument(operand *Operand, builtin_name String, elem *Type) bool {
	if !is_type_valid_atomic_type(elem) {
		error(operand.Expr, "Only an integer, floating-point, boolean, or pointer can be used as an atomic for '%.*s'", builtin_name.Len, builtin_name.Data)
		return false
	}
	return true
}

