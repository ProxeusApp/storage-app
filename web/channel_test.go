package channelhub

import (
	"testing"
)

func TestHasNamedParameter(t *testing.T) {
	if (&Channel{ID: "abc"}).hasNamedParameter() {
		t.Error("should'nt be true!")
	}
	if !(&Channel{ID: "abc:p"}).hasNamedParameter() {
		t.Error("should be true!")
	}
}
