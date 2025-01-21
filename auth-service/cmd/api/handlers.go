package main

import (
	"auth-service/data"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"
)

type UserData struct {
	ID                 int
	IP                 string
	RefreshToken       string
	HashedRefreshToken string
	AccessToken        string
}

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Password  string    `json:"-"`
	Active    int       `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type contextKey string

const userIDKey contextKey = "userID"

// Registrate insert new user to the database
func (app *Config) Registrate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Password  string `json:"password"`
		Active    int    `json:"active"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	user := User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password,
		Active:    requestPayload.Active,
	}
	id, err := app.Repo.Insert(data.User(user))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	err = app.logRequest("registrations", fmt.Sprintf("%s has been registrated", user.Email))
	if err != nil {
		fmt.Println("Error logging of user has benn authenticated:", err)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Succesfully created new user, id: %d", id),
	}

	app.
		writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		fmt.Println("Error in auth service during reading the payload")
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate the user against the database
	user, err := app.Repo.GetByEmail(requestPayload.Email)

	if err != nil {
		fmt.Println("Error in auth service, invalid credentials email")
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := app.Repo.PasswordMatches(requestPayload.Password, *user)
	if err != nil || !valid {
		fmt.Println("Error in auth service, password inmatches")
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	ip := strings.Split(r.RemoteAddr, ":")[0]
	secretKey := "some_secret_key"
	userData, err := generateTokens(user.ID, ip, secretKey)
	if err != nil {
		fmt.Println("Error generating tokens:", err)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	err = app.logRequest("authentication", fmt.Sprintf("%s logged in", user.Email))
	if err != nil {
		fmt.Println("Error logging of user has benn authenticated:", err)
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	err = validateRefreshToken(userData.HashedRefreshToken, userData.RefreshToken)
	if err != nil {
		fmt.Println("Error in auth service, invalid refresh token")
		return
	}
	fmt.Println("Refresh tokens validated")
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    userData.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}
	app.writeJSON(w, http.StatusAccepted, payload)
}

func generateTokens(userID int, ip, secretKey string) (*UserData, error) {
	accessToken, err := generateAccessToken(userID, ip, secretKey)
	if err != nil {
		return nil, err
	}

	refreshToken, hashedRefreshToken, err := generateRefreshToken(secretKey, ip)
	if err != nil {
		return nil, err
	}

	return &UserData{
		ID:                 userID,
		IP:                 ip,
		RefreshToken:       refreshToken,
		HashedRefreshToken: hashedRefreshToken,
		AccessToken:        accessToken,
	}, nil
}

func (app *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = name
	entry.Data = data

	jsonData, err := json.MarshalIndent(entry, "", "\t") //change to marshall??
	if err != nil {
		log.Println("Error marshalling json Data in auth-service")
		return err
	}

	logServiceURL := "http://log-service:82/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))

	_, err = app.Client.Do(request)
	if err != nil {
		log.Println("Error during doing log request in auth-service")
		return err
	}
	return nil
}

func validateRefreshToken(hashedRefreshToken, refreshToken string) error {
	decodedBytes, err := base64.StdEncoding.DecodeString(hashedRefreshToken)
	err = bcrypt.CompareHashAndPassword(decodedBytes, []byte(refreshToken))
	if err != nil {
		return fmt.Errorf("invalid refresh token: %w", err)
	}
	return nil
}

func generateRefreshToken(secretKey, ip string) (string, string, error) {
	refreshToken := secretKey + ip // Создайте Refresh-токен из secretKey и ip

	hashedRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refreshToken), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return refreshToken, base64.StdEncoding.EncodeToString(hashedRefreshToken), nil // Преобразование хеша в base64 перед хранением
}

func generateAccessToken(userID int, ip, secretKey string) (string, error) {
	expirationTime := time.Now().Add(time.Minute * 15) // Access-токен действителен в течение 15 минут

	claims := &jwt.MapClaims{
		"sub": userID,
		"exp": expirationTime.Unix(),
		"ip":  ip,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims) // Используйте метод подписи HS512 для Access-токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// authTokenMiddleware auths users to get access to some pages only by having access token
func (app *Config) authTokenMiddleware(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			cookie, err := r.Cookie("access_token")
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}
			tokenString := cookie.Value
			claims := &jwt.MapClaims{
				"sub": userIDKey,
			}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secretKey), nil
			})

			if err != nil || !token.Valid {
				app.errorJSON(w, errors.New("token is not valid"), http.StatusUnauthorized)
				return
			}

			userID, ok := (*claims)["sub"].(float64)
			if !ok {
				app.errorJSON(w, errors.New("invalid token claims"), http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, int(userID))
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// GetAllUsers retrieves all users from the database, sort them by points
func (app *Config) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := app.Repo.GetAll()
	if err != nil {
		app.errorJSON(w, errors.New("couldn't fetch All users"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Fetched all users"),
		Data:    users,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}
