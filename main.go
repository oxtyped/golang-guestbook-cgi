package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"time"

	"github.com/pkg/errors"
)

var f *os.File

// A GuestbookLine object contains a single guestbook entry
type GuestbookLine struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func ParseGuestbook(f *os.File) ([]GuestbookLine, error) {
	scanner := bufio.NewScanner(f)
	f.Seek(0, 0)
	guestbooks := []GuestbookLine{}
	for scanner.Scan() {
		t := []byte(scanner.Text())
		gbk := GuestbookLine{}
		err := json.Unmarshal(t, &gbk)
		if err != nil {
			return guestbooks, err
		}

		guestbooks = append(guestbooks, gbk)

	}

	return guestbooks, nil

}

// WriteGuestbook takes in a string line that contains the json output of a
// guestbook entry and saves it to disk
func WriteGuestbook(line *GuestbookLine, file *os.File) error {
	jsonBytes, err := json.Marshal(line)
	if err != nil {
		return errors.Wrapf(err, "error marshalling")
	}

	_, err = file.Write(jsonBytes)
	if err != nil {
		return errors.Wrapf(err, "error writing file")
	}
	_, err = file.WriteString("\n")
	if err != nil {
		return errors.Wrapf(err, "error writing string")
	}
	return nil
}

func doPost(w http.ResponseWriter, r *http.Request, f *os.File) {
	r.ParseForm()

	post := r.PostForm
	name := post.Get("name")
	email := post.Get("email")
	message := post.Get("message")
	date := time.Now()

	// save the content into a file
	line := &GuestbookLine{
		Name:      name,
		Email:     email,
		Message:   message,
		Timestamp: date,
	}

	err := WriteGuestbook(line, f)
	if err != nil {
		log.Fatal(err)
	}

	for k := range post {
		fmt.Fprintln(w, "PostForm", k+":", post.Get(k))
	}

}

func doGet(w http.ResponseWriter, r *http.Request, f *os.File) {

	guestbooks, err := ParseGuestbook(f)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range guestbooks {
		fmt.Fprintf(w, "<div><p>Name: %s<br />Time: %s<br />Message: %s</p></div>", v.Name, v.Timestamp, v.Message)
	}

}

func handler(w http.ResponseWriter, r *http.Request) {
	f, err := os.OpenFile("guestbook.dat", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	header := w.Header()
	header.Set("Content-Type", "text/html; charset=utf-8")

	switch r.Method {

	case "GET":
		doGet(w, r, f)
	case "POST":
		doPost(w, r, f)
	default:
		fmt.Fprintln(w, "no")

	}
}

func main() {

	err := cgi.Serve(http.HandlerFunc(handler))
	if err != nil {
		fmt.Println(err)
	}
}
