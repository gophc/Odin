package util

import (
	"reflect"
	"strconv"
	"testing"
)

type testMapToCase[T any, V any] struct {
	s    []T
	f    func(T) (V, bool)
	want []V
}

func testMapTo[T any, V any](t *testing.T, pre string, tests []testMapToCase[T, V]) {
	for i, tt := range tests {
		t.Run(pre+strconv.Itoa(i+1)+">", func(t *testing.T) {
			if got := MapTo(tt.s, tt.f); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapTo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapTo(t *testing.T) {
	testMapTo(t, "MapTo<T1_", []testMapToCase[int, float64]{
		{
			s:    []int{1, 2, 3, 4},
			want: []float64{2.0, 4.0},
			f: func(n int) (float64, bool) {
				return float64(n), n%2 == 0
			},
		},
		{
			s:    []int{1, 2, 3, 4},
			want: []float64{1.0, 3.0},
			f: func(n int) (float64, bool) {
				return float64(n), n%2 == 1
			},
		},
		{
			s:    []int{1, 2, 3, 4},
			want: []float64{},
			f: func(n int) (float64, bool) {
				return float64(n), n%2 == 2
			},
		},
	})

	testMapTo(t, "MapTo<T2_", []testMapToCase[int, string]{
		{
			s:    []int{1, 2, 3, 4},
			want: []string{"2", "4"},
			f: func(n int) (string, bool) {
				return strconv.Itoa(n), n%2 == 0
			},
		},
		{
			s:    []int{1, 2, 3, 4},
			want: []string{"1", "3"},
			f: func(n int) (string, bool) {
				return strconv.Itoa(n), n%2 == 1
			},
		},
		{
			s:    []int{1, 2, 3, 4},
			want: []string{},
			f: func(n int) (string, bool) {
				return strconv.Itoa(n), n%2 == 2
			},
		},
	})
}

func testMaxMin[T Ordered](t *testing.T, x T, y T, want T) {
	t.Run("Max", func(t *testing.T) {
		if got := Max(x, y); !reflect.DeepEqual(got, want) {
			t.Errorf("Max() = %v, want %v", got, want)
		}
	})

	t.Run("Min", func(t *testing.T) {
		switch want {
		case y:
			if got := Min(x, y); !reflect.DeepEqual(got, x) {
				t.Errorf("Min() = %v, want %v", got, x)
			}
		case x:
			if got := Min(x, y); !reflect.DeepEqual(got, y) {
				t.Errorf("Min() = %v, want %v", got, y)
			}
		}
	})
}

func TestMaxMin(t *testing.T) {
	testMaxMin(t, 1, 2, 2)
	testMaxMin(t, 1, 0, 1)
	testMaxMin(t, -1, 0, 0)
	testMaxMin(t, 1.0, 0.0, 1.0)

	testMaxMin(t, 'a', 'b', 'b')
	testMaxMin(t, "ab", "b", "b")

	testMaxMin(t, "ab", "aa", "ab")
}

type testTryNCase[T any] struct {
	s    []T
	i    int
	v    T
	want T
}

func testTryN[T any](t *testing.T, pre string, tests []testTryNCase[T]) {
	for i, tt := range tests {
		t.Run(pre+strconv.Itoa(i)+">", func(t *testing.T) {
			if got := TryN(tt.s, tt.i, tt.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TryN() = %v, want %v", got, tt.want)
			}
		})

		switch tt.i {
		case 0:
			t.Run(pre+strconv.Itoa(i)+">", func(t *testing.T) {
				if got := TryFirst(tt.s, tt.v); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TryFirst() = %v, want %v", got, tt.want)
				}
			})
		case 1:
			t.Run(pre+strconv.Itoa(i)+">", func(t *testing.T) {
				if got := TrySecond(tt.s, tt.v); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TrySecond() = %v, want %v", got, tt.want)
				}
			})
		case 2:
			t.Run(pre+strconv.Itoa(i)+">", func(t *testing.T) {
				if got := TryThird(tt.s, tt.v); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("TryThird() = %v, want %v", got, tt.want)
				}
			})
		}
	}
}

func TestTryN(t *testing.T) {
	testTryN(t, "TryN<T1_", []testTryNCase[int]{
		{
			s: []int{1, 2, 3},
			i: 0, v: -1, want: 1,
		},
		{
			s: []int{1, 2, 3},
			i: 1, v: -1, want: 2,
		},
		{
			s: []int{1, 2, 3},
			i: 2, v: -1, want: 3,
		},
		{
			s: []int{1, 2, 3},
			i: 3, v: -1, want: -1,
		},
	})

	testTryN(t, "TryN<T2_", []testTryNCase[int]{
		{
			s: []int{},
			i: 0, v: -1, want: -1,
		},
		{
			s: []int{},
			i: 1, v: -1, want: -1,
		},
		{
			s: []int{},
			i: -1, v: -1, want: -1,
		},
		{
			s: []int{},
			i: 3, v: -2, want: -2,
		},
	})

	testTryN(t, "TryN<T3_", []testTryNCase[float64]{
		{
			s: []float64{1.0, 2.0, 3.0},
			i: 0, v: -1.0, want: 1.0,
		},
		{
			s: []float64{1.0, 2.0, 3.0},
			i: 1, v: -1.0, want: 2.0,
		},
		{
			s: []float64{1.0, 2.0, 3.0},
			i: 2, v: -1.0, want: 3.0,
		},
		{
			s: []float64{1.0, 2.0, 3.0},
			i: 3, v: -1.0, want: -1.0,
		},
	})
}

type testBinarySearchCase[T Ordered] struct {
	arr    []T
	target T
	want   int
	found  bool
}

func testBinarySearch[T Ordered](t *testing.T, pre string, tests []testBinarySearchCase[T]) {
	for i, tt := range tests {
		t.Run(pre+strconv.Itoa(i)+">", func(t *testing.T) {
			got, found := BinarySearch(tt.arr, tt.target)
			if got != tt.want {
				t.Errorf("BinarySearch(%v, %v) got = %v, want %v", tt.arr, tt.target, got, tt.want)
			}
			if found != tt.found {
				t.Errorf("BinarySearch(%v, %v) got1 = %v, want %v", tt.arr, tt.target, found, tt.found)
			}
		})
	}
}

func TestBinarySearch(t *testing.T) {
	testBinarySearch(t, "BinarySearch<T1_", []testBinarySearchCase[int]{
		{
			arr: []int{1, 2, 3, 4}, target: 1,
			want: 0, found: true,
		},
		{
			arr: []int{1, 2, 3, 4}, target: 2,
			want: 1, found: true,
		},
		{
			arr: []int{1, 2, 3, 4}, target: 3,
			want: 2, found: true,
		},
		{
			arr: []int{1, 2, 3, 4}, target: 4,
			want: 3, found: true,
		},
	})

	testBinarySearch(t, "BinarySearch<T2_", []testBinarySearchCase[int]{
		{
			arr: []int{2, 4, 6, 8}, target: 0,
			want: -1, found: false,
		},
		{
			arr: []int{2, 4, 6, 8}, target: 9,
			want: -1, found: false,
		},
		{
			arr: []int{2, 4, 6, 8}, target: 4,
			want: 1, found: true,
		},
		{
			arr: []int{2, 4, 6, 8}, target: 5,
			want: 1, found: false,
		},
		{
			arr: []int{2, 4, 6, 8}, target: 6,
			want: 2, found: true,
		},
	})
	testBinarySearch(t, "BinarySearch<T3_", []testBinarySearchCase[int]{
		{
			arr: []int{1, 2, 3, 3, 3, 3, 4}, target: 3,
			want: 5, found: true,
		},
		{
			arr: []int{1, 2, 3, 3, 3, 3, 4}, target: 2,
			want: 1, found: true,
		},
		{
			arr: []int{1, 2, 3, 3, 3, 3, 3, 3, 3, 4}, target: 3,
			want: 8, found: true,
		},
		{
			arr: []int{1, 2, 3, 3, 3, 3, 3, 3, 3, 3}, target: 3,
			want: 9, found: true,
		},
		{
			arr: []int{3, 3, 3, 3, 3, 3, 3, 3, 3, 4}, target: 3,
			want: 8, found: true,
		},
		{
			arr: []int{1, 2, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 4}, target: 3,
			want: 11, found: true,
		},
	})

	testBinarySearch(t, "BinarySearch<T4_", []testBinarySearchCase[int]{
		{
			arr: []int{2}, target: 1,
			want: -1, found: false,
		},
		{
			arr: []int{2}, target: 2,
			want: 0, found: true,
		},
		{
			arr: []int{2}, target: 3,
			want: -1, found: false,
		},
	})
}
