package flag_test

import (
	"testing"

	"github.com/Conansgithub/due-private/v2/flag"
)

func TestString(t *testing.T) {
	t.Log(flag.Bool("test.v"))
	t.Log(flag.String("config", "./config"))
}
