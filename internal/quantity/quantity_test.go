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
		inputA       resource.Quantity
		inputB       resource.Quantity
		want         resource.Quantity
		wantOverflow bool
	}{
		{
			inputA:       *resource.NewQuantity(15000, resource.BinarySI),
			inputB:       *resource.NewQuantity(17000, resource.BinarySI),
			want:         *resource.NewQuantity(32000, resource.BinarySI),
			wantOverflow: false,
		},
		{
			// binary + decimal = binary
			inputA:       *resource.NewQuantity(11, resource.BinarySI),
			inputB:       *resource.NewQuantity(34, resource.DecimalSI),
			want:         *resource.NewQuantity(45, resource.BinarySI),
			wantOverflow: false,
		},
		{
			// 1200 + 0.01 = 1200.01
			// inf.Scale is negated, -2 means *100, 2 means *0.01
			inputA:       *resource.NewDecimalQuantity(*inf.NewDec(12, -2), resource.DecimalSI),
			inputB:       *resource.NewDecimalQuantity(*inf.NewDec(1, 2), resource.DecimalSI),
			want:         *resource.NewDecimalQuantity(*inf.NewDec(120001, 2), resource.DecimalSI),
			wantOverflow: false,
		},
		{
			// -10 + 0.00028 = -9.99972
			inputA:       *resource.NewScaledQuantity(-1, 1),
			inputB:       *resource.NewScaledQuantity(28, -5),
			want:         *resource.NewScaledQuantity(-999972, -5),
			wantOverflow: false,
		},
		{
			// overflow testing, max int64 is 9.223372036854775807x10^18
			// use 123400*10^18 + 987600*10^18
			inputA:       *resource.NewScaledQuantity(123400, resource.Exa),
			inputB:       *resource.NewScaledQuantity(987600, resource.Exa),
			want:         *resource.NewScaledQuantity(1111000, resource.Exa),
			wantOverflow: false,
		},
		{
			// demonstrates for practical uses, cannot overflow
			// [max(int64) x10^max(int32)] + [max(int64) x10^max(int32)]
			inputA:       *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			inputB:       *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			want:         *resource.NewDecimalQuantity(*inf.NewDecBig(bigInt, -2147483647), resource.DecimalSI),
			wantOverflow: false,
		},
	}

	for _, test := range tests {
		result, overflow := Add(test.inputA, test.inputB)

		assert.Equal(t, test.want, result)
		assert.Equal(t, test.wantOverflow, overflow)
	}
}

func TestSub(t *testing.T) {
	t.Parallel()

	// big.Int can represent numbers smaller than min(int64)
	bigInt := big.NewInt(0).SetUint64(18446744073709551615)
	bigInt = bigInt.Neg(bigInt)

	tests := []struct {
		inputA       resource.Quantity
		inputB       resource.Quantity
		want         resource.Quantity
		wantOverflow bool
	}{
		{
			inputA:       resource.MustParse("31Gi"),
			inputB:       resource.MustParse("27.2Gi"),
			want:         resource.MustParse("3.8Gi"),
			wantOverflow: false,
		},
		{
			// binary - decimal = binary
			inputA:       resource.MustParse("55Ki"),
			inputB:       resource.MustParse("5000"),
			want:         *resource.NewQuantity(51320, resource.BinarySI),
			wantOverflow: false,
		},
		{
			// 1200 - 0.01 = 1199.99
			// inf.Scale is negated, -2 means *100, 2 mean *0.01
			inputA:       *resource.NewDecimalQuantity(*inf.NewDec(12, -2), resource.DecimalSI),
			inputB:       *resource.NewDecimalQuantity(*inf.NewDec(1, 2), resource.DecimalSI),
			want:         *resource.NewDecimalQuantity(*inf.NewDec(119999, 2), resource.DecimalSI),
			wantOverflow: false,
		},
		{
			// 10 - 10.00028 = -0.00028
			inputA:       *resource.NewScaledQuantity(1, 1),
			inputB:       *resource.NewScaledQuantity(1000028, -5),
			want:         *resource.NewScaledQuantity(-28, -5),
			wantOverflow: false,
		},
		{
			// shouldn't overflow, min int64 is -9.223372036854775808x10^18
			// use 123400*10^18 - 987600*10^18
			inputA:       *resource.NewScaledQuantity(123400, resource.Exa),
			inputB:       *resource.NewScaledQuantity(987600, resource.Exa),
			want:         *resource.NewScaledQuantity(-864200, resource.Exa),
			wantOverflow: false,
		},
		{
			// demonstrates for practical uses, cannot overflow
			// [min(int64) x10^max(int32)] - [max(int64) x10^max(int32)]
			inputA:       *resource.NewScaledQuantity(-9223372036854775808, 2147483647),
			inputB:       *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			want:         *resource.NewDecimalQuantity(*inf.NewDecBig(bigInt, -2147483647), resource.DecimalSI),
			wantOverflow: false,
		},
	}

	for _, test := range tests {
		result, overflow := Sub(test.inputA, test.inputB)

		assert.Equal(t, test.want, result)
		assert.Equal(t, test.wantOverflow, overflow)
	}
}

func TestMulFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA    resource.Quantity
		inputB    float64
		want      resource.Quantity
		wantError error
	}{
		{
			inputA:    resource.MustParse("31Gi"),
			inputB:    2.525873,
			want:      *resource.NewDecimalQuantity(*inf.NewDec(84076199948582912, 6), resource.BinarySI),
			wantError: nil,
		},
	}

	for _, test := range tests {
		result, err := MulFloat64(test.inputA, test.inputB)

		assert.Equal(t, test.want, result)
		assert.Equal(t, test.wantError, err)
	}
}

func TestDivFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA            resource.Quantity
		inputB            float64
		roundingPrecision int32
		want              resource.Quantity
		wantError         error
	}{
		{
			inputA:            resource.MustParse("31Gi"),
			inputB:            2.525873,
			roundingPrecision: 3, // KiB precision
			want:              *resource.NewDecimalQuantity(*inf.NewDec(13178016687, 0), resource.BinarySI),
			wantError:         nil,
		},
	}

	for _, test := range tests {
		result, err := DivFloat64(test.inputA, test.inputB)

		result.RoundUp(resource.Scale(test.roundingPrecision))
		test.want.RoundUp(resource.Scale(test.roundingPrecision))

		assert.Equal(t, test.want, result)
		assert.Equal(t, test.wantError, err)
	}
}
