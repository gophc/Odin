package odingo

// ============================================================================
// Types
// ============================================================================

ParseFieldPrefixMapping :: struct {
    name:       string,
    token_kind: TokenKind,
    flag:       FieldFlag,
}

AstAndFlags :: struct {
    node:  ^Ast,
    flags: u32,
}

// ============================================================================
// FieldFlag enum — individual flags used as u32 bit set
// ============================================================================

FieldFlag :: enum u32 {
    Invalid      = 0,
    Unknown      = 1 << 0,
    using        = 1 << 1,
    no_alias     = 1 << 2,
    no_capture   = 1 << 3,
    c_vararg     = 1 << 4,
    const        = 1 << 5,
    any_int      = 1 << 6,
    subtype      = 1 << 7,
    by_ptr       = 1 << 8,
    no_broadcast = 1 << 9,
    ellipsis     = 1 << 10,

    // Composite permission masks (defined in build_settings or similar; placeholders here)
    Struct    = 1 << 20,
    Signature = 1 << 21,
    Tags      = 1 << 22,
    Results   = 1 << 23,
}

// ============================================================================
// Static prefix-mapping table
// ============================================================================

@(private)
parse_field_prefix_mappings := [?]ParseFieldPrefixMapping{
    {"using",        .Token_using,     .using},
    {"no_alias",     .Token_Hash,      .no_alias},
    {"no_capture",   .Token_Hash,      .no_capture},
    {"c_vararg",     .Token_Hash,      .c_vararg},
    {"const",        .Token_Hash,      .const},
    {"any_int",      .Token_Hash,      .any_int},
    {"subtype",      .Token_Hash,      .subtype},
    {"by_ptr",       .Token_Hash,      .by_ptr},
    {"no_broadcast", .Token_Hash,      .no_broadcast},
}

// ============================================================================
// is_token_field_prefix
// ============================================================================

is_token_field_prefix :: proc(f: ^AstFile) -> FieldFlag {
    switch f.curr_token.kind {
    case .Token_EOF:
        return .Invalid

    case .Token_using:
        return .using

    case .Token_Hash:
        advance_token(f)
        if f.curr_token.kind != .Token_Ident {
            return .Unknown
        }
        for mapping in parse_field_prefix_mappings {
            if mapping.token_kind == .Token_Hash && f.curr_token.string == mapping.name {
                return mapping.flag
            }
        }
        return .Unknown
    }

    return .Invalid
}

// ============================================================================
// parse_field_prefixes
// ============================================================================

parse_field_prefixes :: proc(f: ^AstFile) -> u32 {
    counts: [len(parse_field_prefix_mappings)]int

    for {
        flag := is_token_field_prefix(f)
        if u32(flag) & u32(FieldFlag.Invalid) != 0 {
            break
        }
        if u32(flag) & u32(FieldFlag.Unknown) != 0 {
            syntax_error(f.curr_token, "Unknown prefix kind '#%v'", f.curr_token.string)
            advance_token(f)
            continue
        }
        for mapping, i in parse_field_prefix_mappings {
            if mapping.flag == flag {
                counts[i] += 1
                advance_token(f)
                break
            }
        }
    }

    field_flags: u32 = 0
    for mapping, i in parse_field_prefix_mappings {
        if counts[i] > 0 {
            field_flags |= u32(mapping.flag)
            if counts[i] != 1 {
                prefix := "" if mapping.token_kind != .Token_Hash else "#"
                syntax_error(
                    f.curr_token,
                    "Multiple '%s%v' in this field list",
                    prefix,
                    mapping.name,
                )
            }
        }
    }

    return field_flags
}

// ============================================================================
// check_field_prefixes — exported
// ============================================================================

check_field_prefixes :: proc(f: ^AstFile, name_count: int, allowed_flags: u32, set_flags: u32) -> u32 {
    result := set_flags

    for m in parse_field_prefix_mappings {
        err := false
        if (result & u32(m.flag)) != 0 {
            if m.flag == .using && name_count > 1 {
                err = true
                syntax_error(f.curr_token, "Cannot apply 'using' to more than one of the same type")
            }
            if (allowed_flags & u32(m.flag)) == 0 {
                err = true
                prefix := "" if m.token_kind != .Token_Hash else "#"
                syntax_error(
                    f.curr_token,
                    "'%s%v' is not allowed within this field list",
                    prefix,
                    m.name,
                )
            }
        }
        if err {
            result &= ~u32(m.flag)
        }
    }

    return result
}

// ============================================================================
// convert_to_ident_list — exported
// ============================================================================

convert_to_ident_list :: proc(f: ^AstFile, list: [dynamic]AstAndFlags, ignore_flags: bool, allow_poly_names: bool) -> [dynamic]^Ast {
    allocator := context.allocator
    idents := make([dynamic]^Ast, 0, len(list), allocator)

    for item, i in list {
        ident := item.node

        if !ignore_flags && i != 0 {
            syntax_error(ident, "Illegal use of prefixes in parameter list")
        }

        switch ident.kind {
        case .Ast_Ident, .Ast_BadExpr:
            // OK

        case .Ast_Implicit:
            begin_error_block()
            syntax_error(
                ident,
                "Expected an identifier, '%v' which is a keyword",
                ident.implicit.string,
            )
            if ident.implicit.kind == .Token_context {
                error_line("\tSuggestion: Would 'ctx' suffice as an alternative name?\n")
            }
            end_error_block()
            ident = ast_ident(f, blank_token)

        case .Ast_PolyType:
            if allow_poly_names {
                if ident.poly_type.specialization == nil {
                    // OK
                } else {
                    syntax_error(ident, "Expected a polymorphic identifier without any specialization")
                }
            } else {
                syntax_error(ident, "Expected a non-polymorphic identifier")
            }

        case:
            syntax_error(ident, "Expected an identifier")
            ident = ast_ident(f, blank_token)
        }

        append(&idents, ident)
    }

    return idents
}

// ============================================================================
// allow_field_separator — exported
// ============================================================================

allow_field_separator :: proc(f: ^AstFile) -> bool {
    token := f.curr_token

    if allow_token(f, .Token_Comma) {
        return true
    }

    if token.kind == .Token_Semicolon {
        ok := false
        if file_allow_newline(f) && token_is_newline(token) {
            next := peek_token(f).kind
            switch next {
            case .Token_CloseBrace, .Token_CloseParen:
                ok = true
            }
        }
        if !ok {
            p := token_to_string(token)
            syntax_error(
                token_end_of_line(f, f.prev_token),
                "Expected a comma, got a %v",
                p,
            )
        }
        advance_token(f)
        return true
    }

    return false
}

// ============================================================================
// parse_struct_field_list — exported
// ============================================================================

parse_struct_field_list :: proc(f: ^AstFile, name_count_: ^int) -> ^Ast {
    start_token := f.curr_token
    total_name_count: int
    params := parse_field_list(f, &total_name_count, u32(FieldFlag.Struct), .Token_CloseBrace, false, false)
    if name_count_ != nil {
        name_count_^ = total_name_count
    }
    return params
}

// ============================================================================
// check_procedure_name_list — exported
// ============================================================================

check_procedure_name_list :: proc(names: []^Ast) -> bool {
    if len(names) == 0 {
        return false
    }

    first_is_polymorphic := names[0].kind == .Ast_PolyType
    any_polymorphic_names := first_is_polymorphic

    for i in 1 ..< len(names) {
        name := names[i]
        if first_is_polymorphic {
            if name.kind == .Ast_PolyType {
                any_polymorphic_names = true
            } else {
                syntax_error(name, "Mixture of polymorphic and non-polymorphic identifiers")
                return any_polymorphic_names
            }
        } else {
            if name.kind == .Ast_PolyType {
                any_polymorphic_names = true
                syntax_error(name, "Mixture of polymorphic and non-polymorphic identifiers")
                return any_polymorphic_names
            }
        }
    }

    return any_polymorphic_names
}

// ============================================================================
// parse_field_list — exported (main field-list parser)
// ============================================================================

parse_field_list :: proc(
    f:                       ^AstFile,
    name_count_:             ^int,
    allowed_flags:           u32,
    follow:                  TokenKind,
    allow_default_parameters: bool,
    allow_typeid_token:      bool,
) -> ^Ast {
    allocator := context.allocator

    prev_allow_newline := f.allow_newline
    defer f.allow_newline = prev_allow_newline
    f.allow_newline = file_allow_newline(f)

    start_token := f.curr_token
    docs := f.lead_comment
    params := make([dynamic]^Ast, allocator)
    list := make([dynamic]AstAndFlags, temporary_allocator())
    allow_poly_names := allow_typeid_token
    total_name_count: int
    allow_ellipsis := (allowed_flags & u32(FieldFlag.ellipsis)) != 0
    seen_ellipsis := false
    is_signature := (allowed_flags & u32(FieldFlag.Signature)) == u32(FieldFlag.Signature)

    // ----- Phase 1: collect names/prefixes before colon -----
    for f.curr_token.kind != follow &&
        f.curr_token.kind != .Token_Colon &&
        f.curr_token.kind != .Token_EOF {

        if !is_signature {
            parse_enforce_tabs(f)
        }

        flags := parse_field_prefixes(f)
        param := parse_var_type(f, allow_ellipsis, allow_typeid_token)

        if param.kind == .Ast_Ellipsis {
            if seen_ellipsis {
                syntax_error(param, "Extra variadic parameter after ellipsis")
            }
            seen_ellipsis = true
        } else if seen_ellipsis {
            syntax_error(param, "Extra parameter after ellipsis")
        }

        append(&list, AstAndFlags{node = param, flags = flags})

        if !allow_field_separator(f) {
            break
        }
    }

    // ----- No colon → treat each item as field with implicit names -----
    if f.curr_token.kind != .Token_Colon {
        for item in list {
            ty := item.node
            token := blank_token
            if (allowed_flags & u32(FieldFlag.Results)) != 0 {
                token.string = ""
            }
            token.pos = ast_token(ty).pos

            names := make([dynamic]^Ast, 1, allocator)
            names[0] = ast_ident(f, token)

            flags := check_field_prefixes(f, len(list), allowed_flags, item.flags)
            tag: Token
            param := ast_field(f, names[:], ty, nil, flags, tag, docs, f.line_comment)
            append(&params, param)
        }

        if name_count_ != nil {
            name_count_^ = total_name_count
        }
        return ast_field_list(f, start_token, params[:])
    }

    // ----- Trailing-comma-before-colon check -----
    if f.prev_token.kind == .Token_Comma {
        syntax_error(f.prev_token, "Trailing comma before a colon is not allowed")
    }

    // ----- Convert first group to ident list -----
    names := convert_to_ident_list(f, list, true, allow_poly_names)
    if len(names) == 0 {
        syntax_error(f.curr_token, "Empty field declaration")
    }

    any_polymorphic_names := check_procedure_name_list(names[:])

    set_flags: u32
    if len(list) > 0 {
        set_flags = list[0].flags
    }
    set_flags = check_field_prefixes(f, len(names), allowed_flags, set_flags)
    total_name_count += len(names)

    // ----- Parse  type  [= default]  [tag] -----
    ty: ^Ast
    default_value: ^Ast
    tag: Token

    expect_token_after(f, .Token_Colon, "field list")

    if f.curr_token.kind != .Token_Eq {
        ty = parse_var_type(f, allow_ellipsis, allow_typeid_token)
        tt := unparen_expr(ty)
        if tt == nil {
            syntax_error(f.prev_token, "Invalid type expression in field list")
        } else if is_signature && !any_polymorphic_names &&
            tt.kind == .Ast_TypeidType &&
            tt.typeid_type.specialization != nil {
            syntax_error(ty, "Specialization of typeid is not allowed without polymorphic names")
        }
    }

    if allow_token(f, .Token_Eq) {
        default_value = parse_expr(f, false)
        if !allow_default_parameters {
            syntax_error(f.curr_token, "Default parameters are only allowed for procedures")
            default_value = nil
        }
    }

    if default_value != nil && len(names) > 1 {
        syntax_error(f.curr_token, "Default parameters can only be applied to single values")
    }

    if allowed_flags == u32(FieldFlag.Struct) && default_value != nil {
        syntax_error(default_value, "Default parameters are not allowed for structs")
        default_value = nil
    }

    if ty != nil && ty.kind == .Ast_Ellipsis {
        if seen_ellipsis {
            syntax_error(ty, "Extra variadic parameter after ellipsis")
        }
        seen_ellipsis = true
        if len(names) != 1 {
            syntax_error(ty, "Variadic parameters can only have one field name")
        }
    } else if seen_ellipsis && default_value == nil {
        syntax_error(f.curr_token, "Extra parameter after ellipsis without a default value")
    }

    if ty != nil && default_value == nil {
        if f.curr_token.kind == .Token_String {
            tag = expect_token(f, .Token_String)
            if (allowed_flags & u32(FieldFlag.Tags)) == 0 {
                syntax_error(tag, "Field tags are only allowed within structures")
            }
        }
    }

    more_fields := allow_field_separator(f)
    param := ast_field(f, names[:], ty, default_value, set_flags, tag, docs, f.line_comment)
    append(&params, param)

    if !more_fields {
        if name_count_ != nil {
            name_count_^ = total_name_count
        }
        return ast_field_list(f, start_token, params[:])
    }

    // ----- Phase 2: subsequent comma-separated fields -----
    for f.curr_token.kind != follow &&
        f.curr_token.kind != .Token_EOF &&
        f.curr_token.kind != .Token_Semicolon {

        loop_docs := f.lead_comment

        if !is_signature {
            parse_enforce_tabs(f)
        }

        loop_set_flags := parse_field_prefixes(f)
        loop_tag: Token
        loop_names := parse_ident_list(f, allow_poly_names)

        if len(loop_names) == 0 {
            syntax_error(f.curr_token, "Empty field declaration")
            break
        }

        loop_any_poly := check_procedure_name_list(loop_names[:])
        loop_set_flags = check_field_prefixes(f, len(loop_names), allowed_flags, loop_set_flags)
        total_name_count += len(loop_names)

        loop_type: ^Ast
        loop_default: ^Ast

        expect_token_after(f, .Token_Colon, "field list")

        if f.curr_token.kind != .Token_Eq {
            loop_type = parse_var_type(f, allow_ellipsis, allow_typeid_token)
            ltt := unparen_expr(loop_type)
            if is_signature && !loop_any_poly &&
                ltt != nil &&
                ltt.kind == .Ast_TypeidType &&
                ltt.typeid_type.specialization != nil {
                syntax_error(loop_type, "Specialization of typeid is not allowed without polymorphic names")
            }
        }

        if allow_token(f, .Token_Eq) {
            loop_default = parse_expr(f, false)
            if !allow_default_parameters {
                syntax_error(f.curr_token, "Default parameters are only allowed for procedures")
                loop_default = nil
            }
        }

        if loop_default != nil && len(loop_names) > 1 {
            syntax_error(f.curr_token, "Default parameters can only be applied to single values")
        }

        if loop_type != nil && loop_type.kind == .Ast_Ellipsis {
            if seen_ellipsis {
                syntax_error(loop_type, "Extra variadic parameter after ellipsis")
            }
            seen_ellipsis = true
            if len(loop_names) != 1 {
                syntax_error(loop_type, "Variadic parameters can only have one field name")
            }
        } else if seen_ellipsis && loop_default == nil {
            syntax_error(f.curr_token, "Extra parameter after ellipsis without a default value")
        }

        if loop_type != nil && loop_default == nil {
            if f.curr_token.kind == .Token_String {
                loop_tag = expect_token(f, .Token_String)
                if (allowed_flags & u32(FieldFlag.Tags)) == 0 {
                    syntax_error(loop_tag, "Field tags are only allowed within structures")
                }
            }
        }

        ok := allow_field_separator(f)
        loop_param := ast_field(f, loop_names[:], loop_type, loop_default, loop_set_flags, loop_tag, loop_docs, f.line_comment)
        append(&params, loop_param)

        if !ok {
            break
        }
    }

    if name_count_ != nil {
        name_count_^ = total_name_count
    }
    return ast_field_list(f, start_token, params[:])
}

// ============================================================================
// parse_type_or_ident — exported
// Tries to parse a type expression; falls back to an identifier if the
// current token does not start a type.
// ============================================================================

parse_type_or_ident :: proc(f: ^AstFile) -> ^Ast {
    token := f.curr_token

    // Tokens that unambiguously start a type expression
    switch token.kind {
    case .Token_Hat,       // ^T
        .Token_OpenBracket,  // [N]T, [dynamic]T
        .Token_Map,          // map[K]V
        .Token_Struct,       // struct{...}
        .Token_Enum,         // enum{...}
        .Token_Union,        // union{...}
        .Token_BitSet,       // bit_set[...]
        .Token_Proc,         // proc(...)
        .Token_Distinct,     // distinct T
        .Token_Opaque,       // opaque T
        .Token_Matrix,       // matrix[...]T
        .Token_BitField:     // bit_field N
        return parse_type(f)

    case .Token_Ident:
        // An identifier could be a type name or a regular ident.
        // Heuristic: peek ahead; if followed by a period (selector),
        // open-paren (call / poly-type), or type-introducing token,
        // parse as a type; otherwise parse as an identifier.
        next := peek_token(f).kind
        switch next {
        case .Token_Period,       // pkg.Type
            .Token_OpenParen,     // Type(params) specialization / poly-type
            .Token_OpenBracket,   // array index or type param
            .Token_OpenBrace:     // struct literal
            return parse_type(f)
        }
        return parse_ident(f)
    }

    return parse_ident(f)
}
