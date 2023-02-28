package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/bernardn38/socialsphere/friend-service/helpers"
	"github.com/bernardn38/socialsphere/friend-service/models"
	"github.com/bernardn38/socialsphere/friend-service/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	friendService *service.FriendService
}

func NewHandler(friendService *service.FriendService) *Handler {
	return &Handler{friendService: friendService}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	// read and validate request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		http.Error(w, "error reading body", http.StatusBadRequest)
		return
	}
	userForm, err := ValidateUserForm(body)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err = h.friendService.CreateUser(models.CreateUserForm(userForm))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, http.StatusCreated)

}

func (h *Handler) CreateFollow(w http.ResponseWriter, r *http.Request) {
	friendId := chi.URLParam(r, "friendId")
	if len(friendId) < 1 {
		log.Println("friend id not found")
		http.Error(w, "friend id not provided", http.StatusBadRequest)
		return
	}
	friendIdi64, err := strconv.ParseInt(friendId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	username, ok := r.Context().Value("username").(string)
	if !ok {
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	_, err = h.friendService.CreateUserFollow(int32(userIdi64), username, int32(friendIdi64), r.Header.Get("Cookie"))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseNoPayload(w, 201)
}

func (h *Handler) FindFriends(w http.ResponseWriter, r *http.Request) {
	findFriendsForm, err := ValidateFindFriendsForm(r)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	friends, err := h.friendService.FindFriends(findFriendsForm)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if len(friends) == 0 {
		http.Error(w, "No users found", http.StatusNotFound)
		return
	}
	respPayload, err := json.Marshal(friends)
	if err != nil {
		http.Error(w, "error writing json resp", http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, http.StatusOK, respPayload)
}

func (h *Handler) CheckFollow(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	friendB := chi.URLParam(r, "friendId")
	friendBi64, err := strconv.ParseInt(friendB, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	followStatus, err := h.friendService.CheckFollow(int32(userIdi64), int32(friendBi64))
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.ResponseWithPayload(w, http.StatusOK, []byte(fmt.Sprintf("%v", followStatus)))
}

func (h *Handler) GetFriendsLastestPhotos(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	userFollows, err := h.friendService.GetLatestFriendPhotos(int32(userIdi64))
	if err != nil {
		log.Println(err)
		http.Error(w, "no follows found", http.StatusNotFound)
		return
	}
	resp, err := json.Marshal(userFollows)
	if err != nil {
		log.Println(err)
	}
	helpers.ResponseWithPayload(w, http.StatusOK, resp)
}

func (h *Handler) GetFriends(w http.ResponseWriter, r *http.Request) {
	userId, ok := r.Context().Value("userId").(string)
	if !ok {
		log.Println("error parsing userId to string")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	userIdi64, err := strconv.ParseInt(userId, 10, 32)
	if err != nil {
		log.Println(err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	friendIds, err := h.friendService.GetFriends(int32(userIdi64))
	if err != nil {
		log.Println(err)
		http.Error(w, "no friends found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string][]int32{"friendIds": friendIds})

	w.WriteHeader(http.StatusOK)
}
