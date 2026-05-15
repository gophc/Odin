// Depends on: common.odin (String, Array, isize, u64, f64, heap_allocator, array_init, array_add, array_free, gb_printf_err, gb_memset, gb_assert_handler)
package cmd

import (
	"golang.org/x/sys/windows"
)

type TimeStamp struct {
	Start  uint64
	Finish uint64
	Label  String
}

type Timings struct {
	Total            TimeStamp
	Sections         []TimeStamp
	Freq             uint64
	TotalTimeSeconds float64
}

func win32_time_stamp_time_now() uint64 {
	var counter windows.LargeInteger
	windows.QueryPerformanceCounter(&counter)
	return uint64(counter)
}

var win32_perf_count_freq windows.LargeInteger

func win32_time_stamp_freq() uint64 {
	if win32_perf_count_freq == 0 {
		windows.QueryPerformanceFrequency(&win32_perf_count_freq)
		if !(win32_perf_count_freq != 0) {
			gb_assert_handler("Assertion Failure", "win32_perf_count_freq.QuadPart != 0", "timings.go", 24, 0)
		}
	}
	return uint64(win32_perf_count_freq)
}

func time_stamp_time_now() uint64 {
	return win32_time_stamp_time_now()
}

func time_stamp_freq() uint64 {
	return win32_time_stamp_freq()
}

func make_time_stamp(label String) TimeStamp {
	return TimeStamp{
		Start: time_stamp_time_now(),
		Label: label,
	}
}

func timings_init(t *Timings, label String, buffer_size isize) {
	t.Sections = make([]TimeStamp, 0, buffer_size)
	t.Total = make_time_stamp(label)
	t.Freq = time_stamp_freq()
}

func timings_destroy(t *Timings) {
	t.Sections = nil
}

func timings_stop_current_section(t *Timings) {
	if len(t.Sections) > 0 {
		t.Sections[len(t.Sections)-1].Finish = time_stamp_time_now()
	}
}

func timings_start_section(t *Timings, label String) {
	timings_stop_current_section(t)
	t.Sections = append(t.Sections, make_time_stamp(label))
}

func time_stamp_as_s(ts TimeStamp, freq uint64) float64 {
	if !(ts.Finish >= ts.Start) {
		gb_assert_handler("Assertion Failure", "ts.finish >= ts.start", "timings.go", 73, "time_stamp_as_ms")
	}
	return float64(ts.Finish-ts.Start) / float64(freq)
}

func time_stamp_as_ms(ts TimeStamp, freq uint64) float64 {
	return 1000.0 * time_stamp_as_s(ts, freq)
}

func time_stamp_as_us(ts TimeStamp, freq uint64) float64 {
	return 1000000.0 * time_stamp_as_s(ts, freq)
}

type TimingUnit int

const (
	TimingUnitSecond      TimingUnit = 0
	TimingUnitMillisecond TimingUnit = 1
	TimingUnitMicrosecond TimingUnit = 2
	TimingUnitCOUNT
)

var timing_unit_strings = [TimingUnitCOUNT]string{"s", "ms", "us"}

func time_stamp(ts TimeStamp, freq uint64, unit TimingUnit) float64 {
	switch unit {
	case TimingUnitMillisecond:
		return time_stamp_as_ms(ts, freq)
	case TimingUnitMicrosecond:
		return time_stamp_as_us(ts, freq)
	default:
	case TimingUnitSecond:
		return time_stamp_as_s(ts, freq)
	}
}

func timings_print_all(t *Timings, unit TimingUnit, timings_are_finalized bool) {
	const SPACES_LEN isize = 256
	SPACES := make([]byte, SPACES_LEN+1)
	for i := isize(0); i < SPACES_LEN; i++ {
		SPACES[i] = ' '
	}
	SPACES[SPACES_LEN] = 0

	if !timings_are_finalized {
		timings_stop_current_section(t)
		t.Total.Finish = time_stamp_time_now()
	}

	max_len := isize(36)
	if len(t.Total.Label) > int(max_len) {
		max_len = isize(len(t.Total.Label))
	}
	for _, ts := range t.Sections {
		if len(ts.Label) > int(max_len) {
			max_len = isize(len(ts.Label))
		}
	}
	if !(max_len <= SPACES_LEN) {
		gb_assert_handler("Assertion Failure", "max_len <= SPACES_LEN", "timings.go", 133, 0)
	}

	t.TotalTimeSeconds = time_stamp_as_s(t.Total, t.Freq)
	total_time := time_stamp(t.Total, t.Freq, unit)

	gb_printf_err("%s%*s - % 9.3f %s - %6.2f%%\n",
		t.Total.Label,
		int(max_len)-len(t.Total.Label), "",
		total_time,
		timing_unit_strings[unit],
		100.0)

	for _, ts := range t.Sections {
		section_time := time_stamp(ts, t.Freq, unit)
		gb_printf_err("%s%*s - % 9.3f %s - %6.2f%%\n",
			ts.Label,
			int(max_len)-len(ts.Label), "",
			section_time,
			timing_unit_strings[unit],
			100.0*section_time/total_time,
		)
	}
}
