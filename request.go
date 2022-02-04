package tritonhttp

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type Request struct {
	Method string // e.g. "GET"
	URL    string // e.g. "/path/to/a/file"
	Proto  string // e.g. "HTTP/1.1"

	// Header stores misc headers excluding "Host" and "Connection",
	// which are stored in special fields below.
	// Header keys are case-incensitive, and should be stored
	// in the canonical format in this map.
	Header map[string]string

	Host  string // determine from the "Host" header
	Close bool   // determine from the "Connection" header
}

// ReadRequest tries to read the next valid request from br.
//
// If it succeeds, it returns the valid request read. In this case,
// bytesReceived should be true, and err should be nil.
//
// If an error occurs during the reading, it returns the error,
// and a nil request. In this case, bytesReceived indicates whether or not
// some bytes are received before the error occurs. This is useful to determine
// the timeout with partial request received condition.
func ReadRequest(br *bufio.Reader) (req *Request, bytesReceived bool, err error) {
	// Read start line
	var request Request
	firstLine, err := ReadLine(br)
	if err != nil {
		return nil, false, err
	}
	firstLineArray := strings.Split(firstLine, " ")
	if len(firstLineArray) != 3 || strings.TrimSpace(firstLineArray[0]) != "GET" || strings.TrimSpace(firstLineArray[2]) != "HTTP/1.1" {
		return nil, true, errors.New("invalid request")
	}
	request.Method = strings.TrimSpace(firstLineArray[0])
	request.URL = strings.TrimSpace(firstLineArray[1])
	request.Proto = strings.TrimSpace(firstLineArray[2])

	// Read headers
	request.Header = make(map[string]string)
	for {
		line, err := ReadLine(br)
		if line == "" {
			break
		}
		if err != nil {
			if err != io.EOF {
				return nil, true, err
			}
			break
		}
		lineArray := strings.Split(line, ":")
		feature := strings.TrimSpace(lineArray[0])
		value := strings.TrimSpace(lineArray[1])
		request.Header[feature] = value
	}

	// Check required headers
	request.Host = request.Header["Host"]
	delete(request.Header, "Host")
	// Handle special headers
	if _, ok := request.Header["Connection"]; ok {
		if request.Header["Connection"] == "close" {
			request.Close = true
		} else {
			request.Close = false
		}
		delete(request.Header, "Connection")
	} else {
		request.Close = false
	}
	return &request, true, nil
}
