package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bernardn38/socialsphere/identity-service/helpers"
	"github.com/bernardn38/socialsphere/identity-service/models"

	"github.com/bernardn38/socialsphere/identity-service/service"
)

//	type Handler struct {
//		UserDb    *users.Queries
//		Emitter   *imageServiceBroker.RabbitBroker
//		RpcClient *rpcbroker.RpcClient
//	}
type Handler struct {
	identityService *service.IdentityService
}

func NewHandler(service *service.IdentityService) *Handler {
	return &Handler{
		identityService: service,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// parse user form from request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}
	userForm := models.UserForm{}
	json.Unmarshal(body, &userForm)

	//validate user form
	err = userForm.Validate()
	if err != nil {
		log.Println(err)
		http.Error(w, "user form invalid", http.StatusBadRequest)
		return
	}

	err = h.identityService.CreateUser(userForm)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, http.StatusCreated)
}

func (h *Handler) CreateUserProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	//parse image from form
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		http.Error(w, "File too large", http.StatusRequestEntityTooLarge)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	err = h.identityService.CreateUserProfileImage(userId, file, header.Header.Get("Content-Type"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, http.StatusCreated)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing return error
	userId, err := helpers.GetUserIdFromRequest(r, false)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.identityService.GetUser(userId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//respond with json payload of user data
	jsonResponse, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, 200, []byte(jsonResponse))
}

func (h *Handler) GetUserProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, err := helpers.GetUserIdFromRequest(r, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	imageId, err := h.identityService.GetImageIdByUserId(userId)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fmt.Sprintf(`{"profileImageId":"%s"}`, imageId))
}

// send image to client for userid found in jwt token
func (h *Handler) GetOwnProfileImage(w http.ResponseWriter, r *http.Request) {
	// get user id from url param if missing use jwt token user id
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("could not get userId from context")
		http.Error(w, "", http.StatusInternalServerError)
	}
	convertedUserId, err := helpers.ConvertUserId(userId)
	if err != nil {
		log.Println(err)
	}
	imageBytes, err := h.identityService.GetUserProfileImage(convertedUserId)
	if err != nil {
		log.Println(err)
		http.Error(w, "error getting image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "max-age=86400") //
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(imageBytes)
}
