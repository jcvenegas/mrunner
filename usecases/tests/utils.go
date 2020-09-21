package tests

import (
	"fmt"
	"strings"
)

//TODO: move to kernel
func stringToTomlList(str string) string {
	qStr := []string{}
	for _, a := range strings.Fields(str) {
		qStr = append(qStr, fmt.Sprintf("%q", a))
	}

	l := strings.Join(qStr, ", ")
	l = "[" + l + "]"
	return l
}
