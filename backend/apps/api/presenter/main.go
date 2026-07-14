package presenter

type SuccessBody struct {
	Message string `json:"message" doc:"Response message"`
	Success bool   `json:"success" doc:"Response status"`
	Data    any    `json:"data,omitempty" doc:"Response data"`
}

type SuccessResponse struct {
	Body SuccessBody
}

type ErrorData map[string]any

type ErrorResponse struct {
	Message string      `json:"message" doc:"Response message"`
	Success bool        `json:"success" doc:"Response status"`
	Errors  []ErrorData `json:"errors" doc:"Error data"`
}
