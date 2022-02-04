# OwnHTTP

---

This project is a course project of **Computer Network**.

Clients connect to this socket and use the OwnHTTP protocol to retrieve files from the server. The server will read data from the client, using the framing and parsing techniques discussed in class to interpret one or more requests (if the client is using pipelined requests).  Every time the server reads in a full request, it will service that request and send a response back to the client.  After sending back one (or more) responses, the server will either close the connection (if instructed to do so by the client via the “Connection: close” header, described below), or after an appropriate timeout occurs (also described below).  The web server will then continue waiting for future client connections. The server should be implemented in a concurrent manner, so that it can process multiple client requests overlapping in time.



**Note:** these code now can send request and response under OwnHTTP protocol properly. But it still need to be optimized if receiving request with some melformed start lines or headers.
