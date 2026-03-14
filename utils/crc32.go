package utils

import (
	"fmt"
	"hash/crc32"
)

// CRC32 computes a CRC32 checksum of data using the IEEE polynomial, which is
// the standard polynomial used by most CRC32 implementations including the
// FastReport .NET Crc32 class.
func CRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// CRC32String computes a CRC32 checksum of data and returns it as an
// eight-character lowercase hexadecimal string.
func CRC32String(data []byte) string {
	return fmt.Sprintf("%08x", CRC32(data))
}
