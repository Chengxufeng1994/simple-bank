package gapi

import (
	"context"
	db "github.com/Chengxufeng1994/simple-bank/db/sqlc"
	"github.com/Chengxufeng1994/simple-bank/pb"
	"github.com/Chengxufeng1994/simple-bank/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (srv *Server) LoginUser(ctx context.Context, in *pb.LoginUserRequest) (*pb.LoginUserResponse, error) {
	hashedPassword, err := util.HashPassword(in.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password: %s", err)
	}

	user, err := srv.store.GetUser(ctx, in.Username)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to login user: %s", err)
	}

	err = util.CheckPassword(in.Password, hashedPassword)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to login user: %s", err)
	}

	accessToken, accessPayload, err := srv.tokenMaker.CreateToken(user.Username, srv.config.AccessTokenDuration)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create access token")
	}

	refreshToken, refreshPayload, err := srv.tokenMaker.CreateToken(
		user.Username,
		srv.config.RefreshTokenDuration,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create refresh token")
	}

	mtdt := srv.extractMetadata(ctx)
	session, err := srv.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    mtdt.UserAgent,
		ClientIp:     mtdt.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create session")
	}

	rsp := &pb.LoginUserResponse{
		User:                  convertUser(user),
		SessionId:             session.ID.String(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  timestamppb.New(accessPayload.ExpiredAt),
		RefreshTokenExpiresAt: timestamppb.New(refreshPayload.ExpiredAt),
	}
	return rsp, nil
}
