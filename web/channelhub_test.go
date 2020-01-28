package channelhub

import (
	"fmt"
	"testing"
)

func TestRights(t *testing.T) {
	rights := "rw-w"
	if rights[3:4] == "w" {
		//has write rights
		fmt.Println("has write rights", rights[3:4])
	} else {
		fmt.Println("no write rights", rights[3:4])
	}

	r := ""
	for len(r) < 4 {
		r += "-"
	}

	fmt.Println(r)
}

func TestOwnerChannel(t *testing.T) {
	fmt.Println()
}
