package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

const NUMBER_OF_IP_ADDRESSES = 256 * 256 * 256 * 256

func parseFileName(args []string) string {
	if len(args) == 2 && args[0] == "-file" {
		return args[1]
	}
	return ""
}

func main() {
	startTime := time.Now()
	fileName := parseFileName(os.Args[1:])
	if fileName == "" {
		fmt.Println("Wrong arguments. Use '-file file_name' to specify file for processing")
		return
	}

	counter := NewBitSetUniqueIpCounter()
	numberOfUniqueIPs, err := counter.CountUniqueIP(fileName)
	if err != nil {
		fmt.Println("Some errors here. Check log for details.")
	} else {
		fmt.Printf("Found %d unique IPs\n", numberOfUniqueIPs)
		duration := time.Since(startTime)
		fmt.Printf("Processing time: %.2f seconds\n", duration.Minutes())
	}

}

// Converts an IP address in string format to a 32-bit integer representation
func ToLongValue(ipString string) (int64, error) {
	fmt.Println(ipString)
	segments := strings.Split(ipString, ".")
	if len(segments) != 4 {
		return 0, errors.New("invalid IP address format")
	}

	var result int64
	for i := 0; i < 4; i++ {
		value, err := strconv.Atoi(segments[i])
		if err != nil {
			return 0, err
		}
		result += int64(value) * int64(math.Pow(256, float64(3-i)))
	}

	return result, nil
}

type BitSetUniqueIpCounter struct {
	bitSetLow []uint64 // Lower half of IP range (0 - 2,147,483,647)
	bitSetHi  []uint64 // Upper half of IP range (2,147,483,648 - 4,294,967,295)
	counter   int64
}

func NewBitSetUniqueIpCounter() *BitSetUniqueIpCounter {
	return &BitSetUniqueIpCounter{
		bitSetLow: make([]uint64, math.MaxInt32/64+1),
		bitSetHi:  make([]uint64, math.MaxInt32/64+1),
		counter:   0,
	}
}

// Helper function to set a bit in the bitmap and return if it was newly set
func setBit(bitmap []uint64, pos int) bool {
	index := pos / 64     // Determine which uint64 element to use
	bit := uint(pos % 64) // Determine the bit position within the uint64
	if (bitmap[index] & (1 << bit)) != 0 {
		return false // Bit already set
	}
	bitmap[index] |= (1 << bit) // Set the bit
	return true                 // Bit was not set previously
}

func (b *BitSetUniqueIpCounter) registerLongValue(longValue int64) {
	var workingSet []uint64
	intValue := int(longValue)

	if longValue > math.MaxInt32 {
		intValue = int(longValue - math.MaxInt32)
		workingSet = b.bitSetHi
	} else {
		workingSet = b.bitSetLow
	}

	// Only increment counter if the bit was not already set
	if setBit(workingSet, intValue) {
		b.counter++
	}
}

// CountUniqueIP processes the file to count unique IPs
func (b *BitSetUniqueIpCounter) CountUniqueIP(fileName string) (int64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Printf("Error opening file: %v\n", err)
		return -1, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ip := scanner.Text()
		longValue, err := ToLongValue(ip)
		if err != nil {
			log.Printf("Error parsing IP address: %v\n", err)
			continue
		}
		b.registerLongValue(longValue)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading file: %v\n", err)
		return -1, err
	}

	return b.counter, nil
}
