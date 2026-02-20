package calculator

import "testing"

func TestCacheKey(t *testing.T) {
	tests := []struct {
		name      string
		number1   float64
		number2   float64
		operation string
		want      string
	}{
		{
			name:      "сложение целых",
			number1:   10,
			number2:   5,
			operation: "+",
			want:      "10 + 5",
		},
		{
			name:      "вычитание целых",
			number1:   100,
			number2:   50,
			operation: "-",
			want:      "100 - 50",
		},
		{
			name:      "умножение с дробными",
			number1:   3.14,
			number2:   2,
			operation: "*",
			want:      "3.14 * 2",
		},
		{
			name:      "деление с дробным результатом",
			number1:   1,
			number2:   3,
			operation: "/",
			want:      "1 / 3",
		},
		{
			name:      "отрицательные числа",
			number1:   -10,
			number2:   -5,
			operation: "+",
			want:      "-10 + -5",
		},
		{
			name:      "ноль",
			number1:   0,
			number2:   0,
			operation: "+",
			want:      "0 + 0",
		},
		{
			name:      "большие числа",
			number1:   1000000,
			number2:   999999,
			operation: "-",
			want:      "1000000 - 999999",
		},
		{
			name:      "очень маленькое дробное",
			number1:   0.000001,
			number2:   0.000002,
			operation: "+",
			want:      "0.000001 + 0.000002",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cacheKey(tt.number1, tt.number2, tt.operation)
			if got != tt.want {
				t.Errorf("cacheKey(%v, %v, %q) = %q, want %q",
					tt.number1, tt.number2, tt.operation, got, tt.want)
			}
		})
	}
}
