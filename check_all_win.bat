@echo off

if "%1" == "" (
	echo Checking windows_i386
	odin check examples\all -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -target:windows_i386
	echo Checking windows_amd64
	odin check examples\all -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -target:windows_amd64

	echo Checking js_wasm32
	odin check examples\all -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -target:js_wasm32
	echo Checking js_wasm64p32
	odin check examples\all -vet -vet-tabs -strict-style -vet-style -warnings-as-errors -disallow-do -target:js_wasm64p32
)
