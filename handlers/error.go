 package handlers

// import (
// 	"database/sql"
// 	"html/template"
// 	"net/http"
// 	"time"
// )

// // RenderError renders an error page with the given message and status code
// func RenderError(w http.ResponseWriter, r *http.Request, message string, statuscode int) {
// 	var userID string
// 	sessionCookie, err := r.Cookie("session_id")
// 	if err == nil {
// 		err = db.QueryRow("SELECT user_id FROM sessions WHERE session_id = ?", sessionCookie.Value).Scan(&userID)
// 		if err == sql.ErrNoRows {
// 			http.SetCookie(w, &http.Cookie{
// 				Name:     "session_id",
// 				Value:    "",
// 				Path:     "/",
// 				Expires:  time.Unix(0, 0),
// 				MaxAge:   -1,
// 				HttpOnly: true,
// 			})
// 		} else if err != nil {
// 			http.Error(w, "Database Error", http.StatusInternalServerError)
// 			return
// 		}
// 	}
// 	w.WriteHeader(statuscode)
// 	tmpl, err := template.ParseFiles("./templates/error.html")
// 	if err != nil {
// 		http.Error(w, "Error parsing error template", http.StatusInternalServerError)
// 		return
// 	}
// 	tmpl.Execute(w, map[string]interface{}{
// 		"ErrorMessage": message,
// 		"StatusCode":   statuscode,
// 		"IsLoggedIn":   userID != "",
// 	})
// }

// // HandleDatabaseError handles database-related errors
// func HandleDatabaseError(w http.ResponseWriter, r *http.Request, err error) {
// 	if err != nil {
// 		RenderError(w, r, "Database Error", http.StatusInternalServerError)
// 		return
// 	}
// }
