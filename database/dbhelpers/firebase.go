package dbhelpers

import (
	"bytes"
	"net/http"
	"os"

	"github.com/Shubhouy1/asset-management/models"
	"github.com/go-jose/go-jose/v4/json"
)

func FirebaseEmailLogin(email, password string) (string, error) {

	apiKey := os.Getenv("FIREBASE_API_KEY")

	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + apiKey

	reqBody := models.FirebaseLoginRequest{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result models.FirebaseLoginResponse
	json.NewDecoder(resp.Body).Decode(&result)

	return result.IDToken, nil
}
