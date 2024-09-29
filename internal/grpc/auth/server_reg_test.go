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

// Мок для Auth (реализует интерфейс Auth)
type MockAuthReg struct {
	mock.Mock
}

func (m *MockAuthReg) Login(ctx context.Context, email string, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthReg) RegisterNewUser(ctx context.Context, email string, password string) (string, error) {
	args := m.Called(ctx, email, password)
	return args.String(0), args.Error(1)
}

func Test_serverAPI_Register(t *testing.T) {
	tests := []struct {
		name        string
		mockAuth    func() *MockAuthReg
		args        *sso.RegisterRequest
		want        *sso.RegisterResponse
		wantErr     bool
		wantErrCode codes.Code
	}{
		{
			name: "Successful registration",
			mockAuth: func() *MockAuthReg {
				m := new(MockAuthReg)
				m.On("RegisterNewUser", mock.Anything, "user@example.com", "password123").Return("valid-token", nil)
				return m
			},
			args: &sso.RegisterRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			want: &sso.RegisterResponse{
				Token: "valid-token",
			},
			wantErr:     false,
			wantErrCode: codes.OK,
		},
		{
			name: "Missing email",
			mockAuth: func() *MockAuthReg {
				return new(MockAuthReg)
			},
			args: &sso.RegisterRequest{
				Email:    "",
				Password: "password123",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "Missing password",
			mockAuth: func() *MockAuthReg {
				return new(MockAuthReg)
			},
			args: &sso.RegisterRequest{
				Email:    "user@example.com",
				Password: "",
			},
			wantErr:     true,
			wantErrCode: codes.InvalidArgument,
		},
		{
			name: "User already exists",
			mockAuth: func() *MockAuthReg {
				m := new(MockAuthReg)
				m.On("RegisterNewUser", mock.Anything, "user@example.com", "password123").Return("", storage.ErrUserExists)
				return m
			},
			args: &sso.RegisterRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			wantErr:     true,
			wantErrCode: codes.AlreadyExists,
		},
		{
			name: "Internal error",
			mockAuth: func() *MockAuthReg {
				m := new(MockAuthReg)
				m.On("RegisterNewUser", mock.Anything, "user@example.com", "password123").Return("", errors.New("internal error"))
				return m
			},
			args: &sso.RegisterRequest{
				Email:    "user@example.com",
				Password: "password123",
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

			got, err := s.Register(context.Background(), tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
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
					t.Errorf("Register() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
