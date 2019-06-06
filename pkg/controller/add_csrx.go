package controller

import (
	"github.com/michaelhenkel/csrx-operator/pkg/controller/csrx"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, csrx.Add)
}
