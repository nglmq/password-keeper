package authgrpc

import (
	"context"
	"errors"
	"github.com/nglmq/password-keeper/internal/domain/models"
	"github.com/nglmq/password-keeper/internal/storage"
	sso "github.com/nglmq/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Auth interface {
	Login(ctx context.Context, email string, password string) (token string, err error)
	RegisterNewUser(ctx context.Context, email string, password string) (token string, err error)
}

type Data interface {
	SaveData(ctx context.Context, token, dataType, data string) (string, error)
	GetData(ctx context.Context, token string) (string, []models.Data, error)
}

type serverAPI struct {
	sso.UnimplementedAuthServer
	sso.UnimplementedUserDataServer
	auth Auth
	data Data
}

func Register(gRPC *grpc.Server, auth Auth, data Data) {
	sso.RegisterAuthServer(gRPC, &serverAPI{auth: auth})
	sso.RegisterUserDataServer(gRPC, &serverAPI{data: data})
}

func (s *serverAPI) Login(ctx context.Context, req *sso.LoginRequest) (*sso.LoginResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email should not be empty")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password should not be empty")
	}

	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &sso.LoginResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) Register(ctx context.Context, req *sso.RegisterRequest) (*sso.RegisterResponse, error) {
	if req.GetEmail() == "" {
		return nil, status.Error(codes.InvalidArgument, "email should not be empty")
	}

	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password should not be empty")
	}

	token, err := s.auth.RegisterNewUser(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &sso.RegisterResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) SaveData(ctx context.Context, req *sso.SaveDataRequest) (*sso.SaveDataResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token should not be empty")
	}
	if req.GetDataType() == "" {
		return nil, status.Error(codes.InvalidArgument, "data type should not be empty")
	}

	if req.GetData() == "" {
		return nil, status.Error(codes.InvalidArgument, "data should not be empty")
	}

	token, err := s.data.SaveData(ctx, req.GetToken(), req.GetDataType(), req.GetData())
	if err != nil {
		//if errors.Is(err, storage.ErrUserExists) {
		//	return nil, status.Error(codes.AlreadyExists, "user already exists")
		//}

		return nil, status.Error(codes.Internal, "internal error")
	}

	return &sso.SaveDataResponse{
		Token: token,
	}, nil
}

func (s *serverAPI) GetData(ctx context.Context, req *sso.GetDataRequest) (*sso.GetDataResponse, error) {
	if req.GetToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "token should not be empty")
	}

	token, data, err := s.data.GetData(ctx, req.GetToken())
	if err != nil {
		//if errors.Is(err, storage.ErrUserExists) {
		//	return nil, status.Error(codes.AlreadyExists, "user already exists")
		//}

		return nil, status.Error(codes.Internal, "internal error")
	}

	var grpcData []*sso.Data
	for _, d := range data {
		grpcData = append(grpcData, &sso.Data{
			DataType: d.DataType,
			Content:  d.Content,
		})
	}

	return &sso.GetDataResponse{
		Token: token,
		Data:  grpcData,
	}, nil
}
