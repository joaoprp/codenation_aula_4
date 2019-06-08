package main

import (
	"apitest/middlewares"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var jwtKey = []byte(os.Getenv("JWT_KEY"))

var connectionString = fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
	os.Getenv("PGQL_USERNAME"), os.Getenv("PGQL_PASSWORD"), os.Getenv("PGQL_DATABASE"))

var db, dberr = sql.Open("postgres", connectionString)

// Credentials is exported
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type author struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Posts []post `json:"posts,omitempty"`
}

type post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

type code struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsCustom    bool   `json:"is_custom,omitempty"`
}

type postposts struct {
	Foo string `json:"foo"`
	Val int    `json:"val"`
}

func main() {
	log.Println(os.Getenv("JWT_KEY"))
	if dberr != nil {
		log.Fatal(dberr)
	}

	router := mux.NewRouter()
	/**
	 * Commented block here is a sample for a global middleware preference
	 * We can actually build it if we want as we can daisy-chain middlewares
	 */
	// router.Use(middleware)
	// router.HandleFunc("/", home)
	router.HandleFunc("/auth", generateToken)
	router.Handle("/posts", middlewares.Auth(http.HandlerFunc(posts))).Methods("GET")
	router.Handle("/posts", middlewares.Auth(http.HandlerFunc(postPosts))).Methods("POST")
	router.Handle("/codes", middlewares.Auth(http.HandlerFunc(getCodes))).Methods("GET")
	router.Handle("/code/{id}", middlewares.Auth(http.HandlerFunc(getCode))).Methods("GET")

	http.ListenAndServe(":8080", router)
}

func generateToken(w http.ResponseWriter, r *http.Request) {
	creds := Credentials{
		Username: "johndoe",
	}

	expirationTime := time.Now().Add(5 * time.Minute)

	claims := &middlewares.Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write([]byte(tokenString))

}

func posts(w http.ResponseWriter, r *http.Request) {
	a := author{
		ID:   1,
		Name: "John Doe",
		Posts: []post{
			post{
				ID:    1,
				Title: "Lorem Ipsum",
				Body:  "Lorem ipsum dolor sit amet",
			},
			post{
				ID:    2,
				Title: "Bacon Ipsum",
				Body:  "Bacon ipsum dolor sit amet",
			},
		},
	}

	response, _ := json.Marshal(a)

	w.Write(response)
}

func postPosts(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)

	p := new(postposts)

	json.Unmarshal(body, p)

	w.Write([]byte(strconv.Itoa(p.Val)))
}

func getCodes(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM ia.code LIMIT $1", 10)

	codenation := code{
		ID:   "CDNT",
		Name: "Codenation",
	}

	db.Exec("insert into ia.code (id, name) values ($1, $2)", codenation.ID, codenation.Name)

	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	codes := []code{}

	for rows.Next() {
		c := new(code)

		rows.Scan(&c.ID, &c.Name, &c.Description, &c.IsCustom)

		codes = append(codes, *c)
	}

	response, _ := json.Marshal(codes)

	w.Write(response)
}

func getCode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	log.Println(r.URL.Query().Get("test"))

	rows, err := db.Query("SELECT * FROM ia.code where id = $1", params["id"])
	defer rows.Close()

	if err != nil {
		log.Fatal(err)
	}

	rows.Next()
	c := new(code)

	rows.Scan(&c.ID, &c.Name, &c.Description, &c.IsCustom)

	response, _ := json.Marshal(c)

	w.Write(response)
}
