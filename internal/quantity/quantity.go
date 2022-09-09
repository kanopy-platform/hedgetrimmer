package quantity

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Add returns the result of x+y.
// If true is returned, the operation overflowed and the result will be equal to x.
func Add(x resource.Quantity, y resource.Quantity) (result resource.Quantity, overflow bool) {
	overflow = false
	result = x.DeepCopy()
	result.Add(y)

	if !y.IsZero() && result.Cmp(x) == 0 {
		// TODO add to PR comment it seems overflow is "possible"
		// https://github.com/kubernetes/apimachinery/blob/master/pkg/api/resource/quantity.go#L573
		// however, like in the _test.go, I don't think it's possible for us to cause it
		overflow = true
	}

	return
}

// Add returns the result of x-y.
// If true is returned, the operation overflowed and the result will be equal to x.
func Sub(x resource.Quantity, y resource.Quantity) (result resource.Quantity, overflow bool) {
	overflow = false
	result = x.DeepCopy()
	result.Sub(y)

	if !y.IsZero() && result.Cmp(x) == 0 {
		overflow = true
	}

	return
}

func Mul(x resource.Quantity, y resource.Quantity) resource.Quantity {
	result := resource.Quantity{}
	result.Format = x.Format

	result.AsDec().Mul(x.AsDec(), y.AsDec())
	return result
}

func MulFloat64(x resource.Quantity, y float64) (resource.Quantity, error) {
	yQuantity, err := resource.ParseQuantity(fmt.Sprintf("%v", y))
	if err != nil {
		return x, err
	}

	return Mul(x, yQuantity), nil
}

func DivFloat64(x resource.Quantity, y float64) (resource.Quantity, error) {
	yInverse := 1 / y
	yQuantity, err := resource.ParseQuantity(fmt.Sprintf("%v", yInverse))
	if err != nil {
		return x, err
	}

	return Mul(x, yQuantity), nil
}
