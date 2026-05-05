package nucleus

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/marstid/nuc/pkg/domain"
)

// apiErrorResponse represents the JSON error response from the Nucleus API.
type apiErrorResponse struct {
	Message   string `json:"message"`
	Error     string `json:"error"`
	RequestID string `json:"request_id"`
}

// handleErrorResponse maps an HTTP error response to a domain error.
func handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &domain.APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: failed to read response body", resp.StatusCode),
		}
	}

	var apiErr apiErrorResponse
	if err := json.Unmarshal(body, &apiErr); err != nil {
		// If we can't parse the JSON, use the raw body.
		return &domain.APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}

	msg := apiErr.Message
	if msg == "" {
		msg = apiErr.Error
	}
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	apiError := &domain.APIError{
		StatusCode: resp.StatusCode,
		Message:    msg,
		RequestID:  apiErr.RequestID,
	}

	// Map to sentinel errors for common cases.
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", domain.ErrUnauthorized, msg)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", domain.ErrForbidden, msg)
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", domain.ErrNotFound, msg)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", domain.ErrRateLimited, msg)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w: %s", domain.ErrValidation, msg)
	default:
		return apiError
	}
}
