package main

import (
	"fmt"

	"github.com/afshin-deriv/go-memsize"
)

type ComplexStruct struct {
	IntValue    int
	StringValue string
	SliceValue  []float64
	MapValue    map[string]interface{}
}

func main() {
	// Create a complex structure
	obj := &ComplexStruct{
		IntValue:    42,
		StringValue: "Hello, World!",
		SliceValue:  []float64{1.1, 2.2, 3.3},
		MapValue: map[string]interface{}{
			"key1": "value1",
			"key2": []int{1, 2, 3},
		},
	}

	// Get size without debug info
	size := memsize.GetTotalSize(obj)
	fmt.Printf("Total size: %d bytes\n", size)

	// Enable debug output and get size again
	memsize.Debug = true
	size = memsize.GetTotalSize(obj)
	fmt.Printf("Size with debug output: %d bytes\n", size)
}
