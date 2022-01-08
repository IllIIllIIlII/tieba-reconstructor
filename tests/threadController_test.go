package test

import (
	"testing"
	thread "tieba-reconstructor/utils/tiezi"
)

func TestController(t *testing.T) {
	c := thread.NewThreadController()
	c.Start()

}
