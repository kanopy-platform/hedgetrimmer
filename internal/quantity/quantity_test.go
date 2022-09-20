package quantity

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/inf.v0"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	// big.Int can represent numbers greater than max(int64)
	bigInt := big.NewInt(0).SetUint64(18446744073709551614)

	tests := []struct {
		inputA resource.Quantity
		inputB resource.Quantity
		want   resource.Quantity
	}{
		{
			inputA: resource.MustParse("31.123456Gi"),
			inputB: resource.MustParse("27.654321Gi"),
			want:   resource.MustParse("58.777777Gi"),
		},
		{
			// binary + decimal = binary
			inputA: *resource.NewQuantity(11, resource.BinarySI),
			inputB: *resource.NewQuantity(34, resource.DecimalSI),
			want:   *resource.NewQuantity(45, resource.BinarySI),
		},
		{
			// 1200 + 0.01 = 1200.01
			// inf.Scale is negated, -2 means *100, 2 means *0.01
			inputA: *resource.NewDecimalQuantity(*inf.NewDec(12, -2), resource.DecimalSI),
			inputB: *resource.NewDecimalQuantity(*inf.NewDec(1, 2), resource.DecimalSI),
			want:   *resource.NewDecimalQuantity(*inf.NewDec(120001, 2), resource.DecimalSI),
		},
		{
			// -10 + 0.00028 = -9.99972
			inputA: *resource.NewScaledQuantity(-1, 1),
			inputB: *resource.NewScaledQuantity(28, -5),
			want:   *resource.NewScaledQuantity(-999972, -5),
		},
		{
			// demonstrates for practical uses, cannot overflow
			// [max(int64) x10^max(int32)] + [max(int64) x10^max(int32)]
			inputA: *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			inputB: *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			want:   *resource.NewDecimalQuantity(*inf.NewDecBig(bigInt, -2147483647), resource.DecimalSI),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, Add(test.inputA, test.inputB))
	}
}

func TestSub(t *testing.T) {
	t.Parallel()

	// big.Int can represent numbers smaller than min(int64)
	bigInt := big.NewInt(0).SetUint64(18446744073709551615)
	bigInt = bigInt.Neg(bigInt)

	tests := []struct {
		inputA resource.Quantity
		inputB resource.Quantity
		want   resource.Quantity
	}{
		{
			inputA: resource.MustParse("31Gi"),
			inputB: resource.MustParse("27.2Gi"),
			want:   resource.MustParse("3.8Gi"),
		},
		{
			// binary - decimal = binary
			inputA: resource.MustParse("55Ki"),
			inputB: resource.MustParse("5000"),
			want:   *resource.NewQuantity(51320, resource.BinarySI),
		},
		{
			// 1200 - 0.01 = 1199.99
			// inf.Scale is negated, -2 means *100, 2 mean *0.01
			inputA: *resource.NewDecimalQuantity(*inf.NewDec(12, -2), resource.DecimalSI),
			inputB: *resource.NewDecimalQuantity(*inf.NewDec(1, 2), resource.DecimalSI),
			want:   *resource.NewDecimalQuantity(*inf.NewDec(119999, 2), resource.DecimalSI),
		},
		{
			// 10 - 10.00028 = -0.00028
			inputA: *resource.NewScaledQuantity(1, 1),
			inputB: *resource.NewScaledQuantity(1000028, -5),
			want:   *resource.NewScaledQuantity(-28, -5),
		},
		{
			// demonstrates for practical uses, cannot overflow
			// [min(int64) x10^max(int32)] - [max(int64) x10^max(int32)]
			inputA: *resource.NewScaledQuantity(-9223372036854775808, 2147483647),
			inputB: *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			want:   *resource.NewDecimalQuantity(*inf.NewDecBig(bigInt, -2147483647), resource.DecimalSI),
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, Sub(test.inputA, test.inputB))
	}
}

func TestMulFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA resource.Quantity
		inputB float64
		want   resource.Quantity
		// roundingPrecision is the decimal place to compare want vs result due to loss of precision
		roundingPrecision int32
		wantError         bool
	}{
		{
			// negative multiplier
			inputA:            resource.MustParse("501Mi"),
			inputB:            -0.123456,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(-64855952326656, 6), resource.BinarySI),
			roundingPrecision: -6, // 0.000001 precision
		},
		{
			// positive multiplier
			inputA:            resource.MustParse("31Gi"),
			inputB:            2.525873,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(84076199948582912, 6), resource.BinarySI),
			roundingPrecision: -6, // 0.000001 precision
		},
		{
			// repeating decimal multiplier
			inputA:            resource.MustParse("31Gi"),
			inputB:            10000 / 7.0,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(475514236344, -2), resource.BinarySI),
			roundingPrecision: 2, // 10^2 precision
		},
		{
			// overflow testing
			inputA:            resource.MustParse("5Ei"),
			inputB:            95.05,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(5479259450644040254, -2), resource.BinarySI),
			roundingPrecision: 2, // 10^2 precision
		},
	}

	for _, test := range tests {
		result, err := MulFloat64(test.inputA, test.inputB)
		if test.wantError {
			assert.Error(t, err)
		} else {
			result.RoundUp(resource.Scale(test.roundingPrecision))
			test.want.RoundUp(resource.Scale(test.roundingPrecision))

			assert.Equal(t, test.want, result)
		}
	}
}

func TestDivFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA            resource.Quantity
		inputB            float64
		want              resource.Quantity
		roundingPrecision int32
		wantError         bool
	}{
		{
			// negative divisor
			inputA:            resource.MustParse("256Mi"),
			inputB:            -1 / 6.0,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(-1610612736, 0), resource.BinarySI),
			roundingPrecision: 0,
		},
		{
			// positive divisor
			inputA:            resource.MustParse("31Gi"),
			inputB:            2.525873,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(13178016687, 0), resource.BinarySI),
			roundingPrecision: 3, // KiB precision
		},
		{
			// repeating decimal divisor
			inputA:            resource.MustParse("5128Ti"),
			inputB:            10.0 / 6.0,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(33829773763411968, 1), resource.BinarySI),
			roundingPrecision: -1, // 0.1 precision
		},
	}

	for _, test := range tests {
		result, err := DivFloat64(test.inputA, test.inputB)
		if test.wantError {
			assert.Error(t, err)
		} else {
			result.RoundUp(resource.Scale(test.roundingPrecision))
			test.want.RoundUp(resource.Scale(test.roundingPrecision))

			assert.Equal(t, test.want, result)
		}
	}
}
