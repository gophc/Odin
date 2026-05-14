package cmd

func does_require_msgSend_stret(return_type *Type) bool {
	if return_type == nil {
		return false
	}
	if buildContext.Metrics.Os != TargetOs_darwin {
		return false
	}
	if buildContext.Metrics.Arch == TargetArch_i386 || buildContext.Metrics.Arch == TargetArch_amd64 {
		struct_limit := type_size_of(t_uintptr) << 1
		return type_size_of(return_type) > struct_limit
	}
	if buildContext.Metrics.Arch == TargetArch_arm64 {
		return false
	}
	if buildContext.Metrics.Arch == TargetArch_riscv64 {
		return false
	}
	gb_assert_handler("Panic", 0, "cmd_check_builtin_objc.go", 0, "unsupported architecture")
	return false
}

func get_objc_proc_kind(return_type *Type) ObjcMsgKind {
	if return_type == nil {
		return ObjcMsg_normal
	}
	if buildContext.Metrics.Arch == TargetArch_i386 || buildContext.Metrics.Arch == TargetArch_amd64 {
		if is_type_float(return_type) {
			return ObjcMsg_fpret
		}
		if buildContext.Metrics.Arch == TargetArch_amd64 {
			if is_type_complex(return_type) {
				return ObjcMsg_fpret
			}
		}
	}
	if buildContext.Metrics.Arch != TargetArch_arm64 {
		if does_require_msgSend_stret(return_type) {
			return ObjcMsg_stret
		}
	}
	return ObjcMsg_normal
}

func add_objc_proc_type(c *CheckerContext, call *Ast, return_type *Type, param_types []*Type) {
	kind := get_objc_proc_kind(return_type)
	scope := create_scope(c.Info, nil)
	params := alloc_type_tuple()
	{
		variables := make([]*Entity, 0, len(param_types))
		for _, type_ := range param_types {
			param := alloc_entity_param(scope, BlankToken, type_, false, true)
			variables = append(variables, param)
		}
		params.Tuple.Variables = variables
	}
	results := alloc_type_tuple()
	if return_type != nil {
		variables := make([]*Entity, 1)
		results.Tuple.Variables = variables
		param := alloc_entity_param(scope, BlankToken, return_type, false, true)
		results.Tuple.Variables[0] = param
	}
	var data ObjcMsgData
	data.Kind = kind
	data.ProcType = alloc_type_proc(scope, params, isize(len(param_types)), results, isize(len(results.Tuple.Variables)), false, ProcCC_CDecl)
	mutex_lock(&c.Info.ObjcObjcMsgSendMutex)
	map_set(&c.Info.ObjcMsgSendTypes, call, data)
	mutex_unlock(&c.Info.ObjcObjcMsgSendMutex)
	try_to_add_package_dependency(c, "runtime", "objc_msgSend")
	try_to_add_package_dependency(c, "runtime", "objc_msgSend_fpret")
	try_to_add_package_dependency(c, "runtime", "objc_msgSend_fp2ret")
	try_to_add_package_dependency(c, "runtime", "objc_msgSend_stret")
	args := call.CallExpr.Args
	if len(args) > 0 && args[0].TAV.ObjCSuperTarget != nil {
		try_to_add_package_dependency(c, "runtime", "objc_msgSendSuper2")
		try_to_add_package_dependency(c, "runtime", "objc_msgSendSuper2_stret")
	}
}

func check_builtin_objc_procedure(c *CheckerContext, operand *Operand, call *Ast, id int32, type_hint *Type) bool {
	builtin_name := builtin_procs[id].Name
	if buildContext.Metrics.Os != TargetOs_darwin {
		if buildContext.CommandKind != Command_doc && buildContext.CommandKind != Command_check {
			error(call, "'%.*s' only works on darwin", builtin_name.Len, builtin_name.Data)
		}
	}
	ce := call.CallExpr
	switch id {
	default:
		gb_assert_handler("Panic", 0, "cmd_check_builtin_objc.go", 0, "Implement objective built-in procedure: %.*s", builtin_name.Len, builtin_name.Data)
		return false

	case BuiltinProc_objc_send:
		var return_type *Type
		var rt Operand
		check_expr_or_type(c, &rt, ce.Args[0])
		if rt.Mode == Addressing_Type {
			return_type = rt.Type
		} else if is_operand_nil(rt) {
			return_type = nil
		} else {
			e := expr_to_string(rt.Expr)
			error(rt.Expr, "'%.*s' expected a type or nil to define the return type of the Objective-C call, got %s", builtin_name.Len, builtin_name.Data, e)
			gb_string_free(e)
			return false
		}
		operand.Type = return_type
		if return_type != nil {
			operand.Mode = Addressing_Value
		} else {
			operand.Mode = Addressing_NoValue
		}
		var class_name String
		var sel_name String
		sel_type := t_objc_sel
		var self Operand
		check_expr_or_type(c, &self, ce.Args[1])
		if self.Mode == Addressing_Type {
			if !is_type_objc_object(self.Type) {
				t := type_to_string(self.Type)
				error(self.Expr, "'%.*s' expected a type or value derived from intrinsics.objc_object, got type %s", builtin_name.Len, builtin_name.Data, t)
				gb_string_free(t)
				return false
			}
			if !has_type_got_objc_class_attribute(self.Type) {
				t := type_to_string(self.Type)
				error(self.Expr, "'%.*s' expected a named type with the attribute @(obj_class=<string>) , got type %s", builtin_name.Len, builtin_name.Data, t)
				gb_string_free(t)
				return false
			}
			sel_type = t_objc_class2
		} else if !is_operand_value(self) || !check_is_assignable_to(c, &self, t_objc_id) {
			e := expr_to_string(self.Expr)
			t := type_to_string(self.Type)
			error(self.Expr, "'%.*s' expected a type or value derived from intrinsics.objc_object, got '%s' of type %s", builtin_name.Len, builtin_name.Data, e, t)
			gb_string_free(t)
			gb_string_free(e)
			return false
		} else if !is_type_pointer(self.Type) {
			e := expr_to_string(self.Expr)
			t := type_to_string(self.Type)
			error(self.Expr, "'%.*s' expected a pointer of a value derived from intrinsics.objc_object, got '%s' of type %s", builtin_name.Len, builtin_name.Data, e, t)
			gb_string_free(t)
			gb_string_free(e)
			return false
		} else {
			type_ := type_deref(self.Type)
			if !(type_.Kind == Type_Named &&
				type_.Named.TypeName != nil &&
				type_.Named.TypeName.TypeName.ObjcClassName != "") {
				t := type_to_string(type_)
				error(self.Expr, "'%.*s' expected a named type with the attribute @(obj_class=<string>) , got type %s", builtin_name.Len, builtin_name.Data, t)
				gb_string_free(t)
				return false
			}
		}
		if !is_constant_string(c, builtin_name, ce.Args[2], &sel_name) {
			return false
		}
		const arg_offset = 1
		param_types := make([]*Type, len(ce.Args)-arg_offset)
		param_types[0] = t_objc_id
		param_types[1] = sel_type
		for i := 2 + arg_offset; i < len(ce.Args); i++ {
			var x Operand
			check_expr(c, &x, ce.Args[i])
			if is_type_untyped(x.Type) {
				e := expr_to_string(x.Expr)
				t := type_to_string(x.Type)
				error(x.Expr, "'%.*s' expects typed parameters, got %s of type %s", builtin_name.Len, builtin_name.Data, e, t)
				gb_string_free(t)
				gb_string_free(e)
			}
			param_types[i-arg_offset] = x.Type
		}
		add_objc_proc_type(c, call, return_type, param_types)
		return true

	case BuiltinProc_objc_find_selector,
		BuiltinProc_objc_find_class,
		BuiltinProc_objc_register_selector,
		BuiltinProc_objc_register_class:

		var sel_name String
		if !is_constant_string(c, builtin_name, ce.Args[0], &sel_name) {
			return false
		}
		switch id {
		case BuiltinProc_objc_find_selector,
			BuiltinProc_objc_register_selector:
			operand.Type = t_objc_sel
		case BuiltinProc_objc_find_class,
			BuiltinProc_objc_register_class:
			operand.Type = t_objc_class2
		}
		operand.Mode = Addressing_Value
		try_to_add_package_dependency(c, "runtime", "objc_lookUpClass")
		try_to_add_package_dependency(c, "runtime", "sel_registerName")
		try_to_add_package_dependency(c, "runtime", "objc_allocateClassPair")
		return true

	case BuiltinProc_objc_ivar_get:
		var self_type *Type
		var self Operand
		check_expr_or_type(c, &self, ce.Args[0])
		if !is_operand_value(self) || !check_is_assignable_to(c, &self, t_objc_id) {
			e := expr_to_string(self.Expr)
			t := type_to_string(self.Type)
			error(self.Expr, "'%.*s' expected a type or value derived from intrinsics.objc_object, got '%s' of type %s", builtin_name.Len, builtin_name.Data, e, t)
			gb_string_free(t)
			gb_string_free(e)
			return false
		} else if !is_type_pointer(self.Type) {
			e := expr_to_string(self.Expr)
			t := type_to_string(self.Type)
			error(self.Expr, "'%.*s' expected a pointer of a value derived from intrinsics.objc_object, got '%s' of type %s", builtin_name.Len, builtin_name.Data, e, t)
			gb_string_free(t)
			gb_string_free(e)
			return false
		}
		self_type = type_deref(self.Type)
		if !(self_type.Kind == Type_Named &&
			self_type.Named.TypeName != nil &&
			self_type.Named.TypeName.TypeName.ObjcClassName != "") {
			t := type_to_string(self_type)
			error(self.Expr, "'%.*s' expected a named type with the attribute @(objc_class=<string>) , got type %s", builtin_name.Len, builtin_name.Data, t)
			gb_string_free(t)
			return false
		}
		ivar_type := self_type.Named.TypeName.TypeName.ObjcIvar
		if ivar_type == nil {
			t := type_to_string(self_type)
			error(self.Expr, "'%.*s' requires that type %s have the attribute @(objc_ivar=<ivar_type_name>).", builtin_name.Len, builtin_name.Data, t)
			gb_string_free(t)
			return false
		}
		if type_hint != nil && type_hint.Kind == Type_Pointer && type_hint.Pointer.Elem == ivar_type {
			operand.Type = type_hint
		} else {
			operand.Type = alloc_type_pointer(ivar_type)
		}
		operand.Mode = Addressing_Value
		return true

	case BuiltinProc_objc_block:
		param_operands := make([]Operand, len(ce.Args))
		capture_arg_count := len(ce.Args) - 1
		param_operands[0] = *operand
		for i := 0; i < len(ce.Args)-1; i++ {
			var x Operand
			check_expr(c, &x, ce.Args[i])
			switch x.Mode {
			case Addressing_Value,
				Addressing_Context,
				Addressing_Variable,
				Addressing_Constant:
				param_operands[i] = x
			default:
				e := expr_to_string(x.Expr)
				t := type_to_string(x.Type)
				error(x.Expr, "'%.*s' capture arguments must be values, but got %s of type %s", builtin_name.Len, builtin_name.Data, e, t)
				gb_string_free(t)
				gb_string_free(e)
				return false
			}
		}
		var handler Operand
		if capture_arg_count == 0 {
			handler = param_operands[0]
		} else {
			check_expr_or_type(c, &handler, ce.Args[capture_arg_count])
			param_operands[capture_arg_count] = handler
		}
		if !is_operand_value(handler) || handler.Type.Kind != Type_Proc {
			e := expr_to_string(handler.Expr)
			t := type_to_string(handler.Type)
			error(handler.Expr, "'%.*s' expected a procedure, but got '%s' of type %s", builtin_name.Len, builtin_name.Data, e, t)
			gb_string_free(t)
			gb_string_free(e)
			return false
		}
		handler_node := unparen_expr(handler.Expr)
		switch handler_node.Kind {
		case Ast_ProcLit:
		case Ast_Ident:
			ident := handler_node.Ident
			if ident.Entity == nil {
				error(handler.Expr, "'%.*s' failed to resolve entity from expression", builtin_name.Len, builtin_name.Data)
				return false
			}
			if ident.Entity.Load().Kind != Entity_Procedure {
				e := expr_to_string(handler_node)
				begin_error_block()
				error(handler.Expr, "'%.*s' expected a direct reference to a procedure", builtin_name.Len, builtin_name.Data)
				if ident.Entity.Load().Kind == Entity_Variable {
					error_line("\tSuggestion: Variables referencing a procedure are not allowed, they are not a direct procedure reference.")
				} else {
					error_line("\tSuggestion: Ensure '%s' is not a runtime-evaluated expression.", e)
				}
				error_line("\n\t            Refer to a procedure directly by its name or declare it anonymously: %.*s(proc(){})", builtin_name.Len, builtin_name.Data)
				gb_string_free(e)
				end_error_block()
				return false
			}
		default:
			e := expr_to_string(handler_node)
			begin_error_block()
			error(handler.Expr, "'%.*s' expected a direct reference to a procedure", builtin_name.Len, builtin_name.Data)
			if handler_node.Kind == Ast_CallExpr {
				error_line("\tSuggestion: Do not use a procedure returned from another procedure.")
			} else {
				error_line("\tSuggestion: Ensure '%s' is not a runtime-evaluated expression.", e)
			}
			error_line("\n\t            Refer to a procedure directly by its name or declare it anonymously: %.*s(proc(){})", builtin_name.Len, builtin_name.Data)
			gb_string_free(e)
			end_error_block()
			return false
		}
		handler_type_proc := handler.Type.Proc
		if capture_arg_count > int(handler_type_proc.ParamCount) {
			error(handler.Expr, "'%.*s' captured arguments exceeded the handler's parameter count", builtin_name.Len, builtin_name.Data)
			return false
		}
		if handler_type_proc.CallingConvention == ProcCC_Odin {
			if (c.Scope.Flags & ScopeFlag_ContextDefined) == 0 {
				begin_error_block()
				error(handler.Expr, "The handler procedure for '%.*s' requires a context, but no context is defined in the current scope", builtin_name.Len, builtin_name.Data)
				error_line("\tSuggestion: 'context = runtime.default_context()', or use the \"c\" calling convention for the handler procedure")
				end_error_block()
				return false
			}
		}
		if handler_type_proc.ResultCount > 1 {
			error(handler_type_proc.Node.ProcType.Results, "Handler procedures for '%.*s' cannot have multiple return values", builtin_name.Len, builtin_name.Data)
			return false
		}
		if handler_type_proc.ParamCount > 0 {
			handler_param_types := handler.Type.Proc.Params.Tuple.Variables
			handler_capture_param_types := handler_param_types[len(handler_param_types)-capture_arg_count:]
			for i := 0; i < capture_arg_count; i++ {
				op := param_operands[i]
				if !check_is_assignable_to(c, &op, handler_capture_param_types[i].Type) {
					e := expr_to_string(op.Expr)
					src := type_to_string(op.Type)
					dst := type_to_string(handler_capture_param_types[i].Type)
					error(op.Expr, "'%.*s' captured value '%s' of type '%s' is not assignable to type '%s'", builtin_name.Len, builtin_name.Data, e, src, dst)
					gb_string_free(e)
					gb_string_free(src)
					gb_string_free(dst)
					return false
				}
			}
		}
		cc := handler_type_proc.CallingConvention
		switch cc {
		case ProcCC_Odin, ProcCC_Contextless, ProcCC_CDecl:
		default:
			begin_error_block()
			error(handler.Expr, "'%.*s' Invalid calling convention for block procedure.", builtin_name.Len, builtin_name.Data)
			error_line("\tSuggestion: Do not specify a calling convention or else use \"c\" or \"contextless\"")
			end_error_block()
			return false
		}
		if handler_type_proc.IsPolymorphic {
			error(handler.Expr, "'%.*s' Unspecialized polymorphic procedures are not allowed.", builtin_name.Len, builtin_name.Data)
			return false
		}
		ident := Token{}
		ident.Kind = Token_Ident
		ident.String = S("Objc_Block")
		ident.Pos = ast_token(call).Pos
		l_paren := Token{}
		l_paren.Kind = Token_OpenParen
		l_paren.String = S("(")
		l_paren.Pos = ident.Pos
		r_paren := Token{}
		r_paren.Kind = Token_CloseParen
		r_paren.String = S(")")
		r_paren.Pos = ident.Pos
		handler_proc_type_copy := clone_ast(handler_type_proc.Node, nil)
		handler_proc_type_copy.ProcType.Params.FieldList.List = handler_proc_type_copy.ProcType.Params.FieldList.List[:len(handler_proc_type_copy.ProcType.Params.FieldList.List)-capture_arg_count]
		handler_proc_type_copy.ProcType.CallingConvention = ProcCC_CDecl
		poly_args := make([]*Ast, 1)
		poly_args[0] = handler_proc_type_copy
		t_Objc_Block := find_core_type(c.Checker, S("Objc_Block"))
		var poly_op Operand
		poly_op.Type = t_Objc_Block
		poly_op.Mode = Addressing_Type
		poly_call := ast_call_expr(nil, ast_ident(nil, ident), poly_args, l_paren, r_paren, Token{})
		err := check_polymorphic_record_type(c, &poly_op, poly_call)
		if err != 0 {
			operand.Mode = Addressing_Invalid
			operand.Type = t_invalid
			error(handler.Expr, "'%.*s' failed to determine resulting Objc_Block handler procedure", builtin_name.Len, builtin_name.Data)
			return false
		}
		gb_assert_handler("Assertion Failure", "poly_op.Type != t_Objc_Block", "cmd_check_builtin_objc.go", 0, 0)
		gb_assert_handler("Assertion Failure", "poly_op.Mode == Addressing_Type", "cmd_check_builtin_objc.go", 0, 0)
		is_global_block := capture_arg_count == 0 && handler_type_proc.CallingConvention != ProcCC_Odin
		if is_global_block {
			try_to_add_package_dependency(c, "runtime", "_NSConcreteGlobalBlock")
		} else {
			try_to_add_package_dependency(c, "runtime", "_NSConcreteStackBlock")
		}
		*operand = poly_op
		operand.Type = alloc_type_pointer(operand.Type)
		operand.Mode = Addressing_Value
		return true

	case BuiltinProc_objc_super:
		objc_obj := operand.Type
		if !is_type_objc_ptr_to_object(objc_obj) {
			e := expr_to_string(operand.Expr)
			t := type_to_string(objc_obj)
			error(operand.Expr, "'%.*s' expected a pointer to an Objective-C object, but got '%s' of type %s", builtin_name.Len, builtin_name.Data, e, t)
			gb_string_free(t)
			gb_string_free(e)
			return false
		}
		if operand.Mode != Addressing_Value && operand.Mode != Addressing_Variable {
			e := expr_to_string(operand.Expr)
			t := type_to_string(operand.Type)
			error(operand.Expr, "'%.*s' expression '%s', of type %s, must be a value or variable.", builtin_name.Len, builtin_name.Data, e, t)
			gb_string_free(t)
			gb_string_free(e)
			return false
		}
		obj_type := type_deref(objc_obj)
		gb_assert_handler("Assertion Failure", "obj_type.Kind == Type_Named", "cmd_check_builtin_objc.go", 0)
		call.TAV.ObjCSuperTarget = obj_type
		superclass := obj_type.Named.TypeName.TypeName.ObjcSuperclass
		if superclass == nil {
			t := type_to_string(obj_type)
			error(operand.Expr, "'%.*s' target object '%.*s' does not have an Objective-C superclass. One must be set via the @(objc_superclass) attribute", builtin_name.Len, builtin_name.Data, t.Len, t.Data)
			gb_string_free(t)
			return false
		}
		gb_assert_handler("Assertion Failure", "superclass.Named.TypeName.TypeName.ObjcClassName.Len > 0", "cmd_check_builtin_objc.go", 0)
		operand.Type = alloc_type_pointer(superclass)
		return true
	}
}

