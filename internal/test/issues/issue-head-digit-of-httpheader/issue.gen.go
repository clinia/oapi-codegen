// Package headdigitofhttpheader provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/clinia/oapi-codegen version (devel) DO NOT EDIT.
package headdigitofhttpheader

import (
	"context"
	"fmt"
	"net/http"
)

type N200ResponseHeaders struct {
	N000Foo string
}
type N200Response struct {
	Headers N200ResponseHeaders
}

type GetFooRequestObject struct {
}

type GetFooResponseObject interface {
	VisitGetFooResponse(w http.ResponseWriter) error
}

type GetFoo200Response = N200Response

func (response GetFoo200Response) VisitGetFooResponse(w http.ResponseWriter) error {
	w.Header().Set("000-foo", fmt.Sprint(response.Headers.N000Foo))
	w.WriteHeader(200)
	return nil
}

// StrictServerInterface represents all server handlers.
type StrictServerInterface interface {

	// (GET /foo)
	GetFoo(ctx context.Context, request GetFooRequestObject) (GetFooResponseObject, error)
}
