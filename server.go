package tritonhttp

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// DocRoot specifies the path to the directory to serve static files from.
	DocRoot string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	addr := s.Addr
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Listening on %s\n", addr)

	for {
		conn, err := listen.Accept()
		log.Println("Received a new connection")
		if err != nil {
			log.Print(err)
			continue
		}
		go s.HandleConnection(conn)
	}
	// Hint: call HandleConnection
}

// HandleConnection reads requests from the accepted conn and handles them.
func (s *Server) HandleConnection(conn net.Conn) {
	// Hint: use the other methods below
	bufReader := bufio.NewReader(conn)
	bufWriter := io.Writer(conn)
	var res Response
	var byteRecieved bool
	var err error
	var req *Request
	for {
		// Set timeout
		err = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if err != nil {
			log.Println(err)
			log.Println("1")
		}
		// Try to read next request
		req, byteRecieved, err = ReadRequest(bufReader)
		log.Println(err)
		// Handle EOF
		if err == io.EOF {
			res.HandleBadRequest()
			log.Println(res)
			res.Write(bufWriter)
			break
		} else if err != nil {
			// Handle timeout
			nErr, ok := err.(net.Error)
			if ok {
				if nErr.Timeout() {
					if byteRecieved {
						res.HandleBadRequest()
						res.Write(bufWriter)
						break
					} else {
						break
					}
				}
			} else {
				// Handle bad request
				res.HandleBadRequest()
				res.Write(bufWriter)
				break
			}
		} else {
			// Handle good request
			res = *(s.HandleGoodRequest(req))
			res.Write(bufWriter)
			// Close conn if requested
			if req.Close {
				break
			}
		}
	}
	conn.Close()
}

// HandleGoodRequest handles the valid req and generates the corresponding res.
func (s *Server) HandleGoodRequest(req *Request) (res *Response) {
	var response Response
	var filePath string
	fileInfo, err := os.Stat(s.DocRoot + req.URL)
	if err != nil {
		response.HandleNotFound(req)
	} else {
		if fileInfo.IsDir() {
			filePath = filepath.Join(s.DocRoot + req.URL + "/index.html")
		} else {
			filePath = s.DocRoot + req.URL
		}
		response.HandleOK(req, filePath)
	}
	return &response
	// Hint: use the other methods below
}

// HandleOK prepares res to be a 200 OK response
// ready to be written back to client.
func (res *Response) HandleOK(req *Request, path string) {
	res.Header = make(map[string]string)
	tempPath := strings.Split(path, "/")
	relPath := tempPath[len(tempPath)-1]
	res.StatusCode = 200
	res.Proto = req.Proto
	res.Request = req
	res.FilePath = path
	res.Header["Date"] = FormatTime(time.Now())
	fileInfo, _ := os.Stat(path)
	modTime := fileInfo.ModTime()
	res.Header["Last-Modified"] = FormatTime(modTime)
	tempExt := strings.Split(relPath, ".")[1]
	actualExt := "." + tempExt
	res.Header["Content-Type"] = MIMETypeByExtension(actualExt)
	res.Header["Content-Length"] = fileLength(path)
	if req.Close {
		res.Header["Connection"] = "close"
	}
}

// HandleBadRequest prepares res to be a 400 Bad Request response
// ready to be written back to client.
func (res *Response) HandleBadRequest() {
	res.Header = make(map[string]string)
	res.StatusCode = 400
	res.Proto = "HTTP/1.1"
	res.Header["Date"] = FormatTime(time.Now())
	res.Header["Connection"] = "close"
}

// HandleNotFound prepares res to be a 404 Not Found response
// ready to be written back to client.
func (res *Response) HandleNotFound(req *Request) {
	res.Header = make(map[string]string)
	res.StatusCode = 404
	res.Proto = "HTTP/1.1"
	res.Request = req
	res.FilePath = ""
	res.Header["Date"] = FormatTime(time.Now())
	if req.Close {
		res.Header["Connection"] = "close"
	}
}
