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
		logger.Error("error", err)
		return
	}
	defer r.Body.Close()

	var user storage.User

	newerr := json.Unmarshal(body, &user)
	if newerr != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		logger.Error("error", err)
		return
	}

	trueUser, err := api.users.GetUserByName(user.Name)
	if err != nil {
		http.Error(w, `{"error":"not found"}`, 404)
		logger.Error("error", err)
		return
	}
	if trueUser.Password != user.Password {
		http.Error(w, `bad pass`, 400)
		logger.Error("error", err)
		return
	}

	SID, err := api.session.SetSession(trueUser.ID)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
		logger.Error("error", err)
		return
	}
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   SID,
		Expires: time.Now().Add(10 * time.Hour),
	}

	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, &cookie)
	// r.AddCookie(&cookie)

	http.Redirect(w, r, "/products", 300)
}

func (api *Api) RegistrationUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "server error"}`, 500)
		logger.Error("error", err)
		return
	}
	// fmt.Printf("body: %s\n", string(body))
	defer r.Body.Close()

	var user storage.User

	err = json.Unmarshal(body, &user) //! распаковали json
	if err != nil {
		logger.Error("error", err)
		http.Error(w, `{"error":"incorrect input"}`, 402)
		return
	}

	user.Cart = storage.NewProductStorage()
	// fmt.Println(user.Cart)

	user, err = api.users.AddUser(user) //! добавили пользоавтеля в бд
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
		logger.Error("error", err)
		return
	}

	// fmt.Println(user)

	SID, err := api.session.SetSession(user.ID) //! добавили сессию
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
		logger.Error("error", err)
		return
	}

	// fmt.Println(api.users)
	// fmt.Println(api.session)

	cookie := http.Cookie{
		Name:    "session_id",
		Value:   SID,
		Expires: time.Now().Add(10 * time.Hour),
	}

	w.Header().Set("Content-Type", "application/json")
	http.SetCookie(w, &cookie)
	// r.AddCookie(&cookie)

	http.Redirect(w, r, "/products", 300)
}

func (api *Api) LogoutUser(w http.ResponseWriter, r *http.Request) {

	sess, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, `{"error":"you dont auth"}`, 500)
		logger.Error("error", err)
		return
	}

	if _, err = api.session.GetSession(sess.Value); err != nil { //* если не нашли сессию в бд
		http.Error(w, `{"error":"no session}"`, 401)
		logger.Error("error", err)
		return
	}

	_, err = api.session.DeleteSession(sess.Value)
	if err != nil {
		http.Error(w, `{"error":"db error"}`, 500)
		logger.Error("error", err)
		return
	}

	sess.Expires = time.Now().AddDate(0, 0, -1)
	http.SetCookie(w, sess)
}
