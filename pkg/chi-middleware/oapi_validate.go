// Package middleware implements middleware function for go-chi or net/http,
// which validates incoming HTTP requests to make sure that they conform to the given OAPI 3.0 specification.
// When OAPI validation failes on the request, we return an HTTP/400.
package middleware

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/clinia/oapi-codegen/pkg/codegen"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

// ErrorHandler is called when there is an error in validation
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error, statusCode int)

// MultiErrorHandler is called when oapi returns a MultiError type
type MultiErrorHandler func(openapi3.MultiError) (int, error)

// Options to customize request validation, openapi3filter specified options will be passed through.
type Options struct {
	Options           openapi3filter.Options
	ErrorHandler      ErrorHandler
	MultiErrorHandler MultiErrorHandler
	// SilenceServersWarning allows silencing a warning for https://github.com/clinia/oapi-codegen/issues/882 that reports when an OpenAPI spec has `spec.Servers != nil`
	SilenceServersWarning bool
	// If true, the next handler will be called even when validation fails
	ContinueOnError bool
}

// OapiRequestValidator Creates middleware to validate request by swagger spec.
// This middleware is good for net/http either since go-chi is 100% compatible with net/http.
func OapiRequestValidator(swagger *openapi3.T) func(next http.Handler) http.Handler {
	return OapiRequestValidatorWithOptions(swagger, nil)
}

// OapiRequestValidatorWithOptions Creates middleware to validate request by swagger spec
// against an OpenAPI 3 specification.

// shouldWarnAboutServers checks if a warning about the `Servers` field in the OpenAPI spec should be logged.
func shouldWarnAboutServers(swagger *openapi3.T, options *Options) bool {
	return swagger.Servers != nil && (options == nil || options.SilenceServersWarning)
}

// This middleware is good for net/http either since go-chi is 100% compatible with net/http.
//
// Parameters:
//   - swagger: Pointer to an OpenAPI 3 specification object
//   - options: Optional configuration parameters for the validator
//
// Returns a middleware function that can be used with HTTP handlers.
//
// The middleware performs the following:
//   - Validates requests against the provided OpenAPI specification
//   - Supports URL-encoded path parameters
//   - Handles validation errors through a custom error handler if provided in options
//   - Can continue processing despite validation errors if ContinueOnError is set
//
// Warning: If the OpenAPI spec includes Servers configuration, the middleware performs
// Host header validation which may result in 400 Bad Request responses for otherwise valid requests.
// This behavior can be silenced by setting Options.SilenceServersWarning to true.
//
// The returned middleware function wraps the next handler in the chain and performs validation
// before passing the request through.
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) func(next http.Handler) http.Handler {
	if shouldWarnAboutServers(swagger, options) {
		log.Println("WARN: OapiRequestValidatorWithOptions called with an OpenAPI spec that has `Servers` set. This may lead to an HTTP 400 with `no matching operation was found` when sending a valid request, as the validator performs `Host` header validation. If you're expecting `Host` header validation, you can silence this warning by setting `Options.SilenceServersWarning = true`. See https://github.com/clinia/oapi-codegen/issues/882 for more information.")
	}

	swagger.Paths = SwaggerPathsToGorillaPaths(swagger.Paths)

	router, err := gorillamux.NewRouter(swagger)
	if err != nil {
		panic(err)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Decode the path to support encoded path variables
			// This is a hack to avoid modifying the openapi3filter dependency
			// Reference: [ENG-1396]
			r.URL.RawPath, _ = url.QueryUnescape(r.URL.RawPath)

			// validate request
			statusCode, err := validateRequest(r, router, options)
			if err != nil {
				if options != nil && options.ErrorHandler != nil {
					options.ErrorHandler(w, r, err, statusCode)
				} else {
					http.Error(w, err.Error(), statusCode)
				}

				// In some instances, we want to continue processing the request
				// even if validation fails. This is useful for logging or
				// debugging purposes, or when the caller wants to handle the
				// error in a different way.
				if options == nil || !options.ContinueOnError {
					return
				}
			}

			// serve
			next.ServeHTTP(w, r)
		})
	}

}

func SwaggerPathsToGorillaPaths(paths *openapi3.Paths) *openapi3.Paths {
	newPaths := &openapi3.Paths{}
	for path, pathItem := range paths.Map() {
		newPaths.Set(codegen.SwaggerUriToGorillaUri(path), pathItem)
	}
	return newPaths
}

// This function is called from the middleware above and actually does the work
// of validating a request.
func validateRequest(r *http.Request, router routers.Router, options *Options) (int, error) {

	// Find route
	route, pathParams, err := router.FindRoute(r)
	if err != nil {
		return http.StatusBadRequest, err // We failed to find a matching route for the request.
	}

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    r,
		PathParams: pathParams,
		Route:      route,
	}

	if options != nil {
		requestValidationInput.Options = &options.Options
	}

	if err := openapi3filter.ValidateRequest(context.Background(), requestValidationInput); err != nil {
		me := openapi3.MultiError{}
		if errors.As(err, &me) && options.Options.MultiError {
			errFunc := getMultiErrorHandlerFromOptions(options)
			return errFunc(me)
		}

		switch e := err.(type) {
		case *openapi3filter.RequestError:
			// We've got a bad request
			// Remove reference to schema
			e.Reason = ""
			// Split up the verbose error by lines and return the first one
			// openapi errors seem to be multi-line with a decent message on the first
			errorLines := strings.Split(e.Error(), "\n")
			return http.StatusBadRequest, fmt.Errorf(errorLines[0])
		case *openapi3filter.SecurityRequirementsError:
			return http.StatusUnauthorized, err
		default:
			// This should never happen today, but if our upstream code changes,
			// we don't want to crash the server, so handle the unexpected error.
			return http.StatusInternalServerError, fmt.Errorf("error validating route: %s", err.Error())
		}
	}

	return http.StatusOK, nil
}

// attempt to get the MultiErrorHandler from the options. If it is not set,
// return a default handler
func getMultiErrorHandlerFromOptions(options *Options) MultiErrorHandler {
	if options == nil {
		return defaultMultiErrorHandler
	}

	if options.MultiErrorHandler == nil {
		return defaultMultiErrorHandler
	}

	return options.MultiErrorHandler
}

// defaultMultiErrorHandler returns a StatusBadRequest (400) and a list
// of all the errors. This method is called if there are no other
// methods defined on the options.
func defaultMultiErrorHandler(me openapi3.MultiError) (int, error) {
	return http.StatusBadRequest, me
}
