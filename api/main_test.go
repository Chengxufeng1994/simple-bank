package api

import (
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	"github.com/Chengxufeng1994/simple-bank/util"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

func newTestSrv(t *testing.T, store db.Store) *Server {
	config := util.Config{
		TokenSymmetricKey:   util.RandomString(32),
		AccessTokenDuration: time.Minute,
	}
	srv, err := NewServer(config, store)
	require.NoError(t, err)

	return srv
}
