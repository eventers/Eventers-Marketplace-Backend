package healthcheck

import (
	"io"
	"net/http"
)

// Self writes an alive message to the ResponseWriter.
func Self(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "I am alive")
}
