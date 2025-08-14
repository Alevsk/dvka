package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var signingKey = getSigningKey()

func app(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "static/index.html")
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Sorry, only GET method is supported.")
	}
}

func assets(w http.ResponseWriter, r *http.Request) {
	dir, file := filepath.Split(fmt.Sprintf("static/%s", r.URL.Path))
	f, err := http.Dir(dir).Open(file)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Sorry, resource not found.")
		return
	}
	defer f.Close()
	fi, err := f.Stat()

	if err != nil || fi.IsDir() {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Sorry, resource not found.")
		return
	}
	modTime := fi.ModTime()
	http.ServeContent(w, r, r.URL.Path, modTime, f)
}

func apiV1(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		switch r.URL.Path {
		case "/api/v1/nft-buy":
			keys, ok := r.URL.Query()["id"]
			if !ok || len(keys[0]) < 1 {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			id, err := strconv.Atoi(keys[0])
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			if id < 0 || id >= len(NFTList) {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			cookieHeader := r.Header.Get("Cookie")
			if cookieHeader == "" {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			cookies := strings.Split(cookieHeader, " ")
			if len(cookies) == 0 {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			lasCookie := cookies[len(cookies)-1]
			jwtCookie := strings.Split(lasCookie, "=")
			if len(jwtCookie) != 2 {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			token, err := jwt.ParseWithClaims(
				jwtCookie[1],
				&customClaims{},
				func(token *jwt.Token) (interface{}, error) {
					return []byte(signingKey), nil
				},
			)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			claims, ok := token.Claims.(*customClaims)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}

			if claims.Balance-NFTList[id].Price <= 0 {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, you dont have enough money.")
				return
			}

			updatedNFT := append(claims.NFTs, NFTList[id])
			updatedBalance := claims.Balance - NFTList[id].Price

			updatedClaims := customClaims{
				NFTs:    updatedNFT,
				Balance: updatedBalance,
				StandardClaims: jwt.StandardClaims{
					Issuer: "NFT-Museum",
				},
			}

			newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, updatedClaims)
			signedToken, err := newToken.SignedString([]byte(signingKey))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			expire := time.Now().Add(1 * time.Hour)
			cookie := http.Cookie{Name: "jwt", Value: signedToken, Path: "/", Expires: expire}
			http.SetCookie(w, &cookie)
			fmt.Fprintf(w, "NFT note: "+NFTList[id].comment)
		case "/api/v1/account-reset":
			claims := customClaims{
				NFTs:    []NFT{},
				Balance: 999,
				StandardClaims: jwt.StandardClaims{
					Issuer: "NFT-Museum",
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			signedToken, err := token.SignedString([]byte(signingKey))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			expire := time.Now().Add(1 * time.Hour)
			cookie := http.Cookie{Name: "jwt", Value: signedToken, Path: "/", Expires: expire}
			http.SetCookie(w, &cookie)
		case "/api/v1/download-more":
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			r.Body.Close()
			var params DownloadMoreNFTRequest
			err = json.Unmarshal(body, &params)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			req, err := http.NewRequest("GET", params.URL, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			for key, value := range params.Request {
				req.Header.Set(key, value)
			}
			httpClient := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, server appear to be offline.")
				return
			}
			respBody, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			resp.Body.Close()
			response := &DownloadMoreNFTResponse{
				Status:  resp.Status,
				Message: string(respBody),
			}
			w.Header().Set("Content-Type", "application/json")
			bodyResponse, err := json.Marshal(response)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Sorry, something went wrong.")
				return
			}
			fmt.Fprintf(w, string(bodyResponse))

		default:
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "Sorry, unsupported api.")
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Sorry, only POST method is supported.")
	}
}

func robots(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/robots.txt")
}

func main() {
	r := mux.NewRouter()
	r.UseEncodedPath()
	r.PathPrefix("/assets").HandlerFunc(assets)
	r.PathPrefix("/api/v1").HandlerFunc(apiV1)
	r.HandleFunc("/", app)
	r.HandleFunc("/robots.txt", robots)
	addr := "0.0.0.0:8080"
	fmt.Printf("Starting server at %s\n", addr)
	srv := &http.Server{
		Handler: r,
		Addr:    addr,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 3600 * time.Second,
		ReadTimeout:  3600 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
