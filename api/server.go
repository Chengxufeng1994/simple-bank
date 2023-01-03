package api

import (
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	store  db.Store
	router *gin.Engine
}

func NewServer(store db.Store) *Server {
	srv := &Server{store: store}
	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	router.POST("/users", srv.createUser)

	router.POST("/accounts", srv.createAccount)
	router.GET("/accounts/:id", srv.getAccount)
	router.GET("/accounts", srv.listAccounts)

	router.POST("/transfers", srv.createTransfer)

	srv.router = router
	return srv
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
