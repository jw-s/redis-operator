package cr

import "k8s.io/apimachinery/pkg/api/errors"

func ResourceAlreadyExistError(err error) bool {
	return errors.IsAlreadyExists(err)
}

func ResourceNotFoundError(err error) bool {
	return errors.IsNotFound(err)
}
