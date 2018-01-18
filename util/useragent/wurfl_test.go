package useragent

import (
	"fmt"
	"testing"
)

func TestWurfl(t *testing.T) {
	model := "SPH-A620"
	brand := GetBrand(model)
	fmt.Printf("Brand: %v\n", brand)
}
