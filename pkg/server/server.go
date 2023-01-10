package server

import (
	"github.com/zhashkevych/go-pocket-sdk"
	"net/http"
	"strconv"
	"telegram-bot-pocket/pkg/repository"
)

// тут создаётся новый сервер, который будет слушать поступающие запросы на порт :8000

type AuthorizationServer struct {
	server          *http.Server
	pocketClient    *pocket.Client
	tokenRepository repository.TokenRepository
	redirectURL     string
}

func NewAuthorizationServer(pocketClient *pocket.Client, tokenRepository repository.TokenRepository, redirectURL string) *AuthorizationServer {
	return &AuthorizationServer{
		pocketClient:    pocketClient,
		tokenRepository: tokenRepository,
		redirectURL:     redirectURL}
}

func (s *AuthorizationServer) Start() error {
	s.server = &http.Server{
		Addr:    ":8080",
		Handler: s, // вот тут интересно.
		/*
			для создания сервера надо передать Handler ( обработчик )
			обработчик должен реализовавыват интерфейс
			ServeHTTP(ResponseWriter, *Request)

			т.к. у структуры AuthorizationServer реализован метод ServeHTTP
			с указанными параметрами, то и вся структура подходит под Handler
		*/
	}

	return s.server.ListenAndServe()
}

// это обработчик всех входящих запросов
func (s *AuthorizationServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// тут мы говорим, что кроме метода Get ничего не принимаем
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// тут проверяем что в строке URL есть chat_id и он не пустой
	chatIDParam := r.URL.Query().Get("chat_id")
	if chatIDParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	chatID, err := strconv.ParseInt(chatIDParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	requestToken, err := s.tokenRepository.Get(chatID, repository.RequestTokens)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	authResp, err := s.pocketClient.Authorize(r.Context(), requestToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = s.tokenRepository.Save(chatID, authResp.AccessToken, repository.AccessTokens)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Header() возвращает хэадеры, которые будут записаны в ответ
	w.Header().Add("Location", s.redirectURL)
	w.WriteHeader(http.StatusMovedPermanently)
}
