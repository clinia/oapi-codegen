// Package issue_832 provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/clinia/oapi-codegen version (devel) DO NOT EDIT.
package issue_832

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Defines values for Document_Status.
const (
	Four  Document_Status = "four"
	One   Document_Status = "one"
	Three Document_Status = "three"
	Two   Document_Status = "two"
)

// Document defines model for Document.
type Document struct {
	Name   *string          `json:"name,omitempty"`
	Status *Document_Status `json:"status,omitempty"`
}

// Document_Status defines model for Document.status.
type Document_Status string

// DocumentStatus defines model for DocumentStatus.
type DocumentStatus struct {
	Value *string `json:"value,omitempty"`
}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/7ySz0/rMAzH/5XK7x27tq/vljMIIYQ47AgIhdRbM9o4Styxaer/jpxuReOHxIlL4zq2",
	"vx/HPoCh3pNDxxHUAaJpsdfJvCAz9OhYbB/IY2CL6cbpHuXkvUdQEDlYt4Yxh8iahxSCbuhB3QM5hBz4",
	"leTbBpS/FQ0BHvMP6TnsFmtaiHMxCcwET8up7jjOSfS8QcOieQpaztrnsFvdDV/Rfq4lLutWJMENRhOs",
	"Z0sOFNzqF8ziEDDjVnMW0Awh2i1mUiFmOmDWatd02GSTeLd/cNKx5U4UcKd730nvWwxxqlkVVfFPGiCP",
	"TnsLCv4XVVFDDl5zm9jLU6I6wBrTJKS6FqzrBhRcTvdXyJBDwOjJxantuqrkMOT4OEPtfWdNyi03URhO",
	"4xbrb8AVKPhTvu9DeVyGct6E9ETnT3N3I94xn1nrH8DWv0E7L813zOP4FgAA//8tucJ2/gIAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
