package mytest

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"testing"

	mypkg "example.com/mymodule/mypkg"
	"github.com/stretchr/testify/assert"
)

// searches for the first occurrence of each element in the array substrs within the string a.
// and returns a new slice arranged in the order of appearance.
// if any element does not exist in a, it returns an error.
func orderSubstrings(a string, substrs []string) ([]string, error) {
	// structure for storing each element along with its occurrence position in a
	type posPair struct {
		sub string // elements in substrs
		pos int    // first position in string a
	}

	var pairs []posPair

	for _, sub := range substrs {
		// searches for the position of the substring sub (the first occurrence)
		index := strings.Index(a, sub)
		if index == -1 {
			return nil, fmt.Errorf("substring %q not found in string", sub)
		}
		pairs = append(pairs, posPair{sub: sub, pos: index})
	}

	// sorts in ascending order by positions of pos
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].pos < pairs[j].pos
	})

	// creates a slice containing only the substrings in the sorted order
	ordered := make([]string, len(pairs))
	for i, p := range pairs {
		ordered[i] = p.sub
	}

	return ordered, nil
}

func TestBuiltinLogger(t *testing.T) {
	logger := mypkg.NewBuiltinLogger("stdout")

	tests := [5]struct {
		method string
		input  string
	}{
		{"Debug", "Debug message"},
		{"Info", "Info message"},
		{"Warning", "Warning message"},
		{"Error", "Error message"},
		{"Fatal", "Fatal message"},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			switch tt.method {
			case "Debug":
				logger.Debug(tt.input)
			case "Info":
				logger.Info(tt.input)
			case "Warning":
				logger.Warning(tt.input)
			case "Error":
				logger.Error(tt.input)
			case "Fatal":
				// asserts the statement on panic with defer and recover()
				defer func() {
					err := recover()
					pc, file, line, ok := runtime.Caller(0)
					fn := runtime.FuncForPC(pc)
					expectedMsg := fmt.Sprintf(tt.input)
					if ok {
						caller := fmt.Sprintf("@%s:%d %s(): ", file, line, fn.Name())
						expectedMsg = fmt.Sprintf(caller + tt.input)
					}

					expectedStringOrder := []string{"logger_test.go", "TestBuiltinLogger.func", "Fatal message"}
					actualMsg := fmt.Sprintf("%v", err)
					actualOrder, err := orderSubstrings(actualMsg, expectedStringOrder)
					expectedOrder, err := orderSubstrings(expectedMsg, expectedStringOrder)
					if err != nil {
						fmt.Println(err)
					}

					assert.ElementsMatch(t, actualOrder, expectedOrder)
				}()
				logger.Fatal(tt.input)
			}
		})
	}
}
