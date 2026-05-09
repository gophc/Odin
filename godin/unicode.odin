// unicode.odin - Pure Odin rewrite of unicode.i.cpp
package godin

import "core:mem"

// ---------------------------------------------------------------------------
// Constants (utf8proc options flags as bit set)
// ---------------------------------------------------------------------------
utf8proc_option :: enum u32 {
    NULLTERM  =  0,
    STABLE    =  1,
    COMPAT    =  2,
    COMPOSE   =  3,
    DECOMPOSE =  4,
    IGNORE    =  5,
    REJECTNA  =  6,
    NLF2LS    =  7,
    NLF2PS    =  8,
    NLF2LF    = NLF2LS | NLF2PS,
    STRIPCC   =  9,
    CASEFOLD  =  10,
    CHARBOUND =  11,
    LUMP      =  12,
    STRIPMARK =  13,
}
utf8proc_option_t :: bit_set[utf8proc_option]

// ---------------------------------------------------------------------------
// utf8proc Category & property enums (as in header)
// ---------------------------------------------------------------------------
utf8proc_category_t :: enum i16 {
    CN = 0, LU, LL, LT, LM, LO,
    MN, MC, ME, ND, NL, NO,
    PC, PD, PS, PE, PI, PF, PO,
    SM, SC, SK, SO, ZS, ZL, ZP,
    CC, CF, CS, CO,
}

utf8proc_bidi_class_t :: enum i16 {
    _L   = 1, _LRE, _LRO, _R, _AL,
    _RLE = 6, _RLO, _PDF, _EN, _ES,
    _ET  = 11, _AN, _CS, _NSM, _BN,
    _B   = 16, _S, _WS, _ON, _LRI,
    _RLI = 21, _FSI, _PDI,
}

utf8proc_decomp_type_t :: enum i16 {
    FONT      = 1,
    NOBREAK   = 2,
    INITIAL   = 3,
    MEDIAL    = 4,
    FINAL     = 5,
    ISOLATED  = 6,
    CIRCLE    = 7,
    SUPER     = 8,
    SUB       = 9,
    VERTICAL  = 10,
    WIDE      = 11,
    NARROW    = 12,
    SMALL     = 13,
    SQUARE    = 14,
    FRACTION  = 15,
    COMPAT    = 16,
}

utf8proc_boundclass_t :: enum i16 {
    START              = 0,
    OTHER              = 1,
    CR                 = 2,
    LF                 = 3,
    CONTROL            = 4,
    EXTEND             = 5,
    L                  = 6,
    V                  = 7,
    T                  = 8,
    LV                 = 9,
    LVT                = 10,
    REGIONAL_INDICATOR = 11,
    SPACINGMARK        = 12,
    PREPEND            = 13,
    ZWJ                = 14,
    E_BASE             = 15,
    E_MODIFIER         = 16,
    GLUE_AFTER_ZWJ     = 17,
    E_BASE_GAZ         = 18,
}

// ---------------------------------------------------------------------------
// Basic type aliases matching utf8proc types
// ---------------------------------------------------------------------------

utf8proc_property_t :: struct #packed {
    category: i16,
    combining_class: i16,
    bidi_class: i16,
    decomp_type: i16,
    decomp_seqindex: u16,
    casefold_seqindex: u16,
    uppercase_seqindex: u16,
    lowercase_seqindex: u16,
    titlecase_seqindex: u16,
    comb_index: u16,

    // bit-fields (u16 内打包)
    bidi_mirrored: bool `bit_field:"0:1"`,
    comp_exclusion: bool `bit_field:"1:1"`,
    ignorable: bool `bit_field:"2:1"`,
    control_boundary: bool `bit_field:"3:1"`,
    charwidth: u8 `bit_field:"4:2"`,
    pad: u8 `bit_field:"6:2"`,
    boundclass: u8 `bit_field:"8:8"`,
}

utf8proc_utf8class :: [256] i8 {
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
    2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
    3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
    4, 4, 4, 4, 4, 4, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0
};

// ---------------------------------------------------------------------------
// utf8proc functions
// ---------------------------------------------------------------------------
utf8proc_version :: proc() -> string {
    return "2.1.0"
}

utf8proc_errmsg :: proc(errcode: int) -> string {
    switch errcode {
    case -1: return "Memory for processing UTF-8 data could not be allocated."
    case -2: return "UTF-8 string is too long to be processed."
    case -3: return "Invalid UTF-8 string"
    case -4: return "Unassigned Unicode code point found in UTF-8 string."
    case -5: return "Invalid options for UTF-8 processing chosen."
    case: return "An unknown error occurred while processing UTF-8 data."
    }
}


utf8proc_iterate :: proc(str: []u8, strlen: int, dst: ^i32) -> int {
    end_ptr: ^u8
    dst^ = -1
    if strlen == 0 do return 0
    if strlen < 0 {
        end_ptr = mem.ptr_offset(&str[0], 4) // unsafe, but logic says up to 4 bytes
    } else {
        end_ptr = mem.ptr_offset(&str[0], strlen)
    }

    uc := u32(str[0])
    str := str[1:]
    p := &str[0]
    if uc < 0x80 {
        dst^ = i32(uc)
        return 1
    }
    if (uc - 0xc2) > (0xf4 - 0xc2) {
        return -3
    }
    if uc < 0xe0 {
        if p >= end_ptr || (p^ & 0xc0) != 0x80 {
            return -3
        }
        dst^ = i32((uc & 0x1f) << 6 | u32(p^ & 0x3f))
        return 2
    }
    if uc < 0xf0 {
        if mem.ptr_offset(p, 1) >= end_ptr || (p^ & 0xc0) != 0x80 || (str[1] & 0xc0) != 0x80 {
            return -3
        }
        if uc == 0xed && p^ > 0x9f {
            return -3
        }
        uc = (uc & 0xf) << 12 | u32(p^ & 0x3f) << 6 | u32(str[1] & 0x3f)
        if uc < 0x800 {
            return -3
        }
        dst^ = i32(uc)
        return 3
    }
    // 4-byte sequence
    if mem.ptr_offset(p, 2) >= end_ptr || (p^ & 0xc0) != 0x80 || (str[1] & 0xc0) != 0x80 || (str[2] & 0xc0) != 0x80 {
        return -3
    }
    if uc == 0xf0 {
        if p^ < 0x90 {
            return -3
        }
    } else if uc == 0xf4 {
        if p^ > 0x8f {
            return -3
        }
    }
    dst^ = i32((uc & 7) << 18 | u32(p^ & 0x3f) << 12 | u32(str[1] & 0x3f) << 6 | u32(str[2] & 0x3f))
    return 4
}

utf8proc_codepoint_valid :: proc(uc: i32) -> bool {
    return (u32(uc) - 0xd800 > 0x07ff) && (u32(uc) < 0x110000)
}

utf8proc_encode_char :: proc(uc: i32, dst: []u8) -> int {
    switch {
    case uc < 0:
        return 0
    case uc < 0x80:
        dst[0] = u8(uc)
        return 1
    case uc < 0x800:
        dst[0] = u8(0xC0 + (uc >> 6))
        dst[1] = u8(0x80 + (uc & 0x3F))
        return 2
    case uc < 0x10000:
        dst[0] = u8(0xE0 + (uc >> 12))
        dst[1] = u8(0x80 + ((uc >> 6) & 0x3F))
        dst[2] = u8(0x80 + (uc & 0x3F))
        return 3
    case uc < 0x110000:
        dst[0] = u8(0xF0 + (uc >> 18))
        dst[1] = u8(0x80 + ((uc >> 12) & 0x3F))
        dst[2] = u8(0x80 + ((uc >> 6) & 0x3F))
        dst[3] = u8(0x80 + (uc & 0x3F))
        return 4
    }
    return 0
}

@(private)
unsafe_encode_char :: proc(uc: i32, dst: []u8) -> int {
    switch {
    case uc < 0:
        return 0
    case uc < 0x80:
        dst[0] = u8(uc)
        return 1
    case uc < 0x800:
        dst[0] = u8(0xC0 + (uc >> 6))
        dst[1] = u8(0x80 + (uc & 0x3F))
        return 2
    case uc == 0xFFFF:
        dst[0] = 0xFF
        return 1
    case uc == 0xFFFE:
        dst[0] = 0xFE
        return 1
    case uc < 0x10000:
        dst[0] = u8(0xE0 + (uc >> 12))
        dst[1] = u8(0x80 + ((uc >> 6) & 0x3F))
        dst[2] = u8(0x80 + (uc & 0x3F))
        return 3
    case uc < 0x110000:
        dst[0] = u8(0xF0 + (uc >> 18))
        dst[1] = u8(0x80 + ((uc >> 12) & 0x3F))
        dst[2] = u8(0x80 + ((uc >> 6) & 0x3F))
        dst[3] = u8(0x80 + (uc & 0x3F))
        return 4
    }
    return 0
}

@(private)
unsafe_get_property :: proc(uc: i32) -> ^utf8proc_property_t {
// These tables are indexed the same way as in the original C code.
    stage1 := &utf8proc_stage1table[uc >> 8]
    idx := utf8proc_stage2table[stage1^ + u16(uc & 0xFF)]
    return &utf8proc_properties[idx]
}

utf8proc_get_property :: proc(uc: i32) -> ^utf8proc_property_t {
    if uc < 0 || uc >= 0x110000 {
        return &utf8proc_properties[0] // default property
    }
    return unsafe_get_property(uc)
}

@(private)
grapheme_break_simple :: proc(lbc, tbc: i32) -> bool {
    switch {
    case lbc == i32(utf8proc_boundclass_t.START):
        return true
    case lbc == i32(utf8proc_boundclass_t.CR) && tbc == i32(utf8proc_boundclass_t.LF):
        return false
    case lbc >= i32(utf8proc_boundclass_t.CR) && lbc <= i32(utf8proc_boundclass_t.CONTROL):
        return true
    case tbc >= i32(utf8proc_boundclass_t.CR) && tbc <= i32(utf8proc_boundclass_t.CONTROL):
        return true
    case lbc == i32(utf8proc_boundclass_t.L) &&
    (tbc == i32(utf8proc_boundclass_t.L) || tbc == i32(utf8proc_boundclass_t.V) ||
    tbc == i32(utf8proc_boundclass_t.LV) || tbc == i32(utf8proc_boundclass_t.LVT)):
        return false
    case (lbc == i32(utf8proc_boundclass_t.LV) || lbc == i32(utf8proc_boundclass_t.V)) &&
    (tbc == i32(utf8proc_boundclass_t.V) || tbc == i32(utf8proc_boundclass_t.T)):
        return false
    case (lbc == i32(utf8proc_boundclass_t.LVT) || lbc == i32(utf8proc_boundclass_t.T)) &&
    tbc == i32(utf8proc_boundclass_t.T):
        return false
    case tbc == i32(utf8proc_boundclass_t.EXTEND) || tbc == i32(utf8proc_boundclass_t.ZWJ) ||
    tbc == i32(utf8proc_boundclass_t.SPACINGMARK) || lbc == i32(utf8proc_boundclass_t.PREPEND):
        return false
    case (lbc == i32(utf8proc_boundclass_t.E_BASE) || lbc == i32(utf8proc_boundclass_t.E_BASE_GAZ)) &&
    tbc == i32(utf8proc_boundclass_t.E_MODIFIER):
        return false
    case lbc == i32(utf8proc_boundclass_t.ZWJ) &&
    (tbc == i32(utf8proc_boundclass_t.GLUE_AFTER_ZWJ) || tbc == i32(utf8proc_boundclass_t.E_BASE_GAZ)):
        return false
    case lbc == i32(utf8proc_boundclass_t.REGIONAL_INDICATOR) && tbc == i32(utf8proc_boundclass_t.REGIONAL_INDICATOR):
        return false
    }
    return true
}

@(private)
grapheme_break_extended :: proc(lbc, tbc: i32, state: ^i32) -> bool {
    lbc_override := lbc
    if state != nil && state^ != i32(utf8proc_boundclass_t.START) {
        lbc_override = state^
    }
    break_permitted := grapheme_break_simple(lbc_override, tbc)
    if state != nil {
        if state^ == tbc && tbc == i32(utf8proc_boundclass_t.REGIONAL_INDICATOR) {
            state^ = i32(utf8proc_boundclass_t.OTHER)
        } else if (state^ == i32(utf8proc_boundclass_t.E_BASE) || state^ == i32(utf8proc_boundclass_t.E_BASE_GAZ)) &&
        tbc == i32(utf8proc_boundclass_t.EXTEND) {
            state^ = i32(utf8proc_boundclass_t.E_BASE)
        } else {
            state^ = tbc
        }
    }
    return break_permitted
}

utf8proc_grapheme_break_stateful :: proc(c1, c2: i32, state: ^i32) -> bool {
    return grapheme_break_extended(
    i32(utf8proc_get_property(c1).boundclass),
    i32(utf8proc_get_property(c2).boundclass),
    state,
    )
}

utf8proc_grapheme_break :: proc(c1, c2: i32) -> bool {
    return utf8proc_grapheme_break_stateful(c1, c2, nil)
}

@(private)
seqindex_decode_entry :: proc(seqindex: u32) -> (next_seqindex: u32, codepoint: i32) {
    entry_cp := i32(utf8proc_sequences[seqindex])
    next_seqindex = seqindex + 1  // 默认前进一步
    if (entry_cp & 0xF800) == 0xD800 {
        high := u32(entry_cp & 0x03FF) << 10
        low := u32(utf8proc_sequences[next_seqindex])
        codepoint = i32(high | low) + 0x10000
        next_seqindex += 1 // 消耗两个 u16
    } else {
        codepoint = entry_cp
    }
    return
}

@(private)
seqindex_decode_index :: proc(seqindex: u32) -> (codepoint: i32) {
    _, codepoint = seqindex_decode_entry(seqindex)
    return
}

@(private)
seqindex_write_char_decomposed :: proc(
seqindex: u16,
dst: []i32,
bufsize: int,
options: utf8proc_option_t,
last_boundclass: ^i32,
) -> int {
    written := 0
    seqindex := u32(seqindex & 0x1FFF)
    len_ := int(seqindex >> 13)
    if len_ >= 7 {
        len_ = int(utf8proc_sequences[seqindex])
        seqindex += 1
    }
    codepoint :i32
    for idx := 0; idx <= len_; idx += 1 {
        _, codepoint = seqindex_decode_entry(seqindex)
        n := utf8proc_decompose_char(codepoint, dst[written:], max(0, bufsize - written), options, last_boundclass)
        if n < 0 {
            return -2
        }
        written += n
    }
    return written
}

utf8proc_tolower :: proc(c: i32) -> i32 {
    cl := utf8proc_get_property(c).lowercase_seqindex
    if cl != 0xffff {
        return seqindex_decode_index(u32(cl))
    }
    return c
}
utf8proc_toupper :: proc(c: i32) -> i32 {
    cu := utf8proc_get_property(c).uppercase_seqindex
    if cu != 0xffff {
        return seqindex_decode_index(u32(cu))
    }
    return c
}
utf8proc_totitle :: proc(c: i32) -> i32 {
    cu := utf8proc_get_property(c).titlecase_seqindex
    if cu != 0xffff {
        return seqindex_decode_index(u32(cu))
    }
    return c
}
utf8proc_charwidth :: proc(c: i32) -> int {
    return int(utf8proc_get_property(c).charwidth)
}
utf8proc_category :: proc(c: i32) -> utf8proc_category_t {
    return utf8proc_category_t(utf8proc_get_property(c).category)
}
utf8proc_category_string :: proc(c: i32) -> string {
    s := [30]string{
        "Cn", "Lu", "Ll", "Lt", "Lm", "Lo", "Mn", "Mc", "Me", "Nd", "Nl", "No",
        "Pc", "Pd", "Ps", "Pe", "Pi", "Pf", "Po", "Sm", "Sc", "Sk", "So", "Zs", "Zl", "Zp",
        "Cc", "Cf", "Cs", "Co",
    }
    return s[utf8proc_category(c)]
}

utf8proc_decompose_char :: proc(
uc: i32,
dst: []i32,
bufsize: int,
options: utf8proc_option_t,
last_boundclass: ^i32,
) -> int {
    if uc < 0 || uc >= 0x110000 {
        return -4
    }
    property := unsafe_get_property(uc)
    category := utf8proc_category_t(property.category)
    hangul_sindex := uc - 0xAC00

    // Hangul decomposition
    if .COMPOSE in options || .DECOMPOSE in options {
        if hangul_sindex >= 0 && hangul_sindex < 11172 {
            hangul_tindex: i32
            if bufsize >= 1 {
                dst[0] = 0x1100 + hangul_sindex / 588
                if bufsize >= 2 {
                    dst[1] = 0x1161 + (hangul_sindex % 588) / 28
                }
            }
            hangul_tindex = hangul_sindex % 28
            if hangul_tindex == 0 {
                return 2
            }
            if bufsize >= 3 {
                dst[2] = 0x11A7 + hangul_tindex
            }
            return 3
        }
    }

    if .REJECTNA in options {
        if category == .CN {
            return -4
        }
    }

    if .IGNORE in options {
        if property.ignorable {
            return 0
        }
    }

    if .LUMP in options {
        sub_options := options - { .LUMP }
        #partial switch category {
        case .ZS:
            return utf8proc_decompose_char(0x0020, dst, bufsize, sub_options, last_boundclass)
        case .PD:
            return utf8proc_decompose_char(0x002D, dst, bufsize, sub_options, last_boundclass)
        case .PC:
            return utf8proc_decompose_char(0x005F, dst, bufsize, sub_options, last_boundclass)
        }
        switch uc {
        case 0x2018, 0x2019, 0x02BC, 0x02C8:
            return utf8proc_decompose_char(0x0027, dst, bufsize, sub_options, last_boundclass)
        case 0x2212:
            return utf8proc_decompose_char(0x002D, dst, bufsize, sub_options, last_boundclass)
        case 0x2044, 0x2215:
            return utf8proc_decompose_char(0x002F, dst, bufsize, sub_options, last_boundclass)
        case 0x2236:
            return utf8proc_decompose_char(0x003A, dst, bufsize, sub_options, last_boundclass)
        case 0x2039, 0x2329, 0x3008:
            return utf8proc_decompose_char(0x003C, dst, bufsize, sub_options, last_boundclass)
        case 0x203A, 0x232A, 0x3009:
            return utf8proc_decompose_char(0x003E, dst, bufsize, sub_options, last_boundclass)
        case 0x2216:
            return utf8proc_decompose_char(0x005C, dst, bufsize, sub_options, last_boundclass)
        case 0x02C4, 0x02C6, 0x2038, 0x2303:
            return utf8proc_decompose_char(0x005E, dst, bufsize, sub_options, last_boundclass)
        case 0x02CD:
            return utf8proc_decompose_char(0x005F, dst, bufsize, sub_options, last_boundclass)
        case 0x02CB:
            return utf8proc_decompose_char(0x0060, dst, bufsize, sub_options, last_boundclass)
        case 0x2223:
            return utf8proc_decompose_char(0x007C, dst, bufsize, sub_options, last_boundclass)
        case 0x223C:
            return utf8proc_decompose_char(0x007E, dst, bufsize, sub_options, last_boundclass)
        }
        if (.NLF2LS in options) && (.NLF2PS in options) {
            if category == .ZL || category == .ZP {
                return utf8proc_decompose_char(0x000A, dst, bufsize, sub_options, last_boundclass)
            }
        }
    }

    if .STRIPMARK in options {
        if category == .MN || category == .MC || category == .ME {
            return 0
        }
    }

    if .CASEFOLD in options {
        if property.casefold_seqindex != 0xffff {
            return seqindex_write_char_decomposed(property.casefold_seqindex, dst, bufsize, options, last_boundclass)
        }
    }

    if .COMPOSE in options || .DECOMPOSE in options {
        if property.decomp_seqindex != 0xffff &&
        (property.decomp_type == 0 || (.COMPAT in options)) {
            return seqindex_write_char_decomposed(property.decomp_seqindex, dst, bufsize, options, last_boundclass)
        }
    }

    if .CHARBOUND in options {
        tbc := i32(property.boundclass)
        boundary := grapheme_break_extended(last_boundclass^, tbc, last_boundclass)
        if boundary {
            if bufsize >= 1 {
                dst[0] = 0xFFFF
            }
            if bufsize >= 2 {
                dst[1] = uc
            }
            return 2
        }
    }

    if bufsize >= 1 {
        dst[0] = uc
    }
    return 1
}

utf8proc_decompose :: proc(
str: []u8, strlen: int,
buffer: []i32, bufsize: int,
options: utf8proc_option_t,
) -> int {
    return utf8proc_decompose_custom(str, strlen, buffer, bufsize, options, nil, nil)
}

utf8proc_decompose_custom :: proc(
str: []u8, strlen: int,
buffer: []i32, bufsize: int,
options: utf8proc_option_t,
custom_func: proc(i32, rawptr) -> i32,
custom_data: rawptr,
) -> int {
    wpos := 0
    if (.COMPOSE in options) && (.DECOMPOSE in options) {
        return -5
    }
    if (.STRIPMARK in options) && !(.COMPOSE in options) && !(.DECOMPOSE in options) {
        return -5
    }

    boundclass := i32(utf8proc_boundclass_t.START)
    rpos := 0
    for {
        uc: i32
        if .NULLTERM in options {
            n := utf8proc_iterate(str[rpos:], -1, &uc)
            if uc < 0 {
                return -3
            }
            if n < 0 {
                return -2
            }
            if uc == 0 {
                break
            }
            rpos += n
        } else {
            if rpos >= int(strlen) {
                break
            }
            n := utf8proc_iterate(str[rpos:], int(strlen) - rpos, &uc)
            if uc < 0 {
                return -3
            }
            rpos += n
        }
        if custom_func != nil {
            uc = custom_func(uc, custom_data)
        }
        decomp_result := utf8proc_decompose_char(uc, buffer[wpos:], max(0, bufsize - wpos), options, &boundclass)
        if decomp_result < 0 {
            return decomp_result
        }
        wpos += decomp_result
        if wpos < 0 || wpos > (max(int) / 2 / size_of(i32)) {
            return -2
        }
    }

    // canonical ordering
    if (.COMPOSE in options || .DECOMPOSE in options) && bufsize >= wpos {
        pos := 0
        for pos < wpos - 1 {
            uc1, uc2 := buffer[pos], buffer[pos + 1]
            prop1 := unsafe_get_property(uc1)
            prop2 := unsafe_get_property(uc2)
            if prop1.combining_class > prop2.combining_class && prop2.combining_class > 0 {
                buffer[pos], buffer[pos + 1] = uc2, uc1
                if pos > 0 {
                    pos -= 1
                } else {
                    pos += 1
                }
            } else {
                pos += 1
            }
        }
    }
    return wpos
}

utf8proc_normalize_utf32 :: proc(buffer: []i32, length: int, options: utf8proc_option_t) -> int {
    wpos := 0
    rpos := 0
    length := length
    buf := buffer

    // line break and strip control characters
    if (.NLF2LS in options) || (.NLF2PS in options) || (.STRIPCC in options) {
        wpos = 0
        for rpos = 0; rpos < length; rpos += 1 {
            uc := buf[rpos]
            if uc == 0x000D && rpos < length - 1 && buf[rpos + 1] == 0x000A {
                rpos += 1; continue
            }
            if uc == 0x000A || uc == 0x000D || uc == 0x0085 ||
            ((.STRIPCC in options) && (uc == 0x000B || uc == 0x000C)) {
                if .NLF2LS in options {
                    if .NLF2PS in options {
                        buf[wpos] = 0x000A; wpos += 1
                    }
                    else {
                        buf[wpos] = 0x2028; wpos += 1
                    }
                } else {
                    if .NLF2PS in options {
                        buf[wpos] = 0x2029; wpos += 1
                    }
                    else {
                        buf[wpos] = 0x0020; wpos += 1
                    }
                }
            } else if (.STRIPCC in options) && (uc < 0x0020 || (uc >= 0x007F && uc < 0x00A0)) {
                if uc == 0x0009 {
                    buf[wpos] = 0x0020; wpos += 1
                }
            } else {
                buf[wpos] = uc; wpos += 1
            }
        }
        length = wpos
    }

    // compose
    if .COMPOSE in options {
        starter: ^i32
        starter_property: ^utf8proc_property_t
        max_combining_class : i32 = -1
        wpos = 0
        for rpos = 0; rpos < length; rpos += 1 {
            current_char := buf[rpos]
            current_property := unsafe_get_property(current_char)
            if starter != nil && i32(current_property.combining_class) > max_combining_class {
            // Hangul composition L+V
                hangul_lindex := starter^ - 0x1100
                if hangul_lindex >= 0 && hangul_lindex < 19 {
                    hangul_vindex := current_char - 0x1161
                    if hangul_vindex >= 0 && hangul_vindex < 21 {
                        starter^ = 0xAC00 + (hangul_lindex * 21 + hangul_vindex) * 28
                        starter_property = nil
                        continue
                    }
                }
                // Hangul composition LV+T
                hangul_sindex := starter^ - 0xAC00
                if hangul_sindex >= 0 && hangul_sindex < 11172 && (hangul_sindex % 28) == 0 {
                    hangul_tindex := current_char - 0x11A7
                    if hangul_tindex >= 0 && hangul_tindex < 28 {
                        starter^ += hangul_tindex
                        starter_property = nil
                        continue
                    }
                }
                if starter_property == nil {
                    starter_property = unsafe_get_property(starter^)
                }
                if starter_property.comb_index < 0x8000 &&
                current_property.comb_index != 0xffff &&
                current_property.comb_index >= 0x8000 {
                    sidx := int(starter_property.comb_index)
                    idx := (int(current_property.comb_index) & 0x3FFF) - int(utf8proc_combinations[sidx])
                    if idx >= 0 && idx <= int(utf8proc_combinations[sidx + 1]) {
                        idx += sidx + 2
                        composition: i32
                        if (current_property.comb_index & 0x4000) != 0 {
                            composition = i32(utf8proc_combinations[idx]) << 16 | i32(utf8proc_combinations[idx + 1])
                        } else {
                            composition = i32(utf8proc_combinations[idx])
                        }
                        if composition > 0 &&
                        (!(.STABLE in options) || !unsafe_get_property(composition).comp_exclusion) {
                            starter^ = composition
                            starter_property = nil
                            continue
                        }
                    }
                }
            }
            buf[wpos] = current_char
            if current_property.combining_class != 0 {
                if i32(current_property.combining_class) > max_combining_class {
                    max_combining_class = i32(current_property.combining_class)
                }
            } else {
                starter = &buf[wpos]
                starter_property = nil
                max_combining_class = -1
            }
            wpos += 1
        }
        length = wpos
    }
    return length
}

utf8proc_reencode :: proc(buffer: []i32, length: int, options: utf8proc_option_t) -> int {
    len := utf8proc_normalize_utf32(buffer, length, options)
    if len < 0 {
        return len
    }

    tmp := mem.slice_ptr((^u8)(&buffer[0]), length * 4)
    rpos, wpos := 0, 0
    if .CHARBOUND in options {
        for rpos < len {
            wpos += unsafe_encode_char(buffer[rpos], tmp[wpos:])
            rpos += 1
        }
    } else {
        for rpos < len {
            wpos += utf8proc_encode_char(buffer[rpos], tmp[wpos:])
            rpos += 1
        }
    }
    (^u8)(&buffer[0])^ = 0
    return wpos
}

utf8proc_map :: proc(str: []u8, strlen: int, options: utf8proc_option_t) -> (data: [dynamic]u8, err: int) {
    return utf8proc_map_custom(str, strlen, options, nil, nil)
}

utf8proc_map_custom :: proc(
str: []u8, strlen: int,
options: utf8proc_option_t,
custom_func: proc(i32, rawptr) -> i32,
custom_data: rawptr,
) -> (data: [dynamic]u8, err: int) {
    result := utf8proc_decompose_custom(str, strlen, nil, 0, options, custom_func, custom_data)
    if result < 0 {
        err = result
        return
    }
    // Allocate dynamic array for codepoints (as i32) and later reencode to bytes
    buf_dyn := make([dynamic]u8, result * size_of(i32) + 1, allocator = context.allocator)
    codepoints := mem.slice_ptr((^i32)(&buf_dyn[0]), result)
    result = utf8proc_decompose_custom(str, strlen, codepoints, result, options, custom_func, custom_data)
    if result < 0 {
        delete(buf_dyn)
        err = result
        return
    }
    byte_len := utf8proc_reencode(codepoints, result, options)
    if byte_len < 0 {
        delete(buf_dyn)
        err = byte_len
        return
    }
    resize(&buf_dyn, byte_len)
    return buf_dyn, byte_len
}

utf8proc_NFD :: proc(str: []u8) -> [dynamic]u8 {
    data, _ := utf8proc_map(str, 0, { .NULLTERM, .STABLE, .DECOMPOSE })
    return data
}
utf8proc_NFC :: proc(str: []u8) -> [dynamic]u8 {
    data, _ := utf8proc_map(str, 0, { .NULLTERM, .STABLE, .COMPOSE })
    return data
}
utf8proc_NFKD :: proc(str: []u8) -> [dynamic]u8 {
    data, _ := utf8proc_map(str, 0, { .NULLTERM, .STABLE, .DECOMPOSE, .COMPAT })
    return data
}
utf8proc_NFKC :: proc(str: []u8) -> [dynamic]u8 {
    data, _ := utf8proc_map(str, 0, { .NULLTERM, .STABLE, .COMPOSE, .COMPAT })
    return data
}

// ---------------------------------------------------------------------------
// Character classification helpers (originally in unicode.i.cpp)
// ---------------------------------------------------------------------------
rune_is_letter :: proc(r: rune) -> bool {
    if r < 0x80 {
        if r == '_' {
            return true
        }
        return (u32(r) | 0x20) - 0x61 < 26
    }
    #partial switch utf8proc_category(i32(r)) {
    case .LU, .LL, .LT, .LM, .LO: return true
    case: return false
    }
}
rune_is_digit :: proc(r: rune) -> bool {
    if r < 0x80 {
        return u32(r) - '0' < 10
    }
    return utf8proc_category(i32(r)) == .ND
}
rune_is_letter_or_digit :: proc(r: rune) -> bool {
    if r < 0x80 {
        if r == '_' {
            return true
        }
        if (u32(r) | 0x20) - 0x61 < 26 {
            return true
        }
        return u32(r) - '0' < 10
    }
    #partial switch utf8proc_category(i32(r)) {
    case .LU, .LL, .LT, .LM, .LO: return true
    case .ND: return true
    }
    return false
}
rune_is_whitespace :: proc(r: rune) -> bool {
    switch r {
    case ' ', '\t', '\n', '\r': return true
    }
    return false
}

// ---------------------------------------------------------------------------
// UTF-8 decoder using the table from original code
// ---------------------------------------------------------------------------
@(private)
global__utf8_first := [256]u8{
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0, 0xf0,
    0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
    0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
    0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
    0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
    0xf1, 0xf1, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
    0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02,
    0x13, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x03, 0x23, 0x03, 0x03,
    0x34, 0x04, 0x04, 0x04, 0x44, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1, 0xf1,
}
@(private)
Utf8AcceptRange :: struct {
    lo, hi: u8
}
@(private)
global__utf8_accept_ranges := [?]Utf8AcceptRange{
    { 0x80, 0xbf }, { 0xa0, 0xbf }, { 0x80, 0x9f }, { 0x90, 0xbf }, { 0x80, 0x8f },
}

utf8_decode :: proc(str: []u8, str_len: int, codepoint_out: ^rune) -> (width: int) {
    codepoint, invalid := rune(0xfffd), false
    width = 0
    if str_len <= 0 {
        if codepoint_out != nil {
            codepoint_out^ = codepoint
        }
        return width
    }

    for {
    // 用 for-break 模拟 C 的 goto end/invalid 结构
        s0 := str[0]
        x := global__utf8_first[s0]
        if x >= 0xf0 {
            mask := -rune(x & 1)
            codepoint = (rune(s0) & (~mask)) | (rune(0xfffd) & mask)
            width = 1
            break
        }
        if s0 < 0x80 {
            codepoint = rune(s0)
            width = 1
            break
        }
        sz := u8(x & 7)
        accept := global__utf8_accept_ranges[x >> 4]
        if str_len < int(sz) {
            invalid = true; break
        } // invalid
        b1 := str[1]
        if b1 < accept.lo || accept.hi < b1 {
            invalid = true; break
        }// invalid
        if sz == 2 {
            codepoint = (rune(s0) & 0x1f) << 6 | (rune(b1) & 0x3f)
            width = 2
            break
        }
        b2 := str[2]
        if !(0x80 <= b2 && b2 <= 0xbf) {
            invalid = true; break
        } // invalid
        if sz == 3 {
            codepoint = (rune(s0) & 0x1f) << 12 | (rune(b1) & 0x3f) << 6 | (rune(b2) & 0x3f)
            width = 3
            break
        }
        b3 := str[3]
        if !(0x80 <= b3 && b3 <= 0xbf) {
            invalid = true; break
        } // invalid
        codepoint = (rune(s0) & 0x07) << 18 | (rune(b1) & 0x3f) << 12 | (rune(b2) & 0x3f) << 6 | (rune(b3) & 0x3f)
        width = 4
        break
    }
    if invalid {
        codepoint = rune(0xfffd)
        width = 1
    }
    if codepoint_out != nil {
        codepoint_out^ = codepoint
    }
    return width
}

// ---------------------------------------------------------------------------
// ucg (Unicode Grapheme Cluster) routines
// ---------------------------------------------------------------------------
ucg_binary_search :: proc(value: i32, table: []i32, length: int, stride: int) -> int {
    assert(table != nil)
    assert(length > 0)
    assert(stride > 0)
    n := length
    t := 0
    for n > 1 {
        m := n / 2
        p := t + m * stride
        if value >= table[p] {
            t = p
            n -= m
        } else {
            n = m
        }
    }
    if n != 0 && value >= table[t] {
        return t
    }
    return -1
}

ucg_is_control :: proc(r: i32) -> bool {
    return r <= 0x1F || (0x7F <= r && r <= 0x9F)
}
ucg_is_emoji_modifier :: proc(r: i32) -> bool {
    return 0x1F3FB <= r && r <= 0x1F3FF
}
ucg_is_regional_indicator :: proc(r: i32) -> bool {
    return 0x1F1E6 <= r && r <= 0x1F1FF
}
ucg_is_enclosing_mark :: proc(r: i32) -> bool {
    switch r {
    case 0x0488, 0x0489, 0x1ABE: return true
    }
    if 0x20DD <= r && r <= 0x20E0 {
        return true
    }
    if 0x20E2 <= r && r <= 0x20E4 {
        return true
    }
    if 0xA670 <= r && r <= 0xA672 {
        return true
    }
    return false
}
ucg_is_prepended_concatenation_mark :: proc(r: i32) -> bool {
    switch r {
    case 0x006DD, 0x0070F, 0x008E2, 0x110BD, 0x110CD: return true
    }
    if 0x00600 <= r && r <= 0x00605 {
        return true
    }
    if 0x00890 <= r && r <= 0x00891 {
        return true
    }
    return false
}
ucg_is_spacing_mark :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_spacing_mark_ranges[:], len(ucg_spacing_mark_ranges) / 2, 2)
    if p >= 0 && ucg_spacing_mark_ranges[p] <= r && r <= ucg_spacing_mark_ranges[p + 1] {
        return true
    }
    return false
}
ucg_is_nonspacing_mark :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_nonspacing_mark_ranges[:], len(ucg_nonspacing_mark_ranges) / 2, 2)
    if p >= 0 && ucg_nonspacing_mark_ranges[p] <= r && r <= ucg_nonspacing_mark_ranges[p + 1] {
        return true
    }
    return false
}
ucg_is_emoji_extended_pictographic :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_emoji_extended_pictographic_ranges[:], len(ucg_emoji_extended_pictographic_ranges) / 2, 2)
    if p >= 0 && ucg_emoji_extended_pictographic_ranges[p] <= r && r <= ucg_emoji_extended_pictographic_ranges[p + 1] {
        return true
    }
    return false
}
ucg_is_grapheme_extend :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_grapheme_extend_ranges[:], len(ucg_grapheme_extend_ranges) / 2, 2)
    if p >= 0 && ucg_grapheme_extend_ranges[p] <= r && r <= ucg_grapheme_extend_ranges[p + 1] {
        return true
    }
    return false
}
ucg_is_hangul_syllable_leading :: proc(r: i32) -> bool {
    return (0x1100 <= r && r <= 0x115F) || (0xA960 <= r && r <= 0xA97C)
}
ucg_is_hangul_syllable_vowel :: proc(r: i32) -> bool {
    return (0x1160 <= r && r <= 0x11A7) || (0xD7B0 <= r && r <= 0xD7C6)
}
ucg_is_hangul_syllable_trailing :: proc(r: i32) -> bool {
    return (0x11A8 <= r && r <= 0x11FF) || (0xD7CB <= r && r <= 0xD7FB)
}
ucg_is_hangul_syllable_lv :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_hangul_syllable_lv_singlets[:], len(ucg_hangul_syllable_lv_singlets), 1)
    if p >= 0 && r == ucg_hangul_syllable_lv_singlets[p] {
        return true
    }
    return false
}
ucg_is_hangul_syllable_lvt :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_hangul_syllable_lvt_ranges[:], len(ucg_hangul_syllable_lvt_ranges) / 2, 2)
    if p >= 0 && ucg_hangul_syllable_lvt_ranges[p] <= r && r <= ucg_hangul_syllable_lvt_ranges[p + 1] {
        return true
    }
    return false
}
ucg_is_indic_consonant_preceding_repha :: proc(r: i32) -> bool {
    switch r {
    case 0x00D4E, 0x11941, 0x11D46, 0x11F02: return true
    }
    return false
}
ucg_is_indic_consonant_prefixed :: proc(r: i32) -> bool {
    switch r {
    case 0x1193F, 0x11A3A: return true
    }
    if 0x111C2 <= r && r <= 0x111C3 {
        return true
    }
    if 0x11A84 <= r && r <= 0x11A89 {
        return true
    }
    return false
}
ucg_is_indic_conjunct_break_linker :: proc(r: i32) -> bool {
    switch r {
    case 0x094D, 0x09CD, 0x0ACD, 0x0B4D, 0x0C4D, 0x0D4D: return true
    }
    return false
}
ucg_is_indic_conjunct_break_consonant :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_indic_conjunct_break_consonant_ranges[:], len(ucg_indic_conjunct_break_consonant_ranges) / 2, 2)
    if p >= 0 && ucg_indic_conjunct_break_consonant_ranges[p] <= r && r <= ucg_indic_conjunct_break_consonant_ranges[p + 1] {
        return true
    }
    return false
}
ucg_is_indic_conjunct_break_extend :: proc(r: i32) -> bool {
    p := ucg_binary_search(r, ucg_indic_conjunct_break_extend_ranges[:], len(ucg_indic_conjunct_break_extend_ranges) / 2, 2)
    if p >= 0 && ucg_indic_conjunct_break_extend_ranges[p] <= r && r <= ucg_indic_conjunct_break_extend_ranges[p + 1] {
        return true
    }
    return false
}
ucg_is_gcb_prepend_class :: proc(r: i32) -> bool {
    return ucg_is_indic_consonant_preceding_repha(r) || ucg_is_indic_consonant_prefixed(r) || ucg_is_prepended_concatenation_mark(r)
}
ucg_is_gcb_extend_class :: proc(r: i32) -> bool {
    return ucg_is_grapheme_extend(r) || ucg_is_emoji_modifier(r)
}

ucg_normalized_east_asian_width :: proc(r: i32) -> int {
    if ucg_is_control(r) {
        return 0
    }
    if r <= 0x10FF {
        return 1
    }
    switch r {
    case 0xFEFF, 0x200B, 0x200C, 0x200D, 0x2060: return 0
    }
    len_ := len(ucg_normalized_east_asian_width_ranges)
    p := ucg_binary_search(r, ucg_normalized_east_asian_width_ranges[:], len_ / 3, 3)
    if p >= 0 && ucg_normalized_east_asian_width_ranges[p] <= r && r <= ucg_normalized_east_asian_width_ranges[p + 1] {
        return int(ucg_normalized_east_asian_width_ranges[p + 2])
    }
    return 1
}

// ---------------------------------------------------------------------------
// Grapheme cluster decoder state and helpers
// ---------------------------------------------------------------------------
Grapheme_Cluster_Sequence :: enum {
    None, Indic, Emoji, Regional
}

ucg_grapheme :: struct {
    byte_index: i32,
    rune_index: i32,
    width: i32,
}

ucg_decoder_state :: struct {
    graphemes: ^[dynamic]ucg_grapheme,
    rune_count: i32,
    grapheme_count: i32,
    width: i32,
    last_rune: i32,
    last_rune_breaks_forward: bool,
    last_width: i32,
    last_grapheme_count: i32,
    bypass_next_rune: bool,
    regional_indicator_counter: int,
    current_sequence: Grapheme_Cluster_Sequence,
    continue_sequence: bool,
}

_ucg_decode_grapheme_clusters_deferred_step :: proc(state: ^ucg_decoder_state, byte_index: i32, this_rune: i32) {
    if state.rune_count == 0 && state.grapheme_count == 0 {
        state.grapheme_count += 1
    }
    if state.grapheme_count > state.last_grapheme_count {
        state.width += i32(ucg_normalized_east_asian_width(this_rune))
        // Append new grapheme to dynamic array
        append(state.graphemes, ucg_grapheme{
            byte_index = byte_index,
            rune_index = state.rune_count,
            width      = state.width - state.last_width,
        })
        state.last_grapheme_count = i32(len(state.graphemes))
        state.last_width = state.width
    }
    state.last_rune = this_rune
    state.rune_count += 1
    if !state.continue_sequence {
        state.current_sequence = .None
        state.regional_indicator_counter = 0
    }
    state.continue_sequence = false
}

ucg_decode_grapheme_clusters :: proc(
str: []u8, str_len: int,
) -> (graphemes: [dynamic]ucg_grapheme, rune_count, grapheme_count, width: i32, err: int) {
    state: ucg_decoder_state
    dyn_arr := make([dynamic]ucg_grapheme, allocator = context.allocator)
    state.graphemes = &dyn_arr

    for byte_index : i32 = 0; byte_index < i32(str_len); {
        this_rune: rune
        bytes_advanced := utf8_decode(str[byte_index:], str_len - int(byte_index), &this_rune)
        if this_rune == 0xfffd || bytes_advanced == 0 {
            return dyn_arr, state.rune_count, i32(len(dyn_arr)), state.width, -1
        }
        r := i32(this_rune)
        // CRLF handling
        if r == '\n' && state.last_rune == '\r' {
            state.last_rune_breaks_forward = false
            state.bypass_next_rune = false
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        if ucg_is_control(r) {
            state.grapheme_count += 1
            state.last_rune_breaks_forward = true
            state.bypass_next_rune = true
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        if state.bypass_next_rune {
            if state.last_rune_breaks_forward {
                state.grapheme_count += 1
                state.last_rune_breaks_forward = false
            }
            state.bypass_next_rune = false
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        if r != 0xA9 && r != 0xAE && r <= 0x2FF {
            state.grapheme_count += 1
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        // Hangul
        if 0x1100 <= r && r <= 0xD7FB {
            if ucg_is_hangul_syllable_leading(r) || ucg_is_hangul_syllable_lv(r) || ucg_is_hangul_syllable_lvt(r) {
                if !ucg_is_hangul_syllable_leading(state.last_rune) {
                    state.grapheme_count += 1
                }
                _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                byte_index += i32(bytes_advanced)
                continue
            }
            if ucg_is_hangul_syllable_vowel(r) {
                if ucg_is_hangul_syllable_leading(state.last_rune) || ucg_is_hangul_syllable_vowel(state.last_rune) || ucg_is_hangul_syllable_lv(state.last_rune) {
                    _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                    byte_index += i32(bytes_advanced)
                    continue
                }
                state.grapheme_count += 1
                _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                byte_index += i32(bytes_advanced)
                continue
            }
            if ucg_is_hangul_syllable_trailing(r) {
                if ucg_is_hangul_syllable_trailing(state.last_rune) || ucg_is_hangul_syllable_lvt(state.last_rune) || ucg_is_hangul_syllable_lv(state.last_rune) || ucg_is_hangul_syllable_vowel(state.last_rune) {
                    _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                    byte_index += i32(bytes_advanced)
                    continue
                }
                state.grapheme_count += 1
                _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                byte_index += i32(bytes_advanced)
                continue
            }
        }
        // ZWJ
        if r == 0x200D {
            state.continue_sequence = true
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        // Extend class
        if ucg_is_gcb_extend_class(r) {
            if state.current_sequence == .Indic {
                if ucg_is_indic_conjunct_break_extend(r) && (ucg_is_indic_conjunct_break_linker(state.last_rune) || ucg_is_indic_conjunct_break_consonant(state.last_rune)) {
                    state.continue_sequence = true
                    _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                    byte_index += i32(bytes_advanced)
                    continue
                }
                if ucg_is_indic_conjunct_break_linker(r) && (ucg_is_indic_conjunct_break_linker(state.last_rune) || ucg_is_indic_conjunct_break_extend(state.last_rune) || ucg_is_indic_conjunct_break_consonant(state.last_rune)) {
                    state.continue_sequence = true
                    _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                    byte_index += i32(bytes_advanced)
                    continue
                }
                _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
                byte_index += i32(bytes_advanced)
                continue
            }
            if state.current_sequence == .Emoji && (ucg_is_gcb_extend_class(state.last_rune) || ucg_is_emoji_extended_pictographic(state.last_rune)) {
                state.continue_sequence = true
            }
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        if ucg_is_spacing_mark(r) {
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        // Prepend class
        if ucg_is_gcb_prepend_class(r) {
            state.grapheme_count += 1
            state.bypass_next_rune = true
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        // Indic consonant
        if ucg_is_indic_conjunct_break_consonant(r) {
            if state.current_sequence == .Indic {
                if state.last_rune == 0x200D || ucg_is_indic_conjunct_break_linker(state.last_rune) {
                    state.continue_sequence = true
                } else {
                    state.grapheme_count += 1
                }
            } else {
                state.grapheme_count += 1
                state.current_sequence = .Indic
                state.continue_sequence = true
            }
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        if ucg_is_indic_conjunct_break_extend(r) {
            if state.current_sequence == .Indic {
                if ucg_is_indic_conjunct_break_consonant(state.last_rune) || ucg_is_indic_conjunct_break_linker(state.last_rune) {
                    state.continue_sequence = true
                } else {
                    state.grapheme_count += 1
                }
            }
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        if ucg_is_indic_conjunct_break_linker(r) {
            if state.current_sequence == .Indic {
                if ucg_is_indic_conjunct_break_extend(state.last_rune) || ucg_is_indic_conjunct_break_linker(state.last_rune) {
                    state.continue_sequence = true
                } else {
                    state.grapheme_count += 1
                }
            }
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        // Emoji extended pictographic
        if ucg_is_emoji_extended_pictographic(r) {
            if state.current_sequence != .Emoji || state.last_rune != 0x200D {
                state.grapheme_count += 1
            }
            state.current_sequence = .Emoji
            state.continue_sequence = true
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        // Regional indicator
        if ucg_is_regional_indicator(r) {
            if (state.regional_indicator_counter & 1) == 0 {
                state.grapheme_count += 1
            }
            state.current_sequence = .Regional
            state.continue_sequence = true
            state.regional_indicator_counter += 1
            _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
            byte_index += i32(bytes_advanced)
            continue
        }
        // Fallback
        state.grapheme_count += 1
        _ucg_decode_grapheme_clusters_deferred_step(&state, byte_index, r)
        byte_index += i32(bytes_advanced)
    }

    return dyn_arr, state.rune_count, i32(len(dyn_arr)), state.width, 0
}