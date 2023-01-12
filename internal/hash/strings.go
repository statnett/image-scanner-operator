package hash

import (
	"crypto/md5" //#nosec G501 -- Weak cryptographic primitive only used for hashing
	"fmt"
)

func NewString(objs ...interface{}) string {
	digester := md5.New() //#nosec G401 -- Weak cryptographic primitive only used for hashing
	for _, v := range objs {
		digester.Write([]byte(fmt.Sprintf("%s", v)))
	}
	return fmt.Sprintf("%x", digester.Sum(nil))
}
