// Package head_digit_of_operation_id provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/clinia/oapi-codegen version (devel) DO NOT EDIT.
package head_digit_of_operation_id

import (
	"context"
	"net/http"
)

type N3GPPFooRequestObject struct {
}

type N3GPPFooResponseObject interface {
	VisitN3GPPFooResponse(w http.ResponseWriter) error
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (GET /3gpp/foo)
	N3GPPFoo(ctx context.Context, request N3GPPFooRequestObject) (N3GPPFooResponseObject, error)
}
