package utils

import "hash/crc32"

// HashCode hashes a string to a unique hashcode.
func HashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	return -v
}
