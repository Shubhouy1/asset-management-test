package models

type FirebaseLoginRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	ReturnSecureToken bool   `json:"returnSecureToken"`
}

type FirebaseLoginResponse struct {
	IDToken string `json:"idToken"`
}
