package golumns

import "fmt"

// Display is the function to actually print to the console
func Display(entries []string) error {
	lengths := lengthOfEntryNames(entries)
	columnLength := columnLength(lengths)
	entries = resizeEntries(columnLength, entries)
	termWidth, err := terminalWidth()
	if err != nil {
		return err
	}
	columns := termWidth / columnLength
	if columns == 0 {
		columns++
	}
	for i, entry := range entries {
		if i%columns == 0 {
			if i != 0 {
				fmt.Println()
			}
		}
		fmt.Printf("%s", entry)
	}
	fmt.Printf("%s", "\n")
	return nil
}
