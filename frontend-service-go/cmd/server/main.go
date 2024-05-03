package main

import (
	"bytes"
	"html/template"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/store", handleStore)
	http.HandleFunc("/library", handleLibrary)

	http.ListenAndServe(":3000", nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderTemplate(w, "index.html", map[string]interface{}{
			"Content": template.HTML(getContentHTML("login.html")),
		})
	} else {
		// Handle the login form submission
		// ...
	}
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", map[string]interface{}{
		"Content": template.HTML(getContentHTML("register.html")),
	})
}

func handleStore(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", map[string]interface{}{
		"Content": template.HTML(getContentHTML("store.html")),
	})
}

func handleLibrary(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "index.html", map[string]interface{}{
		"Content": template.HTML(getContentHTML("library.html")),
	})
}

func getContentHTML(tmpl string) string {
	t, err := template.ParseFiles("templates/" + tmpl)
	if err != nil {
		return ""
	}

	var contentBuf bytes.Buffer
	err = t.Execute(&contentBuf, nil)
	if err != nil {
		return ""
	}

	return contentBuf.String()
}

func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	t, err := template.ParseFiles("templates/" + templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
