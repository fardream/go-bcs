package bcs_test

import (
	"fmt"
	"testing"

	"github.com/fardream/go-bcs/bcs"
)

func BenchmarkDecodeSlice(b *testing.B) {
	type TestStruct struct {
		Value int32
		Name  string
	}

	sizes := []int{16, 256, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Create test data with specified size
			testData := make([]TestStruct, size)
			for i := 0; i < size; i++ {
				testData[i] = TestStruct{
					Value: int32(i),
					Name:  fmt.Sprintf("item_%d", i),
				}
			}

			// Marshal the data once
			encoded, err := bcs.Marshal(testData)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var result []TestStruct
				_, err := bcs.Unmarshal(encoded, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeString(b *testing.B) {
	sizes := []int{16, 256, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Create test string of specified size
			testString := make([]byte, size)
			for i := range testString {
				testString[i] = byte('a' + (i % 26))
			}

			// Marshal the string once
			encoded, err := bcs.Marshal(string(testString))
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var result string
				_, err := bcs.Unmarshal(encoded, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkDecodeByteSlice(b *testing.B) {
	sizes := []int{16, 256, 4096}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
			// Create test byte slice of specified size
			testBytes := make([]byte, size)
			for i := range testBytes {
				testBytes[i] = byte(i % 256)
			}

			// Marshal the byte slice once
			encoded, err := bcs.Marshal(testBytes)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var result []byte
				_, err := bcs.Unmarshal(encoded, &result)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}