package bcs_test

import (
	"testing"

	"github.com/fardream/go-bcs/bcs"
)

func BenchmarkDecodeSlice(b *testing.B) {
	type TestStruct struct {
		Value int32
		Name  string
	}

	// Create test data with a slice of structs
	testData := []TestStruct{
		{Value: 1, Name: "first"},
		{Value: 2, Name: "second"},
		{Value: 3, Name: "third"},
		{Value: 4, Name: "fourth"},
		{Value: 5, Name: "fifth"},
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
}

func BenchmarkDecodeString(b *testing.B) {
	// Create test string
	testString := "This is a test string for benchmarking BCS string deserialization"

	// Marshal the string once
	encoded, err := bcs.Marshal(testString)
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
}

func BenchmarkDecodeByteSlice(b *testing.B) {
	// Create test byte slice
	testBytes := make([]byte, 256)
	for i := range testBytes {
		testBytes[i] = byte(i)
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
}