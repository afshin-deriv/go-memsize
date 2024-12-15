# go-memsize
A Go library for calculating the actual memory size of objects, including all indirect allocations.

## Features
- Accurate memory size calculation of Go objects
- Supports complex data structures with circular references
- Debug mode for detailed size breakdowns
- Handles all Go types including:
 - Pointers and interfaces
 - Slices and arrays
 - Maps and structs
 - Basic types and strings

 ## Installation
 ```
 go get github.com/afshin-deriv/go-memsize
 ```

 ## Usage
 ```
 package main

import (
    "fmt"
    "github.com/afshin-deriv/go-memsize"
)

type Person struct {
    Name    string
    Friends []*Person
    Data    map[string]interface{}
}

func main() {
    person := &Person{
        Name: "John Doe",
        Friends: []*Person{},
        Data: map[string]interface{}{
            "age": 30,
            "hobbies": []string{"reading", "coding"},
        },
    }

    size := memsize.GetTotalSize(person)
    fmt.Printf("Total memory size: %d bytes\n", size)

    // Enable debug output
    memsize.Debug = true
    size = memsize.GetTotalSize(person)
}
 ```

 ## How It Works
The library calculates memory size by:

1. Using reflection to traverse object structures
2. Tracking visited pointers to handle circular references
3. Including backing array sizes for slices
4. Accounting for map bucket allocations
5. Measuring actual string data sizes

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

## License
MIT License - see LICENSE file
