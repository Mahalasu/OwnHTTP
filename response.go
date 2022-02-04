package tritonhttp

import (
	"errors"
	"io"
	"os"
	"sort"
	"strconv"
)

type Response struct {
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"

	// Header stores all headers to write to the response.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	// Request is the valid request that leads to this response.
	// It could be nil for responses not resulting from a valid request.
	Request *Request

	// FilePath is the local path to the file to serve.
	// It could be "", which means there is no file to serve.
	FilePath string
}

// Write writes the res to the w.
func (res *Response) Write(w io.Writer) error {
	if err := res.WriteStatusLine(w); err != nil {
		return err
	}
	if err := res.WriteSortedHeaders(w); err != nil {
		return err
	}
	if err := res.WriteBody(w); err != nil {
		return err
	}
	return nil
}

// WriteStatusLine writes the status line of res to w, including the ending "\r\n".
// For example, it could write "HTTP/1.1 200 OK\r\n".
func (res *Response) WriteStatusLine(w io.Writer) error {
	var result string
	if res.StatusCode == 200 {
		result = res.Proto + " " + strconv.Itoa(res.StatusCode) + " " + "OK"
	}
	if res.StatusCode == 400 {
		result = res.Proto + " " + strconv.Itoa(res.StatusCode) + " " + "Bad Request"
	}
	if res.StatusCode == 404 {
		result = res.Proto + " " + strconv.Itoa(res.StatusCode) + " " + "Not Found"
	}
	result += "\r\n"
	_, err := w.Write([]byte(result))
	if err != nil {
		return err
	}
	return nil
}

// WriteSortedHeaders writes the headers of res to w, including the ending "\r\n".
// For example, it could write "Connection: close\r\nDate: foobar\r\n\r\n".
// For HTTP, there is no need to write headers in any particular order.
// TritonHTTP requires to write in sorted order for the ease of testing.
func fileLength(filePath string) string {
	f, _ := os.Stat(filePath)
	return strconv.Itoa(int(f.Size()))
}

func (res *Response) WriteSortedHeaders(w io.Writer) error {
	var keys []string
	for k := range res.Header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		header := CanonicalHeaderKey(k)
		value := res.Header[k]
		result := header + ": " + value + "\r\n"
		_, err := w.Write([]byte(result))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

// WriteBody writes res' file content as the response body to w.
// It doesn't write anything if there is no file to serve.
func (res *Response) WriteBody(w io.Writer) error {
	if res.StatusCode == 400 {
		return errors.New("no need to return body")
	}
	f, err := os.Open(res.FilePath)
	if err != nil {
		return nil
	}

	defer f.Close()

	for {
		var buf = make([]byte, 1)
		_, err := f.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		_, err1 := w.Write([]byte(buf))
		if err1 != nil {
			return err1
		}
	}
	return nil
}
