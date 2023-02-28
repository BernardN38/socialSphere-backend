package handler

// import (
// 	"bytes"
// 	"context"
// 	"database/sql"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"net/http/httptest"
// 	"strings"
// 	"testing"

// 	"github.com/bernardn38/socialsphere/authentication-service/models"
// 	"github.com/bernardn38/socialsphere/authentication-service/sql/users"
// 	"github.com/cristalhq/jwt/v4"
// 	_ "github.com/lib/pq"
// 	"github.com/testcontainers/testcontainers-go"
// 	"github.com/testcontainers/testcontainers-go/wait"
// 	"golang.org/x/crypto/bcrypt"
// )

// func Setup() *Handler {
// 	ctx := context.Background()
// 	// Create a new Postgres container with a random name
// 	req := testcontainers.ContainerRequest{
// 		Image:        "postgres:14-alpine",
// 		ExposedPorts: []string{"5432/tcp"},
// 		Env: map[string]string{
// 			"POSTGRES_USER":     "postgres",
// 			"POSTGRES_PASSWORD": "test",
// 			"POSTGRES_DB":       "test",
// 		},
// 		WaitingFor: wait.ForLog("database system is ready to accept connections"),
// 	}
// 	// Create a RabbitMQ test container
// 	rabbitReq := testcontainers.ContainerRequest{
// 		Image:        "rabbitmq:3-management",
// 		ExposedPorts: []string{"5672/tcp"},
// 		WaitingFor:   wait.ForLog("Server startup complete"),
// 	}

// 	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: req,
// 		Started:          true,
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	rabbitContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
// 		ContainerRequest: rabbitReq,
// 		Started:          true,
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Get the host and port of the container's Postgres server
// 	host, err := postgresContainer.Host(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	port, err := postgresContainer.MappedPort(ctx, "5432")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	pgConnString := fmt.Sprintf("postgresql://postgres:test@%s:%s/test?sslmode=disable", host, port.Port())
// 	// Connect to the container's Postgres server and perform some tests
// 	db, err := sql.Open("postgres", pgConnString)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	_, err = db.Exec(`
// 	CREATE TABLE IF NOT EXISTS users
// 	(
// 		id         serial PRIMARY KEY,
// 		username   text NOT NULL UNIQUE,
// 		email      text NOT NULL UNIQUE,
// 		password   text NOT NULL
// 	)
// 	`)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	rabbitmqIP, err := rabbitContainer.Host(ctx)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	rabbitmqPort, err := rabbitContainer.MappedPort(ctx, "5672")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	rabbitUrl := fmt.Sprintf("amqp://%s:%s@%s:%s", "guest", "guest", rabbitmqIP, rabbitmqPort.Port())
// 	log.Println(rabbitUrl)
// 	h, err := NewHandler(models.Config{
// 		JwtSecretKey:     "secretKey",
// 		JwtSigningMethod: jwt.Algorithm(jwt.HS256),
// 		PostgresUrl:      pgConnString,
// 		RabbitmqUrl:      rabbitUrl,
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	return h
// }
// func TestRegisterUser(t *testing.T) {
// 	h := Setup()
// 	// Define the test cases
// 	testCases := []struct {
// 		name          string
// 		requestBody   models.RegisterForm
// 		expectedCode  int
// 		expectedBody  string
// 		expectedError error
// 	}{
// 		{
// 			name: "valid request",
// 			requestBody: models.RegisterForm{
// 				Username:  "testUser",
// 				Password:  "password",
// 				Email:     "testemail@test.com",
// 				FirstName: "testFirstName",
// 				LastName:  "testLastName",
// 			},
// 			expectedCode:  201,
// 			expectedBody:  "Register Success",
// 			expectedError: nil,
// 		},
// 		{
// 			name: "missing fields",
// 			requestBody: models.RegisterForm{
// 				Username: "username",
// 			},
// 			expectedCode:  http.StatusBadRequest,
// 			expectedBody:  "register form invalid\n",
// 			expectedError: nil,
// 		},
// 		{
// 			name: "invalid email",
// 			requestBody: models.RegisterForm{
// 				Username:  "testUser",
// 				Password:  "password",
// 				Email:     "invalidEmail",
// 				FirstName: "testFirstName",
// 				LastName:  "testLastName",
// 			},
// 			expectedCode:  http.StatusBadRequest,
// 			expectedBody:  "register form invalid\n",
// 			expectedError: nil,
// 		},
// 	}

// 	// Run the test cases
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Create the request
// 			reqBody, _ := json.Marshal(tc.requestBody)
// 			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(reqBody))
// 			recorder := httptest.NewRecorder()

// 			// Call the RegisterUser handler
// 			h.RegisterUser(recorder, req)

// 			// Check the response code
// 			if recorder.Code != tc.expectedCode {
// 				t.Errorf("expected status code %d but got %d", tc.expectedCode, recorder.Code)
// 			}

// 			// Check the response body
// 			if recorder.Body.String() != tc.expectedBody {
// 				t.Errorf("expected body %s but got %s", tc.expectedBody, recorder.Body.String())
// 			}

// 			// Check for errors
// 			if tc.expectedError != nil {
// 				if recorder.Body.String() != tc.expectedError.Error() {
// 					t.Errorf("expected error %s but got %s", tc.expectedError.Error(), recorder.Body.String())
// 				}
// 			}
// 			if tc.expectedCode == 201 {
// 				user, err := h.AuthService.UserDb.GetUserByUsername(context.Background(), tc.requestBody.Username)
// 				if err != nil {
// 					t.Errorf("got error while looking for username:%s in database", tc.requestBody.Username)
// 				}
// 				if user.Username != tc.requestBody.Username {
// 					t.Errorf("expected username %s but got %s", tc.requestBody.Username, user.Username)
// 				}
// 				if user.Email != tc.requestBody.Email {
// 					t.Errorf("expected email %s but got %s", tc.requestBody.Email, user.Email)
// 				}
// 				if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tc.requestBody.Password)) != nil {
// 					t.Errorf("expected password %s but got %s", tc.requestBody.Password, user.Password)
// 				}

// 			}
// 		})
// 	}
// }

// func TestLoginUser(t *testing.T) {
// 	// Define the test cases
// 	testCases := []struct {
// 		name          string
// 		requestBody   models.LoginForm
// 		expectedCode  int
// 		expectedBody  string
// 		expectedError error
// 	}{
// 		{
// 			name: "valid request",
// 			requestBody: models.LoginForm{
// 				Username: "testUser",
// 				Password: "password",
// 			},
// 			expectedCode:  http.StatusOK,
// 			expectedBody:  "Login Success",
// 			expectedError: nil,
// 		},
// 		{
// 			name: "missing password",
// 			requestBody: models.LoginForm{
// 				Username: "testUser",
// 				Password: "",
// 			},
// 			expectedCode:  http.StatusBadRequest,
// 			expectedBody:  "login form invalid\n",
// 			expectedError: nil,
// 		},
// 		{
// 			name: "missing username",
// 			requestBody: models.LoginForm{
// 				Username: "",
// 				Password: "password",
// 			},
// 			expectedCode:  http.StatusBadRequest,
// 			expectedBody:  "login form invalid\n",
// 			expectedError: nil,
// 		},
// 		{
// 			name: "empty",
// 			requestBody: models.LoginForm{
// 				Username: "",
// 				Password: "",
// 			},
// 			expectedCode:  http.StatusBadRequest,
// 			expectedBody:  "login form invalid\n",
// 			expectedError: nil,
// 		},
// 		{
// 			name: "wrong password",
// 			requestBody: models.LoginForm{
// 				Username: "testUser",
// 				Password: "",
// 			},
// 			expectedCode:  http.StatusBadRequest,
// 			expectedBody:  "login form invalid\n",
// 			expectedError: nil,
// 		},
// 	}
// 	h := Setup()
// 	hash, _ := bcrypt.GenerateFromPassword([]byte("password"), 12)
// 	_, err := h.AuthService.UserDb.CreateUser(context.Background(), users.CreateUserParams{
// 		Username: "testUser",
// 		Password: string(hash),
// 		Email:    "email@test.com",
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	// Run the test cases
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			// Create the request
// 			reqBody, _ := json.Marshal(tc.requestBody)
// 			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
// 			recorder := httptest.NewRecorder()

// 			// Call the RegisterUser handler
// 			h.LoginUser(recorder, req)

// 			// Check the response code
// 			if recorder.Code != tc.expectedCode {
// 				t.Errorf("expected status code %d but got %d", tc.expectedCode, recorder.Code)
// 			}

// 			// Check the response body
// 			if !strings.Contains(recorder.Body.String(), tc.expectedBody) {
// 				t.Errorf("expected body %s but got %s", tc.expectedBody, recorder.Body.String())
// 			}

// 			// Check for errors
// 			if tc.expectedError != nil {
// 				if recorder.Body.String() != tc.expectedError.Error() {
// 					t.Errorf("expected error %s but got %s", tc.expectedError.Error(), recorder.Body.String())
// 				}
// 			}
// 		})
// 	}
// }
