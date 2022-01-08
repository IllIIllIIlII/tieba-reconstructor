package test

import (
	"fmt"
	"testing"
	thread "tieba-reconstructor/utils/tiezi"
)

func TestThread(t *testing.T) {
	fmt.Println("Begin testing thread task generation.")
	thread := thread.New("1766018024", 0, false, false)

	t.Log("floor limit: ", thread.FloorLimit)
}
