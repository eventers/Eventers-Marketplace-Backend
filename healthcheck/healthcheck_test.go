package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	r := httptest.NewRequest("GET", "/healthcheck", nil)
	w := httptest.NewRecorder()

	Self(w, r)

	assert.Equal(t, http.StatusOK, w.Code, "Invalid HTTP response code")
	assert.Equal(t, "I am alive", w.Body.String(), "Invalid HTTP response body")
}
