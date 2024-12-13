// memsize.go
package memsize

import (
	"fmt"
	"reflect"
	"runtime"
	"unsafe"
)

// visited keeps track of addresses we've already counted
type visited map[uintptr]bool

// Debug enables detailed size calculation logging
var Debug bool = false

func debugPrint(format string, args ...interface{}) {
	if Debug {
		fmt.Printf(format+"\n", args...)
	}
}

// GetTotalSize returns the total memory size including indirect allocations
func GetTotalSize(v interface{}) uint64 {
	runtime.GC()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	initialHeap := stats.HeapAlloc

	val := reflect.ValueOf(v)
	size := getTotalSize(val, make(visited), "root")

	if Debug {
		fmt.Printf("Initial heap: %d, Final size: %d\n", initialHeap, size)
	}

	return size
}

func getTotalSize(v reflect.Value, seen visited, path string) uint64 {
	if !v.IsValid() {
		debugPrint("%s: Invalid value", path)
		return 0
	}

	var size uint64

	// Special handling for primitive types
	switch v.Kind() {
	case reflect.Bool:
		size := uint64(1) // 1 byte
		debugPrint("%s: Bool size %d", path, size)
		return size

	case reflect.Int8, reflect.Uint8:
		size := uint64(1) // 1 byte
		debugPrint("%s: Int8/Uint8 size %d", path, size)
		return size

	case reflect.Int16, reflect.Uint16:
		size := uint64(2) // 2 bytes
		debugPrint("%s: Int16/Uint16 size %d", path, size)
		return size

	case reflect.Int32, reflect.Uint32, reflect.Float32:
		size := uint64(4) // 4 bytes
		debugPrint("%s: Int32/Uint32/Float32 size %d", path, size)
		return size

	case reflect.Int64, reflect.Uint64, reflect.Float64:
		size := uint64(8) // 8 bytes
		debugPrint("%s: Int64/Uint64/Float64 size %d", path, size)
		return size

	case reflect.Int, reflect.Uint:
		// Size depends on platform (usually 8 bytes on 64-bit systems)
		size := uint64(v.Type().Size())
		debugPrint("%s: Int/Uint size %d", path, size)
		return size
	}

	// Handle special cases first
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			size = uint64(unsafe.Sizeof(v.Interface()))
			debugPrint("%s: Nil interface, size %d", path, size)
			return size
		}
		elemSize := getTotalSize(v.Elem(), seen, path+".elem")
		debugPrint("%s: Interface elem size %d", path, elemSize)
		return elemSize + uint64(unsafe.Sizeof(v.Interface()))

	case reflect.Ptr:
		if v.IsNil() {
			size = uint64(unsafe.Sizeof(v.Interface()))
			debugPrint("%s: Nil pointer, size %d", path, size)
			return size
		}

		// Get pointer address
		addr := uintptr(v.UnsafePointer())
		ptrSize := uint64(unsafe.Sizeof(v.Interface()))

		// Even if we've seen this pointer, we still count the pointer itself
		if seen[addr] {
			debugPrint("%s: Already seen pointer %x, size %d", path, addr, ptrSize)
			return ptrSize
		}

		// Mark as seen
		seen[addr] = true

		// Get the element size
		elemSize := getTotalSize(v.Elem(), seen, path+".ptr")
		totalSize := ptrSize + elemSize
		debugPrint("%s: Pointer to new address %x (size: %d) + elem (size: %d) = %d",
			path, addr, ptrSize, elemSize, totalSize)
		return totalSize

	case reflect.Slice:
		if v.IsNil() {
			debugPrint("%s: Nil slice", path)
			return 0
		}

		headerSize := uint64(unsafe.Sizeof(v.Interface()))
		arraySize := uint64(0)
		if v.Cap() > 0 {
			arraySize = uint64(v.Cap()) * uint64(v.Type().Elem().Size())
		}

		elementsSize := uint64(0)
		for i := 0; i < v.Len(); i++ {
			elemSize := getTotalSize(v.Index(i), seen, fmt.Sprintf("%s[%d]", path, i))
			elementsSize += elemSize
		}

		size = headerSize + arraySize + elementsSize
		debugPrint("%s: Slice header(%d) + array(%d) + elements(%d) = %d",
			path, headerSize, arraySize, elementsSize, size)
		return size

	case reflect.String:
		headerSize := uint64(unsafe.Sizeof(v.Interface()))
		dataSize := uint64(v.Len())
		size = headerSize + dataSize
		debugPrint("%s: String header(%d) + data(%d) = %d", path, headerSize, dataSize, size)
		return size

	case reflect.Map:
		if v.IsNil() {
			debugPrint("%s: Nil map", path)
			return 0
		}

		headerSize := uint64(unsafe.Sizeof(v.Interface()))
		bucketSize := uint64(48) // approximate bucket overhead
		bucketsSize := (uint64(v.Len())/8 + 1) * bucketSize

		contentSize := uint64(0)
		iter := v.MapRange()
		for iter.Next() {
			keySize := getTotalSize(iter.Key(), seen, path+".key")
			valSize := getTotalSize(iter.Value(), seen, path+".value")
			contentSize += keySize + valSize
		}

		size = headerSize + bucketsSize + contentSize
		debugPrint("%s: Map header(%d) + buckets(%d) + content(%d) = %d",
			path, headerSize, bucketsSize, contentSize, size)
		return size

	case reflect.Struct:
		structSize := uint64(unsafe.Sizeof(v.Interface()))
		fieldsSize := uint64(0)

		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldName := v.Type().Field(i).Name
			fieldSize := getTotalSize(field, seen, fmt.Sprintf("%s.%s", path, fieldName))
			fieldsSize += fieldSize
		}

		size = structSize + fieldsSize
		debugPrint("%s: Struct size(%d) + fields(%d) = %d", path, structSize, fieldsSize, size)
		return size

	default:
		size = uint64(unsafe.Sizeof(v.Interface()))
		debugPrint("%s: Basic type size %d", path, size)
		return size
	}
}
