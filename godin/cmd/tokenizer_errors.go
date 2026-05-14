package cmd

import (
	"sort"
	"unicode/utf8"
	"unsafe"
)

const (
	showErrorLineMaxLineLength     = 80
	showErrorLineMaxTabWidth       = 8
	showErrorLineEllipsisPadding   = 8
	showErrorLineMinLeftView       = 8
	showErrorLineMaxInsertedWidth  = showErrorLineMaxTabWidth + showErrorLineEllipsisPadding
	showErrorLineMaxLineLengthPadded = showErrorLineMaxLineLength - showErrorLineMaxInsertedWidth
)

func push_error_value(pos TokenPos, kind ...ErrorValueKind) {
	if globalErrorCollector.CurrErrorValueSet.Load() {
		gb_assert_handler("Assertion Failure", "globalErrorCollector.CurrErrorValueSet false", "cmd_tokenizer_errors.go", 0)
	}
	k := ErrorValueError
	if len(kind) > 0 {
		k = kind[0]
	}
	ev := ErrorValue{Kind: k, Pos: pos, End: TokenPos{}, Msg: nil, SeenNewline: false}
	globalErrorCollector.CurrErrorValue = ev
	globalErrorCollector.CurrErrorValueSet.Store(true)
}

func pop_error_value() {
	mutex_lock(&globalErrorCollector.Mutex)
	if globalErrorCollector.CurrErrorValueSet.Load() {
		globalErrorCollector.ErrorValues = append(globalErrorCollector.ErrorValues, globalErrorCollector.CurrErrorValue)
		globalErrorCollector.CurrErrorValue = ErrorValue{}
		globalErrorCollector.CurrErrorValueSet.Store(false)
	}
	mutex_unlock(&globalErrorCollector.Mutex)
}

func try_pop_error_value() {
	if !globalErrorCollector.InBlock.Load() {
		pop_error_value()
	}
}

func get_error_value() *ErrorValue {
	if !globalErrorCollector.CurrErrorValueSet.Load() {
		gb_assert_handler("Assertion Failure", "globalErrorCollector.CurrErrorValueSet true", "cmd_tokenizer_errors.go", 0)
	}
	return &globalErrorCollector.CurrErrorValue
}

func any_errors() bool {
	return globalErrorCollector.Count.Load() != 0
}

func any_warnings() bool {
	return globalErrorCollector.WarningCount.Load() != 0
}

type ErrorOutProc func(format string, args ...any)

func default_error_out_va(format string, args ...any) {
	buf := gb_bprintf_va(format, args...)
	n := len(buf)
	if n > 0 {
		ev := get_error_value()
		if terse_errors() {
			for i := 0; i < n && !ev.SeenNewline; i++ {
				c := buf[i]
				if c == '\n' {
					ev.SeenNewline = true
				}
				ev.Msg = append(ev.Msg, byte(c))
			}
		} else {
			ev.Msg = append(ev.Msg, []byte(buf)...)
		}
	}
}

var errorOutVa ErrorOutProc = default_error_out_va

func begin_error_block() {
	mutex_lock(&globalErrorCollector.Mutex)
	globalErrorCollector.InBlock.Store(true)
}

func end_error_block() {
	pop_error_value()
	globalErrorCollector.InBlock.Store(false)
	mutex_unlock(&globalErrorCollector.Mutex)
}

func error_out(format string, args ...any) {
	errorOutVa(format, args...)
}

type TerminalStyle int32

const (
	TerminalStyleNormal   TerminalStyle = 0
	TerminalStyleBold     TerminalStyle = 1
	TerminalStyleUnderline TerminalStyle = 2
)

type TerminalColour int32

const (
	TerminalColourWhite  TerminalColour = 0
	TerminalColourRed    TerminalColour = 1
	TerminalColourYellow  TerminalColour = 2
	TerminalColourGreen  TerminalColour = 3
	TerminalColourCyan   TerminalColour = 4
	TerminalColourBlue   TerminalColour = 5
	TerminalColourPurple TerminalColour = 6
	TerminalColourBlack  TerminalColour = 7
	TerminalColourGrey   TerminalColour = 8
)

func terminal_set_colours(style TerminalStyle, foreground TerminalColour) {
	if has_ansi_terminal_colours() {
		ss := "0"
		switch style {
		case TerminalStyleNormal:
			ss = "0"
		case TerminalStyleBold:
			ss = "1"
		case TerminalStyleUnderline:
			ss = "4"
		}
		switch foreground {
		case TerminalColourWhite:
			error_out("\x1b[%s;37m", ss)
		case TerminalColourRed:
			error_out("\x1b[%s;31m", ss)
		case TerminalColourYellow:
			error_out("\x1b[%s;33m", ss)
		case TerminalColourGreen:
			error_out("\x1b[%s;32m", ss)
		case TerminalColourCyan:
			error_out("\x1b[%s;36m", ss)
		case TerminalColourBlue:
			error_out("\x1b[%s;34m", ss)
		case TerminalColourPurple:
			error_out("\x1b[%s;35m", ss)
		case TerminalColourBlack:
			error_out("\x1b[%s;30m", ss)
		case TerminalColourGrey:
			error_out("\x1b[%s;90m", ss)
		}
	}
}

func terminal_reset_colours() {
	if has_ansi_terminal_colours() {
		error_out("\x1b[0m")
	}
}

type ucgGrapheme struct {
	byteIndex int32
	runeIndex int32
	width     int32
}

func ucg_decode_grapheme_clusters(text string) (graphemes []ucgGrapheme, lineLengthRunes int32, lineLengthGraphemes int32, lineWidth int32) {
	b := []byte(text)
	graphemes = make([]ucgGrapheme, 0, len(b))
	var cur ucgGrapheme
	i := 0
	for i < len(b) {
		r, size := utf8.DecodeRune(b[i:])
		w := runeWidth(r)
		lineLengthRunes++
		if size == 0 {
			break
		}
		if w > 0 || (cur.width == 0 && cur.runeIndex == 0) {
			if cur.width > 0 || cur.runeIndex > 0 {
				graphemes = append(graphemes, cur)
			}
			cur = ucgGrapheme{
				byteIndex: int32(i),
				runeIndex: 1,
				width:     int32(w),
			}
		} else {
			cur.runeIndex++
		}
		i += size
	}
	if cur.width > 0 || cur.runeIndex > 0 {
		graphemes = append(graphemes, cur)
	}
	lineLengthGraphemes = int32(len(graphemes))
	for _, g := range graphemes {
		lineWidth += g.width
	}
	return
}

func runeWidth(r rune) int {
	if r == 0 || r == '\t' {
		return 1
	}
	if r < 0x20 || (r >= 0x7f && r < 0xa0) {
		return 0
	}
	if r >= 0x1100 &&
		(r <= 0x115f ||
			r == 0x2329 || r == 0x232a ||
			(r >= 0x2e80 && r <= 0x2e99) ||
			(r >= 0x2e9b && r <= 0x2ef3) ||
			(r >= 0x2f00 && r <= 0x2fd5) ||
			(r >= 0x2ff0 && r <= 0x2fff) ||
			(r >= 0x3000 && r <= 0x303e) ||
			(r >= 0x3041 && r <= 0x3096) ||
			(r >= 0x3099 && r <= 0x30ff) ||
			(r >= 0x3105 && r <= 0x312f) ||
			(r >= 0x3131 && r <= 0x318e) ||
			(r >= 0x3190 && r <= 0x31bf) ||
			(r >= 0x31c0 && r <= 0x31ef) ||
			(r >= 0x31f0 && r <= 0x321e) ||
			(r >= 0x3220 && r <= 0x3247) ||
			(r >= 0x3250 && r <= 0x4dbf) ||
			(r >= 0x4e00 && r <= 0xa48c) ||
			(r >= 0xa490 && r <= 0xa4c6) ||
			(r >= 0xa960 && r <= 0xa97c) ||
			(r >= 0xac00 && r <= 0xd7a3) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xfe10 && r <= 0xfe19) ||
			(r >= 0xfe30 && r <= 0xfe6b) ||
			(r >= 0xff01 && r <= 0xff60) ||
			(r >= 0xffe0 && r <= 0xffe6) ||
			(r >= 0x1b000 && r <= 0x1b0ff) ||
			(r >= 0x1b100 && r <= 0x1b12f) ||
			(r >= 0x1f004 && r <= 0x1f004) ||
			(r >= 0x1f0cf && r <= 0x1f0cf) ||
			(r >= 0x1f18e && r <= 0x1f18e) ||
			(r >= 0x1f191 && r <= 0x1f19a) ||
			(r >= 0x20000 && r <= 0x2fffd) ||
			(r >= 0x30000 && r <= 0x3fffd)) {
		return 2
	}
	return 1
}

func show_error_on_line(pos, end TokenPos) isize {
	get_error_value().End = end
	if !show_error_line() {
		return -1
	}

	var errorStartIndexBytes int32
	theLine := get_file_line_as_string(pos, &errorStartIndexBytes)
	if theLine == nil || gb_string_length(theLine) == 0 {
		terminal_set_colours(TerminalStyleNormal, TerminalColourGrey)
		error_out("\t( empty line )\n")
		terminal_reset_colours()
		if theLine == nil {
			return -1
		}
		return isize(errorStartIndexBytes)
	}
	defer gb_string_free(theLine)

	lineStr := goStr(make_string_c(theLine))
	lineLengthBytes := int32(gb_string_length(theLine))

	graphemes, lineLengthRunes, lineLengthGraphemes, lineWidth := ucg_decode_grapheme_clusters(lineStr)
	_ = lineLengthRunes

	if len(graphemes) == 0 || int32(len(graphemes)) < lineLengthGraphemes {
		graphemes = append(graphemes, ucgGrapheme{
			byteIndex: errorStartIndexBytes,
			runeIndex: lineLengthRunes,
			width:     1,
		})
		lineLengthGraphemes = int32(len(graphemes))
	}

	errorStartIndexGraphemes := int32(0)
	for i, g := range graphemes {
		if g.byteIndex == errorStartIndexBytes {
			errorStartIndexGraphemes = int32(i)
			break
		}
	}
	if errorStartIndexGraphemes == 0 && errorStartIndexBytes != 0 && lineLengthGraphemes != 0 {
		errorStartIndexGraphemes = lineLengthGraphemes
	}

	error_out("\t")

	showRightEllipsis := false
	squigglePadding := int32(0)
	windowOpenBytes := int32(0)
	windowCloseBytes := int32(0)

	if lineWidth > showErrorLineMaxLineLengthPadded {
		windowSizeLeft := int32(0)
		windowSizeRight := int32(0)
		windowOpenGraphemes := int32(0)

		for i := errorStartIndexGraphemes - 1; i > 0; i-- {
			windowSizeLeft += graphemes[i].width
			if windowSizeLeft >= showErrorLineMinLeftView {
				windowOpenGraphemes = i
				windowOpenBytes = graphemes[i].byteIndex
				break
			}
		}

		for i := errorStartIndexGraphemes; i < lineLengthGraphemes; i++ {
			windowSizeRight += graphemes[i].width
			if windowSizeRight >= showErrorLineMaxLineLengthPadded-showErrorLineMinLeftView {
				windowCloseBytes = graphemes[i].byteIndex
				break
			}
		}
		if windowCloseBytes == 0 {
			windowCloseBytes = lineLengthBytes
		}
		if windowSizeRight < showErrorLineMaxLineLengthPadded-showErrorLineMinLeftView {
			for i := windowOpenGraphemes - 1; i > 0; i-- {
				windowSizeLeft += graphemes[i].width
				if windowSizeLeft+windowSizeRight >= showErrorLineMaxLineLengthPadded {
					windowOpenGraphemes = i
					windowOpenBytes = graphemes[i].byteIndex
					break
				}
			}
		}

		if windowCloseBytes < windowOpenBytes {
			gb_assert_handler("Assertion Failure", "window_close_bytes >= window_open_bytes",
				"cmd_tokenizer_errors.go", 0,
				"Error line truncation window has wrong byte indices. (open, close: %i, %i)",
				windowOpenBytes, windowCloseBytes)
		}

		if windowCloseBytes != lineLengthBytes {
			showRightEllipsis = true
		}
		lineLengthBytes = windowCloseBytes
		lineStr = lineStr[windowOpenBytes:]
		lineLengthBytes -= windowOpenBytes

		if windowOpenBytes > 0 {
			error_out("... ")
			squigglePadding += 4
		}
	} else {
		windowOpenBytes = 0
		windowCloseBytes = lineLengthBytes
	}

	for i := errorStartIndexGraphemes; i > 0; i-- {
		if int32(i) < lineLengthGraphemes && graphemes[i].byteIndex == windowOpenBytes {
			break
		}
		squigglePadding += graphemes[i].width
	}

	terminal_set_colours(TerminalStyleNormal, TerminalColourWhite)
	if lineLengthBytes > 0 && int32(len(lineStr)) >= lineLengthBytes {
		error_out("%.*s", int(lineLengthBytes), lineStr[:lineLengthBytes])
	}

	squiggleLength := int32(0)
	trailingSquiggle := false
	if end.FileID == pos.FileID {
		if end.Line > pos.Line {
			showRightEllipsis = true
			for i := errorStartIndexGraphemes; i < lineLengthGraphemes; i++ {
				squiggleLength += graphemes[i].width
				trailingSquiggle = true
			}
		} else if end.Line == pos.Line && end.Column > pos.Column {
			adjustedEndIndex := graphemes[errorStartIndexGraphemes].byteIndex + end.Column - pos.Column
			for i := errorStartIndexGraphemes; i < lineLengthGraphemes; i++ {
				if graphemes[i].byteIndex >= adjustedEndIndex {
					break
				} else if graphemes[i].byteIndex >= windowCloseBytes {
					trailingSquiggle = true
					break
				}
				squiggleLength += graphemes[i].width
			}
		}
	} else {
		squiggleLength = 1
	}

	if showRightEllipsis {
		error_out(" ...")
	}
	error_out("\n\t")

	for i := squigglePadding; i > 0; i-- {
		error_out(" ")
	}

	terminal_set_colours(TerminalStyleBold, TerminalColourGreen)
	if squiggleLength > 0 {
		error_out("^")
		squiggleLength--
	}
	for ; squiggleLength > 1; squiggleLength-- {
		error_out("~")
	}
	if squiggleLength > 0 {
		if trailingSquiggle {
			error_out("~ ...")
		} else {
			error_out("^")
		}
	}
	error_out("\n")
	terminal_reset_colours()
	return isize(squigglePadding)
}

func error_out_empty() {
	error_out("")
}

func error_out_pos(pos TokenPos) {
	terminal_set_colours(TerminalStyleBold, TerminalColourWhite)
	error_out("%s ", token_pos_to_string(pos))
	terminal_reset_colours()
}

func error_out_coloured(str string, style TerminalStyle, foreground TerminalColour) {
	terminal_set_colours(style, foreground)
	error_out(str)
	terminal_reset_colours()
}

func error_va(pos, end TokenPos, format string, args ...any) {
	globalErrorCollector.Count.Add(1)
	mutex_lock(&globalErrorCollector.Mutex)

	if globalErrorCollector.Count.Load() > int64(MAX_ERROR_COLLECTOR_COUNT()) {
		print_all_errors()
		gb_exit(1)
	}

	push_error_value(pos, ErrorValueError)

	if pos.Line == 0 {
		error_out_empty()
		error_out_coloured("Error: ", TerminalStyleNormal, TerminalColourRed)
		errorOutVa(format, args...)
		error_out("\n")
	} else {
		if json_errors() {
			error_out_empty()
		} else {
			error_out_pos(pos)
			error_out_coloured("Error: ", TerminalStyleNormal, TerminalColourRed)
		}
		errorOutVa(format, args...)
		error_out("\n")
		show_error_on_line(pos, end)
	}

	try_pop_error_value()
	mutex_unlock(&globalErrorCollector.Mutex)
}

func warning_va(pos, end TokenPos, format string, args ...any) {
	if global_warnings_as_errors() {
		error_va(pos, end, format, args...)
		return
	}
	if global_ignore_warnings() {
		return
	}

	globalErrorCollector.WarningCount.Add(1)
	mutex_lock(&globalErrorCollector.Mutex)

	push_error_value(pos, ErrorValueWarning)

	if pos.Line == 0 {
		error_out_empty()
		error_out_coloured("Warning: ", TerminalStyleNormal, TerminalColourYellow)
		errorOutVa(format, args...)
		error_out("\n")
	} else {
		if json_errors() {
			error_out_empty()
		} else {
			error_out_pos(pos)
			error_out_coloured("Warning: ", TerminalStyleNormal, TerminalColourYellow)
		}
		errorOutVa(format, args...)
		error_out("\n")
		show_error_on_line(pos, end)
	}

	try_pop_error_value()
	mutex_unlock(&globalErrorCollector.Mutex)
}

func error_line_va(format string, args ...any) {
	errorOutVa(format, args...)
}

func error_no_newline_va(pos TokenPos, format string, args ...any) {
	globalErrorCollector.Count.Add(1)
	mutex_lock(&globalErrorCollector.Mutex)

	if globalErrorCollector.Count.Load() > int64(MAX_ERROR_COLLECTOR_COUNT()) {
		print_all_errors()
		gb_exit(1)
	}

	push_error_value(pos, ErrorValueError)

	if pos.Line == 0 {
		error_out_empty()
		error_out_coloured("Error: ", TerminalStyleNormal, TerminalColourRed)
		errorOutVa(format, args...)
	} else {
		if json_errors() {
			error_out_empty()
		} else {
			error_out_pos(pos)
		}
		if has_ansi_terminal_colours() {
			error_out_coloured("Error: ", TerminalStyleNormal, TerminalColourRed)
		}
		errorOutVa(format, args...)
	}

	try_pop_error_value()
	mutex_unlock(&globalErrorCollector.Mutex)
}

func syntax_error_va(pos, end TokenPos, format string, args ...any) {
	globalErrorCollector.Count.Add(1)
	mutex_lock(&globalErrorCollector.Mutex)

	if globalErrorCollector.Count.Load() > int64(MAX_ERROR_COLLECTOR_COUNT()) {
		print_all_errors()
		gb_exit(1)
	}

	push_error_value(pos, ErrorValueWarning)

	if pos.Line == 0 {
		error_out_empty()
		error_out_coloured("Syntax Error: ", TerminalStyleNormal, TerminalColourRed)
		errorOutVa(format, args...)
		error_out("\n")
	} else {
		if json_errors() {
			error_out_empty()
		} else {
			error_out_pos(pos)
		}
		error_out_coloured("Syntax Error: ", TerminalStyleNormal, TerminalColourRed)
		errorOutVa(format, args...)
		error_out("\n")
		show_error_on_line(pos, end)
	}

	try_pop_error_value()
	mutex_unlock(&globalErrorCollector.Mutex)
}

func syntax_error_with_verbose_va(pos, end TokenPos, format string, args ...any) {
	globalErrorCollector.Count.Add(1)
	mutex_lock(&globalErrorCollector.Mutex)

	if globalErrorCollector.Count.Load() > int64(MAX_ERROR_COLLECTOR_COUNT()) {
		print_all_errors()
		gb_exit(1)
	}

	push_error_value(pos, ErrorValueWarning)

	if pos.Line == 0 {
		error_out_empty()
		error_out_coloured("Syntax Error: ", TerminalStyleNormal, TerminalColourRed)
		errorOutVa(format, args...)
		error_out("\n")
	} else {
		if json_errors() {
			error_out_empty()
		} else {
			error_out_pos(pos)
		}
		if has_ansi_terminal_colours() {
			error_out_coloured("Syntax Error: ", TerminalStyleNormal, TerminalColourRed)
		}
		errorOutVa(format, args...)
		error_out("\n")
		show_error_on_line(pos, end)
	}

	try_pop_error_value()
	mutex_unlock(&globalErrorCollector.Mutex)
}

func syntax_warning_va(pos, end TokenPos, format string, args ...any) {
	if global_warnings_as_errors() {
		syntax_error_va(pos, end, format, args...)
		return
	}
	if global_ignore_warnings() {
		return
	}

	mutex_lock(&globalErrorCollector.Mutex)
	globalErrorCollector.WarningCount.Add(1)

	push_error_value(pos, ErrorValueWarning)

	if pos.Line == 0 {
		error_out_empty()
		error_out_coloured("Syntax Warning: ", TerminalStyleNormal, TerminalColourYellow)
		errorOutVa(format, args...)
		error_out("\n")
	} else {
		if json_errors() {
			error_out_empty()
		} else {
			error_out_pos(pos)
		}
		error_out_coloured("Syntax Warning: ", TerminalStyleNormal, TerminalColourYellow)
		errorOutVa(format, args...)
		error_out("\n")
	}

	try_pop_error_value()
	mutex_unlock(&globalErrorCollector.Mutex)
}

func warning(token Token, format string, args ...any) {
	warning_va(token.Pos, TokenPos{}, format, args...)
}

func error_pos(pos TokenPos, format string, args ...any) {
	error_va(pos, TokenPos{}, format, args...)
}

func error_line(format string, args ...any) {
	error_line_va(format, args...)
}

func syntax_error_pos(pos TokenPos, format string, args ...any) {
	syntax_error_va(pos, TokenPos{}, format, args...)
}

func syntax_warning(token Token, format string, args ...any) {
	syntax_warning_va(token.Pos, TokenPos{}, format, args...)
}

func syntax_error_with_verbose(pos, end TokenPos, format string, args ...any) {
	syntax_error_with_verbose_va(pos, end, format, args...)
}

func compiler_error(format string, args ...any) {
	if any_errors() || any_warnings() {
		print_all_errors()
	}
	gb_printf_err("Internal Compiler Error: %s\n", gb_bprintf_va(format, args...))
	gb_exit(1)
}

func exit_with_errors() {
	if any_errors() || any_warnings() {
		print_all_errors()
	}
	gb_exit(1)
}

func error_value_cmp(a, b unsafe.Pointer) int {
	x := (*ErrorValue)(a)
	y := (*ErrorValue)(b)
	return int(token_pos_cmp(x.Pos, y.Pos))
}

var errorArticleTable = [][2]String{
	{S("a "), S("bit_set literal")},
	{S("a "), S("constant declaration")},
	{S("a "), S("dynamic array literal")},
	{S("a "), S("map index")},
	{S("a "), S("map literal")},
	{S("a "), S("matrix literal")},
	{S("a "), S("polymorphic type argument")},
	{S("a "), S("procedure argument")},
	{S("a "), S("simd vector literal")},
	{S("a "), S("slice literal")},
	{S("a "), S("structure literal")},
	{S("a "), S("variable declaration")},
	{S("an "), S("'any' literal")},
	{S("an "), S("array literal")},
	{S("an "), S("enumerated array literal")},
}

func error_article(contextName String) String {
	for _, entry := range errorArticleTable {
		if contextName == entry[1] {
			return entry[0]
		}
	}
	return String{}
}

func print_all_errors() {
	if errorsAlreadyPrinted {
		if globalErrorCollector.WarningCount.Load() == int64(len(globalErrorCollector.ErrorValues)) {
			for i := range globalErrorCollector.ErrorValues {
				globalErrorCollector.ErrorValues[i].Msg = nil
			}
			globalErrorCollector.ErrorValues = globalErrorCollector.ErrorValues[:0]
			errorsAlreadyPrinted = false
		}
		return
	}

	if !any_errors() && !any_warnings() {
		gb_assert_handler("Assertion Failure", "any_errors() || any_warnings()", "cmd_tokenizer_errors.go", 0)
	}

	sort.Slice(globalErrorCollector.ErrorValues, func(i, j int) bool {
		return token_pos_cmp(globalErrorCollector.ErrorValues[i].Pos, globalErrorCollector.ErrorValues[j].Pos) < 0
	})

	{
		defaultLinesToSkip := isize(1)
		if show_error_line() {
			defaultLinesToSkip += 2
		}
		var prevEv *ErrorValue
		i := 0
		for i < len(globalErrorCollector.ErrorValues) {
			ev := &globalErrorCollector.ErrorValues[i]
			if prevEv != nil && prevEv.Pos == ev.Pos {
				it := String_Iterator{Str: String{Data: &ev.Msg[0], Len: isize(len(ev.Msg))}, Pos: 0}
				for linesToSkip := defaultLinesToSkip; linesToSkip > 0; linesToSkip-- {
					line := string_split_iterator(&it, '\n')
					if line.Len == 0 {
						break
					}
				}
				current := String{Data: &prevEv.Msg[0], Len: isize(len(prevEv.Msg))}
				addition := String{Data: (*byte)(unsafe.Add(unsafe.Pointer(it.Str.Data), it.Pos)), Len: it.Str.Len - it.Pos}
				if addition.Len > 0 && !string_contains_string(current, addition) {
					prevEv.Msg = append(prevEv.Msg, unsafe.Slice(addition.Data, addition.Len)...)
				}
				ev.Msg = nil
				globalErrorCollector.ErrorValues = append(globalErrorCollector.ErrorValues[:i], globalErrorCollector.ErrorValues[i+1:]...)
			} else {
				prevEv = ev
				i++
			}
		}
	}

	res := gb_string_make(heap_allocator(), "")
	defer gb_string_free(res)

	if json_errors() {
		res = gb_string_append_fmt(res, "{\n")
		res = gb_string_append_fmt(res, "\t\"error_count\": %d,\n", len(globalErrorCollector.ErrorValues))
		res = gb_string_append_fmt(res, "\t\"errors\": [\n")
		for i, ev := range globalErrorCollector.ErrorValues {
			res = gb_string_append_fmt(res, "\t\t{\n")
			res = gb_string_append_fmt(res, "\t\t\t\"type\": \"")
			if ev.Kind == ErrorValueWarning {
				res = gb_string_append_fmt(res, "warning")
			} else {
				res = gb_string_append_fmt(res, "error")
			}
			res = gb_string_append_fmt(res, "\",\n")
			if ev.Pos.FileID != 0 {
				res = gb_string_append_fmt(res, "\t\t\t\"pos\": {\n")
				res = gb_string_append_fmt(res, "\t\t\t\t\"file\": \"")
				file := get_file_path_string(ev.Pos.FileID)
				for k := isize(0); k < file.Len; k++ {
					res = escapeChar(res, *(*byte)(unsafe.Add(unsafe.Pointer(file.Data), k)))
				}
				res = gb_string_append_fmt(res, "\",\n")
				res = gb_string_append_fmt(res, "\t\t\t\t\"offset\": %d,\n", ev.Pos.Offset)
				res = gb_string_append_fmt(res, "\t\t\t\t\"line\": %d,\n", ev.Pos.Line)
				res = gb_string_append_fmt(res, "\t\t\t\t\"column\": %d,\n", ev.Pos.Column)
				endColumn := ev.End.Column
				if ev.End.Column < ev.Pos.Column {
					endColumn = ev.Pos.Column
				}
				res = gb_string_append_fmt(res, "\t\t\t\t\"end_column\": %d\n", endColumn)
				res = gb_string_append_fmt(res, "\t\t\t},\n")
			} else {
				res = gb_string_append_fmt(res, "\t\t\t\"pos\": null,\n")
			}
			res = gb_string_append_fmt(res, "\t\t\t\"msgs\": [\n")
			lines := split_lines_from_array(ev.Msg, heap_allocator())
			if len(lines) > 0 {
				res = gb_string_append_fmt(res, "\t\t\t\t\"")
				for j := isize(0); j < isize(len(lines)); j++ {
					line := lines[j]
					for k := isize(0); k < line.Len; k++ {
						c := *(*byte)(unsafe.Add(unsafe.Pointer(line.Data), k))
						res = escapeChar(res, c)
					}
					if j+1 < isize(len(lines)) {
						res = gb_string_append_fmt(res, "\",\n")
						res = gb_string_append_fmt(res, "\t\t\t\t\"")
					}
				}
				res = gb_string_append_fmt(res, "\"\n")
			}
			array_free(&lines)
			res = gb_string_append_fmt(res, "\t\t\t]\n")
			res = gb_string_append_fmt(res, "\t\t}")
			if i+1 != len(globalErrorCollector.ErrorValues) {
				res = gb_string_append_fmt(res, ",")
			}
			res = gb_string_append_fmt(res, "\n")
		}
		res = gb_string_append_fmt(res, "\t]\n")
		res = gb_string_append_fmt(res, "}\n")
	} else {
		for _, ev := range globalErrorCollector.ErrorValues {
			it := String_Iterator{Str: String{Data: &ev.Msg[0], Len: isize(len(ev.Msg))}, Pos: 0}
			for lineIdx := isize(0); ; lineIdx++ {
				line := string_split_iterator(&it, '\n')
				if line.Len == 0 {
					break
				}
				line = string_trim_trailing_whitespace(line)
				res = gb_string_append_length(res, line.Data, line.Len)
				res = gb_string_append_length(res, unsafe.StringData(" \n"), 2)
				if lineIdx == 0 && terse_errors() {
					break
				}
			}
		}
	}

	f := gb_file_get_standard(gbFileStandard_Error)
	gb_file_write(f, res, gb_string_length(res))
	errorsAlreadyPrinted = true
}

func escapeChar(res gbString, c u8) gbString {
	switch c {
	case '\n':
		return gb_string_append_length(res, unsafe.StringData("\\n"), 2)
	case '"':
		return gb_string_append_length(res, unsafe.StringData("\\\""), 2)
	case '\\':
		return gb_string_append_length(res, unsafe.StringData("\\\\"), 2)
	case '\b':
		return gb_string_append_length(res, unsafe.StringData("\\b"), 2)
	case '\f':
		return gb_string_append_length(res, unsafe.StringData("\\f"), 2)
	case '\r':
		return gb_string_append_length(res, unsafe.StringData("\\r"), 2)
	case '\t':
		return gb_string_append_length(res, unsafe.StringData("\\t"), 2)
	default:
		if c <= '\x1f' {
			return gb_string_append_fmt(res, "\\u%04x", c)
		}
		return gb_string_append_length(res, unsafe.StringData(string(rune(c))), 1)
	}
}
