package httpserver

import (
	"chatserver/internal/client"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type HttpServer struct {
	Router      *mux.Router
	RedisClient *redis.Client
	clients     map[string]*client.Client
	mu          *sync.Mutex
}

type IChatAPI interface {
	PostMessages(w http.ResponseWriter, r *http.Request)
	QueryMessages(w http.ResponseWriter, r *http.Request)
}

type PostMessageRequest struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type PostMessageResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type QueryMessageRequest struct {
	Name string `json:"name"`
}

type QueryMessageResponse struct {
	Msgs []string `json:"messages"`
}

func NewHttpServer(rdc *redis.Client, clients map[string]*client.Client, mu *sync.Mutex) *HttpServer {
	return &HttpServer{
		Router:      mux.NewRouter(),
		RedisClient: rdc,
		clients:     clients,
		mu:          mu,
	}
}

func (h *HttpServer) AddRoutes() {
	h.Router.Methods("POST").Path("/message").HandlerFunc(h.PostMessages)
	h.Router.Methods("Get").Path("/message/{userName}").HandlerFunc(h.QueryMessages)
	h.Router.Methods("Get").Path("/").HandlerFunc(h.Home)
}

func (h *HttpServer) PostMessages(w http.ResponseWriter, r *http.Request) {
	req := PostMessageRequest{}
	json.NewDecoder(r.Body).Decode(&req)
	resp := PostMessageResponse{}
	log.Println(req)
	h.mu.Lock()
	c, ok := h.clients[req.Name]
	h.mu.Unlock()
	log.Printf("\n clients %+v", h.clients)
	if !ok {
		resp.Status = "Failed"
		resp.Error = "client not found"
		j, _ := json.Marshal(&resp)
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(j)
		return
	}
	go c.Send(req.Message, req.Name)
	resp.Status = "Success"
	j, _ := json.Marshal(&resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)
}

func (h *HttpServer) Home(w http.ResponseWriter, r *http.Request) {
	log.Println("in home")
}

func (h *HttpServer) QueryMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := vars["userName"]
	if userName == "" {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	msgs, err := h.RedisClient.LRange(userName, 0, -1).Result()
	if err != nil {
		http.Error(w, "Failed to get messages", http.StatusInternalServerError)
	}
	resp := QueryMessageResponse{
		Msgs: msgs,
	}
	j, _ := json.Marshal(&resp)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(j)
}
