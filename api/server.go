package api

import (
	"fmt"
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	"github.com/Chengxufeng1994/simple-bank/token"
	"github.com/Chengxufeng1994/simple-bank/util"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	srv := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}

	srv.setupRouter()
	return srv, nil
}

func (srv *Server) setupRouter() {
	router := gin.Default()

	router.POST("/users", srv.createUser)
	router.POST("/users/login", srv.loginUser)
	router.POST("/token/renew_access", srv.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(srv.tokenMaker))

	authRoutes.POST("/accounts", srv.createAccount)
	authRoutes.GET("/accounts/:id", srv.getAccount)
	authRoutes.GET("/accounts", srv.listAccounts)

	authRoutes.POST("/transfers", srv.createTransfer)

	srv.router = router
}

func (srv *Server) Start(address string) error {
	return srv.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
