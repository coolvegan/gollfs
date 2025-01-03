package gollfs

import (
	"testing"
)

func TestComment(t *testing.T) {
	s := "#"
	result := comment(s)
	if !result {
		t.Errorf("Must be a comment")
	}
}
func TestCommentWhitespace(t *testing.T) {
	s := "#"
	result := comment(s)
	if !result {
		t.Errorf("Must be a comment")
	}
}
func TestCommentTabulator(t *testing.T) {
	s := string('\t') + string('\t') + "  #"
	result := comment(s)
	if !result {
		t.Errorf("Must be a comment")
	}
}

func TestCommentIsNoComment(t *testing.T) {
	s := string('\t') + "watchdog=yes"
	result := comment(s)
	if result {
		t.Errorf("Must be not a comment")
	}
}

func TestCommentIsNoComment2(t *testing.T) {
	s := "# watchdog=yes"
	result := comment(s)
	if !result {
		t.Errorf("Must be not a comment")
	}
}
