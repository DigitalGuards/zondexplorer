package utils

import (
	"math/big"
	"strings"
)

// HexToInt converts a hex string to a *big.Int
func HexToInt(hex string) *big.Int {
	if len(hex) <= 2 {
		return big.NewInt(0)
	}
	hex = strings.TrimPrefix(hex, "0x")
	n := new(big.Int)
	n.SetString(hex, 16)
	return n
}

// CompareHexNumbers compares two hex numbers, returns:
// -1 if a < b
//
//	0 if a == b
//	1 if a > b
func CompareHexNumbers(a, b string) int {
	aInt := HexToInt(a)
	bInt := HexToInt(b)
	return aInt.Cmp(bInt)
}

// AddHexNumbers adds two hex numbers and returns the result as a hex string
func AddHexNumbers(a, b string) string {
	aInt := HexToInt(a)
	bInt := HexToInt(b)
	result := new(big.Int).Add(aInt, bInt)
	if result.Sign() == 0 {
		return "0x0"
	}
	return "0x" + result.Text(16)
}

// SubtractHexNumbers subtracts two hex numbers and returns the result as a hex string
func SubtractHexNumbers(a, b string) string {
	aInt := HexToInt(a)
	bInt := HexToInt(b)
	result := new(big.Int).Sub(aInt, bInt)
	if result.Sign() == 0 {
		return "0x0"
	}
	return "0x" + result.Text(16)
}

// IntToHex converts an int to a hex string
func IntToHex(n int) string {
	if n == 0 {
		return "0x0"
	}
	return "0x" + new(big.Int).SetInt64(int64(n)).Text(16)
}
