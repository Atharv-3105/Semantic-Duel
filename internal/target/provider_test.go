package target

import (
	"testing"
)

func TestProvider_ReturnsWordFromList(t *testing.T) {
	words := []string{"fire", "water", "earth"}
	p := New(words)

	got := p.Random()

	found := false
	for _, w := range words {
		if w == got {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Random() returned %q which is not in the word list", got)
	}
}

func TestProvider_EmptyList_ReturnsDefault(t *testing.T) {
	p := New([]string{})
	got := p.Random()

	if got != "default" {
		t.Errorf("expected \"default\" for empty word list, got %q", got)
	}
}

func TestProvider_SingleWord_AlwaysReturnsThatWord(t *testing.T) {
	p := New([]string{"ocean"})

	for i := 0; i < 20; i++ {
		if got := p.Random(); got != "ocean" {
			t.Errorf("expected \"ocean\", got %q on iteration %d", got, i)
		}
	}
}

func TestProvider_NilList_ReturnsDefault(t *testing.T) {
	p := New(nil)
	got := p.Random()

	if got != "default" {
		t.Errorf("expected \"default\" for nil word list, got %q", got)
	}
}

func TestProvider_DistributionCoversAllWords(t *testing.T) {
	// Run enough iterations to expect every word to appear at least once.
	words := []string{"a", "b", "c", "d", "e"}
	p := New(words)
	seen := make(map[string]bool)

	for i := 0; i < 500; i++ {
		seen[p.Random()] = true
	}

	for _, w := range words {
		if !seen[w] {
			t.Errorf("word %q never returned in 500 iterations — distribution may be broken", w)
		}
	}
}
