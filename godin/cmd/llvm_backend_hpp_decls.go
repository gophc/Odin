package cmd

func lb_init_generator(gen *lbGenerator, c *Checker) bool { return false }
func lb_mangle_name(e *Entity) string { return "" }
func lb_get_entity_name(m *lbModule, e *Entity) string { return "" }
func lb_create_string_attribute(ctx LLVMContextRef, key string, value string) LLVMAttributeRef { return 0 }
func lb_create_enum_attribute(ctx LLVMContextRef, name string, value ...u64) LLVMAttributeRef { return 0 }
func lb_create_enum_attribute_with_type(ctx LLVMContextRef, name string, typ LLVMTypeRef) LLVMAttributeRef { return 0 }
func lb_add_proc_attribute_at_index(p *lbProcedure, index isize, name string, value ...u64) {}
func lb_add_nocapture_proc_attribute_at_index(p *lbProcedure, index isize) {}
func lb_create_procedure(module *lbModule, entity *Entity, ignore_body ...bool) *lbProcedure { return nil }
func lb_type(m *lbModule, typ *Type) LLVMTypeRef { return 0 }
func llvm_get_element_type(typ LLVMTypeRef) LLVMTypeRef { return 0 }
func lb_create_block(p *lbProcedure, name string, append ...bool) *lbBlock { return nil }
func lb_const_nil(m *lbModule, typ *Type) lbValue { return lbValue{} }
func lb_const_undef(m *lbModule, typ *Type) lbValue { return lbValue{} }
func lb_const_value(m *lbModule, typ *Type, value ExactValue, cc ...lbConstContext) lbValue { return lbValue{} }
func lb_const_bool(m *lbModule, typ *Type, value bool) lbValue { return lbValue{} }
func lb_const_int(m *lbModule, typ *Type, value u64) lbValue { return lbValue{} }
func lb_addr(addr lbValue) lbAddr { return lbAddr{} }
func lb_addr_type(addr lbAddr) *Type { return nil }
func llvm_addr_type(module *lbModule, addr_val lbValue) LLVMTypeRef { return 0 }
func lb_addr_store(p *lbProcedure, addr lbAddr, value lbValue) {}
func lb_addr_load(p *lbProcedure, addr lbAddr) lbValue { return lbValue{} }
func lb_emit_load(p *lbProcedure, v lbValue) lbValue { return lbValue{} }
func lb_emit_store(p *lbProcedure, ptr lbValue, value lbValue) {}
func lb_build_stmt(p *lbProcedure, stmt *Ast) {}
func lb_build_expr(p *lbProcedure, expr *Ast) lbValue { return lbValue{} }
func lb_build_addr(p *lbProcedure, expr *Ast) lbAddr { return lbAddr{} }
func lb_build_stmt_list(p *lbProcedure, stmts []*Ast) {}
func lb_emit_epi(p *lbProcedure, value lbValue, index isize) lbValue { return lbValue{} }
func lb_emit_epi_module(m *lbModule, value lbValue, index isize) lbValue { return lbValue{} }
func lb_emit_array_epi(m *lbModule, s lbValue, index isize) lbValue { return lbValue{} }
func lb_emit_array_epi_proc(p *lbProcedure, value lbValue, index isize) lbValue { return lbValue{} }
func lb_emit_struct_ep(p *lbProcedure, s lbValue, index i32) lbValue { return lbValue{} }
func lb_emit_struct_ev(p *lbProcedure, s lbValue, index i32) lbValue { return lbValue{} }
func lb_emit_tuple_ev(p *lbProcedure, value lbValue, index i32) lbValue { return lbValue{} }
func lb_emit_array_ep(p *lbProcedure, s lbValue, index lbValue) lbValue { return lbValue{} }
func lb_emit_deep_field_gep(p *lbProcedure, e lbValue, sel Selection) lbValue { return lbValue{} }
func lb_emit_deep_field_ev(p *lbProcedure, e lbValue, sel Selection) lbValue { return lbValue{} }
func lb_emit_matrix_ep(p *lbProcedure, s lbValue, row lbValue, column lbValue) lbValue { return lbValue{} }
func lb_emit_matrix_epi(p *lbProcedure, s lbValue, row isize, column isize) lbValue { return lbValue{} }
func lb_emit_matrix_ev(p *lbProcedure, s lbValue, row isize, column isize) lbValue { return lbValue{} }
func lb_emit_arith(p *lbProcedure, op TokenKind, lhs lbValue, rhs lbValue, typ *Type) lbValue { return lbValue{} }
func lb_emit_byte_swap(p *lbProcedure, value lbValue, end_type *Type) lbValue { return lbValue{} }
func lb_emit_defer_stmts(p *lbProcedure, kind lbDeferExitKind, block *lbBlock, pos_or_node any) {}
func lb_emit_transmute(p *lbProcedure, value lbValue, t *Type) lbValue { return lbValue{} }
func lb_emit_comp(p *lbProcedure, op_kind TokenKind, left lbValue, right lbValue) lbValue { return lbValue{} }
func lb_emit_call(p *lbProcedure, value lbValue, args []lbValue, inlining ...ProcInlining) lbValue { return lbValue{} }
func lb_emit_conv(p *lbProcedure, value lbValue, t *Type) lbValue { return lbValue{} }
func lb_emit_comp_against_nil(p *lbProcedure, op_kind TokenKind, x lbValue) lbValue { return lbValue{} }
func lb_emit_jump(p *lbProcedure, target_block *lbBlock) {}
func lb_emit_if(p *lbProcedure, cond lbValue, true_block *lbBlock, false_block *lbBlock) {}
func lb_start_block(p *lbProcedure, b *lbBlock) {}
func lb_build_call_expr(p *lbProcedure, expr *Ast, sret_dst ...*lbValue) lbValue { return lbValue{} }
func lb_create_dummy_procedure(m *lbModule, link_name string, typ *Type) *lbProcedure { return nil }
func lb_begin_procedure_body(p *lbProcedure) {}
func lb_end_procedure_body(p *lbProcedure) {}
func lb_find_or_generate_context_ptr(p *lbProcedure) lbAddr { return lbAddr{} }
func lb_push_context_onto_stack(p *lbProcedure, ctx lbAddr) *lbContextData { return nil }
func lb_push_context_onto_stack_from_implicit_parameter(p *lbProcedure) *lbContextData { return nil }
func lb_add_global_generated_from_procedure(p *lbProcedure, typ *Type, value ...lbValue) lbAddr { return lbAddr{} }
func lb_add_global_generated_with_name(m *lbModule, typ *Type, value lbValue, name string, entity ...**Entity) lbAddr { return lbAddr{} }
func lb_add_local(p *lbProcedure, typ *Type, e ...*Entity) lbAddr { return lbAddr{} }
func lb_add_foreign_library_path(m *lbModule, e *Entity) {}
func lb_typeid(m *lbModule, typ *Type) lbValue { return lbValue{} }
func lb_address_from_load_or_generate_local(p *lbProcedure, value lbValue) lbValue { return lbValue{} }
func lb_address_from_load(p *lbProcedure, value lbValue) lbValue { return lbValue{} }
func lb_add_defer_node(p *lbProcedure, scope_index isize, stmt *Ast) {}
func lb_add_local_generated(p *lbProcedure, typ *Type, zero_init bool) lbAddr { return lbAddr{} }
func lb_emit_runtime_call(p *lbProcedure, c_name string, args []lbValue) lbValue { return lbValue{} }
func lb_emit_ptr_offset(p *lbProcedure, ptr lbValue, index lbValue) lbValue { return lbValue{} }
func lb_const_ptr_offset(m *lbModule, ptr lbValue, index lbValue) lbValue { return lbValue{} }
func lb_string_elem(p *lbProcedure, str lbValue) lbValue { return lbValue{} }
func lb_string_len(p *lbProcedure, str lbValue) lbValue { return lbValue{} }
func lb_cstring_len(p *lbProcedure, value lbValue) lbValue { return lbValue{} }
func lb_array_elem(p *lbProcedure, array_ptr lbValue) lbValue { return lbValue{} }
func lb_slice_elem(p *lbProcedure, slice lbValue) lbValue { return lbValue{} }
func lb_slice_len(p *lbProcedure, slice lbValue) lbValue { return lbValue{} }
func lb_dynamic_array_elem(p *lbProcedure, da lbValue) lbValue { return lbValue{} }
func lb_dynamic_array_len(p *lbProcedure, da lbValue) lbValue { return lbValue{} }
func lb_dynamic_array_cap(p *lbProcedure, da lbValue) lbValue { return lbValue{} }
func lb_dynamic_array_allocator(p *lbProcedure, da lbValue) lbValue { return lbValue{} }
func lb_fixed_capacity_dynamic_array_len(p *lbProcedure, da lbValue) lbValue { return lbValue{} }
func lb_map_len(p *lbProcedure, value lbValue) lbValue { return lbValue{} }
func lb_map_cap(p *lbProcedure, value lbValue) lbValue { return lbValue{} }
func lb_soa_struct_len(p *lbProcedure, value lbValue) lbValue { return lbValue{} }
func lb_emit_increment(p *lbProcedure, addr lbValue) {}
func lb_emit_select(p *lbProcedure, cond lbValue, x lbValue, y lbValue) lbValue { return lbValue{} }
func lb_emit_mul_add(p *lbProcedure, a lbValue, b lbValue, c lbValue, t *Type) lbValue { return lbValue{} }
func lb_fill_slice(p *lbProcedure, slice lbAddr, base_elem lbValue, len lbValue) {}
func lb_type_info(p *lbProcedure, typ *Type) lbValue { return lbValue{} }
func lb_find_or_add_entity_string(m *lbModule, str string, custom_link_section bool) lbValue { return lbValue{} }
func lb_generate_anonymous_proc_lit(m *lbModule, prefix_name string, expr *Ast, parent ...*lbProcedure) lbValue { return lbValue{} }
func lb_is_const(value lbValue) bool { return false }
func lb_is_const_or_global(value lbValue) bool { return false }
func lb_is_const_nil(value lbValue) bool { return false }
func lb_get_const_string(m *lbModule, value lbValue) string { return "" }
func lb_generate_local_array(p *lbProcedure, elem_type *Type, count i64, zero_init ...bool) lbValue { return lbValue{} }
func lb_generate_global_array(m *lbModule, elem_type *Type, count i64, prefix string, id i64) lbValue { return lbValue{} }
func lb_gen_map_key_hash(p *lbProcedure, map_ptr lbValue, key lbValue, key_ptr ...*lbValue) lbValue { return lbValue{} }
func lb_gen_map_cell_info_ptr(m *lbModule, typ *Type) lbValue { return lbValue{} }
func lb_gen_map_info_ptr(m *lbModule, map_type *Type) lbValue { return lbValue{} }
func lb_internal_dynamic_map_get_ptr(p *lbProcedure, map_ptr lbValue, key lbValue) lbValue { return lbValue{} }
func lb_internal_dynamic_map_set(p *lbProcedure, map_ptr lbValue, map_type *Type, map_key lbValue, map_value lbValue, node *Ast) {}
func lb_dynamic_map_reserve(p *lbProcedure, map_ptr lbValue, capacity isize, pos TokenPos) lbValue { return lbValue{} }
func lb_find_procedure_value_from_entity(m *lbModule, e *Entity) lbValue { return lbValue{} }
func lb_find_value_from_entity(m *lbModule, e *Entity) lbValue { return lbValue{} }
func lb_store_type_case_implicit(p *lbProcedure, clause *Ast, value lbValue, is_default_case bool) {}
func lb_store_range_stmt_val(p *lbProcedure, stmt_val *Ast, value lbValue) lbAddr { return lbAddr{} }
func lb_emit_source_code_location_const(p *lbProcedure, procedure string, pos TokenPos) lbValue { return lbValue{} }
func lb_const_source_code_location_const(m *lbModule, procedure string, pos TokenPos) lbValue { return lbValue{} }
func lb_handle_param_value(p *lbProcedure, parameter_type *Type, param_value ParameterValue, procedure_type *TypeProc, call_expression *Ast) lbValue { return lbValue{} }
func lb_equal_proc_for_type(m *lbModule, typ *Type) lbValue { return lbValue{} }
func lb_hasher_proc_for_type(m *lbModule, typ *Type) lbValue { return lbValue{} }

func lb_emit_count_ones(p *lbProcedure, x lbValue, typ *Type) lbValue { return lbValue{} }
func lb_emit_count_zeros(p *lbProcedure, x lbValue, typ *Type) lbValue { return lbValue{} }
func lb_emit_count_trailing_zeros(p *lbProcedure, x lbValue, typ *Type) lbValue { return lbValue{} }
func lb_emit_count_leading_zeros(p *lbProcedure, x lbValue, typ *Type) lbValue { return lbValue{} }
func lb_emit_reverse_bits(p *lbProcedure, x lbValue, typ *Type) lbValue { return lbValue{} }
func lb_emit_bit_set_card(p *lbProcedure, x lbValue) lbValue { return lbValue{} }
func lb_mem_zero_addr(p *lbProcedure, ptr LLVMValueRef, typ *Type) {}
func lb_build_nested_proc(p *lbProcedure, pd *AstProcLit, e *Entity) {}
func lb_emit_logical_binary_expr(p *lbProcedure, op TokenKind, left *Ast, right *Ast, typ *Type) lbValue { return lbValue{} }
func lb_build_cond(p *lbProcedure, cond *Ast, true_block *lbBlock, false_block *lbBlock) lbValue { return lbValue{} }
func llvm_const_named_struct(m *lbModule, t *Type, values []LLVMValueRef, value_count isize) LLVMValueRef { return 0 }
func llvm_const_named_struct_internal(m *lbModule, t LLVMTypeRef, values []LLVMValueRef, value_count isize) LLVMValueRef { return 0 }
func lb_set_entity_from_other_modules_linkage_correctly(other_module *lbModule, e *Entity, name string) {}
func lb_expr_untyped_const_to_typed(m *lbModule, expr *Ast, t *Type) lbValue { return lbValue{} }
func lb_is_expr_untyped_const(expr *Ast) bool { return false }
func llvm_alloca(p *lbProcedure, llvm_type LLVMTypeRef, alignment isize, name ...string) LLVMValueRef { return 0 }
func lb_mem_zero_ptr(p *lbProcedure, ptr LLVMValueRef, typ *Type, alignment uint) {}
func lb_emit_init_context(p *lbProcedure, addr lbAddr) {}
func lb_lookup_branch_blocks(p *lbProcedure, ident *Ast) lbBranchBlocks { return lbBranchBlocks{} }
func lb_get_struct_remapping(m *lbModule, t *Type) lbStructFieldRemapping { return nil }
func lb_type_padding_filler(m *lbModule, padding i64, padding_align i64) LLVMTypeRef { return 0 }
func llvm_basic_shuffle(p *lbProcedure, vector LLVMValueRef, mask LLVMValueRef) LLVMValueRef { return 0 }
func lb_call_intrinsic(p *lbProcedure, name string, args []LLVMValueRef, arg_count uint, types []LLVMTypeRef, type_count uint) LLVMValueRef { return 0 }
func lb_mem_copy_overlapping(p *lbProcedure, dst lbValue, src lbValue, len lbValue, is_volatile ...bool) {}
func lb_mem_copy_non_overlapping(p *lbProcedure, dst lbValue, src lbValue, len lbValue, is_volatile ...bool) {}
func lb_mem_zero_ptr_internal(p *lbProcedure, ptr LLVMValueRef, len any, alignment uint, is_volatile bool) LLVMValueRef { return 0 }
func OdinLLVMGetArrayElementType(typ LLVMTypeRef) LLVMTypeRef { return 0 }
func OdinLLVMGetVectorElementType(typ LLVMTypeRef) LLVMTypeRef { return 0 }
func lb_filepath_ll_for_module(m *lbModule) string { return "" }
func lb_type_internal_for_procedures_raw(m *lbModule, typ *Type) LLVMTypeRef { return 0 }
func lb_emit_source_code_location_as_global_ptr(p *lbProcedure, procedure string, pos TokenPos) lbValue { return lbValue{} }
func lb_debug_location_from_token_pos(p *lbProcedure, pos TokenPos) LLVMMetadataRef { return 0 }
func lb_emit_struct_iv(p *lbProcedure, agg lbValue, field lbValue, index i32) lbValue { return lbValue{} }
func lb_build_struct_value(p *lbProcedure, typ *Type, fields []lbValue, count isize) lbValue { return lbValue{} }
func lb_make_slice_value(p *lbProcedure, slice_type *Type, elem lbValue, len lbValue) lbValue { return lbValue{} }
func lb_make_string_value(p *lbProcedure, string_type *Type, elem lbValue, len lbValue) lbValue { return lbValue{} }
func lb_internal_gen_name_from_type(prefix string, typ *Type) string { return "" }
func lb_set_metadata_custom_u64(m *lbModule, v_ref LLVMValueRef, name string, value u64) {}
func lb_get_metadata_custom_u64(m *lbModule, v_ref LLVMValueRef, name string) u64 { return 0 }
func lb_generate_code(gen *lbGenerator) bool { return false }
