# Go Steam Auth
This is a relatively lightweight, zero-dependency library for doing steam authentication.

> [!NOTE]
> This most likely doesn't totally comply with OpenID2, but I couldn't find a half-decent library to do this, and this one has worked fine for me for a while. Please reach out in an issue if something goes wrong.

## Installation
```shell
go get github.com/liondadev/go-steam-auth
```

## Example
See [the examples folder](./examples) for more examples, if any more are ever added.
```go
package main

import (
	"fmt"
	gosteamauth "github.com/liondadev/go-steam-auth"
	"net/http"
	"os"
)

func main() {
	apiKey, ok := os.LookupEnv("STEAM_API_KEY")
	if !ok {
		panic("STEAM_API_KEY is not set")
	}

	auther := gosteamauth.New(apiKey, "http://localhost:8080")

	mux := http.NewServeMux()
	mux.Handle("GET /auth", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, err := auther.GetAuthUrl("http://localhost:8080/auth/callback")
		if err != nil {
			fmt.Println(err)
			return
		}

		http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	}))

	mux.Handle("GET /auth/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		steamid, err := auther.ValidateCallback(r.URL.Query())
		if err != nil {
			fmt.Println(err)
			return
		}

		user, err := auther.GetSteamUser(steamid)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(user)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(user.PersonaName + " - " + user.SteamID))
	}))

	http.ListenAndServe(":8080", mux)
}
```