package scanner

import (
	"github.com/silbinarywolf/compiler-fel/token"
	"testing"
)

//
// todo(Jake): remove the items below and rewrite to check for
//			   UTF-8 support. Might look to see Golang's tests
//			   for this.
//
//			   - Jake B, 4th December, 2017
//

func TestEatCommmentLine(t *testing.T) {
	s := New([]byte(`
		// Test Comment A

		// Test Comment B
		// - Line A
		// - Line B
		// - Line C
	`), "TestEatWhitespace")
	for {
		t := s.GetNextToken()
		if t.Kind == token.EOF {
			break
		}
	}
}

func TestEatCommmentBlock(t *testing.T) {
	s := New([]byte(`
		/*
			Test Comment Block
		*/

		/*
			/*
				Test Nested Comment Block
			*/
		*/
	`), "TestEatWhitespace")
	for {
		t := s.GetNextToken()
		if t.Kind == token.EOF {
			break
		}
	}
}
