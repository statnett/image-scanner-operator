package hash

import (
	"fmt"

	"github.com/mitchellh/hashstructure/v2"
)

func NewString(objs ...interface{}) (string, error) {
	hash, err := hashstructure.Hash(objs, hashstructure.FormatV2, nil)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash), nil
}
