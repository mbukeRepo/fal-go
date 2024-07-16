package fal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type ValueErrorDetail struct {
	Type    string   `json:"type"`
	Message string   `json:"msg"`
	Loc     []string `json:"loc"`
}

type ValueError struct {
	Detail []ValueErrorDetail `json:"detail"`
}

type APIErrorDetail struct {
	Detail string `json:"detail"`
}

type APIError struct {
	Status int         `json:"status"`
	Detail interface{} `json:"detail"`
}

func (e *APIError) Error() string {
	components := []string{}

	if detail, ok := e.Detail.(*APIErrorDetail); ok {
		components = append(components, fmt.Sprintf("status: %d", e.Status))
		components = append(components, fmt.Sprintf("detail: %s", detail.Detail))
	}

	if detail, ok := e.Detail.(*ValueError); ok {
		for _, v := range detail.Detail {
			components = append(components, fmt.Sprintf("status: %d", e.Status))
			components = append(components, fmt.Sprintf("type: %s", v.Type))
			components = append(components, fmt.Sprintf("message: %s", v.Message))
			components = append(components, fmt.Sprintf("location: %s", strings.Join(v.Loc, ", ")))
		}
	}

	output := strings.Join(components, ": ")

	if output == "" {
		output = "unknown error"
	}

	return output
}

func unmarshalAPIError(resp *http.Response, data []byte) *APIError {
	apiErr := &APIError{}
	apiErr.Status = resp.StatusCode

	var apiErrorDetail APIErrorDetail
	if err := json.Unmarshal(data, &apiErrorDetail); err == nil {
		apiErr.Detail = &apiErrorDetail
		return apiErr
	}

	var valueErrorDetail ValueError
	if err := json.Unmarshal(data, &valueErrorDetail); err == nil {
		apiErr.Detail = &valueErrorDetail
		return apiErr
	}

	apiErr.Detail = &APIErrorDetail{
		Detail: "Unknown error occurred",
	}

	return apiErr
}
