package api

import (
	"context"
	"fmt"
	"github.com/nglmq/password-keeper/internal/domain/models"
	"github.com/nglmq/protos/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
)

type Client struct {
	apiAuth sso.AuthClient
	apiData sso.UserDataClient
	log     *slog.Logger
}

func New(log *slog.Logger, addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	//conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	//if err != nil {
	//	return nil, fmt.Errorf("failed to create client: %w", err)
	//}

	return &Client{
		apiAuth: sso.NewAuthClient(conn),
		apiData: sso.NewUserDataClient(conn),
		log:     log,
	}, nil
}

func (c *Client) Register(ctx context.Context, email, password string) (string, error) {
	resp, err := c.apiAuth.Register(ctx, &sso.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to register: %w", err)
	}

	return resp.Token, nil
}

func (c *Client) Login(ctx context.Context, email, password string) (string, error) {
	resp, err := c.apiAuth.Login(ctx, &sso.LoginRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return "", err
	}

	return resp.Token, nil
}

func (c *Client) GetUserData(ctx context.Context, token string) ([]models.Data, error) {
	resp, err := c.apiData.GetData(ctx, &sso.GetDataRequest{
		Token: token,
	})
	if err != nil {
		return []models.Data{}, fmt.Errorf("failed to get user data: %w", err)
	}

	var dataModel []models.Data
	for _, d := range resp.Data {
		dataModel = append(dataModel, models.Data{
			DataType: d.DataType,
			Content:  d.Content,
		})
	}

	return dataModel, nil
}

func (c *Client) SaveUserData(ctx context.Context, token, dataType, data string) error {
	_, err := c.apiData.SaveData(ctx, &sso.SaveDataRequest{
		Token:    token,
		DataType: dataType,
		Data:     data,
	})
	if err != nil {
		return fmt.Errorf("failed to save user data: %w", err)
	}

	return nil
}
