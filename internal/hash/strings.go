package hash

import (
	"crypto/md5"
	"fmt"
)

func NewString(objs ...interface{}) string {
	digester := md5.New()
	for _, v := range objs {
		digester.Write([]byte(fmt.Sprintf("%s", v)))
	}
	return fmt.Sprintf("%x", digester.Sum(nil))
}
