// Command sign-token prints a fresh RS256 JWT compatible with the lighthouse
// demo authn middleware. Run from app/master/service:
//
//	go run ./cmd/sign-token
//
// The token is signed by stubauth.SharedKeypair() — the same Signer the
// servers' verifiers trust — so the printed value passes through the jwt
// engine in authn.Multi end-to-end.
package main

import (
	"fmt"
	"time"

	"github.com/Servora-Kit/servora-example/app/master/service/internal/stubauth"
	gojwt "github.com/golang-jwt/jwt/v5"
)

func main() {
	signer, _ := stubauth.SharedKeypair()
	claims := gojwt.MapClaims{
		"sub":  "lighthouse-user",
		"name": "Lighthouse User",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(1 * time.Hour).Unix(),
	}
	tok, err := signer.Sign(claims)
	if err != nil {
		panic(err)
	}
	fmt.Println(tok)
}
