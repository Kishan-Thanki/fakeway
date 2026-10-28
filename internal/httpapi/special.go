package httpapi

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

func handleCreatedStatus(
	w http.ResponseWriter,
	r *http.Request,
	code int,
	statusText string,
) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Use POST or PUT to simulate a 201 Created response.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	location, ok := responseLocation(r, "/resources/1")
	if !ok {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Location must be a path on the Fakeway server.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	w.Header().Set("Location", location)

	writeJSON(w, r, code, "application/json", StatusResponse{
		Code:     code,
		Status:   statusText,
		Location: location,
	})
}

func handleNoContentStatus(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func handleResetContentStatus(w http.ResponseWriter, code int) {
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(code)
}

func handlePartialContentStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Partial content requires a GET or HEAD request.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	if r.Header.Get("Range") == "" {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Send a Range header to simulate a 206 Partial Content response.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	http.ServeContent(
		w,
		r,
		"fakeway.txt",
		time.Time{},
		strings.NewReader(FAKE_STATUS_REPRESENTATION),
	)
}

func handleMultipleChoicesStatus(
	w http.ResponseWriter,
	r *http.Request,
	code int,
	statusText string,
) {
	location, ok := responseLocation(r, "/health")
	if !ok {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Location must be a path on the Fakeway server.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	alternative := "/version"
	if location == alternative {
		alternative = "/health"
	}

	w.Header().Set("Location", location)

	writeJSON(w, r, code, "application/json", StatusResponse{
		Code:         code,
		Status:       statusText,
		Location:     location,
		Alternatives: []string{location, alternative},
	})
}

func handleRedirectStatus(
	w http.ResponseWriter,
	r *http.Request,
	code int,
) {
	location, ok := responseLocation(r, "/health")
	if !ok {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Location must be a path on the Fakeway server.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	http.Redirect(w, r, location, code)
}

func handleNotModifiedStatus(
	w http.ResponseWriter,
	r *http.Request,
	code int,
) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "A 304 response requires a conditional GET or HEAD request.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	if !matchesFakeETag(r.Header.Get("If-None-Match")) {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:   "about:blank",
				Title:  statusDescription(http.StatusBadRequest),
				Status: http.StatusBadRequest,
				Detail: "Send If-None-Match: " + fakeStatusETag() +
					" to simulate a matching 304 response.",
				Instance: r.URL.Path,
			},
		)
		return
	}

	w.Header().Set("ETag", fakeStatusETag())
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(code)
}

func handleRequestedRangeNotSatisfiableStatus(
	w http.ResponseWriter,
	r *http.Request,
) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Range responses require GET or HEAD.",
				Instance: r.URL.Path,
			},
		)
		return true
	}

	rangeHeader := r.Header.Get("Range")
	if rangeHeader == "" {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Send an unsatisfiable Range header to simulate a 416 response.",
				Instance: r.URL.Path,
			},
		)
		return true
	}

	unsatisfiable, valid := parseUnsatisfiableByteRange(
		rangeHeader,
		int64(len(FAKE_STATUS_REPRESENTATION)),
	)
	if !valid {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "Use a valid single byte range to simulate a 416 response.",
				Instance: r.URL.Path,
			},
		)
		return true
	}

	if !unsatisfiable {
		writeJSON(
			w,
			r,
			http.StatusBadRequest,
			"application/problem+json",
			ProblemResponse{
				Type:     "about:blank",
				Title:    statusDescription(http.StatusBadRequest),
				Status:   http.StatusBadRequest,
				Detail:   "The supplied range is satisfiable. Use a range outside the mock representation, such as bytes=999-.",
				Instance: r.URL.Path,
			},
		)
		return true
	}

	w.Header().Set(
		"Content-Range",
		"bytes */"+strconv.Itoa(len(FAKE_STATUS_REPRESENTATION)),
	)

	return false
}

func prepareStatusHeaders(
	w http.ResponseWriter,
	r *http.Request,
	code int,
) {
	switch code {
	case http.StatusMethodNotAllowed:
		w.Header().Set("Allow", allowedMethodsFor405(r.Method))

	case http.StatusRequestTimeout:
		if r.ProtoMajor == 1 {
			w.Header().Set("Connection", "close")
		}

	case http.StatusUnauthorized:
		w.Header().Set("WWW-Authenticate", `Basic realm="Fakeway"`)

	case http.StatusProxyAuthRequired:
		w.Header().Set("Proxy-Authenticate", `Basic realm="Fakeway"`)

	case http.StatusUpgradeRequired:
		w.Header().Set("Upgrade", "websocket")

		if r.ProtoMajor == 1 {
			w.Header().Set("Connection", "Upgrade")
		}

	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		w.Header().Set("Retry-After", "60")
	}
}
