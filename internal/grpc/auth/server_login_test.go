package authgrpc

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/nglmq/password-keeper/internal/storage"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sso "github.com/nglmq/protos/gen/go/sso"
)

type MockAuthLogin struct {
	mock.Mock
}

func (m *MockAuthLogin) Login(ctx context.Context, email string, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthLogin) RegisterNewUser(ctx context.Context, email string, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func Test_serverAPI_Login(t *testing.T) {
	tests := []struct {
		name        string
		mockAuth    func() *MockAuthLogin
		args        *sso.LoginRequest
		want        *sso.LoginResponse
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "Successful login",
			mockAuth: func() *MockAuthLogin {
				m := new(MockAuthLogin)
				m.On("Login", mock.Anything, "qwerty@gmail.com", "12345").Return("valid-token", nil)
				return m
			},
			args: &sso.LoginRequest{
				Email:    "qwerty@gmail.com",
				Password: "12345",
			},
			want: &sso.LoginResponse{
				Token: "valid-token",
			},
			wantErr:     false,
			wantErrCode: codes.OK,
		},
		{
			name: "Missing email",
			mockAuth: func() *MockAuthLogin {
				return new(MockAuthLogin)
			},
			args: &sso.LoginRequest{
				Email:    "",
				Password: "12345",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Missing password",
			mockAuth: func() *MockAuthLogin {
				return new(MockAuthLogin)
			},
			args: &sso.LoginRequest{
				Email:    "qwerty@gmail.com",
				Password: "",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "User not found",
			mockAuth: func() *MockAuthLogin {
				m := new(MockAuthLogin)
				m.On("Login", mock.Anything, "qwerty@mail.com", "12345678").Return("", storage.ErrUserNotFound)
				return m
			},
			args: &sso.LoginRequest{
				Email:    "qwerty@mail.com",
				Password: "12345678",
			},
			wantErr:     true,
			wantErrCode: codes.NotFound,
		},
		{
			name: "Internal error",
			mockAuth: func() *MockAuthLogin {
				m := new(MockAuthLogin)
				m.On("Login", mock.Anything, "qwerty@gmail.com", "12345").Return("", errors.New("internal error"))
				return m
			},
			args: &sso.LoginRequest{
				Email:    "qwerty@gmail.com",
				Password: "12345",
			},
			wantErr:     true,
			wantErrCode: codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := tt.mockAuth()

			s := &serverAPI{
				auth: mockAuth,
			}

			got, err := s.Login(context.Background(), tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Login() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				st, ok := status.FromError(err)
				if !ok {
					t.Errorf("expected gRPC status error, got %v", err)
					return
				}

				if st.Code() != tt.wantErrCode {
					t.Errorf("expected error code %v, got %v", tt.wantErrCode, st.Code())
				}
			} else {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Login() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
