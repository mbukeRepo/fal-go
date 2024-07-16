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
	Status  int      `json:"status"`
}

type APIErrorDetail struct {
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

type APIError struct {
	Detail interface{} `json:"detail"`
}

func (e *APIError) Error() string {
	components := []string{}

	if detail, ok := e.Detail.(*APIErrorDetail); ok {
		components = append(components, fmt.Sprintf("status: %d", detail.Status))
		components = append(components, fmt.Sprintf("detail: %s", detail.Detail))
	}

	if detail, ok := e.Detail.(*ValueErrorDetail); ok {
		components = append(components, fmt.Sprintf("status: %d", detail.Status))
		components = append(components, fmt.Sprintf("type: %s", detail.Type))
		components = append(components, fmt.Sprintf("message: %s", detail.Message))
		components = append(components, fmt.Sprintf("location: %s", strings.Join(detail.Loc, ", ")))
	}

	output := strings.Join(components, ": ")

	if output == "" {
		output = "unknown error"
	}

	return output
}

func unmarshalAPIError(resp *http.Response, data []byte) *APIError {
	apiErr := &APIError{}

	var apiErrorDetail APIErrorDetail
	if err := json.Unmarshal(data, &apiErrorDetail); err == nil {
		apiErrorDetail.Status = resp.StatusCode
		apiErr.Detail = &apiErrorDetail
		return apiErr
	}

	var valueErrorDetail ValueErrorDetail
	if err := json.Unmarshal(data, &valueErrorDetail); err == nil {
		valueErrorDetail.Status = resp.StatusCode
		apiErr.Detail = &valueErrorDetail
		return apiErr
	}

	apiErr.Detail = &APIErrorDetail{
		Status: resp.StatusCode,
		Detail: "Unknown error occurred",
	}

	return apiErr
}
