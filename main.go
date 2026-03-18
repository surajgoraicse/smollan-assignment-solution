package main

import (
	"fmt"
	"strconv"
	"strings"
)

func Compress(s string) string {
	if len(s) == 0 {
		return ""
	}
	var (
		count   = 0
		pointer = rune(s[0])
	)
	var result strings.Builder

	for _, v := range s {
		if v == pointer {
			count++
		} else {
			result.WriteRune(pointer)
			result.WriteString(strconv.Itoa(count))
			pointer = v
			count = 1
		}
	}
	result.WriteRune(pointer)
	result.WriteString(strconv.Itoa(count))

	return result.String()
}

func main() {
	res := Compress("aaabbbcccd")
	fmt.Println(res)
}
