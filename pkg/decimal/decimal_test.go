package decimal

import (
	"testing"
)

func TestNewFromString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123.4567", "123.4567"},
		{"-123.4567", "-123.4567"},
		{"0.0001", "0.0001"},
		{"100", "100.0000"},
		{"0", "0.0000"},
	}

	for _, tt := range tests {
		d, err := NewFromString(tt.input)
		if err != nil {
			t.Errorf("NewFromString(%s) error: %v", tt.input, err)
			continue
		}
		if d.String() != tt.expected {
			t.Errorf("NewFromString(%s) = %s, want %s", tt.input, d.String(), tt.expected)
		}
	}
}

func TestFromFloat(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{123.4567, "123.4567"},
		{-123.4567, "-123.4567"},
		{0.0001, "0.0001"},
		{100.0, "100.0000"},
	}

	for _, tt := range tests {
		d := NewFromFloat(tt.input)
		if d.String() != tt.expected {
			t.Errorf("NewFromFloat(%f) = %s, want %s", tt.input, d.String(), tt.expected)
		}
	}
}

func TestAdd(t *testing.T) {
	d1, _ := NewFromString("100.5000")
	d2, _ := NewFromString("200.3000")
	result := d1.Add(d2)
	expected := "300.8000"
	if result.String() != expected {
		t.Errorf("Add: %s + %s = %s, want %s", d1, d2, result, expected)
	}
}

func TestSub(t *testing.T) {
	d1, _ := NewFromString("300.8000")
	d2, _ := NewFromString("100.5000")
	result := d1.Sub(d2)
	expected := "200.3000"
	if result.String() != expected {
		t.Errorf("Sub: %s - %s = %s, want %s", d1, d2, result, expected)
	}
}

func TestMul(t *testing.T) {
	d1, _ := NewFromString("100.0000")
	d2, _ := NewFromString("0.0500")
	result := d1.Mul(d2)
	expected := "5.0000"
	if result.String() != expected {
		t.Errorf("Mul: %s * %s = %s, want %s", d1, d2, result, expected)
	}
}

func TestDiv(t *testing.T) {
	d1, _ := NewFromString("1000.0000")
	d2, _ := NewFromString("3.0000")
	result := d1.Div(d2)
	// 1000 / 3 = 333.3333 (四舍五入到4位小数)
	expected := "333.3333"
	if result.String() != expected {
		t.Errorf("Div: %s / %s = %s, want %s", d1, d2, result, expected)
	}
}

func TestRound(t *testing.T) {
	d, _ := NewFromString("123.4567")
	
	tests := []struct {
		places   int
		expected string
	}{
		{0, "123.0000"},
		{1, "123.5000"},
		{2, "123.4600"},
		{3, "123.4570"},
		{4, "123.4567"},
	}

	for _, tt := range tests {
		result := d.Round(tt.places)
		if result.String() != tt.expected {
			t.Errorf("Round(%d) = %s, want %s", tt.places, result, tt.expected)
		}
	}
}

func TestCmp(t *testing.T) {
	d1, _ := NewFromString("100.0000")
	d2, _ := NewFromString("200.0000")
	d3, _ := NewFromString("100.0000")

	if d1.Cmp(d2) != -1 {
		t.Errorf("100 < 200 should be true")
	}
	if d2.Cmp(d1) != 1 {
		t.Errorf("200 > 100 should be true")
	}
	if d1.Cmp(d3) != 0 {
		t.Errorf("100 == 100 should be true")
	}
}

func TestAbs(t *testing.T) {
	d, _ := NewFromString("-123.4567")
	result := d.Abs()
	expected := "123.4567"
	if result.String() != expected {
		t.Errorf("Abs(%s) = %s, want %s", d, result, expected)
	}
}

func TestNeg(t *testing.T) {
	d, _ := NewFromString("123.4567")
	result := d.Neg()
	expected := "-123.4567"
	if result.String() != expected {
		t.Errorf("Neg(%s) = %s, want %s", d, result, expected)
	}
}

func TestFloor(t *testing.T) {
	d, _ := NewFromString("123.4567")
	result := d.Floor()
	expected := "123.0000"
	if result.String() != expected {
		t.Errorf("Floor(%s) = %s, want %s", d, result, expected)
	}
}

func TestCeil(t *testing.T) {
	d, _ := NewFromString("123.4567")
	result := d.Ceil()
	expected := "124.0000"
	if result.String() != expected {
		t.Errorf("Ceil(%s) = %s, want %s", d, result, expected)
	}
}

func TestSum(t *testing.T) {
	d1, _ := NewFromString("100.0000")
	d2, _ := NewFromString("200.0000")
	d3, _ := NewFromString("300.0000")
	result := Sum(d1, d2, d3)
	expected := "600.0000"
	if result.String() != expected {
		t.Errorf("Sum = %s, want %s", result, expected)
	}
}

func TestAvg(t *testing.T) {
	d1, _ := NewFromString("100.0000")
	d2, _ := NewFromString("200.0000")
	d3, _ := NewFromString("300.0000")
	result := Avg(d1, d2, d3)
	expected := "200.0000"
	if result.String() != expected {
		t.Errorf("Avg = %s, want %s", result, expected)
	}
}

func TestZero(t *testing.T) {
	result := Zero()
	expected := "0.0000"
	if result.String() != expected {
		t.Errorf("Zero = %s, want %s", result, expected)
	}
}

func TestOne(t *testing.T) {
	result := One()
	expected := "1.0000"
	if result.String() != expected {
		t.Errorf("One = %s, want %s", result, expected)
	}
}

func TestFloat64(t *testing.T) {
	d, _ := NewFromString("123.4567")
	result := d.Float64()
	expected := 123.4567
	if result != expected {
		t.Errorf("Float64 = %f, want %f", result, expected)
	}
}

func TestInt64(t *testing.T) {
	d, _ := NewFromString("123.4567")
	result := d.Int64()
	expected := int64(123)
	if result != expected {
		t.Errorf("Int64 = %d, want %d", result, expected)
	}
}