package golumns

import (
	"math"
	"sort"

	"golang.org/x/crypto/ssh/terminal"
)

func terminalWidth() (int, error) {
	width, _, err := terminal.GetSize(0)
	return width, err
}

func lengthOfEntryNames(listOfEntryNames []string) []int {
	lengths := []int{}
	for _, entry := range listOfEntryNames {
		lengths = append(lengths, len(entry))
	}
	return lengths
}

// fewestColumns() determins the fewest possible amount of columns that could
// fit the terminal window based on the longest possible combination of elements
func fewestColumns(lengths []int) int {
	termWidth, _ := terminalWidth()

	columns := 1
	prevLength := 0
	for _, length := range lengths {
		if length+prevLength+4 < termWidth {
			columns++
			prevLength += length + 4
		} else {
			break
		}
	}
	return columns
}

func columnLength(lengths []int) int {
	sort.Ints(lengths)
	if len(lengths) == 0 {
		return 4
	}
	return reverse(lengths)[0] + 4
}

// resizeEntries() pads each element to fit correct column length
func resizeEntries(fullSize int, entryNames []string) []string {
	for i, entry := range entryNames {
		for len(entry) != fullSize {
			entry += " "
		}
		entryNames[i] = entry
	}
	return entryNames
}

func numberOfRows(numberOfRows, numberOfNames int) int {
	return int(math.Ceil(float64(numberOfNames) / float64(numberOfRows)))
}

func reverse(input []int) []int {
	if len(input) == 0 {
		return input
	}
	return append(reverse(input[1:]), input[0])
}
