# Front-End Language (FEL)

[![Build Status](https://travis-ci.org/silbinarywolf/compiler-fel.svg?branch=master)](https://travis-ci.org/silbinarywolf/compiler-fel)

**WARNING: This language is in early design and alpha stages, it's recommended that you do not waste your time playing with this**

An experimental programming language in early stages of design and development.
The compiler is written in Golang.

# How to run / testdata project

1) Install `go get github.com/silbinarywolf/compiler-fel`

2) Execute from root directory: `go run fel/fel.go`

3) This will process the `testdata/sampleproject/fel` files and output them in `testdata/sampleproject/public`

4) To run tests, use `go test ./...` from root directory. This will run all project tests, at the time of writing (2017-11-04), there is only `evaluator/css_optimize_test.go`

5) To vet your code, use `go vet ./...` from root directory. This will check for deadcode and incorrect use of fmt.Printf-like functions.

# Proposal / Goals
[https://github.com/SilbinaryWolf/proposal-fel](https://github.com/SilbinaryWolf/proposal-fel)

# Roadmap
[https://github.com/SilbinaryWolf/proposal-fel/blob/master/ROADMAP.md](https://github.com/SilbinaryWolf/proposal-fel/blob/master/ROADMAP.md)