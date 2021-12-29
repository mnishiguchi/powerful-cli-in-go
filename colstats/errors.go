package main

import "errors"

var (
	ErrNotNumber        = errors.New("Data is not numeric")
	ErrInvalidColumn    = errors.New("Invalid column number")
	ErrNoFile           = errors.New("No input file")
	ErrInvalidOperation = errors.New("Invalid operation")
)
