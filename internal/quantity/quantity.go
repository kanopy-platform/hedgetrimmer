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

	result.AsDec().QuoRound(x.AsDec(), y.AsDec(), 6, inf.RoundUp)
	return result
}

func DivFloat64(x resource.Quantity, y float64) (resource.Quantity, error) {
	yQuantity, err := resource.ParseQuantity(fmt.Sprintf("%v", y))
	if err != nil {
		return x, err
	}

	return Div(x, yQuantity), nil
}

func Min(x resource.Quantity, y resource.Quantity) resource.Quantity {
	xCopy := x.DeepCopy()
	yCopy := y.DeepCopy()

	if xCopy.Cmp(yCopy) == -1 {
		return xCopy
	}

	return yCopy
}

func Max(x resource.Quantity, y resource.Quantity) resource.Quantity {
	xCopy := x.DeepCopy()
	yCopy := y.DeepCopy()

	if xCopy.Cmp(yCopy) == -1 {
		return yCopy
	}

	return xCopy
}

// convenience function for testing only
func Ptr(q resource.Quantity) *resource.Quantity {
	return &q
}

// convenience function for testing only
func RoundUp(q resource.Quantity, scale resource.Scale) resource.Quantity {
	rounded := q.DeepCopy()
	rounded.RoundUp(scale)
	return rounded
}
