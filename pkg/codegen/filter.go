package codegen

import "github.com/getkin/kin-openapi/openapi3"

func filterOperationsByTag(swagger *openapi3.T, opts Configuration) {
	if len(opts.OutputOptions.ExcludeTags) > 0 {
		operationsWithTag(swagger.Paths, opts.OutputOptions.ExcludeTags, true)
	}
	if len(opts.OutputOptions.IncludeTags) > 0 {
		operationsWithTag(swagger.Paths, opts.OutputOptions.IncludeTags, false)
	}
}
func operationsWithTag(paths *openapi3.Paths, tags []string, exclude bool) {
	if paths == nil {
		return
	}

	for _, pathItem := range paths.Map() {
		ops := pathItem.Operations()
		names := make([]string, 0, len(ops))
		for name, op := range ops {
			if operationHasTag(op, tags) == exclude {
				names = append(names, name)
			}
		}
		for _, name := range names {
			pathItem.SetOperation(name, nil)
		}
	}
}

// operationHasTag returns true if the operation is tagged with any of tags
func operationHasTag(op *openapi3.Operation, tags []string) bool {
	if op == nil {
		return false
	}
	for _, hasTag := range op.Tags {
		for _, wantTag := range tags {
			if hasTag == wantTag {
				return true
			}
		}
	}
	return false
}
