package cmd

func check_procedure_param_polymorphic_type(ctx *CheckerContext, type_ *Type, type_expr *Ast) {
	if type_ == nil || type_expr == nil || ctx.InPolymorphicSpecialization {
		return
	}
	if !is_type_polymorphic_record_unspecialized(type_) {
		return
	}
	invalid_polymorphic_type_use := false
	switch type_expr.Kind {
	case Ast_Ident:
		invalid_polymorphic_type_use = true
	case Ast_SelectorExpr:
		invalid_polymorphic_type_use = true
	}
	if invalid_polymorphic_type_use {
		expr_str := expr_to_string(type_expr)
		defer gb_string_free(expr_str)
		error(type_expr, "Invalid use of a non-specialized polymorphic type '%s'", goStr(expr_str))
	}
}

func check_procedure_type(ctx *CheckerContext, type_ *Type, proc_type_node *Ast, operands []Operand) bool {
	pt := &proc_type_node.ProcType

	if ctx.PolymorphicScope == nil && ctx.AllowPolymorphicTypes {
		ctx.PolymorphicScope = ctx.Scope
	}

	c := *ctx
	c.CurrProcSig = type_
	c.InProcSig = true

	cc := pt.CallingConvention
	if cc == ProcCC_ForeignBlockDefault {
		cc = ProcCC_CDecl
	}
	if c.ForeignContext.DefaultCC > 0 {
		cc = c.ForeignContext.DefaultCC
	}
	if cc == ProcCC_Odin {
		c.Scope.Flags |= ScopeFlag_ContextDefined
	} else {
		c.Scope.Flags &^= ScopeFlag_ContextDefined
	}

	arch := buildContext.Metrics.Arch
	switch cc {
	case ProcCC_StdCall, ProcCC_FastCall:
		if arch != TargetArchI386 && arch != TargetArchAmd64 {
			error(proc_type_node, "Calling convention '%s' is not supported for this architecture", procCallingConventionStrings[cc])
		}
	case ProcCC_Win64, ProcCC_SysV:
		if arch != TargetArchAmd64 {
			error(proc_type_node, "Calling convention '%s' is only supported for the amd64 architecture", procCallingConventionStrings[cc])
		}
	}

	variadic := false
	variadic_index := isize(-1)
	success := true
	specialization_count := isize(0)

	params := check_get_params(&c, c.Scope, pt.Params, &variadic, &variadic_index, &success, &specialization_count, operands)

	no_poly_return := c.DisallowPolymorphicReturnTypes
	c.DisallowPolymorphicReturnTypes = (c.Scope == c.PolymorphicScope)
	results := check_get_results(&c, c.Scope, pt.Results)
	c.DisallowPolymorphicReturnTypes = no_poly_return

	param_count := isize(0)
	if params != nil {
		param_count = isize(len(params.Tuple.Variables))
	}
	result_count := isize(0)
	if results != nil {
		result_count = isize(len(results.Tuple.Variables))
	}

	if result_count > 0 {
		type_.Proc.HasNamedResults = results.Tuple.Variables[0].Token.String != ""
	}

	optional_ok := (pt.Tags & uint64(ProcTagOptionalOk)) != 0
	if optional_ok {
		if result_count != 2 {
			error(proc_type_node, "#optional_ok requires exactly 2 return values")
		} else if results != nil && results.Tuple.Variables[1] != nil {
			second_type := base_type(results.Tuple.Variables[1].Type)
			if !is_type_boolean(second_type) {
				error(proc_type_node, "#optional_ok expects the second return value to be of type 'bool'")
			}
		}
	}
	if (pt.Tags & uint64(ProcTagOptionalAllocatorError)) != 0 {
		if result_count != 2 {
			error(proc_type_node, "#optional_allocator_error requires exactly 2 return values")
		} else if results != nil && results.Tuple.Variables[1] != nil {
			second_type := base_type(results.Tuple.Variables[1].Type)
			if second_type != t_allocator_error {
				error(proc_type_node, "#optional_allocator_error expects the second return value to be of type 'Allocator_Error'")
			}
		}
	}

	type_.Proc.Node = proc_type_node
	type_.Proc.Scope = c.Scope
	type_.Proc.Params = params
	type_.Proc.ParamCount = int32(param_count)
	type_.Proc.Results = results
	type_.Proc.ResultCount = int32(result_count)
	type_.Proc.Variadic = variadic
	type_.Proc.VariadicIndex = int32(variadic_index)
	type_.Proc.CallingConvention = cc
	type_.Proc.IsPolymorphic = pt.Generic
	type_.Proc.SpecializationCount = specialization_count
	type_.Proc.Diverging = pt.Diverging
	type_.Proc.OptionalOK = optional_ok

	if params != nil {
		for _, param := range params.Tuple.Variables {
			if param == nil {
				continue
			}
			var type_expr *Ast
			if param.DeclInfo != nil {
				type_expr = param.DeclInfo.TypeExpr
			}
			check_procedure_param_polymorphic_type(&c, param.Type, type_expr)

			if is_type_polymorphic(param.Type) {
				type_.Proc.IsPolymorphic = true
			}
			if param.Flags&int32(FieldFlagCVararg) != 0 {
				type_.Proc.CVararg = true
			}
		}
	}
	if results != nil {
		for _, result := range results.Tuple.Variables {
			if result == nil {
				continue
			}
			var type_expr *Ast
			if result.DeclInfo != nil {
				type_expr = result.DeclInfo.TypeExpr
			}
			check_procedure_param_polymorphic_type(&c, result.Type, type_expr)

			if is_type_polymorphic(result.Type) {
				type_.Proc.IsPolymorphic = true
			}
		}
	}

	return success
}
