// docs.odin - Pure Odin rewrite of src/cipp/docs.i.cpp
// Package: godin
// Binary .odin-doc format reader/writer and text doc printer

package godin

import "core:fmt"
import "core:os"
import "core:mem"
import "core:strings"
import "core:slice"
import "core:hash"
import "core:encoding/hex"
import "core:sync/atomic"

// ============================================================================
// Constants
// ============================================================================

ODIN_DOC_MAGIC :: "odindoc\x00"

ODIN_DOC_VERSION_MAJOR :: 0
ODIN_DOC_VERSION_MINOR :: 3
ODIN_DOC_VERSION_PATCH :: 2

ODIN_DOC_TYPE_ELEMS_CAP :: 4

// FNV-1a initial hash value
FNV1A_INIT :: 0x811c9dc5

// ============================================================================
// OdinDoc binary format structures (exact layout for .odin-doc compatibility)
// ============================================================================

OdinDocArray :: struct($T: typeid) {
    offset: u32,
    length: u32,
}

OdinDocString :: OdinDocArray(u8)

OdinDocVersionType :: struct #packed {
    major, minor, patch: u8,
    pad0: u8,
}

OdinDocHeaderBase :: struct #packed {
    magic:       [8]u8,
    padding0:    u32,
    version:     OdinDocVersionType,
    total_size:  u32,
    header_size: u32,
    hash:        u32,
}

OdinDocFileIndex   :: u32
OdinDocPkgIndex    :: u32
OdinDocEntityIndex :: u32
OdinDocTypeIndex   :: u32

OdinDocFile :: struct {
    pkg:  OdinDocPkgIndex,
    name: OdinDocString,
}

OdinDocPosition :: struct {
    file:   OdinDocFileIndex,
    line:   u32,
    column: u32,
    offset: u32,
}

// ============================================================================
// Type kind and flag enums
// ============================================================================

OdinDocTypeKind :: enum u32 {
    Invalid                   = 0,
    Basic                     = 1,
    Named                     = 2,
    Generic                   = 3,
    Pointer                   = 4,
    Array                     = 5,
    EnumeratedArray           = 6,
    Slice                     = 7,
    DynamicArray              = 8,
    Map                       = 9,
    Struct                    = 10,
    Union                     = 11,
    Enum                      = 12,
    Tuple                     = 13,
    Proc                      = 14,
    BitSet                    = 15,
    SimdVector                = 16,
    SOAStructFixed            = 17,
    SOAStructSlice            = 18,
    SOAStructDynamic          = 19,
    MultiPointer              = 22,
    Matrix                    = 23,
    SoaPointer                = 24,
    BitField                  = 25,
    FixedCapacityDynamicArray = 26,
}

OdinDocTypeFlag_Basic :: enum u32 {
    Untyped = 1 << 1,
}

OdinDocTypeFlag_Struct :: enum u32 {
    Polymorphic = 1 << 0,
    Packed      = 1 << 1,
    RawUnion    = 1 << 2,
    AllOrNone   = 1 << 3,
}

OdinDocTypeFlag_Union :: enum u32 {
    Polymorphic = 1 << 0,
    NoNil       = 1 << 1,
    SharedNil   = 1 << 3,
}

OdinDocTypeFlag_Proc :: enum u32 {
    Polymorphic = 1 << 0,
    Diverging   = 1 << 1,
    OptionalOk  = 1 << 2,
    Variadic    = 1 << 3,
    CVararg     = 1 << 4,
}

OdinDocTypeFlag_BitSet :: enum u32 {
    Range          = 1 << 1,
    OpLt           = 1 << 2,
    OpLtEq         = 1 << 3,
    UnderlyingType = 1 << 4,
}

OdinDocType :: struct {
    kind:              OdinDocTypeKind,
    flags:             u32,
    name:              OdinDocString,
    custom_align:      OdinDocString,
    elem_count_len:    u32,
    elem_counts:       [ODIN_DOC_TYPE_ELEMS_CAP]i64,
    calling_convention: OdinDocString,
    types:             OdinDocArray(OdinDocTypeIndex),
    entities:          OdinDocArray(OdinDocEntityIndex),
    polymorphic_params: OdinDocTypeIndex,
    where_clauses:     OdinDocArray(OdinDocString),
    tags:              OdinDocArray(OdinDocString),
}

// ============================================================================
// Entity structures
// ============================================================================

OdinDocEntityKind :: enum u32 {
    Invalid     = 0,
    Constant    = 1,
    Variable    = 2,
    TypeName    = 3,
    Procedure   = 4,
    ProcGroup   = 5,
    ImportName  = 6,
    LibraryName = 7,
    Builtin     = 8,
}

OdinDocEntityFlag :: enum u64 {
    Foreign              = 1 << 0,
    Export               = 1 << 1,
    Param_Using          = 1 << 2,
    Param_Const          = 1 << 3,
    Param_AutoCast       = 1 << 4,
    Param_Ellipsis       = 1 << 5,
    Param_CVararg        = 1 << 6,
    Param_NoAlias        = 1 << 7,
    Param_AnyInt         = 1 << 8,
    Param_ByPtr          = 1 << 9,
    Param_NoBroadcast    = 1 << 10,
    BitField_Field       = 1 << 19,
    Type_Alias           = 1 << 20,
    Builtin_Pkg_Builtin    = 1 << 30,
    Builtin_Pkg_Intrinsics = 1 << 31,
    Var_Thread_Local     = 1 << 40,
    Var_Static           = 1 << 41,
    Private              = 1 << 50,
}

OdinDocAttribute :: struct {
    name:  OdinDocString,
    value: OdinDocString,
}

OdinDocEntity :: struct {
    kind:              OdinDocEntityKind,
    reserved:          u32,
    flags:             u64,
    pos:               OdinDocPosition,
    name:              OdinDocString,
    type:              OdinDocTypeIndex,
    init_string:       OdinDocString,
    reserved_for_init: u32,
    comment:           OdinDocString,
    docs:              OdinDocString,
    field_group_index: i32,
    foreign_library:   OdinDocEntityIndex,
    link_name:         OdinDocString,
    attributes:        OdinDocArray(OdinDocAttribute),
    grouped_entities:  OdinDocArray(OdinDocEntityIndex),
    where_clauses:     OdinDocArray(OdinDocString),
}

// ============================================================================
// Package structures
// ============================================================================

OdinDocPkgFlags :: enum u32 {
    Builtin = 1 << 0,
    Runtime = 1 << 1,
    Init    = 1 << 2,
}

OdinDocScopeEntry :: struct {
    name:   OdinDocString,
    entity: OdinDocEntityIndex,
}

OdinDocPkg :: struct {
    fullpath: OdinDocString,
    name:     OdinDocString,
    flags:    u32,
    docs:     OdinDocString,
    files:    OdinDocArray(OdinDocFileIndex),
    entries:  OdinDocArray(OdinDocScopeEntry),
}

OdinDocHeader :: struct {
    base:      OdinDocHeaderBase,
    files:     OdinDocArray(OdinDocFile),
    pkgs:      OdinDocArray(OdinDocPkg),
    entities:  OdinDocArray(OdinDocEntity),
    types:     OdinDocArray(OdinDocType),
}

// ============================================================================
// Writer item tracker (generic)
// ============================================================================

OdinDocWriterItemTracker :: struct($T: typeid) {
    len:    int,
    cap:    int,
    offset: int,
}

// ============================================================================
// Writer state
// ============================================================================

OdinDocWriterState :: enum {
    Preparing,
    Writing,
}

ODIN_DOC_WRITER_STATE_STRINGS :: [OdinDocWriterState]string{
    .Preparing = "preparing",
    .Writing   = "writing  ",
}

OdinDocWriter :: struct {
    info:          ^CheckerInfo,
    state:         OdinDocWriterState,
    data:          rawptr,
    data_len:      int,
    header:        ^OdinDocHeader,
    string_cache:  map[string]OdinDocString,
    file_cache:    map[rawptr]OdinDocFileIndex,
    pkg_cache:     map[rawptr]OdinDocPkgIndex,
    entity_cache:  map[rawptr]OdinDocEntityIndex,
    type_cache:    map[u64]OdinDocTypeIndex,
    files:         OdinDocWriterItemTracker(OdinDocFile),
    pkgs:          OdinDocWriterItemTracker(OdinDocPkg),
    entities:      OdinDocWriterItemTracker(OdinDocEntity),
    types:         OdinDocWriterItemTracker(OdinDocType),
    strings:       OdinDocWriterItemTracker(u8),
    blob:          OdinDocWriterItemTracker(u8),
}

// Global atomic flag
g_in_doc_writer: atomic.Bool

// ============================================================================
// Entity printing order table (for text output)
// ============================================================================

@(private="file")
print_entity_kind_ordering := [Entity_Count]int{
    -1, 0, 1, 4, 2, 3, -1, -1, -1, -1, -1,
}

@(private="file")
print_entity_names := [Entity_Count]string{
    "",
    "constants",
    "variables",
    "types",
    "procedures",
    "proc_group",
    "",
    "import names",
    "library names",
    "",
    "",
}

// ============================================================================
// Helper: alignment
// ============================================================================

@(private="file")
align_formula_int :: proc(x: int, align: int) -> int {
    return (x + align - 1) & ~(align - 1)
}

@(private="file")
align_of_type :: proc($T: typeid) -> int {
    return align_formula_int(0, align_of(T)) // returns the alignment value
}

// ============================================================================
// Binary data access (from_array, from_string equivalents)
// ============================================================================

@(private="file")
from_array :: proc(base: ^OdinDocHeaderBase, a: OdinDocArray($T)) -> []T {
    if a.length == 0 do return nil
    data := mem.ptr_offset((^T)(base), int(a.offset))
    return mem.slice_ptr(data, int(a.length))
}

@(private="file")
from_string :: proc(base: ^OdinDocHeaderBase, s: OdinDocString) -> string {
    if s.length == 0 do return ""
    data := mem.ptr_offset((^u8)(base), int(s.offset))
    return string(data[:int(s.length)])
}

// ============================================================================
// FNV-1a hash (for binary doc header)
// ============================================================================

@(private="file")
hash_data_after_header :: proc(base: ^OdinDocHeaderBase, data: rawptr, data_len: int) -> u32 {
    start := (^u8)(data)
    h := u32(FNV1A_INIT)
    offset := int(base.header_size)
    end_offset := int(base.total_size)
    for i := offset; i < end_offset; i += 1 {
        h = (h ~ u32(start[i])) * 0x01000193
    }
    return h
}

// ============================================================================
// Writer item tracker init/size
// ============================================================================

@(private="file")
odin_doc_writer_item_tracker_init :: proc(t: ^OdinDocWriterItemTracker($T), size: int) {
    t.len = size
    t.cap = size
}

@(private="file")
odin_doc_writer_tracker_size :: proc(offset: ^int, t: ^OdinDocWriterItemTracker($T), alignment := 1) {
    size := t.cap * size_of(T)
    align_val := max(align_of(T), alignment)
    offset^ = align_formula_int(offset^, align_val)
    t.offset = offset^
    offset^ += size
}

@(private="file")
odin_doc_writer_calc_total_size :: proc(w: ^OdinDocWriter) -> int {
    total_size := size_of(OdinDocHeader)
    odin_doc_writer_tracker_size(&total_size, &w.files)
    odin_doc_writer_tracker_size(&total_size, &w.pkgs)
    odin_doc_writer_tracker_size(&total_size, &w.entities)
    odin_doc_writer_tracker_size(&total_size, &w.types)
    odin_doc_writer_tracker_size(&total_size, &w.strings, 16)
    odin_doc_writer_tracker_size(&total_size, &w.blob, 16)
    return total_size
}

// ============================================================================
// Writer prepare / destroy / start_writing / end_writing
// ============================================================================

@(private="file")
odin_doc_writer_prepare :: proc(w: ^OdinDocWriter, allocator := context.allocator) {
    context.allocator = allocator
    w.state = .Preparing
    w.string_cache = make(map[string]OdinDocString)
    w.file_cache = make(map[rawptr]OdinDocFileIndex)
    w.pkg_cache = make(map[rawptr]OdinDocPkgIndex)
    w.entity_cache = make(map[rawptr]OdinDocEntityIndex)
    w.type_cache = make(map[u64]OdinDocTypeIndex)
    odin_doc_writer_item_tracker_init(&w.files, 1)
    odin_doc_writer_item_tracker_init(&w.pkgs, 1)
    odin_doc_writer_item_tracker_init(&w.entities, 1)
    odin_doc_writer_item_tracker_init(&w.types, 1)
    odin_doc_writer_item_tracker_init(&w.strings, 16)
    odin_doc_writer_item_tracker_init(&w.blob, 16)
}

@(private="file")
odin_doc_writer_destroy :: proc(w: ^OdinDocWriter, allocator := context.allocator) {
    context.allocator = allocator
    free(w.data, allocator)
    delete(w.string_cache)
    delete(w.file_cache)
    delete(w.pkg_cache)
    delete(w.entity_cache)
    delete(w.type_cache)
}

@(private="file")
odin_doc_writer_start_writing :: proc(w: ^OdinDocWriter, allocator := context.allocator) {
    context.allocator = allocator
    w.state = .Writing
    clear(&w.string_cache)
    clear(&w.file_cache)
    clear(&w.pkg_cache)
    clear(&w.entity_cache)
    clear(&w.type_cache)

    total_size := odin_doc_writer_calc_total_size(w)
    total_size = align_formula_int(total_size, 8)
    w.data, _ = mem.alloc(total_size, 8, allocator)
    w.data_len = total_size
    w.header = (^OdinDocHeader)(w.data)
}

@(private="file")
odin_doc_writer_assign_tracker :: proc(array: ^OdinDocArray($T), t: OdinDocWriterItemTracker(T)) {
    array.offset = u32(t.offset)
    array.length = u32(t.len)
}

@(private="file")
odin_doc_writer_end_writing :: proc(w: ^OdinDocWriter) {
    h := w.header
    copy(h.base.magic[:], ODIN_DOC_MAGIC)
    h.base.version.major = ODIN_DOC_VERSION_MAJOR
    h.base.version.minor = ODIN_DOC_VERSION_MINOR
    h.base.version.patch = ODIN_DOC_VERSION_PATCH
    h.base.total_size = u32(w.data_len)
    h.base.header_size = u32(size_of(OdinDocHeader))
    h.base.hash = hash_data_after_header(&h.base, w.data, w.data_len)

    odin_doc_writer_assign_tracker(&h.files, w.files)
    odin_doc_writer_assign_tracker(&h.pkgs, w.pkgs)
    odin_doc_writer_assign_tracker(&h.entities, w.entities)
    odin_doc_writer_assign_tracker(&h.types, w.types)
}

// ============================================================================
// Writer: write_item, get_item (generic)
// ============================================================================

@(private="file")
odin_doc_write_item :: proc(w: ^OdinDocWriter, t: ^OdinDocWriterItemTracker($T), item: ^T, dst: ^^T = nil) -> u32 {
    if w.state == .Preparing {
        t.cap += 1
        if dst != nil do dst^ = nil
        return 0
    }
    assert(t.len < t.cap)
    item_index := t.len
    t.len += 1
    data := mem.ptr_offset((^T)(w.data), t.offset + size_of(T) * item_index)
    if item != nil {
        mem.copy(data, item, size_of(T))
    }
    if dst != nil do dst^ = data
    return u32(item_index)
}

@(private="file")
odin_doc_get_item :: proc(w: ^OdinDocWriter, t: ^OdinDocWriterItemTracker($T), index: u32) -> ^T {
    if w.state != .Writing do return nil
    assert(index < u32(t.len))
    data := mem.ptr_offset((^T)(w.data), t.offset + size_of(T) * int(index))
    return data
}

// ============================================================================
// Writer: string handling
// ============================================================================

@(private="file")
odin_doc_write_string_without_cache :: proc(w: ^OdinDocWriter, str: string) -> OdinDocString {
    res: OdinDocString
    if w.state == .Preparing {
        w.strings.cap += len(str) + 1
    } else {
        assert(w.strings.len + len(str) + 1 <= w.strings.cap)
        offset := w.strings.offset + w.strings.len
        data := mem.ptr_offset((^u8)(w.data), offset)
        mem.copy(data, raw_data(str), len(str))
        data[len(str)] = 0
        w.strings.len += len(str) + 1
        res.offset = u32(offset)
        res.length = u32(len(str))
    }
    return res
}

@(private="file")
odin_doc_write_string :: proc(w: ^OdinDocWriter, str: string) -> OdinDocString {
    if cached, ok := w.string_cache[str]; ok {
        return cached
    }
    res := odin_doc_write_string_without_cache(w, str)
    w.string_cache[str] = res
    return res
}

// ============================================================================
// Writer: slice/blob writing
// ============================================================================

@(private="file")
odin_write_slice :: proc(w: ^OdinDocWriter, data: []$T) -> OdinDocArray(T) {
    if len(data) <= 0 {
        return {0, 0}
    }
    alignment := 4
    if w.state == .Preparing {
        w.blob.cap = align_formula_int(w.blob.cap, alignment)
        w.blob.cap += len(data) * size_of(T)
        return {0, 0}
    }
    w.blob.len = align_formula_int(w.blob.len, alignment)
    offset := w.blob.offset + w.blob.len
    dst := mem.ptr_offset((^u8)(w.data), offset)
    mem.copy(dst, raw_data(data), len(data) * size_of(T))
    w.blob.len += len(data) * size_of(T)
    return {u32(offset), u32(len(data))}
}

@(private="file")
odin_write_item_as_slice :: proc(w: ^OdinDocWriter, data: $T) -> OdinDocArray(T) {
    return odin_write_slice(w, []T{data})
}

// ============================================================================
// Writer: odin_doc_token_pos_cast
// ============================================================================

@(private="file")
odin_doc_token_pos_cast :: proc(w: ^OdinDocWriter, pos: TokenPos) -> OdinDocPosition {
    file_index: OdinDocFileIndex = 0
    if pos.file_id != 0 {
        file := global_files[pos.file_id]
        if file != nil {
            found, ok := w.file_cache[file]
            assert(ok, "file_index_found != nullptr")
            file_index = found
        }
    }
    return OdinDocPosition{
        file   = file_index,
        line   = u32(pos.line),
        column = u32(pos.column),
        offset = u32(pos.offset),
    }
}

// ============================================================================
// Writer: comment group → string
// ============================================================================

@(private="file")
odin_doc_append_comment_group_string :: proc(buf: ^[dynamic]u8, g: ^CommentGroup) -> bool {
    if g == nil do return false

    total_len := 0
    for comment in g.list {
        total_len += len(comment.string) + 1
    }
    if total_len <= len(g.list) do return false

    count := 0
    for comment in g.list {
        s := comment.string
        slash_slash := false

        if len(s) > 1 && s[1] == '/' {
            slash_slash = true
            s = s[2:]
        } else if len(s) > 1 && s[1] == '*' {
            s = s[2:]
            if len(s) >= 2 do s = s[:len(s)-2]
        }
        if len(s) > 0 && s[0] == ' ' {
            s = s[1:]
        }

        if slash_slash {
            if strings.has_prefix(s, "+") do continue
            if strings.has_prefix(s, "@(") do continue
        }

        if slash_slash {
            append_elems(buf, ..transmute([]u8)s)
            append(buf, '\n')
            count += 1
        } else {
            pos := 0
            for pos < len(s) {
                end := pos
                for end < len(s) && s[end] != '\n' {
                    end += 1
                }
                line := s[pos:end]
                pos = end
                trimmed := strings.trim_space(line)
                if len(trimmed) == 0 {
                    if count == 0 do continue
                }
                if strings.has_prefix(line, "* ") {
                    line = line[2:]
                }
                append_elems(buf, ..transmute([]u8)line)
                append(buf, '\n')
                count += 1
                if pos < len(s) do pos += 1 // skip newline
            }
        }
    }

    if count > 0 {
        append(buf, '\n')
        return true
    }
    return false
}

@(private="file")
odin_doc_pkg_doc_string :: proc(w: ^OdinDocWriter, pkg: ^AstPackage, allocator := context.allocator) -> OdinDocString {
    if pkg == nil do return {}
    context.allocator = allocator

    buf := make([dynamic]u8, 0, 0, allocator)
    defer delete(buf)

    for f in pkg.files {
        if f.pkg_decl != nil {
            assert(f.pkg_decl.kind == .PackageDecl)
            odin_doc_append_comment_group_string(&buf, f.pkg_decl.PackageDecl.docs)
        }
    }
    return odin_doc_write_string_without_cache(w, string(buf[:]))
}

@(private="file")
odin_doc_comment_group_string :: proc(w: ^OdinDocWriter, g: ^CommentGroup, allocator := context.allocator) -> OdinDocString {
    if g == nil do return {}
    context.allocator = allocator

    buf := make([dynamic]u8, 0, 0, allocator)
    defer delete(buf)
    odin_doc_append_comment_group_string(&buf, g)
    return odin_doc_write_string_without_cache(w, string(buf[:]))
}

// ============================================================================
// Writer: expr string (stub — depends on compiler internals)
// ============================================================================

@(private="file")
odin_doc_expr_string :: proc(w: ^OdinDocWriter, expr: ^Ast, allocator := context.allocator) -> OdinDocString {
    // Stub: requires write_expr_to_string from compiler internals
    // The original calls write_expr_to_string with optional CmdDocFlag_Short mode
    _ = w
    _ = expr
    _ = allocator
    return {}
}

// ============================================================================
// Writer: attributes, where_clauses
// ============================================================================

@(private="file")
odin_doc_attributes :: proc(w: ^OdinDocWriter, attributes: []^Ast, allocator := context.allocator) -> OdinDocArray(OdinDocAttribute) {
    context.allocator = allocator

    count := 0
    for attr in attributes {
        if attr.kind != .Attribute do continue
        count += len(attr.Attribute.elems)
    }

    attribs := make([dynamic]OdinDocAttribute, 0, count, allocator)
    defer delete(attribs)

    for attr in attributes {
        if attr.kind != .Attribute do continue
        for elem in attr.Attribute.elems {
            name: string
            value: ^Ast
            #partial switch elem.kind {
            case .Ident:
                name = elem.Ident.token.string
            case .Implicit:
                name = elem.Implicit.string
            case .FieldValue:
                fv := &elem.FieldValue
                if fv.field.kind == .Ident {
                    name = fv.field.Ident.token.string
                } else if fv.field.kind == .Implicit {
                    name = fv.field.Implicit.string
                }
                value = fv.value
            case:
                continue
            }

            doc_attrib := OdinDocAttribute{
                name  = odin_doc_write_string(w, name),
                value = odin_doc_expr_string(w, value, allocator),
            }
            append(&attribs, doc_attrib)
        }
    }

    return odin_write_slice(w, attribs[:])
}

@(private="file")
odin_doc_where_clauses :: proc(w: ^OdinDocWriter, where_clauses: []^Ast, allocator := context.allocator) -> OdinDocArray(OdinDocString) {
    if len(where_clauses) == 0 do return {}
    context.allocator = allocator

    clauses := make([dynamic]OdinDocString, len(where_clauses), allocator)
    defer delete(clauses)

    for s, i in where_clauses {
        clauses[i] = odin_doc_expr_string(w, s, allocator)
    }
    return odin_write_slice(w, clauses[:])
}

// ============================================================================
// Writer: type and entity slice helpers
// ============================================================================

@(private="file")
odin_doc_type_as_slice :: proc(w: ^OdinDocWriter, type: ^Type, cache := true) -> OdinDocArray(OdinDocTypeIndex) {
    index := odin_doc_type(w, type, cache)
    return odin_write_item_as_slice(w, index)
}

@(private="file")
odin_doc_add_entity_as_slice :: proc(w: ^OdinDocWriter, e: ^Entity) -> OdinDocArray(OdinDocEntityIndex) {
    index := odin_doc_add_entity(w, e)
    return odin_write_item_as_slice(w, index)
}

// ============================================================================
// Writer: odin_doc_type — the big type serialization switch
// ============================================================================

@(private="file")
odin_doc_type :: proc(w: ^OdinDocWriter, type: ^Type, cache := true) -> OdinDocTypeIndex {
    if type == nil do return 0

    t := type
    if t.kind == .Named {
        e := t.Named.type_name
        if e.TypeName.is_type_alias {
            t = t.Named.base
        }
    }

    type_hash: u64
    if cache {
        type_hash = type_hash_canonical_type(t)
        if found, ok := w.type_cache[type_hash]; ok {
            return found
        }
    }

    doc_type: OdinDocType
    dst: ^OdinDocType
    type_index := odin_doc_write_item(w, &w.types, &doc_type, &dst)

    if cache {
        w.type_cache[type_hash] = type_index
    }

    switch t.kind {
    case .Basic:
        doc_type.kind = .Basic
        doc_type.name = odin_doc_write_string(w, t.Basic.name)
        if is_type_untyped(t) {
            doc_type.flags |= u32(OdinDocTypeFlag_Basic.Untyped)
        }

    case .Named:
        doc_type.kind = .Named
        doc_type.name = odin_doc_write_string(w, t.Named.name)
        doc_type.types = odin_doc_type_as_slice(w, base_type(t))
        doc_type.entities = odin_doc_add_entity_as_slice(w, t.Named.type_name)

    case .Generic:
        name := t.Generic.name
        if t.Generic.entity != nil {
            name = t.Generic.entity.token.string
        }
        doc_type.kind = .Generic
        doc_type.name = odin_doc_write_string(w, name)
        if t.Generic.specialized != nil {
            doc_type.types = odin_doc_type_as_slice(w, t.Generic.specialized, false)
        }

    case .Pointer:
        doc_type.kind = .Pointer
        doc_type.types = odin_doc_type_as_slice(w, t.Pointer.elem)

    case .MultiPointer:
        doc_type.kind = .MultiPointer
        doc_type.types = odin_doc_type_as_slice(w, t.MultiPointer.elem)

    case .SoaPointer:
        doc_type.kind = .SoaPointer
        doc_type.types = odin_doc_type_as_slice(w, t.SoaPointer.elem)

    case .Array:
        doc_type.kind = .Array
        doc_type.elem_count_len = 1
        doc_type.elem_counts[0] = t.Array.count
        if t.Array.generic_count != nil {
            types := [2]OdinDocTypeIndex{
                odin_doc_type(w, t.Array.elem),
                odin_doc_type(w, t.Array.generic_count),
            }
            doc_type.types = odin_write_slice(w, types[:])
        } else {
            doc_type.types = odin_doc_type_as_slice(w, t.Array.elem)
        }

    case .EnumeratedArray:
        doc_type.kind = .EnumeratedArray
        doc_type.elem_count_len = 1
        doc_type.elem_counts[0] = t.EnumeratedArray.count
        types := [2]OdinDocTypeIndex{
            odin_doc_type(w, t.EnumeratedArray.index),
            odin_doc_type(w, t.EnumeratedArray.elem),
        }
        doc_type.types = odin_write_slice(w, types[:])

    case .Slice:
        doc_type.kind = .Slice
        doc_type.types = odin_doc_type_as_slice(w, t.Slice.elem)

    case .DynamicArray:
        doc_type.kind = .DynamicArray
        doc_type.types = odin_doc_type_as_slice(w, t.DynamicArray.elem)

    case .FixedCapacityDynamicArray:
        doc_type.kind = .FixedCapacityDynamicArray
        doc_type.elem_count_len = 1
        doc_type.elem_counts[0] = t.FixedCapacityDynamicArray.capacity
        if t.FixedCapacityDynamicArray.generic_capacity != nil {
            types := [2]OdinDocTypeIndex{
                odin_doc_type(w, t.FixedCapacityDynamicArray.elem),
                odin_doc_type(w, t.FixedCapacityDynamicArray.generic_capacity),
            }
            doc_type.types = odin_write_slice(w, types[:])
        } else {
            doc_type.types = odin_doc_type_as_slice(w, t.FixedCapacityDynamicArray.elem)
        }

    case .Map:
        doc_type.kind = .Map
        types := [2]OdinDocTypeIndex{
            odin_doc_type(w, t.Map.key),
            odin_doc_type(w, t.Map.value),
        }
        doc_type.types = odin_write_slice(w, types[:])

    case .BitField:
        doc_type.kind = .BitField
        alloc := context.allocator
        fields := make([dynamic]OdinDocEntityIndex, 0, len(t.BitField.fields), alloc)
        for field in t.BitField.fields {
            append(&fields, odin_doc_add_entity(w, field))
        }
        doc_type.entities = odin_write_slice(w, fields[:])
        delete(fields)
        doc_type.types = odin_doc_type_as_slice(w, t.BitField.backing_type)

    case .Struct:
        if t.Struct.soa_kind != .None {
            switch t.Struct.soa_kind {
            case .Fixed:
                doc_type.kind = .SOAStructFixed
                doc_type.elem_count_len = 1
                doc_type.elem_counts[0] = t.Struct.soa_count
            case .Slice:
                doc_type.kind = .SOAStructSlice
            case .Dynamic:
                doc_type.kind = .SOAStructDynamic
            }
            doc_type.types = odin_doc_type_as_slice(w, t.Struct.soa_elem)
        } else {
            doc_type.kind = .Struct
            if t.Struct.is_polymorphic { doc_type.flags |= u32(OdinDocTypeFlag_Struct.Polymorphic) }
            if t.Struct.is_packed      { doc_type.flags |= u32(OdinDocTypeFlag_Struct.Packed) }
            if t.Struct.is_raw_union   { doc_type.flags |= u32(OdinDocTypeFlag_Struct.RawUnion) }
            if t.Struct.is_all_or_none { doc_type.flags |= u32(OdinDocTypeFlag_Struct.AllOrNone) }

            if t.Struct.custom_min_field_align > 0 || t.Struct.custom_max_field_align > 0 {
                doc_type.elem_count_len = 2
                doc_type.elem_counts[0] = u32(max(t.Struct.custom_min_field_align, 0))
                doc_type.elem_counts[1] = u32(max(t.Struct.custom_max_field_align, 0))
            }

            alloc := context.allocator
            fields := make([dynamic]OdinDocEntityIndex, 0, len(t.Struct.fields), alloc)
            for field in t.Struct.fields {
                append(&fields, odin_doc_add_entity(w, field))
            }
            doc_type.entities = odin_write_slice(w, fields[:])
            delete(fields)

            doc_type.polymorphic_params = odin_doc_type(w, t.Struct.polymorphic_params)

            if t.Struct.node != nil {
                st := &t.Struct.node.StructType
                if st.align != nil {
                    doc_type.custom_align = odin_doc_expr_string(w, st.align, alloc)
                }
                doc_type.where_clauses = odin_doc_where_clauses(w, st.where_clauses[:], alloc)
            }

            tags := make([dynamic]OdinDocString, 0, len(t.Struct.fields), alloc)
            for tag in t.Struct.tags {
                append(&tags, odin_doc_write_string(w, tag))
            }
            doc_type.tags = odin_write_slice(w, tags[:])
            delete(tags)
        }

    case .Union:
        doc_type.kind = .Union
        if t.Union.is_polymorphic { doc_type.flags |= u32(OdinDocTypeFlag_Union.Polymorphic) }
        switch t.Union.kind {
        case .NoNil:     doc_type.flags |= u32(OdinDocTypeFlag_Union.NoNil)
        case .SharedNil: doc_type.flags |= u32(OdinDocTypeFlag_Union.SharedNil)
        }

        alloc := context.allocator
        variants := make([dynamic]OdinDocTypeIndex, 0, len(t.Union.variants), alloc)
        for v in t.Union.variants {
            append(&variants, odin_doc_type(w, v))
        }
        doc_type.types = odin_write_slice(w, variants[:])
        delete(variants)
        doc_type.polymorphic_params = odin_doc_type(w, t.Union.polymorphic_params)

        if t.Union.node != nil && t.Union.node.kind == .UnionType {
            ut := &t.Union.node.UnionType
            if ut.align != nil {
                doc_type.custom_align = odin_doc_expr_string(w, ut.align, alloc)
            }
            doc_type.where_clauses = odin_doc_where_clauses(w, ut.where_clauses[:], alloc)
        }

    case .Enum:
        doc_type.kind = .Enum
        alloc := context.allocator
        fields := make([dynamic]OdinDocEntityIndex, 0, len(t.Enum.fields), alloc)
        for field in t.Enum.fields {
            append(&fields, odin_doc_add_entity(w, field))
        }
        doc_type.entities = odin_write_slice(w, fields[:])
        delete(fields)
        if t.Enum.base_type != nil {
            doc_type.types = odin_doc_type_as_slice(w, t.Enum.base_type)
        }

    case .Tuple:
        doc_type.kind = .Tuple
        alloc := context.allocator
        variables := make([dynamic]OdinDocEntityIndex, 0, len(t.Tuple.variables), alloc)
        for v in t.Tuple.variables {
            append(&variables, odin_doc_add_entity(w, v))
        }
        doc_type.entities = odin_write_slice(w, variables[:])
        delete(variables)

    case .Proc:
        doc_type.kind = .Proc
        if t.Proc.is_polymorphic { doc_type.flags |= u32(OdinDocTypeFlag_Proc.Polymorphic) }
        if t.Proc.diverging      { doc_type.flags |= u32(OdinDocTypeFlag_Proc.Diverging) }
        if t.Proc.optional_ok    { doc_type.flags |= u32(OdinDocTypeFlag_Proc.OptionalOk) }
        if t.Proc.variadic       { doc_type.flags |= u32(OdinDocTypeFlag_Proc.Variadic) }
        if t.Proc.c_vararg       { doc_type.flags |= u32(OdinDocTypeFlag_Proc.CVararg) }

        types := [2]OdinDocTypeIndex{
            odin_doc_type(w, t.Proc.params),
            odin_doc_type(w, t.Proc.results),
        }
        doc_type.types = odin_write_slice(w, types[:])
        conv := proc_calling_convention_strings[t.Proc.calling_convention]
        doc_type.calling_convention = odin_doc_write_string(w, string(conv))

    case .BitSet:
        doc_type.kind = .BitSet
        type_count := 0
        types: [2]OdinDocTypeIndex
        if t.BitSet.elem != nil {
            types[type_count] = odin_doc_type(w, t.BitSet.elem)
            type_count += 1
        }
        if t.BitSet.underlying != nil {
            types[type_count] = odin_doc_type(w, t.BitSet.underlying)
            type_count += 1
            doc_type.flags |= u32(OdinDocTypeFlag_BitSet.UnderlyingType)
        }
        doc_type.types = odin_write_slice(w, types[:type_count])
        doc_type.elem_count_len = 2
        doc_type.elem_counts[0] = t.BitSet.lower
        doc_type.elem_counts[1] = t.BitSet.upper

    case .SimdVector:
        doc_type.kind = .SimdVector
        doc_type.elem_count_len = 1
        doc_type.elem_counts[0] = t.SimdVector.count
        doc_type.types = odin_doc_type_as_slice(w, t.SimdVector.elem)

    case .Matrix:
        doc_type.kind = .Matrix
        doc_type.elem_count_len = 2
        doc_type.elem_counts[0] = t.Matrix.row_count
        doc_type.elem_counts[1] = t.Matrix.column_count
        doc_type.types = odin_doc_type_as_slice(w, t.Matrix.elem)
    }

    if dst != nil {
        dst^ = doc_type
    }
    return type_index
}

// ============================================================================
// Writer: odin_doc_add_entity
// ============================================================================

@(private="file")
odin_doc_add_entity :: proc(w: ^OdinDocWriter, e: ^Entity) -> OdinDocEntityIndex {
    if e == nil do return 0

    if prev, ok := w.entity_cache[e]; ok {
        return prev
    }
    if e.pkg != nil {
        if _, ok := w.pkg_cache[e.pkg]; !ok {
            return 0
        }
    }

    doc_entity: OdinDocEntity
    dst: ^OdinDocEntity
    doc_entity_index := odin_doc_write_item(w, &w.entities, &doc_entity, &dst)
    w.entity_cache[e] = doc_entity_index

    type_expr: ^Ast
    init_expr: ^Ast
    decl_node: ^Ast
    comment: ^CommentGroup
    docs: ^CommentGroup

    if e.decl_info != nil {
        type_expr = e.decl_info.type_expr
        init_expr = e.decl_info.init_expr
        decl_node = e.decl_info.decl_node
        comment = e.decl_info.comment
        docs = e.decl_info.docs
    }

    if e.kind == .Variable {
        if comment == nil { comment = e.Variable.comment }
        if docs == nil    { docs = e.Variable.docs }
    } else if e.kind == .Constant {
        if comment == nil { comment = e.Constant.comment }
        if docs == nil    { docs = e.Constant.docs }
    }

    name := e.token.string
    link_name: string
    pos := e.token.pos
    kind: OdinDocEntityKind
    flags: u64
    field_group_index: i32 = -1

    kind = OdinDocEntityKind(u32(e.kind)) // Assumes 1:1 mapping with EntityKind

    switch e.kind {
    case .TypeName:
        if e.TypeName.is_type_alias {
            flags |= u64(OdinDocEntityFlag.Type_Alias)
        }
    case .Variable:
        if e.Variable.is_foreign { flags |= u64(OdinDocEntityFlag.Foreign) }
        if e.Variable.is_export  { flags |= u64(OdinDocEntityFlag.Export) }
        if e.Variable.thread_local_model != "" {
            flags |= u64(OdinDocEntityFlag.Var_Thread_Local)
        }
        if .Static in e.flags { flags |= u64(OdinDocEntityFlag.Var_Static) }
        link_name = e.Variable.link_name
        if init_expr == nil {
            init_expr = e.Variable.init_expr
        }
        if .BitFieldField in e.flags {
            field_group_index = -i32(e.Variable.bit_field_bit_size)
        } else {
            field_group_index = e.Variable.field_group_index
        }
    case .Constant:
        field_group_index = e.Constant.field_group_index
    case .Procedure:
        if e.Procedure.is_foreign { flags |= u64(OdinDocEntityFlag.Foreign) }
        if e.Procedure.is_export  { flags |= u64(OdinDocEntityFlag.Export) }
        link_name = e.Procedure.link_name
    case .Builtin:
        bp := builtin_procs[e.Builtin.id]
        pos = {}
        name = bp.name
        switch bp.pkg {
        case .Builtin:
            flags |= u64(OdinDocEntityFlag.Builtin_Pkg_Builtin)
        case .Intrinsics:
            flags |= u64(OdinDocEntityFlag.Builtin_Pkg_Intrinsics)
        }
    }

    if .Using in e.flags       { flags |= u64(OdinDocEntityFlag.Param_Using) }
    if .ConstInput in e.flags  { flags |= u64(OdinDocEntityFlag.Param_Const) }
    if .Ellipsis in e.flags    { flags |= u64(OdinDocEntityFlag.Param_Ellipsis) }
    if .NoAlias in e.flags     { flags |= u64(OdinDocEntityFlag.Param_NoAlias) }
    if .AnyInt in e.flags      { flags |= u64(OdinDocEntityFlag.Param_AnyInt) }
    if .ByPtr in e.flags       { flags |= u64(OdinDocEntityFlag.Param_ByPtr) }
    if .NoBroadcast in e.flags { flags |= u64(OdinDocEntityFlag.Param_NoBroadcast) }

    // Private flag check
    if e.scope != nil && (.File in e.scope.flags || .Pkg in e.scope.flags) && !is_entity_exported(e) {
        flags |= u64(OdinDocEntityFlag.Private)
    }

    init_string: OdinDocString
    if init_expr != nil {
        init_string = odin_doc_expr_string(w, init_expr, context.allocator)
    } else {
        switch e.kind {
        case .Constant:
            if .ImplicitEnumValue in e.Constant.flags {
                init_string = {}
            } else if e.Constant.param_value.original_ast_expr != nil {
                init_string = odin_doc_expr_string(w, e.Constant.param_value.original_ast_expr, context.allocator)
            } else {
                init_string = odin_doc_write_string(w, exact_value_to_string(e.Constant.value))
            }
        case .Variable:
            if e.Variable.param_value.original_ast_expr != nil {
                init_string = odin_doc_expr_string(w, e.Variable.param_value.original_ast_expr, context.allocator)
            }
        }
    }

    doc_entity.kind = kind
    doc_entity.flags = flags
    doc_entity.pos = odin_doc_token_pos_cast(w, pos)
    doc_entity.name = odin_doc_write_string(w, name)
    doc_entity.type = 0
    doc_entity.init_string = init_string
    doc_entity.comment = odin_doc_comment_group_string(w, comment, context.allocator)
    doc_entity.docs = odin_doc_comment_group_string(w, docs, context.allocator)
    doc_entity.field_group_index = field_group_index
    doc_entity.foreign_library = 0
    doc_entity.link_name = odin_doc_write_string(w, link_name)

    if e.decl_info != nil {
        doc_entity.attributes = odin_doc_attributes(w, e.decl_info.attributes[:], context.allocator)
    }
    doc_entity.grouped_entities = {}
    doc_entity.where_clauses = {}

    if dst != nil {
        dst^ = doc_entity
    }
    return doc_entity_index
}

// ============================================================================
// Writer: odin_doc_update_entities
// ============================================================================

@(private="file")
odin_doc_update_entities :: proc(w: ^OdinDocWriter, allocator := context.allocator) {
    context.allocator = allocator

    // First pass: add types for all cached entities
    {
        entities := make([dynamic]^Entity, 0, len(w.entity_cache), allocator)
        for e_ptr := range w.entity_cache {
            append(&entities, cast(^Entity)e_ptr)
        }
        for e in entities {
            assert(e != nil)
            _ = odin_doc_type(w, e.type)
        }
        delete(entities)
    }

    // Second pass: update each entity entry with type, foreign library, grouped entities
    // Stub: original iterates over entity cache entries using OrderedInsertPtrMap internals
    // In Odin we iterate differently due to map semantics
    for e_ptr, entity_index in w.entity_cache {
        e := cast(^Entity)e_ptr
        type_index := odin_doc_type(w, e.type)

        foreign_library: OdinDocEntityIndex
        grouped_entities: OdinDocArray(OdinDocEntityIndex)

        switch e.kind {
        case .Variable:
            if w.state == .Writing {
                assert(type_index != 0)
            }
            foreign_library = odin_doc_add_entity(w, e.Variable.foreign_library)
        case .Procedure:
            foreign_library = odin_doc_add_entity(w, e.Procedure.foreign_library)
        case .ProcGroup:
            pges := make([dynamic]OdinDocEntityIndex, 0, len(e.ProcGroup.entities), allocator)
            for entity in e.ProcGroup.entities {
                append(&pges, odin_doc_add_entity(w, entity))
            }
            grouped_entities = odin_write_slice(w, pges[:])
            delete(pges)
        }

        if w.state == .Preparing {
            assert(entity_index == 0)
        } else {
            assert(entity_index != 0)
        }

        dst := odin_doc_get_item(w, &w.entities, entity_index)
        if dst != nil {
            if dst.kind == .Variable {
                assert(type_index != 0)
            }
            dst.type = type_index
            dst.foreign_library = foreign_library
            dst.grouped_entities = grouped_entities
        }
    }
}

// ============================================================================
// Writer: odin_doc_add_pkg_entries
// ============================================================================

@(private="file")
odin_doc_add_pkg_entries :: proc(w: ^OdinDocWriter, pkg: ^AstPackage, allocator := context.allocator) -> OdinDocArray(OdinDocScopeEntry) {
    if pkg.scope == nil do return {}
    if _, ok := w.pkg_cache[pkg]; !ok do return {}
    context.allocator = allocator

    entries := make([dynamic]OdinDocScopeEntry, 0, len(w.entity_cache), allocator)
    defer delete(entries)

    // Iterate scope elements — stub: depends on scope internals
    // The original iterates scope->elements hash table slots
    // Skipped: non-exported check via is_entity_exported(e, true)
    _ = entries

    return {} // Stub: body depends on scope element internals
}

// ============================================================================
// Writer: odin_doc_write_docs — main package serialization
// ============================================================================

@(private="file")
odin_doc_write_docs :: proc(w: ^OdinDocWriter, allocator := context.allocator) {
    context.allocator = allocator

    pkgs := make([dynamic]^AstPackage, 0, len(w.info.packages), allocator)
    defer delete(pkgs)

    // Collect packages
    for _, pkg in w.info.packages {
        if .AllPackages in build_context.cmd_doc_flags {
            append(&pkgs, pkg)
        } else {
            if pkg.kind == .Init || pkg.is_extra {
                append(&pkgs, pkg)
            }
        }
    }

    // Sort by name — cmp_ast_package_by_name equivalent
    slice.sort_by(pkgs[:], proc(a, b: ^AstPackage) -> bool {
        return strings.compare(a.name, b.name) < 0
    })

    for pkg in pkgs {
        pkg_flags: u32

        switch pkg.kind {
        case .Normal:  // nothing
        case .Runtime: pkg_flags |= u32(OdinDocPkgFlags.Runtime)
        case .Init:    pkg_flags |= u32(OdinDocPkgFlags.Init)
        case .Builtin: pkg_flags |= u32(OdinDocPkgFlags.Builtin)
        }

        doc_pkg: OdinDocPkg
        doc_pkg.fullpath = odin_doc_write_string(w, pkg.fullpath)
        doc_pkg.name     = odin_doc_write_string(w, pkg.name)
        doc_pkg.flags    = pkg_flags
        doc_pkg.docs     = odin_doc_pkg_doc_string(w, pkg, allocator)

        dst: ^OdinDocPkg
        pkg_index := odin_doc_write_item(w, &w.pkgs, &doc_pkg, &dst)
        w.pkg_cache[pkg] = pkg_index

        file_indices := make([dynamic]OdinDocFileIndex, 0, len(pkg.files), allocator)
        for file in pkg.files {
            doc_file: OdinDocFile
            doc_file.pkg = pkg_index
            doc_file.name = odin_doc_write_string(w, file.fullpath)
            file_index := odin_doc_write_item(w, &w.files, &doc_file)
            w.file_cache[file] = file_index
            append(&file_indices, file_index)
        }
        doc_pkg.files = odin_write_slice(w, file_indices[:])
        delete(file_indices)

        doc_pkg.entries = odin_doc_add_pkg_entries(w, pkg, allocator)

        if dst != nil {
            dst^ = doc_pkg
        }
    }

    odin_doc_update_entities(w, allocator)
}

// ============================================================================
// Writer: odin_doc_write_to_file
// ============================================================================

@(private="file")
odin_doc_write_to_file :: proc(w: ^OdinDocWriter, filename: string) {
    fmt.printf("odin_doc_write_to_file %s\n", filename)

    handle, err := os.open(filename, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0o644)
    if err != os.ERROR_NONE {
        fmt.fprintf(os.stderr, "Failed to write .odin-doc to: %s\n", filename)
        // exit_with_errors() stub
        return
    }
    defer os.close(handle)

    data := mem.byte_slice(w.data, w.data_len)
    os.write(handle, data)
    fmt.printf("Wrote .odin-doc file to: %s\n", filename)
}

// ============================================================================
// Writer: odin_doc_write — main entry
// ============================================================================

@(private="file")
odin_doc_write :: proc(info: ^CheckerInfo, filename: string, allocator := context.allocator) {
    context.allocator = allocator
    atomic.store(&g_in_doc_writer, true)

    w: OdinDocWriter
    w.info = info

    odin_doc_writer_prepare(&w, allocator)
    odin_doc_write_docs(&w, allocator)
    odin_doc_writer_start_writing(&w, allocator)
    odin_doc_write_docs(&w, allocator)
    odin_doc_writer_end_writing(&w)
    odin_doc_write_to_file(&w, filename)

    odin_doc_writer_destroy(&w, allocator)
    atomic.store(&g_in_doc_writer, false)
}

// ============================================================================
// Public: is_in_doc_writer
// ============================================================================

is_in_doc_writer :: proc() -> bool {
    return atomic.load(&g_in_doc_writer)
}

// ============================================================================
// Text-mode doc printer functions
// ============================================================================

@(private="file")
print_doc_line :: proc{
    print_doc_line_string,
    print_doc_line_format,
}

@(private="file")
print_doc_line_string :: proc(indent: i32, data: string) {
    for _ in 0 ..< indent {
        fmt.printf("\t")
    }
    os.write(os.stdout, transmute([]u8)data)
    fmt.printf("\n")
}

@(private="file")
print_doc_line_format :: proc(indent: i32, fmt_str: string, args: ..any) {
    for _ in 0 ..< indent {
        fmt.printf("\t")
    }
    fmt.printfln(fmt_str, ..args)
}

@(private="file")
print_doc_line_no_newline :: proc(indent: i32, data: string) {
    for _ in 0 ..< indent {
        fmt.printf("\t")
    }
    os.write(os.stdout, transmute([]u8)data)
}

@(private="file")
print_doc_comment_group_string :: proc(indent: i32, g: ^CommentGroup) -> bool {
    if g == nil do return false

    total_len := 0
    for comment in g.list {
        total_len += len(comment.string) + 1
    }
    if total_len <= len(g.list) do return false

    count := 0
    for comment in g.list {
        s := comment.string
        slash_slash := false

        if len(s) > 1 && s[1] == '/' {
            slash_slash = true
            s = s[2:]
        } else if len(s) > 1 && s[1] == '*' {
            s = s[2:]
            if len(s) >= 2 do s = s[:len(s)-2]
        }
        if len(s) > 0 && s[0] == ' ' {
            s = s[1:]
        }

        if slash_slash {
            if strings.has_prefix(s, "+") do continue
            if strings.has_prefix(s, "@(") do continue
        }

        if slash_slash {
            print_doc_line_string(indent, s)
            count += 1
        } else {
            pos := 0
            for pos < len(s) {
                end := pos
                for end < len(s) && s[end] != '\n' {
                    end += 1
                }
                line := s[pos:end]
                pos = end
                trimmed := strings.trim_space(line)
                if len(trimmed) == 0 {
                    if count == 0 do continue
                }
                if strings.has_prefix(line, "* ") {
                    line = line[2:]
                }
                print_doc_line_string(indent, line)
                count += 1
                if pos < len(s) do pos += 1
            }
        }
    }

    if count > 0 {
        print_doc_line_string(0, "")
        return true
    }
    return false
}

@(private="file")
print_doc_expr :: proc(expr: ^Ast) {
    // Stub: requires expr_to_string or expr_to_string_shorthand from compiler
    _ = expr
}

// ============================================================================
// Text-mode: print_doc_package
// ============================================================================

@(private="file")
print_doc_package :: proc(info: ^CheckerInfo, pkg: ^AstPackage, allocator := context.allocator) {
    if pkg == nil do return
    context.allocator = allocator

    fmt.printfln("package %s", pkg.name)

    for f in pkg.files {
        if f.pkg_decl != nil {
            assert(f.pkg_decl.kind == .PackageDecl)
            print_doc_comment_group_string(1, f.pkg_decl.PackageDecl.docs)
        }
    }

    if pkg.scope != nil {
        entities := make([dynamic]^Entity, 0, len(pkg.scope.elements), allocator)
        defer delete(entities)

        // Iterate scope elements
        for entry in pkg.scope.elements {
            e := entry.value
            switch e.kind {
            case .Invalid, .Builtin, .Nil, .Label:
                continue
            case .Constant, .Variable, .TypeName, .Procedure, .ProcGroup, .ImportName, .LibraryName:
                // OK
            }
            if e.pkg != pkg do continue
            if !is_entity_exported(e) do continue
            append(&entities, e)
        }

        in_src_order := (.InSourceOrder in build_context.cmd_doc_flags)

        if in_src_order {
            slice.sort_by(entities[:], proc(a, b: ^Entity) -> bool {
                if a.pkg != b.pkg {
                    if a.pkg == nil do return true
                    if b.pkg == nil do return false
                    res := strings.compare(a.pkg.name, b.pkg.name)
                    if res != 0 do return res < 0
                }
                sx := a.order_in_src
                sy := b.order_in_src
                if sx != sy do return sx < sy
                return a.token.pos.offset < b.token.pos.offset
            })
        } else {
            slice.sort_by(entities[:], proc(a, b: ^Entity) -> bool {
                if a.pkg != b.pkg {
                    if a.pkg == nil do return true
                    if b.pkg == nil do return false
                    res := strings.compare(a.pkg.name, b.pkg.name)
                    if res != 0 do return res < 0
                }
                ox := print_entity_kind_ordering[a.kind]
                oy := print_entity_kind_ordering[b.kind]
                res := ox - oy
                if res != 0 do return res < 0
                return strings.compare(a.token.string, b.token.string) < 0
            })
        }

        show_docs := !(.Short in build_context.cmd_doc_flags)

        curr_file: ^AstFile
        curr_entity_kind: EntityKind = .Invalid

        for e in entities {
            if in_src_order {
                if curr_file != e.file {
                    if curr_file != nil {
                        print_doc_line_string(0, "")
                    }
                    curr_file = e.file
                    filename := remove_directory_from_path(curr_file.fullpath)
                    print_doc_line_format(1, "file: %s", filename)
                }
            } else {
                if curr_entity_kind != e.kind {
                    if curr_entity_kind != .Invalid {
                        print_doc_line_string(0, "")
                    }
                    curr_entity_kind = e.kind
                    print_doc_line_format(1, "%s", print_entity_names[e.kind])
                }
            }

            type_expr: ^Ast
            init_expr: ^Ast
            decl_node: ^Ast
            comment: ^CommentGroup
            docs: ^CommentGroup

            if e.decl_info != nil {
                type_expr = e.decl_info.type_expr
                init_expr = e.decl_info.init_expr
                decl_node = e.decl_info.decl_node
                comment = e.decl_info.comment
                docs = e.decl_info.docs
            }

            assert(type_expr != nil || init_expr != nil)

            print_doc_line_no_newline(2, e.token.string)

            if type_expr != nil {
                t := expr_to_string(type_expr)
                fmt.printf(": %s ", t)
                delete(t)
            } else {
                fmt.printf(" :")
            }

            if e.kind == .Variable {
                if init_expr != nil {
                    fmt.printf("= ")
                    print_doc_expr(init_expr)
                }
            } else {
                fmt.printf(": ")
                print_doc_expr(init_expr)
            }
            fmt.printf("\n")

            if show_docs {
                print_doc_comment_group_string(3, docs)
            }
        }
        print_doc_line_string(0, "")
    }

    if len(pkg.fullpath) != 0 {
        print_doc_line_string(0, "")
        print_doc_line_string(1, "fullpath:")
        print_doc_line_format(2, "%s", pkg.fullpath)
        print_doc_line_string(1, "files:")
        for f in pkg.files {
            filename := remove_directory_from_path(f.fullpath)
            print_doc_line_string(2, filename)
        }
    }
}

// ============================================================================
// Public: generate_documentation — main entry point
// ============================================================================

generate_documentation :: proc(c: ^Checker, allocator := context.allocator) {
    context.allocator = allocator
    info := &c.info

    if .DocFormat in build_context.cmd_doc_flags {
        init_fullpath := c.parser.init_fullpath
        output_name: string
        output_base: string

        if len(build_context.out_filepath) == 0 {
            output_name = remove_directory_from_path(init_fullpath)
            output_name = remove_extension_from_path(output_name)
            output_name = strings.trim_space(output_name)
            if len(output_name) == 0 {
                output_name = info.init_scope.pkg.name
            }
            output_base = output_name
        } else {
            output_name = build_context.out_filepath
            output_name = strings.trim_space(output_name)
            if len(output_name) == 0 {
                output_name = info.init_scope.pkg.name
            }
            pos := string_extension_position(output_name)
            if pos < 0 {
                output_base = output_name
            } else {
                output_base = output_name[:pos]
            }
        }

        output_base = path_to_full_path(output_base, allocator)
        output_file_path := strings.concatenate({output_base, ".odin-doc"}, allocator)
        defer delete(output_file_path)

        odin_doc_write(info, output_file_path, allocator)
    } else {
        pkgs := make([dynamic]^AstPackage, 0, len(info.packages), allocator)
        defer delete(pkgs)

        for _, pkg in info.packages {
            if .AllPackages in build_context.cmd_doc_flags {
                append(&pkgs, pkg)
            } else {
                if pkg.kind == .Init || pkg.is_extra {
                    append(&pkgs, pkg)
                }
            }
        }

        slice.sort_by(pkgs[:], proc(a, b: ^AstPackage) -> bool {
            return strings.compare(a.name, b.name) < 0
        })

        for pkg in pkgs {
            print_doc_package(info, pkg, allocator)
        }
    }
}
```