package usecases

import (
	"testing"
)

// Add adds two integers and returns the sum.
func TestAdd(t *testing.T) {
	testCases := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "Positive numbers",
			a:        2,
			b:        3,
			expected: 5,
		},
		{
			name:     "Negative numbers",
			a:        -2,
			b:        -3,
			expected: -5,
		},
		{
			name:     "Zero",
			a:        0,
			b:        0,
			expected: 0,
		},
		{
			name:     "Mixed numbers",
			a:        -2,
			b:        3,
			expected: 1,
		},
		{
			name:     "Large numbers",
			a:        1000000,
			b:        2000000,
			expected: 3000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Add(tc.a, tc.b)
			if got != tc.expected {
				t.Errorf("Add(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.expected)
			}
		})
	}
}

func TestSubstract(t *testing.T) {
	testCases := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "Positive numbers",
			a:        3,
			b:        2,
			expected: 1,
		},
		{
			name:     "Negative numbers",
			a:        -2,
			b:        -3,
			expected: 1,
		},
		{
			name:     "Zero",
			a:        0,
			b:        0,
			expected: 0,
		},
		{
			name:     "Mixed numbers",
			a:        -2,
			b:        3,
			expected: -5,
		},
		{
			name:     "Large numbers",
			a:        1000000,
			b:        2000000,
			expected: -1000000,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Subtract(tc.a, tc.b)
			if got != tc.expected {
				t.Errorf("Subtract(%d, %d) = %d, want %d", tc.a, tc.b, got, tc.expected)
			}
		})
	}
}
