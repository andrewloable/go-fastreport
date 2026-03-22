// Package utils — murmur3.go implements Murmur3-based hash helpers ported from
// FastReport.Utils.Crypter.ComputeHash (Utils/Crypter.cs) and default-password
// helpers for EncryptStringDefault / DecryptStringDefault.
package utils

import (
	"encoding/hex"
	"io"
	"strings"
)

const (
	murmur3C1 uint64 = 0x87c37b91114253d5
	murmur3C2 uint64 = 0x4cf5ad432745937f
)

func murmur3MixKey1(k1 uint64) uint64 {
	k1 *= murmur3C1
	k1 = (k1 << 31) | (k1 >> 33)
	k1 *= murmur3C2
	return k1
}

func murmur3MixKey2(k2 uint64) uint64 {
	k2 *= murmur3C2
	k2 = (k2 << 33) | (k2 >> 31)
	k2 *= murmur3C1
	return k2
}

func murmur3MixFinal(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}

// murmur3Hash computes MurmurHash3_x64_128 of bb with seed 0, matching the
// C# Murmur3 class in Utils/Crypter.cs exactly (including 4-byte-per-slot quirk).
func murmur3Hash(bb []byte) []byte {
	var h1, h2 uint64
	length := uint64(0)
	pos := 0
	remaining := uint64(len(bb))

	for remaining >= 16 {
		// C# reads only 4 bytes per 8-byte slot (uint cast from int ops).
		k1 := uint64(uint32(int(bb[pos]) | int(bb[pos+1])<<8 | int(bb[pos+2])<<16 | int(bb[pos+3])<<24))
		pos += 8
		k2 := uint64(uint32(int(bb[pos]) | int(bb[pos+1])<<8 | int(bb[pos+2])<<16 | int(bb[pos+3])<<24))
		pos += 8
		length += 16
		remaining -= 16
		// MixBody
		h1 ^= murmur3MixKey1(k1)
		h1 = (h1 << 27) | (h1 >> 37)
		h1 += h2
		h1 = h1*5 + 0x52dce729
		h2 ^= murmur3MixKey2(k2)
		h2 = (h2 << 31) | (h2 >> 33)
		h2 += h1
		h2 = h2*5 + 0x38495ab5
	}

	if remaining > 0 {
		var k1, k2 uint64
		length += remaining
		switch remaining {
		case 15:
			k2 ^= uint64(bb[pos+14]) << 48
			fallthrough
		case 14:
			k2 ^= uint64(bb[pos+13]) << 40
			fallthrough
		case 13:
			k2 ^= uint64(bb[pos+12]) << 32
			fallthrough
		case 12:
			k2 ^= uint64(bb[pos+11]) << 24
			fallthrough
		case 11:
			k2 ^= uint64(bb[pos+10]) << 16
			fallthrough
		case 10:
			k2 ^= uint64(bb[pos+9]) << 8
			fallthrough
		case 9:
			k2 ^= uint64(bb[pos+8])
			fallthrough
		case 8:
			k1 ^= uint64(uint32(int(bb[pos]) | int(bb[pos+1])<<8 | int(bb[pos+2])<<16 | int(bb[pos+3])<<24))
		case 7:
			k1 ^= uint64(bb[pos+6]) << 48
			fallthrough
		case 6:
			k1 ^= uint64(bb[pos+5]) << 40
			fallthrough
		case 5:
			k1 ^= uint64(bb[pos+4]) << 32
			fallthrough
		case 4:
			k1 ^= uint64(bb[pos+3]) << 24
			fallthrough
		case 3:
			k1 ^= uint64(bb[pos+2]) << 16
			fallthrough
		case 2:
			k1 ^= uint64(bb[pos+1]) << 8
			fallthrough
		case 1:
			k1 ^= uint64(bb[pos])
		}
		h1 ^= murmur3MixKey1(k1)
		h2 ^= murmur3MixKey2(k2)
	}

	h1 ^= length
	h2 ^= length
	h1 += h2
	h2 += h1
	h1 = murmur3MixFinal(h1)
	h2 = murmur3MixFinal(h2)
	h1 += h2
	h2 += h1

	hash := make([]byte, 16)
	for i := range 8 {
		hash[i] = byte(h1 >> (uint(i) * 8))
		hash[8+i] = byte(h2 >> (uint(i) * 8))
	}
	return hash
}

// ComputeHashBytes returns the 32-char uppercase hex Murmur3 hash of input.
// Mirrors C# Crypter.ComputeHash(byte[]).
func ComputeHashBytes(input []byte) string {
	return strings.ToUpper(hex.EncodeToString(murmur3Hash(input)))
}

// ComputeHashString returns the Murmur3 hash of the UTF-8 bytes of input.
// Mirrors C# Crypter.ComputeHash(string).
func ComputeHashString(input string) string {
	return ComputeHashBytes([]byte(input))
}

// ComputeHashReader reads all bytes from r and returns the Murmur3 hash.
// Mirrors C# Crypter.ComputeHash(Stream).
func ComputeHashReader(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return ComputeHashBytes(data), nil
}

// defaultCrypterPassword is the encryption password used when no explicit
// password is provided. Matches C# typeof(Crypter).FullName.
var defaultCrypterPassword = "FastReport.Utils.Crypter"

// SetDefaultPassword changes the default encryption password.
func SetDefaultPassword(pwd string) { defaultCrypterPassword = pwd }

// EncryptStringDefault encrypts plaintext using the default password.
func EncryptStringDefault(plaintext string) (string, error) {
	return EncryptString(plaintext, defaultCrypterPassword)
}

// DecryptStringDefault decrypts a "rij"-prefixed string using the default password.
func DecryptStringDefault(data string) (string, error) {
	return DecryptString(data, defaultCrypterPassword)
}
