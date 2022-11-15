package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"projectsMod/pkg/user"
	"strconv"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"errors"
)

type service struct {
	store map[string]*user.User
	idstore map[int]*user.User
}



func main() {
	r := chi.NewRouter()
 	r.Use(middleware.Logger)
 	srv:=service{make(map[string]*user.User), make(map[int]*user.User)}
 	r.HandleFunc("/create", srv.Create)
 	r.HandleFunc("/get", srv.GetAll)
	r.HandleFunc("/make_friends", srv.MakeFriends)
	r.Route("/friends", func(r chi.Router) {
        r.HandleFunc("/user_id{id}", srv.GetFriendList)
    })
	r.HandleFunc("/user", srv.DelUser)
	r.HandleFunc("/user_id{id}", srv.NewAge)

 	http.ListenAndServe("localhost:3001", r)
}

func (s *service) Create (w http.ResponseWriter, r *http.Request) {		
	if r.Method == "POST" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var u *user.User
		if err := json.Unmarshal(content,&u); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		s.store[u.Name] = u
		id := len(s.store)
		s.idstore[id] = u

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("User was created "+u.Name + " id:"+ strconv.Itoa(id)))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}	

func (s *service) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method== "GET" {
		resporse := ""
		for _, userVal := range s.store {
			resporse += userVal.ToString()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(resporse))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (s *service) MakeFriends (w http.ResponseWriter, r *http.Request) {		
	if r.Method == "POST" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var ids *user.Ids
		if err := json.Unmarshal(content,&ids); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		_, presensUser := s.idstore[ids.SourseId]
		_, presensFriend := s.idstore[ids.TargetId]
		if  presensUser && presensFriend {
			s.idstore[ids.SourseId].Friends = append(s.idstore[ids.SourseId].Friends, s.idstore[ids.TargetId].Name)
			s.idstore[ids.TargetId].Friends  = append(s.idstore[ids.TargetId].Friends, s.idstore[ids.SourseId].Name)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(s.idstore[ids.TargetId].Name + " и " + s.idstore[ids.SourseId].Name + " теперь друзья."))
			return
		}	
	}
	w.WriteHeader(http.StatusBadRequest)
}	

func (s *service) GetFriendList(w http.ResponseWriter, r *http.Request) {
	if r.Method== "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		if id < 0 {
            http.Error(w, errors.New("Error").Error(), http.StatusUnauthorized)
            return
        }
		
		resporse := ""
		for _, v := range  s.idstore[id].Friends {
			resporse += v + " "
		}
		w.Write([]byte(resporse))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (s *service) DelUser(w http.ResponseWriter, r *http.Request) {
	if r.Method== "DELETE" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()
	
		var ids *user.Ids
		if err := json.Unmarshal(content,&ids); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		id := ids.TargetId

		for i := 1; i<= len(s.idstore); i++ {
			for j, v := range  s.idstore[i].Friends {
			if v == s.idstore[id].Name {
				delj := j
				s.idstore[i].Friends = append(s.idstore[i].Friends[0:delj], s.idstore[i].Friends[delj+1:]...)
			}
		}
		
		}

		for _, userVal := range s.store {
			if s.idstore[id].Name == userVal.Name{
				delete(s.store, userVal.Name)
			}
		}

		w.Write([]byte(s.idstore[id].Name))

		delete(s.idstore, id)
		
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (s *service) NewAge(w http.ResponseWriter, r *http.Request) {
	if r.Method== "PUT" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		if id < 0 {
            http.Error(w, errors.New("Error").Error(), http.StatusUnauthorized)
            return
        }

		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()
	
		var age *user.NewUserAge
		if err := json.Unmarshal(content,&age); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		newAge := age.NewAge

		s.idstore[id].Age = newAge
		for _, userVal := range s.store {
			if s.idstore[id].Name == userVal.Name{
				userVal.Age = newAge
			}
		}

		w.Write([]byte("Возраст пользователя успешно обновлён"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

