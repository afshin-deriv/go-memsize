package memsize

import (
	"fmt"
	"testing"
	"unsafe"
)

type Person struct {
	Name    string
	Friends []*Person
	Data    map[string]interface{}
}

// Adding structs with function pointers
type Handler struct {
	OnSuccess func(data string) error
	OnError   func(err error)
	Process   func(input int) (output int)
}

type ServiceWithCallbacks struct {
	Name           string
	Handlers       []Handler
	DefaultHandler func(string) error
}

func TestGetTotalSize(t *testing.T) {
	Debug = true // Enable debug output

	t.Run("Function Pointers", func(t *testing.T) {
		// Create some test functions
		successFn := func(data string) error {
			return nil
		}

		errorFn := func(err error) {
			// Do nothing
		}

		processFn := func(input int) int {
			return input * 2
		}

		// Test single function pointer
		t.Run("Single Function", func(t *testing.T) {
			size := GetTotalSize(successFn)
			fmt.Printf("Single function pointer size: %d bytes\n", size)
			if size == 0 {
				t.Error("Expected non-zero size for function pointer")
			}
		})

		// Test struct with function pointers
		t.Run("Struct With Functions", func(t *testing.T) {
			handler := Handler{
				OnSuccess: successFn,
				OnError:   errorFn,
				Process:   processFn,
			}

			size := GetTotalSize(handler)
			fmt.Printf("Handler struct with function pointers size: %d bytes\n", size)
			if size == 0 {
				t.Error("Expected non-zero size for handler struct")
			}
		})

		// Test slice of structs with function pointers
		t.Run("Slice of Handlers", func(t *testing.T) {
			service := ServiceWithCallbacks{
				Name:           "TestService",
				Handlers:       make([]Handler, 2),
				DefaultHandler: successFn,
			}

			service.Handlers[0] = Handler{
				OnSuccess: successFn,
				OnError:   errorFn,
				Process:   processFn,
			}

			service.Handlers[1] = Handler{
				OnSuccess: func(data string) error { return nil },
				OnError:   func(err error) {},
				Process:   func(input int) int { return input },
			}

			size := GetTotalSize(service)
			fmt.Printf("Service with handlers size: %d bytes\n", size)
			if size == 0 {
				t.Error("Expected non-zero size for service")
			}
		})

		// Test map with function values
		t.Run("Map with Functions", func(t *testing.T) {
			funcMap := map[string]interface{}{
				"success": successFn,
				"error":   errorFn,
				"process": processFn,
			}

			size := GetTotalSize(funcMap)
			fmt.Printf("Map with function values size: %d bytes\n", size)
			if size == 0 {
				t.Error("Expected non-zero size for function map")
			}
		})
	})

	// Original Person struct test
	t.Run("Complex Structure", func(t *testing.T) {
		person := &Person{
			Name:    "John Doe",
			Friends: make([]*Person, 0),
			Data: map[string]interface{}{
				"age":     30,
				"hobbies": []string{"reading", "coding", "something else!!!"},
			},
		}

		friend := &Person{
			Name:    "Jane Doe",
			Friends: []*Person{person},
			Data: map[string]interface{}{
				"age": 28,
			},
		}
		person.Friends = append(person.Friends, friend)

		size := GetTotalSize(person)
		fmt.Printf("Complex structure size: %d bytes\n", size)

		if size == 0 {
			t.Error("Size should not be zero for complex structure")
		}
		if size > 10*1024*1024 { // 10MB sanity check
			t.Error("Size seems unreasonably large")
		}
	})

	// Test simple types first
	t.Run("Simple Types", func(t *testing.T) {
		cases := []struct {
			name string
			v    interface{}
		}{
			{"int", 42},
			{"string", "hello"},
			{"slice", []int{1, 2, 3}},
			{"map", map[string]int{"a": 1, "b": 2}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				size := GetTotalSize(tc.v)
				fmt.Printf("%s size: %d bytes\n", tc.name, size)
				if size == 0 {
					t.Errorf("Expected non-zero size for %s", tc.name)
				}
			})
		}
	})

	// Test complex structure
	t.Run("Complex Structure", func(t *testing.T) {
		person := &Person{
			Name:    "John Doe",
			Friends: make([]*Person, 0),
			Data: map[string]interface{}{
				"age":     30,
				"hobbies": []string{"reading", "coding", "something else!!!"},
			},
		}

		friend := &Person{
			Name:    "Jane Doe",
			Friends: []*Person{person},
			Data: map[string]interface{}{
				"age": 28,
			},
		}
		person.Friends = append(person.Friends, friend)

		size := GetTotalSize(person)
		fmt.Printf("Complex structure size: %d bytes\n", size)

		if size == 0 {
			t.Error("Size should not be zero for complex structure")
		}
		if size > 10*1024*1024 { // 10MB sanity check
			t.Error("Size seems unreasonably large")
		}
	})
}

func TestComplexStructure(t *testing.T) {
	Debug = true

	person := &Person{
		Name:    "John Doe",
		Friends: make([]*Person, 0),
		Data: map[string]interface{}{
			"age":     30,
			"hobbies": []string{"reading", "coding"},
		},
	}

	// Add a friend to create circular reference
	friend := &Person{
		Name:    "Jane Doe",
		Friends: []*Person{person},
		Data: map[string]interface{}{
			"age": 28,
		},
	}
	person.Friends = append(person.Friends, friend)

	size := GetTotalSize(person)
	fmt.Printf("Complex structure total size: %d bytes\n", size)

	// Basic size validations
	if size <= uint64(unsafe.Sizeof(person)) {
		t.Error("Size should be larger than just the pointer size")
	}

	// Validate size components exist
	if size < uint64(len(person.Name)) {
		t.Error("Size should include string content")
	}

	if size < uint64(cap(person.Friends)*int(unsafe.Sizeof(&Person{}))) {
		t.Error("Size should include Friends slice capacity")
	}

	// The size should be reasonable
	if size < 100 || size > 10000 {
		t.Errorf("Size %d seems unreasonable for this structure", size)
	}
}
