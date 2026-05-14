from __future__ import print_function

REM = """ @CodeSection == @Batch 
@echo off
setlocal EnableDelayedExpansion

where /Q cl.exe || (
    set __VSCMD_ARG_NO_LOGO=1
    for /f "tokens=*" %%i in ('"C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe" -latest -products * -requires Microsoft.VisualStudio.Component.VC.Tools.x86.x64 -property installationPath') do set VS=%%i
    if "!VS!" equ "" (
        echo ERROR: MSVC installation not found
        exit /b 1
    )
    call "!VS!\Common7\Tools\vsdevcmd.bat" -arch=x64 -host_arch=x64 || exit /b 1
)

if "%VSCMD_ARG_TGT_ARCH%" neq "x64" (
    if "%ODIN_IGNORE_MSVC_CHECK%" == "" (
        echo ERROR: please run this from MSVC x64 native tools command prompt, 32-bit target is not supported!
        exit /b 1
    )
)

where /Q git.exe || goto skip_git_hash
if not exist .git\ goto skip_git_hash
for /f "tokens=1,2" %%i IN ('git show "--pretty=%%cd %%h" "--date=format:%%Y-%%m-%%d" --no-patch --no-notes HEAD') do (
    set CURR_DATE_TIME=%%i
    set GIT_SHA=%%j
)
if %ERRORLEVEL% equ 0 (
    goto have_git_hash_and_date
)
:skip_git_hash
for /f %%i in ('get-date') do (
    set CURR_DATE_TIME=%%i
    rem Don't set GIT_SHA
)

:have_git_hash_and_date
set curr_year=%CURR_DATE_TIME:~0,4%
set curr_month=%CURR_DATE_TIME:~5,2%
set curr_day=%CURR_DATE_TIME:~8,2%

:: Make sure this is a decent name and not generic
set exe_name=odin_tmp.exe

:: Debug = 0, Release = 1
if "%1" == "1" (
    set release_mode=1
) else if "%1" == "release" (
    set release_mode=1
) else (
    set release_mode=0
)

:: Normal = 0, CI Nightly = 1
if "%2" == "1" (
    set nightly=1
) else (
    set nightly=0
)

if %release_mode% equ 0 (
    set V1=%curr_year%
    set V2=%curr_month%
    set V3=%curr_day%
) else (
    set V1=%curr_year%
    set V2=%curr_month%
    set V3=0
)
set V4=0
set odin_version_full="%V1%.%V2%.%V3%.%V4%"
set odin_version_raw="dev-%V1%-%V2%"
set compiler_flags= -nologo -Oi -TP -fp:precise -Gm- -MP -FC -EHsc- -GR- -GF
rem Parse source code as utf-8 even on shift-jis and other codepages
rem See https://learn.microsoft.com/en-us/cpp/build/reference/utf-8-set-source-and-executable-character-sets-to-utf-8?view=msvc-170
set compiler_flags= %compiler_flags% /utf-8
set compiler_defines= -DODIN_VERSION_RAW=\"%odin_version_raw%\" -DGIT_SHA=\"%GIT_SHA%\"

rem fileversion is defined as {Major,Minor,Build,Private: u16} so a bit limited
set rc_flags=-nologo "-DGIT_SHA=%GIT_SHA% -DVP=dev-%V1%-%V2%:%GIT_SHA% nologo -DV1=%V1% -DV2=%V2% -DV3=%V3% -DV4=%V4% -DVF=%odin_version_full% -DNIGHTLY=%nightly%"

if %nightly% equ 1 set compiler_defines=%compiler_defines% -DNIGHTLY

if %release_mode% EQU 0 ( rem Debug
    set compiler_flags=%compiler_flags% -Od -MDd -Z7
    set rc_flags=%rc_flags% -D_DEBUG
) else ( rem Release
    set compiler_flags=%compiler_flags% -O2 -MT -Z7
    set compiler_defines=%compiler_defines% -DNO_ARRAY_BOUNDS_CHECK
)

set compiler_warnings=-W4 -WX -wd4003 -wd4100 -wd4101 -wd4127 -wd4146 -wd4324 -wd4505 -wd4456 -wd4457

set compiler_includes=/Isrc\
set libs=kernel32.lib Synchronization.lib bin\llvm\windows\LLVM-C.lib
set odin_res=misc\odin.res
set odin_rc=misc\odin.rc

set linker_flags= -incremental:no -opt:ref -subsystem:console -MANIFEST:EMBED

if %release_mode% EQU 0 ( rem Debug
    set linker_flags=%linker_flags% -debug /NATVIS:src\odin_compiler.natvis
) else ( rem Release
    set linker_flags=%linker_flags% -debug
)

set compiler_settings=%compiler_includes% %compiler_flags% %compiler_warnings% %compiler_defines%
set linker_settings=%libs% %odin_res% %linker_flags%

SET rc_flags
SET compiler_settings
SET linker_settings
SET CURR_DATE_TIME
ECHO. & ECHO ========================================

ECHO.
ECHO $^>_ rc %rc_flags% %odin_rc%
ECHO.
ECHO $^>_ cl %compiler_settings% "src\main.cpp" "src\libtommath.cpp" /link %linker_settings% -OUT:%exe_name%
ECHO.
ECHO $^>_ mt -nologo -inputresource:%exe_name%;#1 -manifest misc\odin.manifest -outputresource:%exe_name%;#1 -validate_manifest -identity:"odin, processorArchitecture=amd64, version=%odin_version_full%, type=win32"
ECHO. & ECHO ========================================

ECHO.
ECHO $^>_ cl %compiler_settings% /P "src\libtommath.cpp" /Fi"src\libtommath.i"
cl %compiler_settings% /P "src\libtommath.cpp" /Fi"src\libtommath.i"
ECHO.
ECHO $^>_ cl %compiler_settings% /P "src\main.cpp" /Fi"src\main.i"
cl %compiler_settings% /P "src\main.cpp" /Fi"src\main.i"
ECHO. & ECHO ========================================

for /f "delims=" %%I in ('python2 "%~f0" clean_ipp "src\libtommath.i" ') do set "%%I" 
for /f "delims=" %%I in ('python2 "%~f0" clean_ipp "src\main.i" ') do set "%%I" 

ECHO. 
ECHO. & ECHO ========================================
ECHO $^>_ cl %compiler_settings% "src\cipp\main.i.cpp" "src\cipp\libtommath.i.cpp" /link %linker_settings% -OUT:%exe_name%
ECHO.

REM goto end_of_build

del *.pdb > NUL 2> NUL
del *.ilk > NUL 2> NUL

rc %rc_flags% %odin_rc%
cl %compiler_settings% "src\cipp\main.i.cpp" "src\cipp\libtommath.i.cpp" /link %linker_settings% -OUT:%exe_name%
if %errorlevel% neq 0 goto end_of_build
mt -nologo -inputresource:%exe_name%;#1 -manifest misc\odin.manifest -outputresource:%exe_name%;#1 -validate_manifest -identity:"odin, processorArchitecture=amd64, version=%odin_version_full%, type=win32"
if %errorlevel% neq 0 goto end_of_build

REM call build_vendor.bat
REM if %errorlevel% neq 0 goto end_of_build

rem If the demo doesn't run for you and your CPU is more than a decade old, try -microarch:native
if %release_mode% EQU 0 odin run examples/demo -vet -strict-style -resource:examples/demo/demo.rc -- Hellope World

rem Many non-compiler devs seem to run debug build but don't realize.
if %release_mode% EQU 0 echo: & echo Debug compiler built. Note: run "build.bat release" if you want a faster, release mode compiler.

del *.obj > NUL 2> NUL

:end_of_build
goto :EOF

#ifdef S_MP_DIV_RECURSIVE_C
    #error "S_MP_DIV_RECURSIVE_C is defined"
#else
    #error "S_MP_DIV_RECURSIVE_C is NOT defined"
#endif
#error "Compilation stopped"
:"""

# -*- coding: utf-8 -*-
import os
import re
import sys
import json
from StringIO import StringIO
from contextlib import closing


# region  ----------- HELP FUNC -----------

def fix_and_get_lines(in_file, tmp_file):
	if os.path.isfile(tmp_file):
		with open(tmp_file, 'r') as rf:
			text = rf.read()
	else:
		with open(in_file, 'r') as rf:
			text = rf.read()

		text = text.replace('\x0c', '\n').replace('\r\n', '\n').replace('\r', '\n')
		text = re.sub(r'[ \t\x0c]+$', '', text, flags=re.MULTILINE)
		text = re.sub(r'\n{2,}', '\n', text).strip('\n')
		text = re.sub(r'^[ \t]+(#line.*?)$', r'\1', text, flags=re.MULTILINE)

		with open(tmp_file, 'w') as wf:
			wf.write(text)

	lines = text.split('\n')
	## prefix = [i for i in lines if '#line' in i and i != i.lstrip()]
	## last = [i for i in lines if i != i.rstrip()]
	return lines


def _listdir(str_dir, filter_func, skips):
	path_files, current_files = {}, os.listdir(str_dir)
	for file_name in current_files:
		if file_name == '.' or file_name == '..':
			continue
		full_name = os.path.join(str_dir, file_name)
		if os.path.isfile(full_name):
			if filter_func(full_name, file_name):
				path_files.setdefault(full_name, file_name)
		elif os.path.isdir(full_name) and file_name not in skips \
				and not file_name.startswith('.') and not file_name.startswith('$'):
			next_files = _listdir(full_name, filter_func, skips)
			for n_full_name, n_file_name in next_files.items():
				path_files.setdefault(n_full_name, n_file_name)

	return path_files


def list_dir_by_name(str_dir, filter_func=None, encoding=None, skips=None):
	skips = {'System Volume Information', '$RECYCLE.BIN'} if skips is None else skips
	_str_dir = str_dir.encode(encoding, 'ignore') if isinstance(str_dir, unicode) and encoding else str_dir

	if not os.path.isdir(_str_dir):
		return {}

	filter_func = filter_func if hasattr(filter_func, '__call__') else lambda f, n: True
	path_files, name_files, _path_files = {}, {}, _listdir(_str_dir, filter_func, skips)
	for n_full_name, n_file_name in _path_files.items():
		stat = os.stat(n_full_name)
		_full_name = n_full_name.decode(encoding) if encoding else n_full_name
		_file_name = n_file_name.decode(encoding) if encoding else n_file_name
		path_files[_full_name] = [_file_name, stat]
		name_files[_file_name] = [_full_name, stat]
	return path_files, name_files


def dump_cleaned_cpp(in_file, out_file, lines, name_files=None):
	include = [
		'<windows.h>', '<string.h>', '<wchar.h>', '<psapi.h>',
		'<stdio.h>', '<math.h>', '<intrin.h>', '<atomic>', '<stdlib.h>',
		'gb/gb.h', 'utf8proc/utf8proc.c', 'ucg/ucg.c',
		'llvm-c/DataTypes.h',
		'llvm-c/ExternC.h',
		'llvm-c/Types.h',
		'llvm-c/Core.h',
		'llvm-c/ExecutionEngine.h',
		'llvm-c/Analysis.h',
		'llvm-c/Object.h',
		'llvm-c/BitWriter.h',
		'llvm-c/DebugInfo.h',
		'llvm-c/Transforms/PassBuilder.h', ] if 'main.' in in_file else [
		'<stddef.h>', '<stdint.h>', '<stdbool.h>', '<stdio.h>', '<limits.h>', '<stdarg.h>'
	]
	include = [i for i in include if i.startswith('<')]

	content, f_name = '', os.path.basename(out_file)
	with closing(StringIO()) as wf:
		wf.write("/* auto gen by ipp from %s */\n" % (os.path.basename(in_file),))
		wf.write("\n")
		[wf.write("#include " + ("%s\n" if i.startswith('<') else '"%s"\n') % (i,)) for i in include]

		wf.write("\n")
		wf.writelines(lines)
		content = wf.getvalue()

	if content and name_files and f_name in name_files and isinstance(name_files[f_name], (tuple, list)):
		s_size, f_size = len(content) + (content.count('\n') if os.name.startswith('nt') else 0), \
			name_files[f_name][1].st_size if name_files[f_name][1] else 0
		if f_size == s_size:
			return

	with open(out_file, 'w') as wf:
		wf.write("/* auto gen by ipp from %s */\n" % (os.path.basename(in_file),))
		wf.write("\n")
		[wf.write("#include " + ("%s\n" if i.startswith('<') else '"%s"\n') % (i,)) for i in include]

		wf.write("\n")
		wf.writelines(lines)


def dump_cleaned_part_cpp(in_file, out_file, lines, name_files=None):
	content, f_name = '', os.path.basename(out_file)
	with closing(StringIO()) as wf:
		wf.write("/* auto gen by ipp %s part of %s */\n" %
				 (os.path.basename(out_file), os.path.basename(in_file)))
		wf.write("\n")
		wf.writelines(lines)
		content = wf.getvalue()

	if content and name_files and f_name in name_files and isinstance(name_files[f_name], (tuple, list)):
		s_size, f_size = len(content) + (content.count('\n') if os.name.startswith('nt') else 0), \
			name_files[f_name][1].st_size if name_files[f_name][1] else 0
		if f_size == s_size:
			return

	with open(out_file, 'w') as wf:
		wf.write("/* auto gen by ipp %s part of %s */\n" %
				 (os.path.basename(out_file), os.path.basename(in_file)))
		wf.write("\n")
		wf.writelines(lines)


def _out_name(s):
	s = os.path.basename(s)
	if s.endswith('.i'):
		s = s.replace('.i', '.i.cpp')
	elif s.endswith('.c'):
		s = s.replace('.c', '.i.c')
	elif s.endswith('.h'):
		s = s.replace('.h', '.i.h')
	else:
		s = s.replace('.cpp', '.i.cpp').replace('.hpp', '.i.hpp')
	return s


# endregion

def clean_ipp(base, in_file, out_file=None, out_folder='cipp'):
	out_dir = os.path.join(os.path.dirname(os.path.abspath(in_file)), out_folder)
	in_file = os.path.abspath(in_file) if os.path.isfile(in_file) else os.path.join(base, in_file)
	out_file = str(out_file) if out_file else os.path.join(out_dir, _out_name(in_file))

	if not os.path.isfile(in_file): raise ValueError("file not found: " + in_file)
	if not os.path.isdir(out_dir): os.mkdir(out_dir)

	base_pre = (base + '/').replace('\\', r'\\').replace('/', r'\\')
	src_pre = (base.rstrip('/').rstrip('\\') + '/src') \
		.replace('\\', r'\\').replace('/', r'\\').split(r':\\', 1)[-1]
	file_pre = in_file.replace('.i', '.cpp') \
		.replace('\\', r'\\').replace('/', r'\\').split(r':\\', 1)[-1]

	part_map = {
		'common.cpp': [
			'gb.h', 'ucg_tables.h', 'utf8proc_data.c', 'unicode.cpp',
			'threading.cpp', 'common_memory.cpp', 'thread_pool.cpp', 'string.cpp',
			## 'array.cpp', 'queue.cpp', 'range_cache.cpp',
			'ptr_map.cpp', 'ptr_set.cpp', 'string_map.cpp', 'string16_map.cpp', 'string_set.cpp',
			## 'priority_queue.cpp', 'string_interner.cpp', 'path.cpp'
		]
		, 'checker.cpp': [
			'types.cpp',
			'check_expr.cpp', 'check_builtin.cpp', 'check_type.cpp', 'name_canonicalization.cpp',
			'check_decl.cpp', 'check_stmt.cpp',
		], 'llvm_backend.cpp': [
			'llvm_backend.hpp', 'llvm_abi.cpp', 'llvm_backend_opt.cpp', 'llvm_backend_general.cpp',
			'llvm_backend_debug.cpp', 'llvm_backend_const.cpp', 'llvm_backend_type.cpp',
			'llvm_backend_utility.cpp', 'llvm_backend_expr.cpp', 'llvm_backend_stmt.cpp',
			'llvm_backend_proc.cpp', 'llvm_backend_passes.cpp'
		], 'build_settings.cpp': ['build_settings_microarch.cpp'],
		'timings.cpp': [], 'cached.cpp': [], 'bundle_command.cpp': [], 'bug_report.cpp': [],
		'parser.cpp': [], 'tokenizer.cpp': [], 'docs.cpp': [],
		'linker.cpp': [], 'big_int.cpp': [], 'exact_value.cpp': [],
		'parser.hpp': [], 'checker.hpp': ['checker_builtin_procs.hpp'],
	} if 'main.' in in_file else {}

	all_files, name_files = list_dir_by_name(out_dir, lambda f, n: '.i.' in n)

	lines = fix_and_get_lines(in_file, out_file.replace('.cpp', '.tmp'))
	ret, sub_map = do_clean_ipp(base_pre, src_pre, file_pre, lines, part_map)

	if len(all_files) != len(name_files):
		dup = [v[0] for k, v in all_files.items() if k not in {f[0]:n for n, f in name_files.items()}]
		_LOG('list_dir not eq %d => %d :' % (len(all_files), len(name_files)))
		[_LOG('  %s :\n    %s\n' % (f, '\n    '.join(
			[k for k, v in all_files.items() if v[0] == f]))) for f in dup]
		return

	dump_cleaned_cpp(in_file, out_file, ret, name_files)

	for part_name, lines in sub_map.items():
		part_file = os.path.join(os.path.dirname(out_file), _out_name(part_name))
		dump_cleaned_part_cpp(in_file, part_file, lines, name_files)

	_LOG('done: ' + out_file)


# region  ----------- LINE FUNC -----------

def _build_lines(src_pre, file_pre, lines):
	lines_i = [None for _ in lines]
	kdx, len_, erase = 0, len(lines), False
	while kdx < len_:
		line_ = lines[kdx]
		if line_.startswith('#line'):
			aa = line_.split(" ", 2)
			ff, line_no = aa[2][1:-1].split(r':\\', 1)[-1], int(aa[1])
			df = ff.rsplit(r'\\', 1)[0]
			lines[kdx] = ''
			if df.startswith(src_pre):
				# noinspection PyTypeChecker
				lines_i[kdx] = (ff, line_no, line_)
				erase = False
			else:
				erase = True
		else:
			if erase: lines[kdx] = ''

		kdx += 1

	return lines_i


def _build_ff_map(file_pre, lines_i):
	def _update_ff_map(ff_, line_no_, kdx_):
		if ff_ in ff_map:
			if 0 <= ff_map[ff_][0] < kdx_: ff_map[ff_][0] = kdx_
			if line_no_ > ff_map[ff_][1]: ff_map[ff_][1] = line_no_
			if kdx_ > nn_map[ff_][1]: nn_map[ff_][1] = kdx_
		else:
			assert line_no_ == 1, "first line_no %d != 1 file: %s" % (line_no_, ff_)
			_nf_ = ff_.rsplit(r'\\', 1)[-1]
			ff_map[ff_] = [-1 if ff_ == file_pre else kdx_, line_no_, _nf_]
			nn_map[ff_] = [kdx_, kdx_]

	ff_map, nn_map = {}, {}
	for kdx, item in enumerate(lines_i, 0):
		if not item: continue
		ff, line_no, line_ = item
		_update_ff_map(ff, line_no, kdx)

	kdx, len_, info_map = 0, len(lines_i), {}
	for k, v in ff_map.items():
		_, line_max_, nf_ = v
		line_num, kdx = nn_map[k][1] - nn_map[k][0], nn_map[k][1]
		while kdx < len_:
			item = lines_i[kdx]
			kdx += 1

			if not item: continue
			ff, line_no, line_ = item
			kdx_max, line_max, nf = ff_map[ff] if ff in ff_map else (0, 0, '')
			if nf_ != nf: break

		info_map[k] = [nf_, line_num + kdx - nn_map[k][1], nn_map[k][0], nn_map[k][1], line_max_]

	files = info_map.values()
	files.sort(key=lambda o: o[1])

	return ff_map, info_map


def __assert_lino_max(ff, last_kdx, kdx_max, line_no, is_need):
	assert kdx_max == -1 or (
			kdx_max >= last_kdx and (line_no >= 1 or (line_no == 1 and is_need))
	), 'start not eq %d file: %s' % (last_kdx, ff)


# endregion

# noinspection PyUnresolvedReferences
def do_clean_ipp(base_pre, src_pre, file_pre, lines, part_map):
	lines_i = _build_lines(src_pre, file_pre, lines)
	ff_map, nn_map = _build_ff_map(file_pre, lines_i)

	kdx, len_, last_ff = 0, len(lines), ''
	for jdx, item in enumerate(lines_i, 0):
		if not item:
			last_ff = ''
			continue

		# noinspection PyTupleAssignmentBalance
		ff, line_no, line_ = item
		kdx_max, line_max, nf = ff_map[ff] if ff in ff_map else (0, 0, '')
		if line_no == 1 or line_no == line_max or kdx_max == -1:
			if last_ff == ff:
				last_ff = ''
				continue
			last_ff, lines[jdx] = ff, line_.replace(base_pre, '').replace(r'\\', '/')

	sub_map, nf_map, last_nf = {k: [] for k in part_map.keys()}, {}, {}
	while kdx < len_ and sub_map:
		last_kdx, item = kdx, lines_i[kdx]
		if last_nf and not item:
			sub_map[last_nf].append([lines[kdx], lines_i[kdx]])
			lines[kdx] = ''

		kdx += 1
		if not item: continue
		# noinspection PyTupleAssignmentBalance
		ff, line_no, line_ = item
		kdx_max, line_max, nf = ff_map[ff] if ff in ff_map else (0, 0, '')
		__assert_lino_max(ff, last_kdx, kdx_max, line_no, nf in sub_map)
		if last_nf and nf not in sub_map: last_nf = ''; continue
		if nf not in sub_map: continue
		sub_map[nf].append([lines[last_kdx], lines_i[last_kdx]])
		lines[last_kdx], last_nf = '#include "%s"' % (_out_name(nf),), nf
		nf_map[nf] = item
		while kdx <= kdx_max:
			sub_map[nf].append([lines[kdx], lines_i[kdx]])
			lines[kdx] = ''
			kdx += 1

	out_map = {}
	for sub_name, sub_item in sub_map.items():
		need_subs = [sub for sub in part_map.get(sub_name, []) if sub]
		if need_subs:
			lines_, lines_i_ = [i[0] for i in sub_item], [i[1] for i in sub_item]
			for need_sub in need_subs:
				# noinspection PyUnresolvedReferences
				tmp = sub_lines_of_part(nf_map[sub_name][0], need_sub, lines_i_, lines_)
				out_map[need_sub] = [i + "\n" for i in tmp if i]
			out_map[sub_name] = [i + "\n" for i in lines_ if i]
		else:
			out_map[sub_name] = [i[0] + "\n" for i in sub_item if i[0]]

	return [i + "\n" for i in lines if i], out_map


def sub_lines_of_part(file_pre, need_sub, lines_i, lines, sub_main=False):
	ff_map, nn_map = _build_ff_map(file_pre, lines_i)

	kdx, len_, erase, last_nf, sub_lines = 0, len(lines), False, '', []
	while kdx < len_:
		last_kdx, item = kdx, lines_i[kdx]
		if last_nf and not item:
			sub_lines.append(lines[kdx])
			lines[kdx] = ''

		kdx += 1
		if not item: continue
		ff, line_no, line_ = item
		kdx_max, line_max, nf = ff_map[ff] if ff in ff_map else (0, 0, '')
		__assert_lino_max(ff, last_kdx, kdx_max, line_no, nf == need_sub)

		lines[last_kdx] = line_ if sub_main and ff == file_pre else lines[last_kdx]

		if last_nf and nf != need_sub: break
		if nf != need_sub: continue
		sub_lines.append(lines[last_kdx])
		lines[last_kdx], last_nf = '#include "%s"' % (_out_name(nf),), nf
		while kdx <= kdx_max:
			sub_lines.append(lines[kdx])
			lines[kdx] = ''
			kdx += 1

	return sub_lines


# region  ----------- TEST FUNC -----------

def _LOG(msg, handle=None):
	print(msg, file=sys.stderr)
	if handle:
		handle.write(msg + '\n')
		handle.flush()


def all_test(test_pre='_test'):
	globals_dict = globals()
	for k, v in globals_dict.items():
		if k.startswith(test_pre):
			_LOG("\n\n>>%s" % (k,))
			if hasattr(v, '__call__'):
				v()


class Error(Exception):
	pass


def _unittest(func, *cases):
	def _functest(func_, is_pass, *args, **kws):
		result = None
		try:
			_LOG('\n%s -> %s' % (is_pass, func_.func_name))
			result = func_(*args, **kws)
			_LOG('=%s' % (json.dumps(result, indent=2),))
		except Error as ex:
			_LOG("%s -> %s:%s" % (is_pass, type(ex), ex))
			if is_pass:
				raise ex
		else:
			if not is_pass:
				raise AssertionError("is_pass:%s but no Exception!!!" % (is_pass,))
		return result

	return [_functest(func, *case) for case in cases]


# endregion

def main(action='clean_ipp', in_file='src/main.i', out_file=None):
	action = sys.argv[1] if len(sys.argv) >= 2 else action
	in_file = sys.argv[2] if len(sys.argv) >= 3 else in_file
	out_file = sys.argv[3] if len(sys.argv) >= 4 else out_file

	base = os.getcwd()

	if action == 'all_test':
		all_test()
	elif action == 'clean_ipp':
		clean_ipp(base, in_file, out_file)
	else:
		_LOG('''
    useage ` python2 "%~f0" all_test | clean_ipp [in_file] [out_file]`

    ''')


if __name__ == '__main__':
	_LOG("\n========== START ===========")
	main()
	_LOG("\n==========  END  ===========")
