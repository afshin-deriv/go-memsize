// doc.go
package memsize

/*
Package memsize provides functionality to calculate the actual memory size of Go objects,
including all indirect allocations.

The main function GetTotalSize accepts any Go value and returns its total memory size
in bytes, including all referenced memory. It correctly handles circular references and
complex data structures.

Example usage:

    type Person struct {
        Name    string
        Friends []*Person
        Data    map[string]interface{}
    }

    person := &Person{
        Name: "John",
        Friends: make([]*Person, 0),
        Data: map[string]interface{}{
            "age": 30,
        },
    }

    size := memsize.GetTotalSize(person)
    fmt.Printf("Total memory size: %d bytes\n", size)

For detailed size calculation information, enable debug mode:

    memsize.Debug = true
    size = memsize.GetTotalSize(person)

The library handles all Go types including:
  - Basic types (int, float64, bool, etc.)
  - Strings
  - Slices and arrays
  - Maps
  - Structs
  - Pointers and interfaces
  - Circular references
*/
