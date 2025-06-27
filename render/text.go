package render

import (
	"fmt"
	"net/http"
)

type Str struct {
	Format string
	Data   []any
}

var textContentType = []string{"text/plain; charset=utf-8"}

func WriteString(w http.ResponseWriter, format string, data []any) (err error) {
	writeContentType(w, textContentType)
	if len(data) > 0 {
		_, err = fmt.Fprintf(w, format, data...)
		return
	}
	_, err = w.Write(StringToBytes(format))
	return
}
