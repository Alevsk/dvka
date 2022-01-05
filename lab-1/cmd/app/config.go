package main

import "github.com/dgrijalva/jwt-go"
import "github.com/minio/pkg/env"

func GetSigningKEy() string {
	return env.Get("DVKA-LAB1-SIGNING-KEY", "334f4b53-609c-4737-9d59-a71720b00cb8")
}

func GetFlag() string {
	return env.Get("DVKA-LAB1-FLAG", "theflag")
}

type customClaims struct {
	Balance float64 `json:"balance"`
	NFTs    []NFT   `json:"NFTs"`
	jwt.StandardClaims
}

type NFT struct {
	ID      int
	Name    string
	Price   float64
	comment string
}

var NFTList = []NFT{
	{
		ID:      0,
		Name:    "Golden bored ape",
		Price:   1000,
		comment: GetFlag(),
	},
	{
		ID:      1,
		Name:    "Cinema bored ape",
		Price:   700,
		comment: "",
	},
	{
		ID:      2,
		Name:    "Cool bored ape",
		Price:   550,
		comment: "",
	},
	{
		ID:      3,
		Name:    "Cyborg bored ape",
		Price:   500,
		comment: "",
	},
	{
		ID:      4,
		Name:    "Sad king bored ape",
		Price:   350,
		comment: "",
	},
	{
		ID:      5,
		Name:    "Punk Van Pelt bored ape",
		Price:   200,
		comment: "",
	},
}

type DownloadMoreNFTRequest struct {
	URL     string            `json:"url,omitempty"`
	Request map[string]string `json:"request,omitempty"`
}

type DownloadMoreNFTResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
