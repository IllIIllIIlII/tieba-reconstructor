package test

import (
	"regexp"
	"testing"
)

func TestRegex(T *testing.T) {
	regex := regexp.MustCompile("\"post_no\":\\d*,")
	str := "\"post_no\":2,"
	r := regex.FindString(str)
	println(r)
}
