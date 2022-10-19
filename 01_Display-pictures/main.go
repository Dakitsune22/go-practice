package main

import (
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	uuid "github.com/satori/go.uuid"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	// Add route to serve pictures:
	http.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("./public"))))
	http.ListenAndServe(":8080", nil)
}

func index(rw http.ResponseWriter, req *http.Request) {
	c := getCookie(rw, req)

	// Process form submission:
	if req.Method == http.MethodPost {
		mf, mh, err := req.FormFile("nf")
		if err != nil {
			log.Println(err)
		}
		defer mf.Close()
		// Create SHA for file name:
		ext := strings.Split(mh.Filename, ".")[1] // Get the extension of file name
		h := sha1.New()
		io.Copy(h, mf)
		fname := fmt.Sprintf("%x", h.Sum(nil)) + "." + ext
		// Create new file:
		wd, err := os.Getwd()
		if err != nil {
			log.Println(err)
		}
		path := filepath.Join(wd, "public", "pics", fname)
		nf, err := os.Create(path)
		if err != nil {
			log.Println(err)
		}
		defer nf.Close()
		// Copy:
		mf.Seek(0, 0)
		io.Copy(nf, mf)
		// Add file name to user's cookie:
		appendValue(rw, c, fname)
	}
	data := strings.Split(c.Value, "|")
	tpl.ExecuteTemplate(rw, "index.gohtml", data[1:])
}

func getCookie(rw http.ResponseWriter, req *http.Request) *http.Cookie {
	c, err := req.Cookie("session")
	if err != nil {
		sId, _ := uuid.NewV4()
		c = &http.Cookie{
			Name:   "session",
			Value:  sId.String(),
			MaxAge: 60,
		}
	}
	// Refresh or create cookie:
	http.SetCookie(rw, c)

	return c
}

func appendValue(rw http.ResponseWriter, c *http.Cookie, fname string) *http.Cookie {
	if !strings.Contains(c.Value, fname) {
		c.Value += "|" + fname
	}
	// Refresh or update cookie:
	http.SetCookie(rw, c)
	return c
}
