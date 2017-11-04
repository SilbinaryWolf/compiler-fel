# Front-End Language (FEL)

An experimental programming language in early stages of design and development.
The compiler is written in Golang.

# How to run / test project

1) Execute from root directory: `go run fel/fel.go`

2) This will process the `testdata/sampleproject/fel` files and output them in `testdata/sampleproject/public`

3) To run tests, use `go test ./...` from root directory. This will run all project tests, at the time of writing (2017-11-04), there is only `evaluator/css_optimize_test.go`

4) To vet your code, use `go vet ./...` from root directory. This will check for deadcode and incorrect use of fmt.Printf-like functions.

# Proposal / Goals
[https://github.com/SilbinaryWolf/proposal-fel](https://github.com/SilbinaryWolf/proposal-fel)

# Roadmap
[https://github.com/SilbinaryWolf/proposal-fel/blob/master/ROADMAP.md](https://github.com/SilbinaryWolf/proposal-fel/blob/master/ROADMAP.md)