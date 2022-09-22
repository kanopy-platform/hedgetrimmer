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
		msg    string
	}{
		{
			inputA: resource.MustParse("31.123456Gi"),
			inputB: resource.MustParse("27.654321Gi"),
			want:   resource.MustParse("58.777777Gi"),
			msg:    "Add(parsed BinarySI, parsed BinarySI)",
		},
		{
			// binary + decimal = binary
			inputA: *resource.NewQuantity(11, resource.BinarySI),
			inputB: *resource.NewQuantity(34, resource.DecimalSI),
			want:   *resource.NewQuantity(45, resource.BinarySI),
			msg:    "Add(BinarySI, DecimalSI)",
		},
		{
			// 1200 + 0.01 = 1200.01
			// inf.Scale is negated, -2 means *100, 2 means *0.01
			inputA: *resource.NewDecimalQuantity(*inf.NewDec(12, -2), resource.DecimalSI),
			inputB: *resource.NewDecimalQuantity(*inf.NewDec(1, 2), resource.DecimalSI),
			want:   *resource.NewDecimalQuantity(*inf.NewDec(120001, 2), resource.DecimalSI),
			msg:    "Add(DecimalSI, DecimalSI with decimal digits)",
		},
		{
			// -10 + 0.00028 = -9.99972
			inputA: *resource.NewScaledQuantity(-1, 1),
			inputB: *resource.NewScaledQuantity(28, -5),
			want:   *resource.NewScaledQuantity(-999972, -5),
			msg:    "Add(negative DecimalSI, DecimalSI with decimal digits)",
		},
		{
			// [max(int64) x10^max(int32)] + [max(int64) x10^max(int32)]
			inputA: *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			inputB: *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			want:   *resource.NewDecimalQuantity(*inf.NewDecBig(bigInt, -2147483647), resource.DecimalSI),
			msg:    "Add(overflow testing)",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, Add(test.inputA, test.inputB), test.msg)
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
		msg    string
	}{
		{
			inputA: resource.MustParse("31Gi"),
			inputB: resource.MustParse("27.2Gi"),
			want:   resource.MustParse("3.8Gi"),
			msg:    "Sub(parsed BinarySI, parsed BinarySI)",
		},
		{
			// binary - decimal = binary
			inputA: resource.MustParse("55Ki"),
			inputB: resource.MustParse("5000"),
			want:   *resource.NewQuantity(51320, resource.BinarySI),
			msg:    "Sub(BinarySI, DecimalSI)",
		},
		{
			// 1200 - 0.01 = 1199.99
			// inf.Scale is negated, -2 means *100, 2 mean *0.01
			inputA: *resource.NewDecimalQuantity(*inf.NewDec(12, -2), resource.DecimalSI),
			inputB: *resource.NewDecimalQuantity(*inf.NewDec(1, 2), resource.DecimalSI),
			want:   *resource.NewDecimalQuantity(*inf.NewDec(119999, 2), resource.DecimalSI),
			msg:    "Sub(DecimalSI, DecimalSI with decimal digits)",
		},
		{
			// 10 - 10.00028 = -0.00028
			inputA: *resource.NewScaledQuantity(1, 1),
			inputB: *resource.NewScaledQuantity(1000028, -5),
			want:   *resource.NewScaledQuantity(-28, -5),
			msg:    "Sub(DecimalSI, DecimalSI with decimal digits), negative result",
		},
		{
			// [min(int64) x10^max(int32)] - [max(int64) x10^max(int32)]
			inputA: *resource.NewScaledQuantity(-9223372036854775808, 2147483647),
			inputB: *resource.NewScaledQuantity(9223372036854775807, 2147483647),
			want:   *resource.NewDecimalQuantity(*inf.NewDecBig(bigInt, -2147483647), resource.DecimalSI),
			msg:    "Sub(overflow testing)",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, Sub(test.inputA, test.inputB), test.msg)
	}
}

func TestMul(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA resource.Quantity
		inputB resource.Quantity
		// roundingPrecision is the decimal place to compare want vs result due to loss of precision
		roundingPrecision int32
		want              resource.Quantity
		msg               string
	}{
		{
			inputA:            resource.MustParse("31.123456Gi"),
			inputB:            resource.MustParse("8.123456789"),
			want:              *resource.NewDecimalQuantity(*inf.NewDec(2714741989849547521, 7), resource.BinarySI),
			roundingPrecision: -7, // 0.0000001 precision
			msg:               "Mul(parsed BinarySI, parsed DecimalSI)",
		},
		{
			// 12345600 * -0.003456 = -42666.3936
			inputA:            *resource.NewDecimalQuantity(*inf.NewDec(123456, -2), resource.DecimalSI),
			inputB:            *resource.NewDecimalQuantity(*inf.NewDec(-3456, 6), resource.DecimalSI),
			want:              *resource.NewDecimalQuantity(*inf.NewDec(-426663936, 4), resource.DecimalSI),
			roundingPrecision: -4, // 0.0001 precision
			msg:               "Mul(DecimalSI, negative DecimalSI with decimal digits)",
		},
		{
			inputA:            resource.MustParse("518Ti"),
			inputB:            resource.MustParse("1234.56789"),
			want:              *resource.NewDecimalQuantity(*inf.NewDec(7031444666729507272, 1), resource.BinarySI),
			roundingPrecision: -1, // 0.1 precision
			msg:               "Mul(overflow testing)",
		},
	}

	for _, test := range tests {
		result := Mul(test.inputA, test.inputB)
		result.RoundUp(resource.Scale(test.roundingPrecision))
		test.want.RoundUp(resource.Scale(test.roundingPrecision))

		assert.Equal(t, test.want, result, test.msg)
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
		msg               string
	}{
		{
			inputA:            resource.MustParse("501Mi"),
			inputB:            -0.123456,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(-64855952326656, 6), resource.BinarySI),
			roundingPrecision: -6, // 0.000001 precision
			msg:               "MulFloat64(parsed BinarySI, negative float)",
		},
		{
			inputA:            resource.MustParse("31Gi"),
			inputB:            2.525873,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(84076199948582912, 6), resource.BinarySI),
			roundingPrecision: -6, // 0.000001 precision
			msg:               "MulFloat64(parsed BinarySI, positive float)",
		},
		{
			inputA:            resource.MustParse("31Gi"),
			inputB:            10000 / 7.0,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(475514236344, -2), resource.BinarySI),
			roundingPrecision: 2, // 10^2 precision
			msg:               "MulFloat64(parsed BinarySI, repeating decimal)",
		},
		{
			inputA:            resource.MustParse("5Ei"),
			inputB:            95.05,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(5479259450644040254, -2), resource.BinarySI),
			roundingPrecision: 2, // 10^2 precision
			msg:               "MulFloat64(overflow testing)",
		},
	}

	for _, test := range tests {
		result, err := MulFloat64(test.inputA, test.inputB)
		if test.wantError {
			assert.Error(t, err)
		} else {
			result.RoundUp(resource.Scale(test.roundingPrecision))
			test.want.RoundUp(resource.Scale(test.roundingPrecision))

			assert.Equal(t, test.want, result, test.msg)
		}
	}
}

func TestDiv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA resource.Quantity
		inputB resource.Quantity
		// roundingPrecision is the decimal place to compare want vs result due to loss of precision
		roundingPrecision int32
		want              resource.Quantity
		msg               string
	}{
		{
			inputA:            resource.MustParse("31.123456Gi"),
			inputB:            resource.MustParse("8.123456789"),
			want:              *resource.NewDecimalQuantity(*inf.NewDec(4113834452825049, 6), resource.BinarySI),
			roundingPrecision: -6,
			msg:               "Div(parsed BinarySI, parsed DecimalSI)",
		},
		{
			// 12345600 / -0.003456 = -3572222222.222222...
			inputA:            *resource.NewDecimalQuantity(*inf.NewDec(123456, -2), resource.DecimalSI),
			inputB:            *resource.NewDecimalQuantity(*inf.NewDec(-3456, 6), resource.DecimalSI),
			want:              *resource.NewDecimalQuantity(*inf.NewDec(-3572222222222223, 6), resource.DecimalSI),
			roundingPrecision: -6,
			msg:               "Div(DecimalSI, negative DecimalSI with decimal digits)",
		},
		{
			inputA:            resource.MustParse("518Ti"),
			inputB:            resource.MustParse("0.123456789"),
			want:              *resource.NewDecimalQuantity(*inf.NewDec(4613330929803852262, 3), resource.BinarySI),
			roundingPrecision: -3,
			msg:               "Div(overflow testing)",
		},
	}

	for _, test := range tests {
		result := Div(test.inputA, test.inputB)
		result.RoundUp(resource.Scale(test.roundingPrecision))
		test.want.RoundUp(resource.Scale(test.roundingPrecision))

		assert.Equal(t, test.want, result, test.msg)
	}
}

func TestDivFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA resource.Quantity
		inputB float64
		want   resource.Quantity
		// roundingPrecision is the decimal place to compare want vs result due to loss of precision
		roundingPrecision int32
		wantError         bool
		msg               string
	}{
		{
			inputA:            resource.MustParse("256Mi"),
			inputB:            -1 / 6.0,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(-1610612736, 0), resource.BinarySI),
			roundingPrecision: 1, // precision to 10s
			msg:               "DivFloat64(parsed BinarySI, negative repeating decimal)",
		},
		{
			inputA:            resource.MustParse("31Gi"),
			inputB:            2.525873,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(13178016688, 0), resource.BinarySI),
			roundingPrecision: 0,
			msg:               "DivFloat64(parsed BinarySI, positive float)",
		},
		{
			inputA:            resource.MustParse("5128Ti"),
			inputB:            10.0 / 3333.0,
			want:              *resource.NewDecimalQuantity(*inf.NewDec(1879243932557534823, 0), resource.BinarySI),
			roundingPrecision: 12, // precision to 10^12
			msg:               "DivFloat64(overflow testing)",
		},
	}

	for _, test := range tests {
		result, err := DivFloat64(test.inputA, test.inputB)
		if test.wantError {
			assert.Error(t, err)
		} else {
			result.RoundUp(resource.Scale(test.roundingPrecision))
			test.want.RoundUp(resource.Scale(test.roundingPrecision))

			assert.Equal(t, test.want, result, test.msg)
		}
	}
}

func TestMin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA resource.Quantity
		inputB resource.Quantity
		want   resource.Quantity
		msg    string
	}{
		{
			inputA: resource.MustParse("256Mi"),
			inputB: resource.MustParse("256.123Mi"),
			want:   *Ptr(resource.MustParse("256Mi")).ToDec(), // for large values it stores as Dec
			msg:    "Min(256Mi, 256.123Mi) = 256Mi",
		},
		{
			inputA: resource.MustParse("0.12345Gi"),
			inputB: resource.MustParse("5Ki"),
			want:   resource.MustParse("5Ki"),
			msg:    "Min(0.12345Gi, 5Ki) = 5Ki",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, Min(test.inputA, test.inputB), test.msg)
	}
}

func TestMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		inputA resource.Quantity
		inputB resource.Quantity
		want   resource.Quantity
		msg    string
	}{
		{
			inputA: resource.MustParse("256Mi"),
			inputB: resource.MustParse("256.123Mi"),
			want:   *Ptr(resource.MustParse("256.123Mi")).ToDec(), // for large values it stores as Dec
			msg:    "Max(256Mi, 256.123Mi) = 256.123Mi",
		},
		{
			inputA: resource.MustParse("0.12345Gi"),
			inputB: resource.MustParse("5Ki"),
			want:   *Ptr(resource.MustParse("0.12345Gi")).ToDec(),
			msg:    "Max(0.12345Gi, 5Ki) = 0.12345Gi",
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.want, Max(test.inputA, test.inputB), test.msg)
	}
}
