package add

import "testing"

func TestAdd(t *testing.T) {
	a := 2
	b := 3

	expected := 5
	actual := add(a, b)

	if expected != actual {
		t.Errorf("Expected %d, got %d.", expected, actual)
	}
}
