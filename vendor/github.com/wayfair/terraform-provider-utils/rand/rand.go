// Package rand contains randomization functions for use in
// unit/regression/integration tests or anywhere where randomized data
// is needed.
package rand

import (
	"fmt"
	"math/rand"
	"time"
)

// Constants for commonly used alphabets to provide to String()
const (
	Lower      = "abcdefghijklmnopqrstuvwxyz"
	Upper      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Digit      = "0123456789"
	Whitespace = "\t\r\n "
	Special    = "`~!@#$%^&*()-_=+[]{}\\|:;'\"/?,.<>"
)

// String creates a random string of n characters with the supplied alphabet.
// It panics if the alphabet is the empty string, or the length is negative.
func String(n int, alphabet string) string {
	if alphabet == "" || n < 0 {
		panic("invalid argument to String")
	}
	runeAlphabet := []rune(alphabet)
	runeAlphabetLen := len(runeAlphabet)
	bStr := make([]rune, n)
	for i := 0; i < n; i++ {
		bStr[i] = runeAlphabet[rand.Intn(runeAlphabetLen)]
	}
	return string(bStr)
}

// Time generates a random time.Time between 1970-01-00 and 2070-01-00,
// inclusive.
func Time() time.Time {
	UnixDateStart := time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	UnixDateEnd := time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
	delta := UnixDateEnd - UnixDateStart
	sec := rand.Int63n(delta) + UnixDateStart
	return time.Unix(sec, 0)
}

// IntArrayUnique creates an array of size n filled with randomly generated
// integers.  Each entry in the array will be unique. It panics if the length
// is negative.
func IntArrayUnique(n int) []int {
	if n < 0 {
		panic("Invalid argument to IntArrayUnique")
	}
	// hash table used to determine uniqueness.  The go compiler will optimize
	// and the struct{} will not require additional space. The map is used
	// to determine if a value has already been chosen with a constant time
	// lookup operation
	valMap := make(map[int]struct{}, n)
	arr := make([]int, n)
	for idx, _ := range arr {
		// keep generating a random int until it is no longer contained in the
		// array
		var i int
		for {
			i = rand.Int()
			if _, ok := valMap[i]; !ok {
				break
			}
		}
		// add the unique int to the array & mark it as contained in the array
		// by adding a map entry
		arr[idx] = i
		valMap[i] = struct{}{}
	}
	return arr
}

// IPv4 network mask constants
const (
	// Class A network mask: 255.0.0.0 => /8
	IPv4ClassAMask uint32 = 0xFF000000
	// Class B network mask: 255.255.0.0 => /16
	IPv4ClassBMask uint32 = 0xFFFF0000
	// CLass C network mask: 255.255.255.0 => /24
	IPv4ClassCMask uint32 = 0xFFFFFF00
	// Private class B network mask: 255.240.0.0 => /12
	IPv4PrivateClassBMask uint32 = 0xFFF00000
	// Private Class C network mask: 255.255.0.0 => /16
	IPv4PrivateClassCMask uint32 = 0xFFFF0000
)

// IPv4 network host starting address constants
const (
	// Class A IP address start: 1.0.0.0
	IPv4ClassAStart uint32 = 0x01000000
	// Class B IP address start: 128.0.0.0
	IPv4ClassBStart uint32 = 0x80000000
	// Class C IP address start: 192.0.0.0
	IPv4ClassCStart uint32 = 0xC0000000
	// Class D IP address start: 224.0.0.0
	IPv4ClassDStart uint32 = 0xE0000000
	// Class E IP address start: 240.0.0.0
	IPv4ClassEStart uint32 = 0xF0000000
	// Class A private IP address start: 10.0.0.0
	IPv4PrivateClassAStart uint32 = 0x0A000000
	// Class B private IP address start: 172.16.0.0
	IPv4PrivateClassBStart uint32 = 0xAC100000
	// Class C private IP address start: 192.168.0.0
	IPv4PrivateClassCStart uint32 = 0xC0A80000
)

// IPv4 generates a random IPv4 address beginning with the starting IP address
// and going up to the broadcast address for that network (inclusive). The
// ending IP address is determined based on the network mask.
func IPv4(startingIP uint32, mask uint32) uint32 {
	// Perform bitwise XOR on the mask.  This will get the inverse of the
	// mask, indicating our IP range. If our delta is 0, return the starting
	// IP address to avoid divide by 0
	var delta uint32 = mask ^ 0xFFFFFFFF
	if delta == 0 {
		return startingIP
	}
	// Generating a random IP between the starting addreses and the
	// ending IP address
	return startingIP + (rand.Uint32() % delta)
}

// IPv4 generates a random IPv4 address and returns the IP address as a string
// in dotted quad notation. Inputs function similarly to IPv4().
func IPv4Str(startingIP uint32, mask uint32) string {
	ip := IPv4(startingIP, mask)
	return fmt.Sprintf(
		"%d.%d.%d.%d",
		(ip>>24)&0xFF,
		(ip>>16)&0xFF,
		(ip>>8)&0xFF,
		ip&0xFF,
	)
}

// MACAddr48 generates a random 48 bit MAC address
func MACAddr48() uint64 {
	// generate a random, unsigned 64 bit number to use as our MAC address
	var mac uint64 = rand.Uint64()
	// Perform bitwise AND on the mask for the MAC (48 bits), this will zero
	// out the leading 16 bits
	return mac & 0x0000FFFFFFFFFFFF
}

// MACAddr48Str generates a random 48 bit MAC address as a string, using the
// desired separator between octets.
func MACAddr48Str(sep string) string {
	mac := MACAddr48()
	return fmt.Sprintf(
		"%x%s%x%s%x%s%x%s%x%s%x",
		(mac>>40)&0xFF,
		sep,
		(mac>>32)&0xFF,
		sep,
		(mac>>24)&0xFF,
		sep,
		(mac>>16)&0xFF,
		sep,
		(mac>>8)&0xFF,
		sep,
		mac&0xFF,
	)
}

// MACAddr64 generates a random 64 bit MAC address
func MACAddr64() uint64 {
	return rand.Uint64()
}

// MACAddr64Str generates a random 64 bit MAC address as a string, using the
// desired separator between octets.
func MACAddr64Str(sep string) string {
	mac := MACAddr48()
	return fmt.Sprintf(
		"%x%s%x%s%x%s%x%s%x%s%x%s%x%s%x",
		(mac>>56)&0xFF,
		sep,
		(mac>>48)&0xFF,
		sep,
		(mac>>40)&0xFF,
		sep,
		(mac>>32)&0xFF,
		sep,
		(mac>>24)&0xFF,
		sep,
		(mac>>16)&0xFF,
		sep,
		(mac>>8)&0xFF,
		sep,
		mac&0xFF,
	)
}
