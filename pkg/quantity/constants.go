package quantity

import "k8s.io/apimachinery/pkg/api/resource"

var (
	One_KiB resource.Quantity = resource.MustParse("1Ki")
	One_MiB resource.Quantity = resource.MustParse("1Mi")
	Ten_MiB resource.Quantity = resource.MustParse("10Mi")
)
