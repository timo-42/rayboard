package shared

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func Operation(method string, path string, tag string, summary string) huma.Operation {
	return huma.Operation{
		OperationID: operationID(method, path),
		Method:      method,
		Path:        path,
		Tags:        []string{tag},
		Summary:     summary,
		Security:    SecurityForMethod(method),
		Responses: map[string]*huma.Response{
			"default": {
				Description: "Error response",
				Content: map[string]*huma.MediaType{
					"application/json": {Schema: &huma.Schema{Type: "object"}},
				},
			},
		},
	}
}

func SecurityForMethod(method string) []map[string][]string {
	return []map[string][]string{
		{"bearerToken": {}},
		{"sessionCookie": {}},
	}
}

func OperationWithStatus(method string, path string, tag string, summary string, status int) huma.Operation {
	op := Operation(method, path, tag, summary)
	op.DefaultStatus = status
	return op
}

func PublicOperation(method string, path string, tag string, summary string) huma.Operation {
	op := Operation(method, path, tag, summary)
	op.Security = []map[string][]string{}
	return op
}

func PublicOperationWithStatus(method string, path string, tag string, summary string, status int) huma.Operation {
	op := PublicOperation(method, path, tag, summary)
	op.DefaultStatus = status
	return op
}

func operationID(method string, path string) string {
	pathID := strings.Trim(path, "/")
	pathID = strings.TrimPrefix(pathID, "api/")
	pathID = strings.ReplaceAll(pathID, "{", "by-")
	pathID = strings.ReplaceAll(pathID, "}", "")
	pathID = strings.ReplaceAll(pathID, "/", "-")
	pathID = strings.ReplaceAll(pathID, "_", "-")
	return strings.ToLower(method) + "-" + pathID
}

func Mutating(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return false
	default:
		return true
	}
}
