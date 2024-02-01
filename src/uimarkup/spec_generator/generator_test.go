package spec_generator

import (
	"testing"
)

func TestGenerateProperties(t *testing.T) {
	t.SkipNow()
	if err := writePropertyFile(); err != nil {
		t.Fatal(err)
	} else if err := writeProperties(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateFunctions(t *testing.T) {
	t.SkipNow()
	if err := writeFunctionFile(); err != nil {
		t.Fatal(err)
	} else if err := writeFunctions(); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateElements(t *testing.T) {
	t.SkipNow()
	if err := writeElementsFile(); err != nil {
		t.Fatal(err)
	} else if err := writeElements(); err != nil {
		t.Fatal(err)
	}
}

func TestGeneratePseudos(t *testing.T) {
	t.SkipNow()
	if err := writePseudoFile(); err != nil {
		t.Fatal(err)
	} else if err := writePseudos(); err != nil {
		t.Fatal(err)
	}
}
