package httpapi

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func responseLocation(r *http.Request, fallback string) (string, bool) {
	raw := r.URL.Query().Get("location")
	if raw == "" {
		return fallback, true
	}

	if strings.ContainsAny(raw, "\\\r\n") {
		return "", false
	}

	location, err := url.Parse(raw)
	if err != nil ||
		location.IsAbs() ||
		location.Host != "" ||
		!strings.HasPrefix(location.Path, "/") ||
		strings.HasPrefix(raw, "//") {
		return "", false
	}

	return location.String(), true
}

func matchesFakeETag(header string) bool {
	for _, tag := range strings.Split(header, ",") {
		tag = strings.TrimSpace(tag)

		if tag == "*" {
			return true
		}

		tag = strings.TrimPrefix(tag, "W/")
		if tag == fakeStatusETag() {
			return true
		}
	}

	return false
}

func parseUnsatisfiableByteRange(header string, size int64) (bool, bool) {
	unit, value, ok := strings.Cut(strings.TrimSpace(header), "=")
	if !ok || !strings.EqualFold(unit, "bytes") || strings.Contains(value, ",") {
		return false, false
	}

	startText, endText, ok := strings.Cut(value, "-")
	if !ok {
		return false, false
	}

	if startText == "" {
		suffixLength, err := strconv.ParseInt(endText, 10, 64)
		if err != nil || suffixLength <= 0 {
			return false, false
		}

		return size == 0, true
	}

	start, err := strconv.ParseInt(startText, 10, 64)
	if err != nil || start < 0 {
		return false, false
	}

	if endText == "" {
		return start >= size, true
	}

	end, err := strconv.ParseInt(endText, 10, 64)
	if err != nil || end < start {
		return false, false
	}

	return start >= size, true
}

func allowedMethodsFor405(requestMethod string) string {
	methods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodOptions,
	}

	allowed := make([]string, 0, len(methods))

	for _, method := range methods {
		if method == requestMethod {
			continue
		}

		if requestMethod == http.MethodGet && method == http.MethodHead {
			continue
		}
		if requestMethod == http.MethodHead && method == http.MethodGet {
			continue
		}

		allowed = append(allowed, method)
	}

	return strings.Join(allowed, ", ")
}

func statusDetail(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "The request could not be processed because it is invalid."

	case http.StatusUnauthorized:
		return "Authentication is required to access this resource."

	case http.StatusForbidden:
		return "The server understood the request but refuses to authorize it."

	case http.StatusNotFound:
		return "The requested resource was not found."

	case http.StatusMethodNotAllowed:
		return "The request method is not allowed for this resource."

	case http.StatusNotAcceptable:
		return "The server cannot provide a representation acceptable to the client."

	case http.StatusRequestTimeout:
		return "The server timed out waiting for the request."

	case http.StatusConflict:
		return "The request conflicts with the current state of the resource."

	case http.StatusGone:
		return "The requested resource is no longer available."

	case http.StatusLengthRequired:
		return "The request must include a Content-Length header."

	case http.StatusRequestEntityTooLarge:
		return "The request content exceeds the accepted size."

	case http.StatusUnsupportedMediaType:
		return "The request content type is not supported."

	case http.StatusRequestedRangeNotSatisfiable:
		return "The requested byte range is not satisfiable."

	case http.StatusUnprocessableEntity:
		return "The request content could not be processed."

	case http.StatusProxyAuthRequired:
		return "Proxy authentication is required."

	case http.StatusTooManyRequests:
		return "Too many requests have been received."

	case http.StatusInternalServerError:
		return "The server encountered an unexpected condition."

	case http.StatusNotImplemented:
		return "The requested functionality is not supported."

	case http.StatusBadGateway:
		return "The server received an invalid response from an upstream server."

	case http.StatusServiceUnavailable:
		return "The service is temporarily unavailable."

	case http.StatusGatewayTimeout:
		return "The server did not receive a timely response from an upstream server."

	default:
		return "Fakeway is simulating the requested HTTP error response."
	}
}
