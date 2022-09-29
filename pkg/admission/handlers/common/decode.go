package common

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Decode struct {
	Decoder *admission.Decoder
}

func (d *Decode) InjectDecoder(decoder *admission.Decoder) error {
	if decoder == nil {
		return fmt.Errorf("decoder cannot be nil")
	}
	d.Decoder = decoder
	return nil
}
