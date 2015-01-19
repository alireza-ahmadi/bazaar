package bazaar

import (
	"bytes"
	"mime/multipart"
)

// Multipart map structure
type Form map[string]string

// Create body of request based on the map fields, Also return content type
func (f Form) Build() (body *bytes.Buffer, contentType string) {
	body = bytes.NewBuffer(nil)
	w := multipart.NewWriter(body)

	for k, v := range f {
		w.WriteField(k, v)
	}

	contentType = w.FormDataContentType()
	w.Close()
	return
}
