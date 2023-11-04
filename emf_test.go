package emf_test

import (
	"testing"

	"github.com/FollowTheProcess/emf"
)

func TestHello(t *testing.T) {
	got := emf.Hello()
	want := "Hello emf"

	if got != want {
		t.Errorf("got %s, wanted %s", got, want)
	}
}
