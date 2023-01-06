package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"projectsMod/pkg/pkg/user"
	"strconv"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v4/pgxpool"
)


const addr2 string = "localhost:8081"

type API struct {
	Ctx     context.Context
	Dbpool  *pgxpool.Pool
	ConnStr string
}

func main() {
	ctx := context.Background()
	connStr := "postgres://postgres:Password10@localhost:5432/base"
	dbpool, err := pgxpool.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	api := API{ctx, dbpool, connStr}

	queryCreateTable := `CREATE TABLE if not exists baseUsers (id SERIAL PRIMARY KEY, name VARCHAR(50), age integer, friends character varying(20)[]) ;`
	_, err = dbpool.Exec(ctx, queryCreateTable)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create baseUsers table: %v\n", err)
		os.Exit(1)
	}
	log.Println("Successfully created relational table baseUsers")

	log.Println("Successfully connected relational table baseUsers")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.HandleFunc("/create", api.Create)
	r.HandleFunc("/get", api.GetAll)
	r.HandleFunc("/make_friends", api.MakeFriends)
	r.Route("/friends", func(r chi.Router) {
		r.HandleFunc("/user_id{id}", api.GetFriendList)
	})
	r.HandleFunc("/user", api.DelUser)
	r.HandleFunc("/user_id{id}", api.NewAge)

	http.ListenAndServe(addr2, r)
}

func (api *API) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var u *user.User
		if err := json.Unmarshal(content, &u); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		conn, err := api.Dbpool.Acquire(api.Ctx)
  		if err != nil {
    		log.Printf("Unable to acquire a database connection: %v\n", err)
    		w.WriteHeader(500)
    	return
  		}
  		defer conn.Release()

		tx, _ := conn.Begin(api.Ctx)
	
		defer tx.Rollback(api.Ctx)
		
		var id int
		row := tx.QueryRow(api.Ctx, "INSERT INTO public.baseusers(name, age) VALUES ($1, $2) RETURNING id", u.Name, u.Age)

		err = row.Scan(&id)   
  		if err != nil {
    		log.Printf("Unable to INSERT: %v\n", err)
   			w.WriteHeader(500)
    	return
  		}

		err = tx.Commit(api.Ctx)
		if err != nil {
			log.Println(err)
		}
  		
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("User was created " + u.Name +  " id:" + strconv.Itoa(int(id))))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (api *API) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		rows, _ := api.Dbpool.Query(api.Ctx, `SELECT id, name, age FROM public.baseusers`)
		defer rows.Close()
		var user user.User //LJ,FDBKFV NJKDN            KHHH
		for rows.Next() {
			if err := pgxscan.ScanRow(&user, rows); err != nil {
				w.WriteHeader(http.StatusInternalServerError) 
				w.Write([]byte(err.Error()))
				return
			}
			w.Write([]byte(user.ToString()))
		}
		if err := rows.Err(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}


func (api *API) MakeFriends(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var ids *user.Ids
		if err := json.Unmarshal(content, &ids); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		query := fmt.Sprintf("SELECT name, friends FROM public.baseusers WHERE id='%d'",ids.SourseId)
		var userSourse user.User
		if err := pgxscan.Get(
			api.Ctx, api.Dbpool, &userSourse, query,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError) 
			w.Write([]byte(err.Error()))
			return
		}

		query = fmt.Sprintf("SELECT name, friends FROM public.baseusers WHERE id='%d'",ids.TargetId)
		var userTarget user.User
		if err := pgxscan.Get(
			api.Ctx, api.Dbpool, &userTarget, query,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var user *user.User
		if user.FindFriend(userSourse, userTarget) && user.FindFriend(userTarget, userSourse) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(userTarget.Name + " и " + userSourse.Name + " уже друзья."))
			return
		} else {
			conn, err := api.Dbpool.Acquire(api.Ctx)
			if err != nil {
				log.Printf("Unable to acquire a database connection: %v\n", err)
				w.WriteHeader(500)
			return
			}
			defer conn.Release()

			tx, _ := conn.Begin(api.Ctx)
		
			defer tx.Rollback(api.Ctx)

			_,err = tx.Exec(api.Ctx, `UPDATE public.baseusers SET friends = array_prepend($1, $2) WHERE id = $3`, userSourse.Name, userTarget.Friends, ids.TargetId)    
			if err != nil {
				log.Printf("Unable to UPDATE %v: %v\n", userTarget.Name, err)
				w.WriteHeader(500)
			return
			}

			_,err = tx.Exec(api.Ctx, `UPDATE public.baseusers SET friends = array_prepend($1, $2) WHERE id = $3`, userTarget.Name, userSourse.Friends, ids.SourseId)    
			if err != nil {
				log.Printf("Unable to UPDATE %v: %v\n", userSourse.Name, err)
				w.WriteHeader(500)
			return
			}

			err = tx.Commit(api.Ctx)
			if err != nil {
				log.Println(err)
			}
		
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(userTarget.Name + " и " + userSourse.Name + " теперь друзья."))
			return
		}
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (api *API) GetFriendList(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		if id < 0 {
			http.Error(w, errors.New("Error").Error(), http.StatusUnauthorized)
			return
		}
		
		query := fmt.Sprintf("SELECT friends FROM public.baseusers WHERE id='%d'",id)
		var user user.User
		if err := pgxscan.Get(
			api.Ctx, api.Dbpool, &user, query,
		); err != nil {
			w.WriteHeader(http.StatusInternalServerError) 
			w.Write([]byte(err.Error()))
			return
		}

		resporse := ""
		for _, v := range user.Friends {
			resporse += v + " "
		}
		w.Write([]byte(resporse))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (api *API) DelUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == "DELETE" {
		content, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		defer r.Body.Close()

		var ids *user.Ids
		if err := json.Unmarshal(content, &ids); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		id := ids.TargetId

		query := fmt.Sprintf("SELECT name FROM public.baseusers WHERE id='%d'",id)
		var user user.User
		if err := pgxscan.Get(api.Ctx, api.Dbpool, &user, query,); err != nil {
			w.WriteHeader(http.StatusInternalServerError) 
			w.Write([]byte(err.Error()))
			return
		}

		conn, err := api.Dbpool.Acquire(api.Ctx)
  		if err != nil {
    		log.Printf("Unable to acquire a database connection: %v\n", err)
   	 		w.WriteHeader(500)
    	return
  		}
  		defer conn.Release()

		  _,err = conn.Exec(api.Ctx, `UPDATE public.baseusers SET friends = array_remove(friends, $1)`, user.Name)    
		  if err != nil {
			  log.Printf("Unable to UPDATE column friends: %v\n", err)
			  w.WriteHeader(500)
		  return
		  }

  		ct, err := conn.Exec(api.Ctx,"DELETE FROM public.baseusers WHERE id = $1", id)
  		if err != nil {
    		log.Printf("Unable to DELETE: %v\n", err)
    		w.WriteHeader(500)
    		return
  		}

  		if ct.RowsAffected() == 0 {
    		w.WriteHeader(404)
    	return
		}

		w.Write([]byte(user.Name))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}

func (api *API) NewAge(w http.ResponseWriter, r *http.Request) {
	if r.Method == "PUT" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		id, _ := strconv.Atoi(chi.URLParam(r, "id"))
		if id <= 0 {
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
		if err := json.Unmarshal(content, &age); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		conn, err := api.Dbpool.Acquire(api.Ctx)
  		if err != nil {
    		log.Printf("Unable to acquire a database connection: %v\n", err)
    		w.WriteHeader(500)
    	return
  		}
  		defer conn.Release()

		tx, _ := conn.Begin(api.Ctx)
	
		defer tx.Rollback(api.Ctx)
		
		_,err = tx.Exec(api.Ctx, "UPDATE public.baseusers SET age = $1 WHERE id = $2", age.NewAge, id)  
  		if err != nil {
    		log.Printf("Unable to UPDATE: %v\n", err)
   			w.WriteHeader(500)
    	return
  		}

		err = tx.Commit(api.Ctx)
		if err != nil {
			log.Println(err)
		}

		w.Write([]byte("Возраст пользователя успешно обновлён"))
		return
	}
	w.WriteHeader(http.StatusBadRequest)
}
