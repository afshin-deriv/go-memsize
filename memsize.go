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

// getValueAddr returns the memory address of a reflect.Value if possible
func getValueAddr(v reflect.Value) uintptr {
	switch v.Kind() {
	case reflect.Ptr, reflect.UnsafePointer:
		if v.IsNil() {
			return 0
		}
		return uintptr(v.UnsafePointer())
	case reflect.Interface:
		if v.IsNil() {
			return 0
		}
		return uintptr(unsafe.Pointer(reflect.ValueOf(v.Interface()).Pointer()))
	case reflect.Slice, reflect.Map:
		if v.IsNil() {
			return 0
		}
		return v.Pointer()
	default:
		if v.CanAddr() {
			return uintptr(unsafe.Pointer(v.UnsafeAddr()))
		}
		return 0
	}
}

// GetTotalSize returns the total memory size including indirect allocations
func GetTotalSize(v interface{}) uint64 {
	runtime.GC()

	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	initialHeap := stats.HeapAlloc

	size := getTotalSize(reflect.ValueOf(v), make(visited), "root")

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

	// Get address if possible to track visited values
	addr := getValueAddr(v)
	if addr != 0 {
		if seen[addr] {
			debugPrint("%s: Already seen address %x", path, addr)
			return 0
		}
		seen[addr] = true
	}

	var size uint64

	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			size = uint64(unsafe.Sizeof(v.Interface()))
			debugPrint("%s: Nil interface, size %d", path, size)
			return size
		}
		size = getTotalSize(v.Elem(), seen, path+".elem")
		debugPrint("%s: Interface elem size %d", path, size)
		return size

	case reflect.Ptr:
		if v.IsNil() {
			size = uint64(unsafe.Sizeof(v.Interface()))
			debugPrint("%s: Nil pointer, size %d", path, size)
			return size
		}
		ptrSize := uint64(unsafe.Sizeof(v.Interface()))
		elemSize := getTotalSize(v.Elem(), seen, path+".ptr")
		size = ptrSize + elemSize
		debugPrint("%s: Pointer (size: %d) + elem (size: %d) = %d", path, ptrSize, elemSize, size)
		return size

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
			fieldSize := getTotalSize(field, seen, fmt.Sprintf("%s.%s", path, v.Type().Field(i).Name))
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
