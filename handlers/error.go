package handlers

import (
	"html/template"
	"net/http"
)

func RenderError(w http.ResponseWriter, message string, statuscode int) {
	w.WriteHeader(statuscode)
	tmpl, err := template.ParseFiles("./templates/error.html")
	if err != nil {
		http.Error(w, "Error parsing error template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]interface{}{
		"ErrorMessage": message,
		"StatusCode":  statuscode,
	})
}
