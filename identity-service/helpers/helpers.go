package helpers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bernardn38/socialsphere/identity-service/models"
	"github.com/cristalhq/jwt/v4"
	"github.com/go-chi/chi/v5"
	amqp "github.com/rabbitmq/amqp091-go"
)

func ResponseWithJson(w http.ResponseWriter, statusCode int, payload models.JsonResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_, _ = w.Write(jsonData)
}

func ResponseWithPayload(w http.ResponseWriter, responseCode int, payload []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	_, _ = w.Write(payload)
}
func ResponseNoPayload(w http.ResponseWriter, responseCode int) {
	w.WriteHeader(responseCode)
}

func ConvertPostId(postId string) (int32, error) {
	parsedPostId, err := strconv.ParseInt(postId, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(parsedPostId), nil
}

func ConvertUserId(userId any) (int32, error) {
	stringUserId, ok := userId.(string)
	if !ok {
		log.Println(stringUserId)
		return 0, errors.New("invalid userId")
	}
	userId64, err := strconv.ParseInt(stringUserId, 10, 32)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return int32(userId64), nil
}

func GetUserIdFromRequest(r *http.Request, checkContext bool) (int32, error) {
	var userId int32

	urlUserId := chi.URLParam(r, "userId")
	log.Println(urlUserId)
	if len(urlUserId) > 0 {
		convertedId, err := ConvertUserId(urlUserId)
		if err != nil {
			log.Println(err)
			return 0, errors.New("invalid user id url")
		}
		userId = convertedId
	} else if checkContext {
		contextUserId := r.Context().Value("userId")
		convertedUserId, err := ConvertUserId(contextUserId)
		if err != nil {
			log.Println(err)
			return 0, errors.New("invalid user id context")
		}
		userId = convertedUserId
	}
	return userId, nil

}

func ConnectToRabbitMQ(rabbitUrl string) *amqp.Connection {
	backOff := time.Second * 5
	for {
		conn, err := amqp.Dial(rabbitUrl)
		if err != nil {
			log.Println("Connection not ready backing off for ", backOff)
			time.Sleep(backOff)
			backOff = backOff + (time.Second * 5)
		} else {
			log.Println("Connected to rabbit ")
			return conn
		}
	}
}

func GetEnvConfig() models.Config {
	//get configuration from enviroment and validate
	postgresUrl := os.Getenv("DSN")
	jwtSecret := os.Getenv("jwtSecret")
	rabbitMQUrl := os.Getenv("rabbitMQUrl")
	port := os.Getenv("port")
	config := models.Config{
		JwtSecretKey:     jwtSecret,
		JwtSigningMethod: jwt.Algorithm(jwt.HS256),
		PostgresUrl:      postgresUrl,
		RabbitmqUrl:      rabbitMQUrl,
		Port:             port,
	}
	err := config.Validate()
	if err != nil {
		log.Fatal(err.Error())
	}
	return config
}
