package circularbuffer

import (
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) {
	circle := New(7)
	mybuf := make([]int, 7)

	for i := 0; i < 10; i++ {
		mybuf[circle.AppendIndex()] = i
	}

	a, b, c, d := circle.SliceIndices()

	got := append(mybuf[a:b], mybuf[c:d]...)
	want := []int{3, 4, 5, 6, 7, 8, 9}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestDifferentParams(t *testing.T) {
	for n := 1; n < 10; n++ {
		for m := 0; m < 3*n; m++ {
			circle := New(n)
			mybuf := make([]int, n)

			var want []int
			for i := 0; i < n; i++ {
				val := m - i - 1
				if val < 0 {
					continue
				}
				want = append([]int{val}, want...)
			}

			for i := 0; i < m; i++ {
				mybuf[circle.AppendIndex()] = i
			}

			a, b, c, d := circle.SliceIndices()

			got := append(mybuf[a:b], mybuf[c:d]...)

			if !((len(want) == 0 && len(got) == 0) || reflect.DeepEqual(want, got)) {
				t.Errorf("for a buffer of capacity %d filling %d values, got %v want %v", n, m, got, want)
			}
		}
	}
}

func TestEmpty(t *testing.T) {
	circle := New(7)
	mybuf := make([]int, 7)

	a, b, c, d := circle.SliceIndices()

	got := len(append(mybuf[a:b], mybuf[c:d]...))
	want := 0

	if got != want {
		t.Errorf("length incorrect: got %v want %v", got, want)
	}

	if circle.Elements != want {
		t.Errorf("Elements incorrect: got %v want %v", circle.Elements, want)
	}
}

func TestOne(t *testing.T) {
	circle := New(7)
	mybuf := make([]int, 7)

	circle.AppendIndex()

	a, b, c, d := circle.SliceIndices()

	got := len(append(mybuf[a:b], mybuf[c:d]...))
	want := 1

	if got != want {
		t.Errorf("length incorrect: got %v want %v", got, want)
	}

	if circle.Elements != want {
		t.Errorf("Elements incorrect: got %v want %v", circle.Elements, want)
	}
}

func TestFull(t *testing.T) {
	circle := New(7)
	mybuf := make([]int, 7)

	for i := 0; i < 10; i++ {
		circle.AppendIndex()
	}

	a, b, c, d := circle.SliceIndices()

	got := len(append(mybuf[a:b], mybuf[c:d]...))
	want := 7

	if got != want {
		t.Errorf("length incorrect: got %v want %v", got, want)
	}

	if circle.Elements != want {
		t.Errorf("Elements incorrect: got %v want %v", circle.Elements, want)
	}
}
