// Package web provides HTTP handler utilities for serving go-fastreport reports
// over HTTP. It offers ready-made handlers that accept a PreparedPages instance
// and write it to an http.ResponseWriter using the appropriate content type.
//
// Usage:
//
//	http.Handle("/report.html", web.HTMLHandler(preparedPages))
//	http.Handle("/report.pdf",  web.PDFHandler(preparedPages))
//	http.Handle("/report.png",  web.ImageHandler(preparedPages))
package web
