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

// Returns x/y rounded to the specified scale s.
// Note that inf.Scale is inverted. -2 means 100, 2 means 0.01.
func Div(x resource.Quantity, y resource.Quantity, s inf.Scale, rounder inf.Rounder) resource.Quantity {
	result := resource.Quantity{}
	result.Format = x.Format

	result.AsDec().QuoRound(x.AsDec(), y.AsDec(), s, rounder)
	return result
}

func DivFloat64(x resource.Quantity, y float64, s inf.Scale, rounder inf.Rounder) (resource.Quantity, error) {
	yQuantity, err := resource.ParseQuantity(fmt.Sprintf("%v", y))
	if err != nil {
		return x, err
	}

	return Div(x, yQuantity, s, rounder), nil
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

// Rounds input q to the nearest BinarySI representation Mi/Ki.
func RoundBinarySI(q resource.Quantity, rounder inf.Rounder) resource.Quantity {
	qCopy := q.DeepCopy()

	// Note the implementation cannot use resource.RoundUp(), it operates using DecimalSI.
	if qCopy.Cmp(TenMi) == 1 {
		qCopy = round(qCopy, OneMi, rounder)
	} else {
		qCopy = round(qCopy, OneKi, rounder)
	}

	qCopy.Format = resource.BinarySI
	return qCopy
}

// Performs integer division to round up q to the given unit.
func round(q resource.Quantity, unit resource.Quantity, rounder inf.Rounder) resource.Quantity {
	return Mul(Div(q, unit, 0, rounder), unit)
}
