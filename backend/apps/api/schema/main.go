package schema

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

var APICheck = huma.Operation{
	// OperationID: "get-api-check",
	Path:        "/check",
	Summary:     "Check",
	Description: "Hit this endpoint to check if this API is up",
	Method:      http.MethodGet,
	Responses: map[string]*huma.Response{
		"200": {
			Description: "OK",
			// Content: map[string]*huma.MediaType{
			// 	"application/json"
			// },
		},
	},
}
