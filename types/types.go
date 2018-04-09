package types

// NOTE(Jake): 2018-02-5
//
// Related types are currently in `parser/typecheck.go`
// The only reason this interface is not is to avoid a cyclic issue
// between `parser` and `ast`.
//

type TypeInfo interface {
	String() string
}

//
// Int
//

type Int struct{}

func (info *Int) String() string { return "int" }
