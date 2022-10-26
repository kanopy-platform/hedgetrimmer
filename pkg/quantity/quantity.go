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

// Returns x/y rounded up to the specified scale s.
// Note that inf.Scale is inverted. -2 means 100, 2 means 0.01.
func Div(x resource.Quantity, y resource.Quantity, s inf.Scale) resource.Quantity {
	result := resource.Quantity{}
	result.Format = x.Format

	result.AsDec().QuoRound(x.AsDec(), y.AsDec(), s, inf.RoundUp)
	return result
}

func DivFloat64(x resource.Quantity, y float64, s inf.Scale) (resource.Quantity, error) {
	yQuantity, err := resource.ParseQuantity(fmt.Sprintf("%v", y))
	if err != nil {
		return x, err
	}

	return Div(x, yQuantity, s), nil
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

func Ptr(q resource.Quantity) *resource.Quantity {
	return &q
}

// Rounds up input q to the nearest BinarySI representation Mi/Ki.
func RoundUpBinarySI(q resource.Quantity) resource.Quantity {
	qCopy := q.DeepCopy()

	// Note the implementation cannot use resource.RoundUp(), it operates using DecimalSI.
	if qCopy.Cmp(Ten_MiB) == 1 {
		qCopy = roundUp(qCopy, One_MiB)
	} else {
		qCopy = roundUp(qCopy, One_KiB)
	}

	qCopy.Format = resource.BinarySI
	return qCopy
}

// Performs integer division to round up q to the given unit.
func roundUp(q resource.Quantity, unit resource.Quantity) resource.Quantity {
	return Mul(Div(q, unit, 0), unit)
}
