package random

import (
	"testing"
)

func TestString(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Log(String(10))
	}
}

func TestRandomLowerString(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Log(RandomLowerString(10))
	}
}

func TestRandomUpperString(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Log(RandomUpperString(10))
	}
}
