package cmd

func is_load_directive_call(call *Ast) bool {
	call = unparen_expr(call)
	if call.Kind != AstCallExpr {
		return false
	}
	ce := &call.CallExpr
	if ce.Proc.Kind != AstBasicDirective {
		return false
	}
	bd := &ce.Proc.BasicDirective
	name := bd.Name.String
	return name == "load"
}

func is_call_expr_field_value(ce *AstCallExpr) bool {
	if len(ce.Args) == 0 {
		return false
	}
	return ce.Args[0].Kind == AstFieldValue
}

func evaluate_where_clauses(ctx *CheckerContext, call_expr *Ast, scope *Scope, clauses []*Ast, print_err bool) bool {
	if len(clauses) != 0 {
		for _, clause := range clauses {
			var o Operand
			check_expr(ctx, &o, clause)
			if o.Mode != AddressingConstant {
				if print_err {
					error(clause, "'where' clauses expect a constant boolean evaluation")
				}
				if print_err && call_expr != nil {
					error(call_expr, "at caller location")
				}
				return false
			} else if o.Value.Kind != ExactValueBool {
				if print_err {
					error(clause, "'where' clauses expect a constant boolean evaluation")
				}
				if print_err && call_expr != nil {
					error(call_expr, "at caller location")
				}
				return false
			} else if !o.Value.ValueBool {
				if print_err {
					begin_errorblock()
					str := expr_to_string(clause)
					error(clause, "'where' clause evaluated to false:\n\t%s", str)
					gb_string_free(str)
					if scope != nil {
						print_count := 0
						for _, entry := range scope.Elements {
							e := entry.Value
							switch e.Kind {
							case Entity_TypeName:
								str := type_to_string(e.Type)
								errorline("\t\t%.*s :: %s;\n", e.Token.String, str)
								gb_string_free(str)
								print_count += 1
							case Entity_Constant:
								if print_count == 0 {
									errorline("\n\tWith the following definitions:\n")
								}
								str := exact_value_to_string(e.Constant.Value)
								if is_type_untyped(e.Type) {
									errorline("\t\t%.*s :: %s;\n", e.Token.String, str)
								} else {
									t := type_to_string(e.Type)
									errorline("\t\t%.*s : %s : %s;\n", e.Token.String, t, str)
									gb_string_free(t)
								}
								gb_string_free(str)
								print_count += 1
							}
						}
					}
					if call_expr != nil {
						pos := ast_token(call_expr).Pos
						errorline("%s at caller location\n", token_pos_to_string(pos))
					}
					end_errorblock()
				}
				return false
			}
			if ast_file_vet_style(ctx.File) {
				c := unparen_expr(clause)
				if c.Kind == AstBinaryExpr && c.BinaryExpr.Op.Kind == TokenCmpAnd {
					begin_errorblock()
					error(c, "Prefer to separate 'where' clauses with a comma rather than '&&'")
					x := expr_to_string(c.BinaryExpr.Left)
					y := expr_to_string(c.BinaryExpr.Right)
					errorline("\tSuggestion: '%s, %s'\n", x, y)
					gb_string_free(y)
					gb_string_free(x)
					end_errorblock()
				}
			}
		}
	}
	return true
}

func check_named_arguments(c *CheckerContext, type_ *Type, named_args []*Ast, named_operands *[]Operand, show_error bool) bool {
	success := true
	type_ = base_type(type_)
	if len(named_args) > 0 {
		var pt *TypeProc
		if is_type_proc(type_) {
			pt = &type_.Proc
		}
		for _, arg := range named_args {
			if arg.Kind != AstFieldValue {
				if show_error {
					error(arg, "Expected a 'field = value'")
				}
				return false
			}
			fv := &arg.FieldValue
			if fv.Field.Kind != AstIdent {
				if show_error {
					expr_str := expr_to_string(fv.Field)
					error(arg, "Invalid parameter name '%s' in procedure call", expr_str)
					gb_string_free(expr_str)
				}
				success = false
				continue
			}
			key := fv.Field.Ident.Token.String
			value := fv.Value
			var type_hint *Type
			if pt != nil {
				param_index := lookup_procedure_parameter(pt, key)
				if param_index < 0 {
					if show_error {
						error(value, "No parameter named '%.*s' for this procedure type", key)
					}
					success = false
					continue
				}
				e := pt.Params.Tuple.Variables[param_index]
				if !is_type_polymorphic(e.Type) {
					type_hint = e.Type
				}
			}
			var o Operand
			check_expr_with_type_hint(c, &o, value, type_hint)
			if o.Mode == AddressingInvalid {
				success = false
			}
			*named_operands = append(*named_operands, o)
		}
	}
	return success
}

func check_assignment_arguments(ctx *CheckerContext, lhs []Operand, operands *[]Operand, rhs []*Ast) bool {
	optional_ok := false
	tuple_index := 0
	for _, rhs_expr := range rhs {
		c_ := *ctx
		c := &c_
		var o Operand
		var type_hint *Type
		if tuple_index < len(lhs) {
			type_hint = lhs[tuple_index].Type
		}
		check_expr_base(c, &o, rhs_expr, type_hint)
		if o.Mode == AddressingNoValue {
			erroroperand_no_value(&o)
			o.Mode = AddressingInvalid
		}
		if o.Type == nil || o.Type.Kind != TypeTuple {
			if len(lhs) == 2 && len(rhs) == 1 &&
				(o.Mode == AddressingMapIndex || o.Mode == AddressingOptionalOk || o.Mode == AddressingOptionalOkPtr) {
				expr := unparen_expr(o.Expr)
				val0 := o
				val1 := o
				val0.Mode = AddressingValue
				val1.Mode = AddressingValue
				val1.Type = t_untyped_bool
				check_promote_optional_ok(c, &o, nil, &val1.Type)
				if expr.Kind == AstTypeAssertion &&
					(o.Mode == AddressingOptionalOk || o.Mode == AddressingOptionalOkPtr) {
					if is_blank_ident(lhs[0].Expr) {
						expr.TypeAssertion.Ignores[0] = true
					}
					if is_blank_ident(lhs[1].Expr) {
						expr.TypeAssertion.Ignores[1] = true
					}
				}
				*operands = append(*operands, val0)
				*operands = append(*operands, val1)
				optional_ok = true
				tuple_index += 2
			} else if o.Mode == AddressingOptionalOk && is_type_tuple(o.Type) {
				tuple := o.Type
				expr := unparen_expr(o.Expr)
				if expr.Kind == AstCallExpr {
					expr.CallExpr.OptionalOkOne = true
				}
				val := o
				val.Type = tuple.Tuple.Variables[0].Type
				val.Mode = AddressingValue
				*operands = append(*operands, val)
				tuple_index += len(tuple.Tuple.Variables)
			} else {
				*operands = append(*operands, o)
				tuple_index += 1
			}
		} else {
			tuple := &o.Type.Tuple
			for _, e := range tuple.Variables {
				o.Type = e.Type
				*operands = append(*operands, o)
			}
			tuple_index += len(tuple.Variables)
		}
	}
	return optional_ok
}

func check_unpack_arguments(ctx *CheckerContext, lhs []*Entity, lhs_count isize, operands *[]Operand, rhs_arguments []*Ast, flags UnpackFlags, variadic_index int) bool {
	allow_ok := (flags & UnpackFlagAllowOk) != 0
	allow_undef := (flags & UnpackFlagAllowUndef) != 0
	is_variadic := variadic_index > -1
	if !is_variadic {
		variadic_index = int(lhs_count)
	}
	optional_ok := false
	tuple_index := 0
	for _, rhs := range rhs_arguments {
		if rhs.Kind == AstFieldValue {
			error(rhs, "Invalid use of 'field = value'")
			rhs = rhs.FieldValue.Value
		}
		c_ := *ctx
		c := &c_
		var o Operand
		var type_hint *Type
		if lhs != nil {
			if tuple_index < variadic_index {
				e := lhs[tuple_index]
				if e != nil {
					type_hint = e.Type
				}
			} else if is_variadic {
				e := lhs[variadic_index]
				if e != nil {
					type_hint = e.Type.Slice.Elem
				}
			}
		}
		rhs_expr := unparen_expr(rhs)
		if allow_undef && rhs_expr != nil && rhs_expr.Kind == AstUninit {
			o.Type = t_untyped_uninit
			o.Mode = AddressingValue
			o.Expr = rhs
			add_type_and_value(c, rhs, o.Mode, o.Type, o.Value)
		} else {
			check_expr_base(c, &o, rhs, type_hint)
		}
		if o.Mode == AddressingNoValue {
			erroroperand_no_value(&o)
			o.Mode = AddressingInvalid
		}
		if o.Type == nil || o.Type.Kind != TypeTuple {
			if allow_ok && lhs_count == 2 && len(rhs_arguments) == 1 &&
				(o.Mode == AddressingMapIndex || o.Mode == AddressingOptionalOk || o.Mode == AddressingOptionalOkPtr) {
				expr := unparen_expr(o.Expr)
				val0 := o
				val1 := o
				val0.Mode = AddressingValue
				val1.Mode = AddressingValue
				val1.Type = t_untyped_bool
				check_promote_optional_ok(c, &o, nil, &val1.Type)
				if expr.Kind == AstTypeAssertion &&
					(o.Mode == AddressingOptionalOk || o.Mode == AddressingOptionalOkPtr) {
					if is_blank_ident(lhs[0].Token) {
						expr.TypeAssertion.Ignores[0] = true
					}
					if is_blank_ident(lhs[1].Token) {
						expr.TypeAssertion.Ignores[1] = true
					}
				}
				*operands = append(*operands, val0)
				*operands = append(*operands, val1)
				optional_ok = true
				tuple_index += add_dependencies_from_unpacking(c, lhs, lhs_count, tuple_index, 2)
			} else {
				*operands = append(*operands, o)
				tuple_index += 1
			}
		} else {
			tuple := &o.Type.Tuple
			for _, e := range tuple.Variables {
				o.Type = e.Type
				*operands = append(*operands, o)
			}
			count := len(tuple.Variables)
			tuple_index += add_dependencies_from_unpacking(c, lhs, lhs_count, tuple_index, count)
		}
	}
	return optional_ok
}

func add_dependencies_from_unpacking(c *CheckerContext, lhs []*Entity, lhs_count isize, tuple_index isize, tuple_count isize) isize {
	if lhs == nil || c.Decl == nil {
		return tuple_count
	}
	for j := 0; isize(j)+tuple_index < lhs_count && isize(j) < tuple_count; j++ {
		e := lhs[tuple_index+isize(j)]
		if e == nil {
			continue
		}
		decl := decl_info_of_entity(e)
		if decl == nil {
			continue
		}
		rw_mutex_shared_lock(&decl.DepsMutex)
		rw_mutex_lock(&c.Decl.DepsMutex)
		for _, dep := range decl.Deps {
			ptr_set_add(&c.Decl.Deps, dep)
		}
		rw_mutex_unlock(&c.Decl.DepsMutex)
		rw_mutex_shared_unlock(&decl.DepsMutex)
	}
	return tuple_count
}

func check_call_arguments_internal(c *CheckerContext, call *Ast,
	entity *Entity, proc_type *Type,
	positional_operands []Operand, named_operands []Operand,
	show_errormode CallArgumentErrorMode,
	data *CallArgumentData,
	checking_proc_group bool) CallArgumentError {

	err := CallArgumentErrorNone
	ce := &call.CallExpr
	proc_type = base_type(proc_type)
	pt := &proc_type.Proc
	param_count := int32(0)
	param_count_excluding_defaults := get_procedure_param_count_excluding_defaults(proc_type, &param_count)
	variadic := pt.Variadic
	vari_expand := ce.Ellipsis.Pos.Line != 0
	score := int64(0)
	show_error := show_errormode == CallArgumentErrorModeShowErrors
	final_proc_type := proc_type
	var gen_entity *Entity

	if vari_expand && !variadic {
		if show_error {
			error_range(ce.Ellipsis.Pos, ce.Ellipsis.Pos,
				"Cannot use '..' in call to a non-variadic procedure: '%.*s'",
				ce.Proc.Ident.Token.String)
		}
		err = CallArgumentErrorNonVariadicExpand
	} else if vari_expand && pt.CVararg {
		if show_error {
			error_range(ce.Ellipsis.Pos, ce.Ellipsis.Pos,
				"Cannot use '..' in call to a '#c_vararg' variadic procedure: '%.*s'",
				ce.Proc.Ident.Token.String)
		}
		err = CallArgumentErrorNonVariadicExpand
	}

	visited := make([]bool, pt.ParamCount)
	ordered_operands := make([]Operand, pt.ParamCount)

	positional_operand_count := isize(len(positional_operands))
	if variadic {
		positional_operand_count = min(positional_operand_count, isize(pt.VariadicIndex))
	} else if positional_operand_count > isize(pt.ParamCount) {
		err = CallArgumentErrorTooManyArguments
		if show_error {
			proc_str := expr_to_string(ce.Proc)
			error(call, "Too many arguments for '%s', expected %d arguments, got %d", proc_str, param_count_excluding_defaults, len(positional_operands))
			gb_string_free(proc_str)
		}
		return err
	}
	positional_operand_count = min(positional_operand_count, isize(pt.ParamCount))

	for i := isize(0); i < positional_operand_count; i++ {
		ordered_operands[i] = positional_operands[i]
		visited[i] = true
	}

	variadic_operands := positional_operands[positional_operand_count:]
	named_variadic_param := false

	if len(named_operands) != 0 {
		for i := 0; i < len(ce.SplitArgs.Named); i++ {
			arg := ce.SplitArgs.Named[i]
			operand := named_operands[i]
			fv := &arg.FieldValue
			if fv.Field.Kind != AstIdent {
				if show_error {
					expr_str := expr_to_string(fv.Field)
					error(arg, "Invalid parameter name '%s' in procedure call", expr_str)
					gb_string_free(expr_str)
				}
				err = CallArgumentErrorInvalidFieldValue
				continue
			}
			name := fv.Field.Ident.Token.String
			param_index := lookup_procedure_parameter(pt, name)
			if param_index < 0 {
				if show_error {
					error(arg, "No parameter named '%.*s' for this procedure type", name)
				}
				err = CallArgumentErrorParameterNotFound
				continue
			}
			if pt.Variadic && param_index == int(pt.VariadicIndex) {
				named_variadic_param = true
			}
			if visited[param_index] {
				if show_error {
					error(arg, "Duplicate parameter '%.*s' in procedure call", name)
				}
				err = CallArgumentErrorDuplicateParameter
				continue
			}
			visited[param_index] = true
			ordered_operands[param_index] = operand
		}
	}

	dummy_argument_count := isize(0)
	actually_variadic := false
	if variadic {
		if visited[pt.VariadicIndex] &&
			positional_operand_count < isize(len(positional_operands)) {
			if show_error {
				name := pt.Params.Tuple.Variables[pt.VariadicIndex].Token.String
				error(call, "Variadic parameters already handled with a named argument '%.*s' in procedure call", name)
			}
			err = CallArgumentErrorDuplicateParameter
		} else if !visited[pt.VariadicIndex] {
			visited[pt.VariadicIndex] = true
			variadic_operand := &ordered_operands[pt.VariadicIndex]
			if vari_expand {
				if len(variadic_operands) == 0 {
					error(call, "'..' in the wrong position")
				} else {
					*variadic_operand = variadic_operands[0]
					variadic_operand.Type = default_type(variadic_operand.Type)
					actually_variadic = true
				}
			} else {
				f := call.File()
				var o Operand
				o.Mode = AddressingValue
				o.Expr = ast_ident(f, make_token_ident("nil"))
				o.Expr.Ident.Token.Pos = ast_token(call).Pos
				if len(variadic_operands) != 0 {
					actually_variadic = true
					o.Expr.Ident.Token.Pos = ast_token(variadic_operands[0].Expr).Pos
					vt := pt.Params.Tuple.Variables[pt.VariadicIndex]
					o.Type = vt.Type
				} else {
					dummy_argument_count += 1
					o.Type = t_untyped_nil
				}
				*variadic_operand = o
			}
		}
	}

	for i := int32(0); i < pt.ParamCount; i++ {
		if !visited[i] {
			e := pt.Params.Tuple.Variables[i]
			context_allocator_error := false
			if e.Kind == Entity_Variable {
				if e.Variable.ParamValue.Kind != ParameterValue_Invalid {
					if ast_file_vet_explicit_allocators(c.File) && !checking_proc_group {
						if e.Variable.ParamValue.OriginalAstExpr.Kind == AstSelectorExpr {
							expr := e.Variable.ParamValue.OriginalAstExpr.SelectorExpr.Expr
							selector := e.Variable.ParamValue.OriginalAstExpr.SelectorExpr.Selector
							if expr.Kind == AstImplicit &&
								expr.Implicit.String == "context" &&
								selector.Kind == AstIdent &&
								(selector.Ident.Token.String == "allocator" ||
									selector.Ident.Token.String == "temp_allocator") {
								context_allocator_error = true
							}
						}
					}
					if !context_allocator_error {
						ordered_operands[i].Mode = AddressingValue
						ordered_operands[i].Type = e.Type
						ordered_operands[i].Expr = e.Variable.ParamValue.OriginalAstExpr
						dummy_argument_count += 1
						score += assign_score_function(1)
						continue
					}
				}
			}
			if show_error {
				if context_allocator_error {
					str := type_to_string(e.Type)
					error(call, "Parameter '%.*s' of type '%s' must be explicitly provided in procedure call",
						e.Token.String, str)
					gb_string_free(str)
				} else if e.Kind == Entity_TypeName {
					error(call, "Type parameter '%.*s' is missing in procedure call",
						e.Token.String)
				} else if e.Kind == Entity_Constant && e.Constant.Value.Kind != ExactValueInvalid {
				} else {
					str := type_to_string(e.Type)
					error(call, "Parameter '%.*s' of type '%s' is missing in procedure call",
						e.Token.String, str)
					gb_string_free(str)
				}
			}
			err = CallArgumentErrorParameterMissing
		}
	}

	eval_param_and_score := func(o *Operand, param_type *Type, param_is_variadic bool, e *Entity) i64 {
		allow_array_programming := !(e != nil && (e.Flags&EntityFlag_NoBroadcast) != 0)
		s := int64(0)
		if !check_is_assignable_to_with_score(c, o, param_type, &s, param_is_variadic, allow_array_programming) {
			ok := false
			if e != nil && (e.Flags&EntityFlag_AnyInt) != 0 {
				if is_type_integer(param_type) {
					ok = check_is_castable_to(c, o, param_type)
				}
			}
			if !allow_array_programming && check_is_assignable_to_with_score(c, o, param_type, nil, param_is_variadic, !allow_array_programming) {
				if show_error {
					error(o.Expr, "'#no_broadcast' disallows automatic broadcasting a value across all elements of an array-like type in a procedure argument")
				}
			}
			if ok {
				s = assign_score_function(10)
			} else {
				if show_error {
					check_assignment(c, o, param_type, S("procedure argument"))
				}
				err = CallArgumentErrorWrongTypes
			}
		} else if show_error {
			check_assignment(c, o, param_type, S("procedure argument"))
		}
		if e != nil && (e.Flags&EntityFlag_ConstInput) != 0 {
			if o.Mode != AddressingConstant {
				if show_error {
					error(o.Expr, "Expected a constant value for the argument '%.*s'", e.Token.String)
				}
				err = CallArgumentErrorNoneConstantParameter
			}
		}
		if e != nil && e.Kind == Entity_Constant && is_type_proc(e.Type) {
			ok := false
			if o.Mode == AddressingConstant {
				ok = true
			} else if o.Value.Kind == ExactValueProcedure {
				ok = true
			}
			if !ok {
				if show_error {
					error(o.Expr, "Expected a constant procedure value for the argument '%.*s'", e.Token.String)
				}
				err = CallArgumentErrorNoneConstantParameter
			}
		}
		if err == CallArgumentErrorNone && is_type_any(param_type) {
			add_type_info_type(c, o.Type)
		}
		if o.Mode == AddressingType && is_type_typeid(param_type) {
			add_type_info_type(c, o.Type)
			add_type_and_value(c, o.Expr, AddressingValue, param_type, exact_value_typeid(o.Type))
		} else if show_error && is_type_untyped(o.Type) {
			update_untyped_expr_type(c, o.Expr, param_type, true)
		}
		return s
	}

	if len(ordered_operands) == 0 && param_count_excluding_defaults == 0 {
		err = CallArgumentErrorNone
		if variadic {
			t := pt.Params.Tuple.Variables[0].Type
			if is_type_polymorphic(t) {
				if show_error {
					error(call, "Ambiguous call to a polymorphic variadic procedure with no variadic input")
				}
				err = CallArgumentErrorAmbiguousPolymorphicVariadic
			}
		}
	} else {
		if pt.IsPolymorphic && !pt.IsPolySpecialized && err == CallArgumentErrorNone {
			var poly_proc_data PolyProcData
			if find_or_generate_polymorphic_procedure_from_parameters(c, entity, &ordered_operands, call, &poly_proc_data) {
				gen_entity = poly_proc_data.GenEntity
				gept := base_type(gen_entity.Type)
				final_proc_type = gen_entity.Type
				pt = &gept.Proc
			} else {
				err = CallArgumentErrorWrongTypes
			}
		}
		for i := int32(0); i < pt.ParamCount; i++ {
			o := &ordered_operands[i]
			if o.Mode == AddressingInvalid {
				continue
			}
			e := pt.Params.Tuple.Variables[i]
			param_is_variadic := pt.Variadic && pt.VariadicIndex == i
			if e.Kind == Entity_TypeName {
				if o.Mode != AddressingType {
					if show_error {
						error(o.Expr, "Expected a type for the argument '%.*s'", e.Token.String)
					}
					err = CallArgumentErrorWrongTypes
				}
				if are_types_identical(e.Type, o.Type) {
					score += assign_score_function(1)
				} else {
					score += assign_score_function(10)
				}
				continue
			}
			if param_is_variadic {
				if !named_variadic_param {
					continue
				}
			}
			score += eval_param_and_score(o, e.Type, false, e)
		}
	}

	if variadic {
		var_entity := pt.Params.Tuple.Variables[pt.VariadicIndex]
		slice := var_entity.Type
		elem := base_type(slice).Slice.Elem
		t := elem
		if is_type_polymorphic(t) {
			if show_error {
				error(call, "Ambiguous call to a polymorphic variadic procedure with no variadic input %s", type_to_string(final_proc_type))
			}
			err = CallArgumentErrorAmbiguousPolymorphicVariadic
		}
		for operand_index := 0; operand_index < len(variadic_operands); operand_index++ {
			o := &variadic_operands[operand_index]
			if vari_expand {
				t = slice
				if operand_index > 0 {
					if show_error {
						error(o.Expr, "'..' in a variadic procedure can only have one variadic argument at the end")
					}
					if data != nil {
						data.Score = score
						data.ResultType = final_proc_type.Proc.Results
						data.GenEntity = gen_entity
					}
					return CallArgumentErrorMultipleVariadicExpand
				}
			}
			score += eval_param_and_score(o, t, true, var_entity)
		}
		if !vari_expand && len(variadic_operands) != 0 {
			if c.Decl != nil {
				found := false
				for _, vr := range c.Decl.VariadicReuses {
					if are_types_identical(slice, vr.SliceType) {
						vr.MaxCount = max(vr.MaxCount, isize(len(variadic_operands)))
						found = true
						break
					}
				}
				if !found {
					c.Decl.VariadicReuses = append(c.Decl.VariadicReuses, VariadicReuseData{SliceType: slice, MaxCount: isize(len(variadic_operands))})
				}
			}
		}
	}

	if data != nil {
		data.Score = score
		data.ResultType = final_proc_type.Proc.Results
		data.GenEntity = gen_entity
		var proc_lit *Ast
		if ce.Proc.TAV.Value.Kind == ExactValueProcedure {
			vp := unparen_expr(ce.Proc.TAV.Value.ValueProcedure)
			if vp != nil && vp.Kind == AstProcLit {
				proc_lit = vp
			}
		}
		if proc_lit == nil {
			add_type_and_value(c, ce.Proc, AddressingValue, final_proc_type, ExactValue{})
		}
	}
	return err
}

func check_call_arguments_single(c *CheckerContext, call *Ast, operand *Operand,
	e *Entity, proc_type *Type,
	positional_operands []Operand, named_operands []Operand,
	show_errormode CallArgumentErrorMode,
	data *CallArgumentData,
	checking_proc_group bool) bool {

	return_on_failure := show_errormode == CallArgumentErrorModeNoErrors
	ident := operand.Expr
	for ident.Kind == AstSelectorExpr {
		s := ident.SelectorExpr.Selector
		ident = s
	}
	if e == nil {
		e = entity_of_node(ident)
		if e != nil {
			proc_type = e.Type
		}
	}
	proc_type = base_type(proc_type)
	if proc_type == t_invalid {
		return false
	}

	err := check_call_arguments_internal(c, call, e, proc_type, positional_operands, named_operands, show_errormode, data, checking_proc_group)
	if return_on_failure && err != CallArgumentErrorNone {
		return false
	}

	entity_to_use := data.GenEntity
	if entity_to_use == nil {
		entity_to_use = e
	}
	if !return_on_failure && entity_to_use != nil {
		add_entity_use(c, ident, entity_to_use)
		update_untyped_expr_type(c, operand.Expr, entity_to_use.Type, true)
		add_type_and_value(c, operand.Expr, operand.Mode, entity_to_use.Type, operand.Value)
	}
	if data.GenEntity != nil {
		e := data.GenEntity
		decl := data.GenEntity.DeclInfo
		ctx := *c
		ctx.Scope = decl.Scope
		ctx.Decl = decl
		ctx.ProcName = e.Token.String
		ctx.CurrProcDecl = decl
		ctx.CurrProcSig = e.Type
		ok := evaluate_where_clauses(&ctx, call, decl.Scope, decl.ProcLit.ProcLit.WhereClauses, !return_on_failure)
		if return_on_failure {
			if !ok {
				return false
			}
		} else {
			decl.WhereClausesEvaluated = true
			if ok && (data.GenEntity.Flags&EntityFlag_ProcBodyChecked) == 0 {
				check_procedure_later_full(c.Checker, e.File, e.Token, decl, e.Type, decl.ProcLit.ProcLit.Body, decl.ProcLit.ProcLit.Tags)
			}
			if is_type_proc(data.GenEntity.Type) {
				t := base_type(entity_to_use.Type)
				data.ResultType = t.Proc.Results
			}
		}
	}
	return true
}

func check_call_arguments_proc_group(c *CheckerContext, operand *Operand, call *Ast) CallArgumentData {
	ce := &call.CallExpr
	positional_args := ce.SplitArgs.Positional
	named_args := ce.SplitArgs.Named

	var data CallArgumentData
	data.ResultType = t_invalid

	procs := proc_group_entities_cloned(c, *operand)
	if len(procs) > 1 {
		max_arg_count := isize(len(positional_args) + len(named_args))
		for _, arg := range positional_args {
			arg = strip_or_return_expr(arg)
			if arg != nil && arg.Kind == AstCallExpr {
				max_arg_count = isize(int64(0x7fffffffffffffff))
				break
			}
		}
		if max_arg_count != isize(int64(0x7fffffffffffffff)) {
			for _, arg := range named_args {
				if arg.Kind == AstFieldValue {
					arg = strip_or_return_expr(arg.FieldValue.Value)
					if arg != nil && arg.Kind == AstCallExpr {
						max_arg_count = isize(int64(0x7fffffffffffffff))
						break
					}
				}
			}
		}
		for _, arg := range named_args {
			if arg.Kind != AstFieldValue {
				continue
			}
			fv := &arg.FieldValue
			if fv.Field.Kind != AstIdent {
				continue
			}
			key := fv.Field.Ident.Token.String
			for proc_index := len(procs) - 1; proc_index >= 0; proc_index-- {
				t := procs[proc_index].Type
				if is_type_proc(t) {
					param_index := lookup_procedure_parameter(t, key)
					if param_index < 0 {
						procs = append(procs[:proc_index], procs[proc_index+1:]...)
					}
				}
			}
		}
		if len(procs) == 0 {
			procs = proc_group_entities_cloned(c, *operand)
		}
		for proc_index := 0; proc_index < len(procs); {
			proc := procs[proc_index]
			pt := base_type(proc.Type)
			if !(pt != nil && is_type_proc(pt)) {
				proc_index++
				continue
			}
			param_count := int32(0)
			param_count_excluding_defaults := get_procedure_param_count_excluding_defaults(pt, &param_count)
			if param_count_excluding_defaults > max_arg_count {
				procs = append(procs[:proc_index], procs[proc_index+1:]...)
				continue
			}
			if !pt.Proc.Variadic && max_arg_count != isize(int64(0x7fffffffffffffff)) && isize(param_count) < max_arg_count {
				procs = append(procs[:proc_index], procs[proc_index+1:]...)
				continue
			}
			proc_index++
		}
	}

	var lhs []*Entity
	lhs_count := isize(-1)
	variadic_index := int32(-1)

	positional_operands := make([]Operand, 0)
	named_operands := make([]Operand, 0)

	if len(procs) == 1 {
		e := procs[0]
		pt := base_type(e.Type)
		if pt != nil && is_type_proc(pt) {
			lhs, lhs_count = populate_proc_parameter_list(c, pt)
			if pt.Proc.Variadic {
				variadic_index = pt.Proc.VariadicIndex
			}
		}
		check_unpack_arguments(c, lhs, lhs_count, &positional_operands, positional_args, UnpackFlagNone, int(variadic_index))
		if check_named_arguments(c, e.Type, named_args, &named_operands, true) {
			check_call_arguments_single(c, call, operand,
				e, e.Type,
				positional_operands, named_operands,
				CallArgumentErrorModeShowErrors,
				&data, false)
		}
		return data
	}

	{
		proc_arg_count := isize(-1)
		for _, p := range procs {
			pt := base_type(p.Type)
			if pt != nil && is_type_proc(pt) {
				if proc_arg_count < 0 {
					proc_arg_count = isize(pt.Proc.ParamCount)
				} else {
					proc_arg_count = min(proc_arg_count, isize(pt.Proc.ParamCount))
				}
			}
		}
		if proc_arg_count >= 0 {
			lhs_count = proc_arg_count
			if lhs_count > 0 {
				lhs = make([]*Entity, lhs_count)
				for param_index := isize(0); param_index < lhs_count; param_index++ {
					var e *Entity
					for _, p := range procs {
						pt := base_type(p.Type)
						if !(pt != nil && is_type_proc(pt)) {
							continue
						}
						if e == nil {
							e = pt.Proc.Params.Tuple.Variables[param_index]
						} else {
							f := pt.Proc.Params.Tuple.Variables[param_index]
							if e == f {
								continue
							}
							if are_types_identical(e.Type, f.Type) {
								ee := (e.Flags & EntityFlag_Ellipsis) != 0
								fe := (f.Flags & EntityFlag_Ellipsis) != 0
								if ee == fe {
									continue
								}
							}
							e = nil
							break
						}
					}
					lhs[param_index] = e
				}
				for _, p := range procs {
					pt := base_type(p.Type)
					if !(pt != nil && is_type_proc(pt)) {
						continue
					}
					if pt.Proc.IsPolymorphic {
						if variadic_index == -1 {
							variadic_index = pt.Proc.VariadicIndex
						} else if variadic_index != pt.Proc.VariadicIndex {
							variadic_index = -1
							break
						}
					} else {
						variadic_index = -1
						break
					}
				}
			}
		}
	}

	check_unpack_arguments(c, lhs, lhs_count, &positional_operands, positional_args, UnpackFlagNone, int(variadic_index))

	for i := 0; i < len(named_args); i++ {
		arg := named_args[i]
		if arg.Kind != AstFieldValue {
			error(arg, "Expected a 'field = value'")
			return data
		}
		fv := &arg.FieldValue
		if fv.Field.Kind != AstIdent {
			expr_str := expr_to_string(fv.Field)
			error(arg, "Invalid parameter name '%s' in procedure call", expr_str)
			gb_string_free(expr_str)
			return data
		}
		key := fv.Field.Ident.Token.String
		value := fv.Value
		var type_hint *Type
		for lhs_idx := isize(0); lhs_idx < lhs_count; lhs_idx++ {
			e := lhs[lhs_idx]
			if e != nil && e.Token.String == key &&
				!is_type_polymorphic(e.Type) {
				type_hint = e.Type
				break
			}
		}
		var o Operand
		check_expr_with_type_hint(c, &o, value, type_hint)
		named_operands = append(named_operands, o)
	}

	valids := make([]ValidIndexAndScore, 0)
	proc_entities := make([]*Entity, 0)
	for _, proc := range procs {
		proc_entities = append(proc_entities, proc)
	}
	max_matched_features := 0
	expr_name := expr_to_string(operand.Expr)
	c.InProcGroup = true
	for i := 0; i < len(procs); i++ {
		p := procs[i]
		if p.Flags&EntityFlag_Disabled != 0 {
			continue
		}
		pt := base_type(p.Type)
		if pt != nil && is_type_proc(pt) {
			var d CallArgumentData
			ctx := *c
			ctx.NoPolymorphicErrors = true
			ctx.AllowPolymorphicTypes = is_type_polymorphic(pt)
			ctx.HidePolymorphicErrors = true
			is_a_candidate := check_call_arguments_single(&ctx, call, operand,
				p, pt,
				positional_operands, named_operands,
				CallArgumentErrorModeNoErrors,
				&d, true)
			if !is_a_candidate {
				continue
			}
			index := i
			var item ValidIndexAndScore
			item.Score = d.Score
			if d.GenEntity != nil {
				proc_entities = append(proc_entities, d.GenEntity)
				index = len(proc_entities) - 1
				item.Score += assign_score_function(1)
			}
			max_matched_features = max(max_matched_features, matched_target_features(&pt.Proc))
			item.Index = index
			valids = append(valids, item)
		}
	}
	c.InProcGroup = false

	if max_matched_features > 0 {
		for i := 0; i < len(valids); i++ {
			p := procs[valids[i].Index]
			t := base_type(p.Type)
			matched := matched_target_features(&t.Proc)
			valids[i].Score += assign_score_function(int64(max_matched_features - matched))
		}
	}

	if len(valids) > 1 {
		sort_valids(valids)
		best_score := valids[0].Score
		best_entity := proc_entities[valids[0].Index]
		for i := 1; i < len(valids); i++ {
			if best_score > valids[i].Score {
				valids = valids[:i]
				break
			}
			if best_entity == proc_entities[valids[i].Index] {
				valids = valids[:i]
				break
			}
		}
	}

	print_argument_types := func() {
		errorline("\tGiven argument types:\n")
		i := 0
		for _, o := range positional_operands {
			type_ := type_to_string(o.Type)
			errorline("\t \x95 %s\n", type_)
			gb_string_free(type_)
		}
		for _, o := range named_operands {
			type_ := type_to_string(o.Type)
			if i < len(ce.SplitArgs.Named) {
				named_field := ce.SplitArgs.Named[i]
				fv := &named_field.FieldValue
				field := expr_to_string(fv.Field)
				errorline("\t \x95 %s = %s\n", field, type_)
				gb_string_free(field)
			} else {
				errorline("\t \x95 %s\n", type_)
			}
			gb_string_free(type_)
		}
	}

	if len(valids) == 0 {
		begin_errorblock()
		error(operand.Expr, "No procedures or ambiguous call for procedure group '%s' that match with the given arguments", expr_name)
		if len(positional_operands) == 0 && len(named_operands) == 0 {
			errorline("\tNo given arguments\n")
		} else {
			print_argument_types()
		}
		if len(procs) == 0 {
			procs = proc_group_entities_cloned(c, *operand)
		}
		possibly_ignore := make([]bool, len(procs))
		possibly_ignore_set := 0
		for i := 0; i < len(procs); i++ {
			proc := procs[i]
			t := base_type(proc.Type)
			if t == nil || t.Kind != TypeProc {
				continue
			}
			pt := &t.Proc
			if pt.ParamCount == 0 {
				continue
			}
			for j := int32(0); j < int32(len(pt.Params.Tuple.Variables)); j++ {
				v := pt.Params.Tuple.Variables[j]
				if v.Kind != Entity_TypeName {
					continue
				}
				dst_t := base_type(v.Type)
				for dst_t.Kind == TypeGeneric && dst_t.Generic.Specialized != nil {
					dst_t = dst_t.Generic.Specialized
				}
				if int(j) >= len(positional_operands) {
					continue
				}
				o := positional_operands[j]
				if o.Mode != AddressingType {
					continue
				}
				t := base_type(o.Type)
				if t.Kind == dst_t.Kind {
					continue
				}
				st := base_type(type_deref(o.Type))
				dt := base_type(type_deref(dst_t))
				if st.Kind == dt.Kind {
					continue
				}
				if is_type_soa_struct(st) {
					possibly_ignore[i] = true
					possibly_ignore_set += 1
					continue
				}
			}
		}
		if possibly_ignore_set == len(procs) {
			possibly_ignore_set = 0
		}
		max_name_length := isize(0)
		max_type_length := isize(0)
		for i := 0; i < len(procs); i++ {
			if possibly_ignore_set != 0 && possibly_ignore[i] {
				continue
			}
			proc := procs[i]
			t := base_type(proc.Type)
			if t == t_invalid {
				continue
			}
			var prefix, prefix_sep string
			if proc.Pkg != nil {
				prefix = proc.Pkg.Name
				prefix_sep = "."
			}
			name := proc.Token.String
			l := isize(len(prefix) + len(prefix_sep) + len(name))
			max_name_length = max(max_name_length, l)
			var pt string
			if t.Proc.Node != nil {
				pt = expr_to_string(t.Proc.Node)
			} else {
				pt = type_to_string(t)
			}
			max_type_length = max(max_type_length, isize(len(pt)))
		}
		max_spaces := max(max_name_length, max_type_length)
		spaces := make([]byte, max_spaces)
		for i := range spaces {
			spaces[i] = ' '
		}

		{
			try_addr := false
			try_addr_idx := isize(-1)
			for i := 0; i < len(procs); i++ {
				if possibly_ignore_set != 0 && possibly_ignore[i] {
					continue
				}
				proc := procs[i]
				pos := proc.Token.Pos
				t := base_type(proc.Type)
				if t == t_invalid {
					continue
				}
				if t.Proc.Params != nil && len(t.Proc.Params.Tuple.Variables) > 0 {
					n := min(isize(len(t.Proc.Params.Tuple.Variables)), isize(len(positional_operands)))
					for i := isize(0); i < n; i++ {
						dst := t.Proc.Params.Tuple.Variables[i].Type
						src := positional_operands[i]
						if check_is_assignable_to(c, &src, dst) {
						} else if check_is_assignable_to(c, &src, type_deref(dst)) {
							try_addr = true
							if try_addr_idx < 0 {
								try_addr_idx = i
							}
						}
					}
				}
			}
			if try_addr {
				errorline("  \n")
				errorline("\tSuggestion:\n")
				errorline("\t\t%s(", expr_name)
				i := 0
				for _, o := range positional_operands {
					if i > 0 {
						errorline(", ")
					}
					expr := expr_to_string(o.Expr)
					if i == int(try_addr_idx) {
						errorline("&")
					}
					errorline("%s", expr)
					gb_string_free(expr)
					i++
				}
				for _, o := range named_operands {
					if i > 0 {
						errorline(", ")
					}
					expr := expr_to_string(o.Expr)
					if i < len(ce.SplitArgs.Named) {
						named_field := ce.SplitArgs.Named[i]
						fv := &named_field.FieldValue
						field := expr_to_string(fv.Field)
						errorline("%s = %s", field, expr)
						gb_string_free(field)
					} else {
						errorline("%s", expr)
					}
					gb_string_free(expr)
					i++
				}
				errorline(")\n")
				errorline("  \n")
			}
		}

		if len(procs) > 0 {
			errorline("Did you mean one of the following overloads?\n")
		}
		for i := 0; i < len(procs); i++ {
			if possibly_ignore_set != 0 && possibly_ignore[i] {
				continue
			}
			proc := procs[i]
			pos := proc.Token.Pos
			t := base_type(proc.Type)
			if t == t_invalid {
				continue
			}
			var pt string
			if t.Proc.Node != nil {
				pt = expr_to_string(t.Proc.Node)
			} else {
				pt = type_to_string(t)
			}
			var prefix, prefix_sep string
			if proc.Pkg != nil {
				prefix = proc.Pkg.Name
				prefix_sep = "."
			}
			name := proc.Token.String
			len_ := isize(len(prefix) + len(prefix_sep) + len(name))
			name_padding := int(max_name_length - len_)
			if name_padding < 0 {
				name_padding = 0
			}
			type_padding := int(max_type_length - isize(len(pt)))
			if type_padding < 0 {
				type_padding = 0
			}
			sep := "::"
			if proc.Kind == Entity_Variable {
				sep = ":="
			}
			errorline("\t%s%s%s %s%s %s %s%sat %s\n",
				prefix, prefix_sep, name,
				string(spaces[:name_padding]),
				sep,
				pt,
				string(spaces[:type_padding]),
				token_pos_to_string(pos),
			)
		}
		if len(procs) > 0 {
			errorline("\n")
		}
		data.ResultType = t_invalid
	} else if len(valids) > 1 {
		begin_errorblock()
		error(operand.Expr, "Ambiguous procedure group call '%s' that match with the given arguments", expr_name)
		if len(positional_operands) == 0 && len(named_operands) == 0 {
			errorline("\tNo given arguments\n")
		} else {
			print_argument_types()
		}
		for _, valid := range valids {
			proc := proc_entities[valid.Index]
			pos := proc.Token.Pos
			t := base_type(proc.Type)
			var pt string
			if t.Proc.Node != nil {
				pt = expr_to_string(t.Proc.Node)
			} else {
				pt = type_to_string(t)
			}
			name := proc.Token.String
			sep := "::"
			if proc.Kind == Entity_Variable {
				sep = ":="
			}
			errorline("\t%.*s %s %s ", name, sep, pt)
			if proc.DeclInfo != nil && proc.DeclInfo.ProcLit != nil {
				pl := &proc.DeclInfo.ProcLit.ProcLit
				if pl.WhereToken.Kind != Token_Invalid {
					errorline("\n\t\twhere ")
					for j, clause := range pl.WhereClauses {
						if j != 0 {
							errorline("\t\t      ")
						}
						str := expr_to_string(clause)
						errorline("%s", str)
						gb_string_free(str)
						if j != len(pl.WhereClauses)-1 {
							errorline(",")
						}
					}
					errorline("\n\t")
				}
			}
			errorline("at %s\n", token_pos_to_string(pos))
		}
		data.ResultType = t_invalid
	} else {
		e := proc_entities[valids[0].Index]
		check_call_arguments_single(c, call, operand,
			e, e.Type,
			positional_operands, named_operands,
			CallArgumentErrorModeShowErrors,
			&data, false)
		return data
	}
	return data
}

func check_call_arguments(c *CheckerContext, operand *Operand, call *Ast) CallArgumentData {
	var proc_type *Type
	var data CallArgumentData
	data.ResultType = t_invalid
	proc_type = base_type(operand.Type)
	var pt *TypeProc
	if proc_type != nil {
		pt = &proc_type.Proc
	}

	ce := &call.CallExpr
	any_failure := false
	var positional_args, named_args []*Ast

	if ce.SplitArgs == nil {
		positional_args = ce.Args
		for i, arg := range ce.Args {
			if arg.Kind == AstFieldValue {
				positional_args = ce.Args[:i]
				break
			}
		}
		named_args = ce.Args[len(positional_args):]
		split_args := &AstSplitArgs{}
		split_args.Positional = positional_args
		split_args.Named = named_args
		ce.SplitArgs = split_args
	} else {
		positional_args = ce.SplitArgs.Positional
		named_args = ce.SplitArgs.Named
	}

	if operand.Mode == AddressingProcGroup {
		return check_call_arguments_proc_group(c, operand, call)
	}

	positional_operands := make([]Operand, 0)
	named_operands := make([]Operand, 0)

	if len(positional_args) > 0 {
		var lhs []*Entity
		lhs_count := isize(-1)
		variadic_index := int32(-1)
		if pt != nil {
			lhs, lhs_count = populate_proc_parameter_list(c, proc_type)
			if pt.Variadic {
				variadic_index = pt.VariadicIndex
			}
		}
		check_unpack_arguments(c, lhs, lhs_count, &positional_operands, positional_args, UnpackFlagNone, int(variadic_index))
	}

	if len(named_args) > 0 {
		for _, arg := range named_args {
			if arg.Kind != AstFieldValue {
				error(arg, "Expected a 'field = value'")
				return data
			}
			fv := &arg.FieldValue
			if fv.Field.Kind != AstIdent {
				expr_str := expr_to_string(fv.Field)
				error(arg, "Invalid parameter name '%s' in procedure call", expr_str)
				any_failure = true
				gb_string_free(expr_str)
				continue
			}
			key := fv.Field.Ident.Token.String
			value := fv.Value
			param_index := lookup_procedure_parameter(pt, key)
			var type_hint *Type
			if param_index >= 0 {
				e := pt.Params.Tuple.Variables[param_index]
				type_hint = e.Type
			}
			var o Operand
			check_expr_with_type_hint(c, &o, value, type_hint)
			if o.Mode == AddressingInvalid {
				any_failure = true
			}
			named_operands = append(named_operands, o)
		}
	}

	if !any_failure {
		check_call_arguments_single(c, call, operand,
			nil, proc_type,
			positional_operands, named_operands,
			CallArgumentErrorModeShowErrors,
			&data, false)
	} else if pt != nil {
		data.ResultType = pt.Results
	}
	return data
}

func lookup_polymorphic_record_parameter(t *Type, parameter_name string) isize {
	if !is_type_polymorphic_record(t) {
		return -1
	}
	params := get_record_polymorphic_params(t)
	if params == nil {
		return -1
	}
	for i, e := range params.Variables {
		name := e.Token.String
		if is_blank_ident(name) {
			continue
		}
		if name == parameter_name {
			return isize(i)
		}
	}
	return -1
}

func check_polymorphic_record_type(c *CheckerContext, operand *Operand, call *Ast) CallArgumentError {
	ce := &call.CallExpr
	original_type := operand.Type
	show_error := true
	var operands []Operand
	err := CallArgumentErrorNone
	named_fields := false

	{
		prev_type_path := c.TypePath
		c.TypePath = new_checker_type_path()

		if is_call_expr_field_value(ce) {
			named_fields = true
			operands = make([]Operand, len(ce.Args))
			for i, arg := range ce.Args {
				fv := &arg.FieldValue
				if fv.Value == nil {
					error_range(fv.Eq.Pos, fv.Eq.Pos, "Expected a value")
					err = CallArgumentErrorInvalidFieldValue
					continue
				}
				if fv.Field.Kind == AstIdent {
					name := fv.Field.Ident.Token.String
					index := lookup_polymorphic_record_parameter(original_type, name)
					if index >= 0 {
						params := get_record_polymorphic_params(original_type)
						e := params.Variables[index]
						if e.Kind == Entity_Constant {
							check_expr_with_type_hint(c, &operands[i], fv.Value, e.Type)
							continue
						}
					}
				}
				check_expr_or_type(c, &operands[i], fv.Value)
			}
			vari_expand := ce.Ellipsis.Pos.Line != 0
			if vari_expand {
				error_range(ce.Ellipsis.Pos, ce.Ellipsis.Pos, "Invalid use of '..' in a polymorphic type call'")
			}
		} else {
			operands = make([]Operand, 0, 2*len(ce.Args))
			var lhs []*Entity
			lhs_count := isize(-1)
			params := get_record_polymorphic_params(original_type)
			if params != nil {
				lhs = params.Variables
				lhs_count = isize(len(params.Variables))
			}
			check_unpack_arguments(c, lhs, lhs_count, &operands, ce.Args, UnpackFlagNone, -1)
		}

		destroy_checker_type_path(c.TypePath)
		c.TypePath = prev_type_path
	}

	if err != CallArgumentErrorNone {
		operand.Mode = AddressingInvalid
		return err
	}

	tuple := get_record_polymorphic_params(original_type)
	param_count := isize(len(tuple.Variables))
	minimum_param_count := param_count
	for ; minimum_param_count > 0; minimum_param_count-- {
		e := tuple.Variables[minimum_param_count-1]
		if e.Kind != Entity_Constant {
			break
		}
		if e.Constant.ParamValue.Kind == ParameterValue_Invalid {
			break
		}
	}

	ordered_operands := operands
	if !named_fields {
		ordered_operands = make([]Operand, len(operands))
		copy(ordered_operands, operands)
	} else {
		visited := make([]bool, param_count)
		ordered_operands = make([]Operand, param_count)
		for i, arg := range ce.Args {
			fv := &arg.FieldValue
			if fv.Field.Kind != AstIdent {
				if show_error {
					expr_str := expr_to_string(fv.Field)
					error(arg, "Invalid parameter name '%s' in polymorphic type call", expr_str)
					gb_string_free(expr_str)
				}
				err = CallArgumentErrorInvalidFieldValue
				continue
			}
			name := fv.Field.Ident.Token.String
			index := lookup_polymorphic_record_parameter(original_type, name)
			if index < 0 {
				if show_error {
					error(arg, "No parameter named '%.*s' for this polymorphic type", name)
				}
				err = CallArgumentErrorParameterNotFound
				continue
			}
			if visited[index] {
				if show_error {
					error(arg, "Duplicate parameter '%.*s' in polymorphic type", name)
				}
				err = CallArgumentErrorDuplicateParameter
				continue
			}
			visited[index] = true
			ordered_operands[index] = operands[i]
		}
		for i := isize(0); i < param_count; i++ {
			if !visited[i] {
				e := tuple.Variables[i]
				if is_blank_ident(e.Token) {
					continue
				}
				if show_error {
					if e.Kind == Entity_TypeName {
						error(call, "Type parameter '%.*s' is missing in polymorphic type call",
							e.Token.String)
					} else {
						str := type_to_string(e.Type)
						error(call, "Parameter '%.*s' of type '%s' is missing in polymorphic type call",
							e.Token.String, str)
						gb_string_free(str)
					}
				}
				err = CallArgumentErrorParameterMissing
			}
		}
	}

	if err != CallArgumentErrorNone {
		operand.Mode = AddressingInvalid
		return err
	}

	for len(ordered_operands) > 0 {
		if ordered_operands[len(ordered_operands)-1].Expr != nil {
			break
		}
		ordered_operands = ordered_operands[:len(ordered_operands)-1]
	}

	if minimum_param_count != param_count {
		if param_count < isize(len(ordered_operands)) {
			error(call, "Too many polymorphic type arguments, expected a maximum of %d, got %d", param_count, len(ordered_operands))
			err = CallArgumentErrorTooManyArguments
		} else if minimum_param_count > isize(len(ordered_operands)) {
			error(call, "Too few polymorphic type arguments, expected a minimum of %d, got %d", minimum_param_count, len(ordered_operands))
			err = CallArgumentErrorTooFewArguments
		}
	} else {
		if param_count < isize(len(ordered_operands)) {
			error(call, "Too many polymorphic type arguments, expected %d, got %d", param_count, len(ordered_operands))
			err = CallArgumentErrorTooManyArguments
		} else if param_count > isize(len(ordered_operands)) {
			error(call, "Too few polymorphic type arguments, expected %d, got %d", param_count, len(ordered_operands))
			err = CallArgumentErrorTooFewArguments
		}
	}
	if err != CallArgumentErrorNone {
		return err
	}

	if minimum_param_count != param_count {
		new_oo := make([]Operand, param_count)
		copy(new_oo, ordered_operands)
		ordered_operands = new_oo
		missing_count := int(0)
		for i, o := range ordered_operands {
			if o.Expr == nil {
				e := tuple.Variables[i]
				if e.Kind == Entity_Constant {
					missing_count += 1
					ordered_operands[i].Mode = AddressingConstant
					ordered_operands[i].Type = default_type(e.Type)
					ordered_operands[i].Expr = unparen_expr(e.Constant.ParamValue.OriginalAstExpr)
					if e.Constant.ParamValue.Kind == ParameterValue_Constant {
						ordered_operands[i].Value = e.Constant.ParamValue.Value
					}
				} else if e.Kind == Entity_TypeName {
					missing_count += 1
					ordered_operands[i].Mode = AddressingType
					ordered_operands[i].Type = e.Type
					ordered_operands[i].Expr = e.Identifier
				}
			}
		}
		_ = missing_count
	}

	oo_count := min(param_count, isize(len(ordered_operands)))
	score := int64(0)
	for i := isize(0); i < oo_count; i++ {
		e := tuple.Variables[i]
		o := &ordered_operands[i]
		if o.Mode == AddressingInvalid {
			continue
		}
		if e.Kind == Entity_TypeName {
			if o.Mode != AddressingType {
				if show_error {
					expr := expr_to_string(o.Expr)
					error(o.Expr, "Expected a type for the argument '%.*s', got %s", e.Token.String, expr)
					gb_string_free(expr)
				}
				err = CallArgumentErrorWrongTypes
			}
			if are_types_identical(e.Type, o.Type) {
				score += assign_score_function(1)
			} else {
				score += assign_score_function(10)
			}
		} else {
			s := int64(0)
			if o.Type.Kind == TypeGeneric {
				score += assign_score_function(1)
				continue
			} else if !check_is_assignable_to_with_score(c, o, e.Type, &s) {
				if show_error {
					check_assignment(c, o, e.Type, S("polymorphic type argument"))
				}
				err = CallArgumentErrorWrongTypes
			}
			o.Type = e.Type
			if o.Mode != AddressingConstant {
				valid := false
				if is_type_proc(o.Type) {
					proc_entity := entity_from_expr(o.Expr)
					valid = proc_entity != nil
				}
				if !valid {
					if show_error {
						error(o.Expr, "Expected a constant value for this polymorphic type argument")
					}
					err = CallArgumentErrorNoneConstantParameter
				}
			}
			score += s
		}
	}
	if show_error && err != CallArgumentErrorNone {
		return err
	}

	{
		found_gen_types := ensure_polymorphic_record_entity_has_gen_types(c, original_type)
		mutex_lock(&found_gen_types.Mutex)
		found_entity := find_polymorphic_record_entity(found_gen_types, param_count, ordered_operands)
		if found_entity != nil {
			mutex_unlock(&found_gen_types.Mutex)
			operand.Mode = AddressingType
			operand.Type = found_entity.Type
			return err
		}
		ctx := *c
		ctx.Scope = polymorphic_record_parent_scope(original_type)
		bt := base_type(original_type)
		generated_name := expr_to_string(call)
		named_type := alloc_type_named(generated_name, nil, nil)

		if bt.Kind == TypeStruct {
			node := clone_ast(bt.Struct.Node)
			struct_type := alloc_type_struct()
			struct_type.Struct.Node = node
			struct_type.Struct.PolymorphicParent = original_type
			set_base_type(named_type, struct_type)
			check_open_scope(&ctx, node)
			check_struct_type(&ctx, struct_type, node, &ordered_operands, named_type, original_type)
			check_close_scope(&ctx)
		} else if bt.Kind == TypeUnion {
			node := clone_ast(bt.Union.Node)
			union_type := alloc_type_union()
			union_type.Union.Node = node
			union_type.Union.PolymorphicParent = original_type
			set_base_type(named_type, union_type)
			check_open_scope(&ctx, node)
			check_union_type(&ctx, union_type, node, &ordered_operands, named_type, original_type)
			check_close_scope(&ctx)
		} else {
			panic("Unsupported parametric polymorphic record type")
		}
		mutex_unlock(&found_gen_types.Mutex)

		bt = base_type(named_type)
		if bt.Kind == TypeStruct || bt.Kind == TypeUnion {
			e := original_type.Named.TypeName
			s := e.Token.String + "("
			tuple := get_record_polymorphic_params(bt)
			if tuple != nil {
				for i, v := range tuple.Variables {
					name := v.Token.String
					if i > 0 {
						s += ", "
					}
					s += "$" + name
					if v.Kind == Entity_TypeName {
						if v.Type != nil && v.Type.Kind != TypeGeneric {
							s += "="
							s += type_to_string(v.Type)
						}
					} else if v.Kind == Entity_Constant {
						if v.Constant.Value.Kind != ExactValueInvalid {
							s += "="
							s += exact_value_to_string(v.Constant.Value)
						}
					}
				}
			}
			s += ")"
			named_type.Named.Name = s
			if named_type.Named.TypeName != nil {
				named_type.Named.TypeName.Token.String = s
			}
		}
		operand.Mode = AddressingType
		operand.Type = named_type
	}
	return err
}

func check_call_parameter_mixture(args []*Ast, context string, allow_mixed ...bool) bool {
	allowMixed := false
	if len(allow_mixed) > 0 {
		allowMixed = allow_mixed[0]
	}
	success := true
	if len(args) > 0 {
		if allowMixed {
			was_named := false
			for _, arg := range args {
				if was_named && arg.Kind != AstFieldValue {
					error(arg, "Non-named parameter is not allowed to follow named parameter i.e. 'field = value' in a %s", context)
					success = false
					break
				}
				was_named = was_named || arg.Kind == AstFieldValue
			}
		} else {
			first_is_field_value := args[0].Kind == AstFieldValue
			for _, arg := range args {
				mix := false
				if first_is_field_value {
					mix = arg.Kind != AstFieldValue
				} else {
					mix = arg.Kind == AstFieldValue
				}
				if mix {
					error(arg, "Mixture of 'field = value' and value elements in a %s is not allowed", context)
					success = false
				}
			}
		}
	}
	return success
}

func check_call_expr_as_type_cast(c *CheckerContext, operand *Operand, call *Ast, args []*Ast, type_hint *Type) ExprKind {
	t := operand.Type
	if is_type_polymorphic_record(t) {
		if !check_call_parameter_mixture(args, "polymorphic type construction") {
			operand.Mode = AddressingInvalid
			operand.Expr = call
			return ExprStmt
		}
		if !is_type_named(t) {
			s := expr_to_string(operand.Expr)
			error(call, "Illegal use of an unnamed polymorphic record, %s", s)
			gb_string_free(s)
			operand.Mode = AddressingInvalid
			operand.Type = t_invalid
			return ExprExpr
		}
		err := check_polymorphic_record_type(c, operand, call)
		if err == CallArgumentErrorNone {
			ident := operand.Expr
			for ident.Kind == AstSelectorExpr {
				s := ident.SelectorExpr.Selector
				ident = s
			}
			ot := operand.Type
			e := ot.Named.TypeName
			add_entity_use(c, ident, e)
			add_type_and_value(c, call, AddressingType, ot, ExactValue{})
		} else {
			operand.Mode = AddressingInvalid
			operand.Type = t_invalid
		}
	} else {
		if !check_call_parameter_mixture(args, "type conversion") {
			operand.Mode = AddressingInvalid
			operand.Expr = call
			return ExprStmt
		}
		operand.Mode = AddressingInvalid
		arg_count := len(args)
		switch arg_count {
		case 0:
			str := type_to_string(t)
			error(call, "Missing argument in conversion to '%s'", str)
			gb_string_free(str)
		default:
			str := type_to_string(t)
			if t.Kind == TypeBasic {
				begin_errorblock()
				switch t.Basic.Kind {
				case Basic_complex32, Basic_complex64, Basic_complex128:
					error(call, "Too many arguments in conversion to '%s'", str)
					errorline("\tSuggestion: %s(1+2i) or construct with 'complex'\n", str)
				case Basic_quaternion64, Basic_quaternion128, Basic_quaternion256:
					error(call, "Too many arguments in conversion to '%s'", str)
					errorline("\tSuggestion: %s(1+2i+3j+4k) or construct with 'quaternion'\n", str)
				default:
					error(call, "Too many arguments in conversion to '%s'", str)
				}
				end_errorblock()
			} else {
				error(call, "Too many arguments in conversion to '%s'", str)
			}
			gb_string_free(str)
		case 1: {
			arg := args[0]
			if arg.Kind == AstFieldValue {
				error(call, "'field = value' cannot be used in a type conversion")
				arg = arg.FieldValue.Value
			}
			check_expr_with_type_hint(c, operand, arg, t)
			if operand.Mode != AddressingInvalid {
				if is_type_polymorphic(t) {
					error(call, "A polymorphic type cannot be used in a type conversion")
				} else {
					check_cast(c, operand, t)
				}
			}
			operand.Type = t
			operand.Expr = call
			if operand.Mode != AddressingInvalid {
				update_untyped_expr_type(c, arg, t, false)
				check_representable_as_constant(c, operand.Value, t, &operand.Value)
			}
		}
		}
	}
	return ExprExpr
}

func check_call_expr(c *CheckerContext, operand *Operand, call *Ast, proc *Ast, args []*Ast, inlining ProcInlining, tailing ProcTailing, type_hint *Type) ExprKind {
	if proc != nil &&
		proc.Kind == AstBasicDirective {
		bd := &proc.BasicDirective
		name := bd.Name.String
		if name == "location" ||
			name == "exists" ||
			name == "assert" ||
			name == "panic" ||
			name == "defined" ||
			name == "config" ||
			name == "load" ||
			name == "load_directory" ||
			name == "load_hash" ||
			name == "hash" ||
			name == "caller_expression" {
			operand.Mode = AddressingBuiltin
			operand.BuiltinID = BuiltinProc_DIRECTIVE
			operand.Expr = proc
			operand.Type = t_invalid
			add_type_and_value(c, proc, operand.Mode, operand.Type, operand.Value)
		} else {
			error(proc, "Unknown directive: #%.*s", name)
			operand.Expr = proc
			operand.Type = t_invalid
			operand.Mode = AddressingInvalid
			return ExprExpr
		}
		if inlining != ProcInliningNone {
			error(call, "Inlining directives are not allowed on built-in procedures")
		}
		if tailing != ProcTailingNone {
			error(call, "Tailing directives are not allowed on built-in procedures")
		}
	} else {
		if proc != nil {
			check_expr_or_type(c, operand, proc)
		}
	}

	if operand.Mode == AddressingInvalid {
		if !check_call_parameter_mixture(args, "procedure call") {
			operand.Mode = AddressingInvalid
			operand.Expr = call
			return ExprStmt
		}
		for _, arg := range args {
			if arg.Kind == AstFieldValue {
				arg = arg.FieldValue.Value
			}
			check_expr_base(c, operand, arg, nil)
		}
		operand.Mode = AddressingInvalid
		operand.Expr = call
		return ExprStmt
	}

	if operand.Mode == AddressingType {
		return check_call_expr_as_type_cast(c, operand, call, args, type_hint)
	}

	if operand.Mode == AddressingBuiltin {
		if !check_call_parameter_mixture(args, "builtin call") {
			operand.Mode = AddressingInvalid
			operand.Expr = call
			return ExprStmt
		}
		id := operand.BuiltinID
		e := entity_of_node(operand.Expr)
		if e != nil && e.Token.String == "expand_to_tuple" {
			error(operand.Expr, "'expand_to_tuple' has been replaced with 'expand_values'")
		}
		if !check_builtin_procedure(c, operand, call, id, type_hint) {
			operand.Mode = AddressingInvalid
			operand.Type = t_invalid
		}
		operand.Expr = call
		return builtin_procs[id].Kind
	}

	{
		context := "procedure call"
		if operand.Mode == AddressingProcGroup {
			context = "procedure group call"
		}
		if !check_call_parameter_mixture(args, context, true) {
			operand.Mode = AddressingInvalid
			operand.Expr = call
			return ExprStmt
		}
	}

	initial_entity := entity_of_node(operand.Expr)
	if initial_entity != nil && initial_entity.Kind == Entity_Procedure {
		if initial_entity.Procedure.DeferredProcedure.Entity != nil {
			call.ViralStateFlags |= ViralStateFlag_ContainsDeferredProcedure
			if c.Decl != nil {
				c.Decl.DeferUsed += 1
			}
		}
		add_entity_use(c, operand.Expr, initial_entity)
		if initial_entity.Procedure.EntryPointOnly {
			if c.CurrProcDecl != nil && c.CurrProcDecl.Entity == c.Info.EntryPoint {
			} else {
				error(operand.Expr, "Procedures with the attribute '@(entry_point_only)' can only be called directly from the user-level entry point procedure")
			}
		}
	}

	if operand.Mode != AddressingProcGroup {
		proc_type := base_type(operand.Type)
		valid_type := (proc_type != nil) && is_type_proc(proc_type)
		valid_mode := is_operand_value(*operand)
		if !valid_type || !valid_mode {
			e := operand.Expr
			str := expr_to_string(e)
			type_str := type_to_string(operand.Type)
			error(e, "Cannot call a non-procedure: '%s' of type '%s'", str, type_str)
			gb_string_free(type_str)
			gb_string_free(str)
			operand.Mode = AddressingInvalid
			operand.Expr = call
			return ExprStmt
		}
	}

	data := check_call_arguments(c, operand, call)
	result_type := data.ResultType

	*operand = Operand{}
	operand.Expr = call

	if result_type == t_invalid {
		operand.Mode = AddressingInvalid
		operand.Type = t_invalid
		return ExprStmt
	}

	pt := base_type(operand.Type)
	if pt == nil {
		pt = t_invalid
	}
	if pt == t_invalid {
		if operand.Expr != nil && operand.Expr.Kind == AstCallExpr {
			pt = type_of_expr(operand.Expr.CallExpr.Proc)
		}
		if pt == t_invalid && data.GenEntity != nil {
			pt = data.GenEntity.Type
		}
	}
	pt = base_type(pt)

	if pt.Kind == TypeProc && pt.Proc.CallingConvention == ProcCCOdin {
		if (c.Scope.Flags & ScopeFlag_ContextDefined) == 0 {
			begin_errorblock()
			if c.Scope.Flags&ScopeFlag_File != 0 {
				error(call, "Procedures requiring a 'context' cannot be called at the global scope")
			} else {
				error(call, "'context' has not been defined within this scope, but is required for this procedure call")
				errorline("\tSuggestion: 'context = runtime.default_context()'")
			}
			end_errorblock()
		}
	}

	if result_type == nil {
		operand.Mode = AddressingNoValue
	} else {
		count := len(result_type.Tuple.Variables)
		switch count {
		case 0:
			operand.Mode = AddressingNoValue
		case 1:
			operand.Mode = AddressingValue
			operand.Type = result_type.Tuple.Variables[0].Type
		default:
			operand.Mode = AddressingValue
			operand.Type = result_type
		}
	}

	is_call_inlined := false
	switch inlining {
	case ProcInliningInline:
		is_call_inlined = true
		if proc != nil {
			e := entity_from_expr(proc)
			if e != nil && e.Kind == Entity_Procedure {
				decl := e.DeclInfo
				if decl.ProcLit != nil {
					pl := &decl.ProcLit.ProcLit
					if pl.Inlining == ProcInliningNoInline {
						error(call, "'#force_inline' cannot be applied to a procedure that has been marked as '#force_no_inline'")
					}
				}
			}
		}
	case ProcInliningNoInline:
	case ProcInliningNone:
		if proc != nil {
			e := entity_from_expr(proc)
			if e != nil && e.Kind == Entity_Procedure {
				decl := e.DeclInfo
				if decl.ProcLit != nil {
					pl := &decl.ProcLit.ProcLit
					if pl.Inlining == ProcInliningInline {
						is_call_inlined = true
					}
				}
			}
		}
	}

	switch tailing {
	case ProcTailingNone:
	case ProcTailingMustTail:
		if c.CurrProcSig == nil || !are_types_identical(c.CurrProcSig, pt) {
			begin_errorblock()
			a := type_to_string(pt)
			b := type_to_string(c.CurrProcSig)
			error(call, "Use of '#must_tail' of a procedure must have the same type as the procedure it was called within")
			errorline("\tCall type: %s, parent type: %s", a, b)
			gb_string_free(b)
			gb_string_free(a)
			end_errorblock()
		}
	}

	{
		var invalid string
		if pt.Kind == TypeProc && len(pt.Proc.RequireTargetFeature) != 0 {
			if !check_target_feature_is_valid_for_target_arch(pt.Proc.RequireTargetFeature, &invalid) {
				error(call, "Called procedure requires target feature '%.*s' which is invalid for the build target", invalid)
			} else if !check_target_feature_is_enabled(pt.Proc.RequireTargetFeature, &invalid) {
				error(call, "Calling this procedure requires target feature '%.*s' to be enabled", invalid)
			}
		}
		if pt.Kind == TypeProc && len(pt.Proc.EnableTargetFeature) != 0 {
			if !check_target_feature_is_valid_for_target_arch(pt.Proc.EnableTargetFeature, &invalid) {
				error(call, "Called procedure enables target feature '%.*s' which is invalid for the build target", invalid)
			}
			if is_call_inlined {
				if c.CurrProcDecl == nil {
					error(call, "Calling a '#force_inline' procedure that enables target features is not allowed at file scope")
				} else {
					e := c.CurrProcDecl.Entity
					scope_features := e.Type.Proc.EnableTargetFeature
					if !check_target_feature_is_superset_of(scope_features, pt.Proc.EnableTargetFeature, &invalid) {
						begin_errorblock()
						error(call, "Inlined procedure enables target feature '%.*s', this requires the calling procedure to at least enable the same feature", invalid)
						errorline("\tSuggested Example: @(enable_target_feature=\"%.*s\")\n", invalid)
						end_errorblock()
					}
				}
			}
		}
	}

	operand.Expr = call
	{
		type_ := type_of_expr(call.CallExpr.Proc)
		if type_ == nil {
			type_ = pt
		}
		type_ = base_type(type_)
		if type_.Kind == TypeProc && type_.Proc.OptionalOK && type_.Proc.ResultCount > 0 {
			operand.Mode = AddressingOptionalOk
			operand.Type = type_.Proc.Results.Tuple.Variables[0].Type
			if operand.Expr != nil && operand.Expr.Kind == AstCallExpr {
				operand.Expr.CallExpr.OptionalOkOne = true
			}
		}
	}

	proc_entity := entity_from_expr(call.CallExpr.Proc)
	is_objc_call := proc_entity != nil && proc_entity.Kind == Entity_Procedure && proc_entity.Procedure.IsObjcImplOrImport
	if is_objc_call {
		check_objc_call_expr(c, operand, call, proc_entity, pt)
	}
	return ExprExpr
}

func sort_valids(valids []ValidIndexAndScore) {
	for i := 0; i < len(valids); i++ {
		for j := i + 1; j < len(valids); j++ {
			if valids[j].Score > valids[i].Score {
				valids[i], valids[j] = valids[j], valids[i]
			}
		}
	}
}
