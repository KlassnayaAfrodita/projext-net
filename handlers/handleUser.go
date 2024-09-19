package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/KlassnayaAfrodita/mylib/storage"
)

func (api *Api) AuthenticationUser(w http.ResponseWriter, r *http.Request) { //! принимаем POST json
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "server error"}`, 500)
	}
	defer r.Body.Close()

	var user storage.User

	newerr := json.Unmarshal(body, &user)
	if newerr != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	trueUser, err := api.users.GetUser(user.ID)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, 404)
	}
	if trueUser.Password != user.Password {
		http.Error(w, `bad pass`, 400)
		return
	}

	SID, err := api.session.SetSession(trueUser.ID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
	}
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   SID,
		Expires: time.Now().Add(10 * time.Hour),
	}

	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/products", 200)
}

func (api *Api) RegistrationUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "server error"}`, 500)
	}
	defer r.Body.Close()

	var user storage.User

	err = json.Unmarshal(body, &user) //! распаковали json
	if err != nil {
		http.Error(w, `{"error":"incorrect input"}`, 402)
	}

	user, err = api.users.AddUser(user) //! добавили пользоавтеля в бд
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
	}

	SID, err := api.session.SetSession(user.ID) //! добавили сессию
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
	}

	cookie := http.Cookie{
		Name:    "session_id",
		Value:   SID,
		Expires: time.Now().Add(10 * time.Hour),
	}

	http.SetCookie(w, &cookie)

	http.Redirect(w, r, "/products", 200)
}

func (api *Api) LogoutUser(w http.ResponseWriter, r *http.Request) {

	session, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, `{"error":"you dont auth"}`, 500)
		return
	}

	if _, ok := api.session[session.Value]; !ok { //* если не нашли сессию в бд
		http.Error(w, `no sess`, 401)
		return
	}

	_, err = api.session.DeleteSession(session.Value)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
	}

	session.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, session)
}
