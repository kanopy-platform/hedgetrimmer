package quantity

import (
	"fmt"

	"gopkg.in/inf.v0"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Add(x resource.Quantity, y resource.Quantity) resource.Quantity {
	result := x.DeepCopy()
	result.Add(y)
	return result
}

func Sub(x resource.Quantity, y resource.Quantity) resource.Quantity {
	result := x.DeepCopy()
	result.Sub(y)
	return result
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

func Div(x resource.Quantity, y resource.Quantity) resource.Quantity {
	result := resource.Quantity{}
	result.Format = x.Format

	result.AsDec().QuoRound(x.AsDec(), y.AsDec(), 0, inf.RoundUp)
	return result
}

func DivFloat64(x resource.Quantity, y float64) (resource.Quantity, error) {
	yQuantity, err := resource.ParseQuantity(fmt.Sprintf("%v", y))
	if err != nil {
		return x, err
	}

	return Div(x, yQuantity), nil
}
