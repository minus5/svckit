package lazy

import "os"

func init() {
	os.Setenv("SVCKIT_DCY_CONSUL", "--")
}
