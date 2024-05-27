package chessengine

import (
	"unsafe"
)

func sizeOfMap(m map[uint64]uint64) float64 {
	// Start with the size of the map header
	size := unsafe.Sizeof(m)

	// Iterate over the map to sum sizes of keys and values
	for key, value := range m {
		size += unsafe.Sizeof(key) + unsafe.Sizeof(value)
	}

	// Convert the size from bytes to kilobytes
	sizeKB := float64(size) / 1024.0
	return sizeKB
}
