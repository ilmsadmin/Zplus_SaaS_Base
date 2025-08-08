//go:build test_import

package main

import (
	"fmt"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

func main() {
	var reg domain.DomainRegistration
	fmt.Printf("Type: %T\n", reg)
}
