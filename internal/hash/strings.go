package hash

import (
	"fmt"

	"github.com/gohugoio/hashstructure"
)

func NewString(objs ...interface{}) (string, error) {
	hash, err := hashstructure.Hash(objs, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash), nil
}
