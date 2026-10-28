package httpapi

import (
	"net/http"
)

const FAKE_STATUS_REPRESENTATION = "Fakeway mock response!"

func fakeStatusETag() string {
	return "\"fakeway-status-" + Version + "\""
}

func handleSpecialStatus(
	w http.ResponseWriter,
	r *http.Request,
	code int,
	statusText string,
) bool {
	switch code {
	case http.StatusCreated:
		handleCreatedStatus(w, r, code, statusText)
		return true

	case http.StatusNoContent:
		handleNoContentStatus(w, code)
		return true

	case http.StatusResetContent:
		handleResetContentStatus(w, code)
		return true

	case http.StatusPartialContent:
		handlePartialContentStatus(w, r)
		return true

	case http.StatusMultipleChoices:
		handleMultipleChoicesStatus(w, r, code, statusText)
		return true

	case http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect:
		handleRedirectStatus(w, r, code)
		return true

	case http.StatusNotModified:
		handleNotModifiedStatus(w, r, code)
		return true

	case http.StatusRequestedRangeNotSatisfiable:
		return handleRequestedRangeNotSatisfiableStatus(w, r)
	}

	prepareStatusHeaders(w, r, code)
	return false
}
