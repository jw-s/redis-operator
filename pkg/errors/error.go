package errors

import (
	"errors"
)

var (
	UnsupportedKubeResource = errors.New("unsupported Kubernetes resource")
)
