package main

import (
	"fmt"
	"errors"
)

var ErrZero = errors.New("zero is not allowed")

func fizzBuzz(n int) (string, error) {
	if (n == 0) {
		return "", ErrZero
	} else if (n % 5 == 0 && n % 3 == 0) {
		return "FizzBuzz", nil
	} else if (n % 3 == 0) {
		return "Fizz", nil
	} else if (n % 5 == 0) {
		return "Buzz", nil
	} else {
		return fmt.Sprintf("%d", n), nil
	}
}
