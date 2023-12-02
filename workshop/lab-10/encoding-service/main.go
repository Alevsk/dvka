package main

import (
	"fmt"
	"net/http"
	"os/exec"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	// http://localhost:1337/run?command=encode&message=hello%20world
	r.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		command := vars["command"]
		message := vars["message"]
		fmt.Fprintf(w,"Here is your encoded message:\n\n")
		out, err := exec.Command(command, message).Output()
		if err != nil {
			fmt.Fprintf(w,err.Error())
		} else {
			fmt.Fprintf(w,string(out))
		}
	}).Queries("command", "{command}","message","{message}")
	http.ListenAndServe(":1337", r)
}