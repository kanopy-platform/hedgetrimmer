package quantity

import "k8s.io/apimachinery/pkg/api/resource"

var (
	OneKi resource.Quantity = resource.MustParse("1Ki")
	OneMi resource.Quantity = resource.MustParse("1Mi")
	TenMi resource.Quantity = resource.MustParse("10Mi")
)
