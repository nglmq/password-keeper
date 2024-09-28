package tui

import (
	"context"
	"errors"
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	api "github.com/nglmq/password-keeper/internal/clients/sso"
	"github.com/nglmq/password-keeper/internal/domain/models"
	"github.com/nglmq/password-keeper/internal/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
)

type Choice int

const (
	Login Choice = iota + 1
	Register
)

func (c Choice) String() string {
	switch c {
	case Login:
		return "Login"
	case Register:
		return "Register"
	default:
		return ""
	}
}

type User struct {
	Choice   Choice
	Continue bool
	Email    string
	Password string
}

func StartCLI(api *api.Client) {
	var user User
	//
	//form := huh.NewForm(
	//	huh.NewGroup(huh.NewNote().
	//		Title("GophKeeper").
	//		Description("Welcome to _GophKeeper™_.\n\n\n").
	//		Next(true).
	//		NextLabel("Next"),
	//	),
	//
	//	huh.NewGroup(
	//		huh.NewSelect[Choice]().
	//			Title("Choose option").
	//			Options(
	//				huh.NewOption("Login", Login),
	//				huh.NewOption("Register", Register)).
	//			Value(&user.Choice),
	//	),
	//
	//	huh.NewGroup(
	//		huh.NewInput().
	//			Value(&user.Email).
	//			Title("Enter your email").
	//			Placeholder("user@mail.ru").
	//			Validate(func(s string) error {
	//				if s == "" {
	//					return errors.New("email is required")
	//				}
	//				return nil
	//			}),
	//
	//		huh.NewInput().
	//			Value(&user.Password).
	//			Placeholder("Password").
	//			Title("Enter password").
	//			EchoMode(huh.EchoModePassword).
	//			Validate(func(s string) error {
	//				if s == "" {
	//					return errors.New("password is required")
	//				}
	//				return nil
	//			}),
	//
	//		huh.NewConfirm().
	//			Title("Continue?").
	//			Value(&user.Continue).
	//			Affirmative("Yes!").
	//			Negative("No."),
	//	),
	//)

	form := renderForm(&user)
	err := form.Run()

	if err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}

	resp, err := sendData(api, user.Choice, user.Email, user.Password)
	if err != nil {
		fmt.Println("\n\n\nTry again\n\n\n")
		StartCLI(api)
		return
	}

	data, err := api.GetUserData(context.Background(), resp)
	if err != nil {
		if errors.Is(err, storage.ErrDataNotFound) {
			printErrorTable(err)
			return
		}
		fmt.Println("Ошибка при получении данных:", err)
		return
	}

	if err := printDataTable(data); err != nil {
		fmt.Println("Ошибка при выводе данных:", err)
		return
	}

}

func renderForm(user *User) *huh.Form {
	//var user User

	form := huh.NewForm(
		huh.NewGroup(huh.NewNote().
			Title("GophKeeper").
			Description("Welcome to _GophKeeper™_.\n\n\n").
			Next(true).
			NextLabel("Next"),
		),

		huh.NewGroup(
			huh.NewSelect[Choice]().
				Title("Choose option").
				Options(
					huh.NewOption("Login", Login),
					huh.NewOption("Register", Register)).
				Value(&user.Choice),
		),

		huh.NewGroup(
			huh.NewInput().
				Value(&user.Email).
				Title("Enter your email").
				Placeholder("user@mail.ru").
				Validate(func(s string) error {
					if s == "" {
						return errors.New("email is required")
					}
					return nil
				}),

			huh.NewInput().
				Value(&user.Password).
				Placeholder("Password").
				Title("Enter password").
				EchoMode(huh.EchoModePassword).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("password is required")
					}
					return nil
				}),

			huh.NewConfirm().
				Title("Continue?").
				Value(&user.Continue).
				Affirmative("Yes!").
				Negative("No."),
		),
	)

	return form
}

func sendData(api *api.Client, userChoice Choice, userEmail, userPassword string) (string, error) {
	switch {
	case userChoice == Login:
		resp, err := api.Login(context.Background(), userEmail, userPassword)
		if err != nil {
			st, ok := status.FromError(err)

			if ok {
				switch st.Code() {
				case codes.InvalidArgument:
					err = fmt.Errorf("wrong login or password")
					printErrorTable(err)
					return "", err

				case codes.NotFound:
					err = fmt.Errorf("user not found")
					printErrorTable(err)
					return "", err

				case codes.Internal:
					err = fmt.Errorf("internal error")
					printErrorTable(err)
					return "", err

				default:
					fmt.Println("Ошибка при входе:", st.Message())
					printErrorTable(err)
					return "", err
				}
			}
			printErrorTable(err)

			return "", err
		}

		return resp, nil

	case userChoice == Register:
		token, err := api.Register(context.Background(), userEmail, userPassword)
		if err != nil {
			st, ok := status.FromError(err)

			if ok {
				switch st.Code() {
				case codes.InvalidArgument:
					err = fmt.Errorf("wrong login or password")
					printErrorTable(err)
					return "", err

				case codes.AlreadyExists:
					err = fmt.Errorf("user already exists")
					printErrorTable(err)
					return "", err

				case codes.Internal:
					err = fmt.Errorf("internal error")
					printErrorTable(err)
					return "", err

				default:
					fmt.Println("Ошибка при входе:", st.Message())
					printErrorTable(err)
					return "", err
				}
			}
			printErrorTable(err)

			return "", err
		}

		return token, nil
	}

	return "", nil
}

func printErrorTable(err error) {
	rows := [][]string{
		{"Error", err.Error()},
	}

	const (
		red      = lipgloss.Color("#DC2E2E")
		lightRed = lipgloss.Color("#E77373")
	)

	re := lipgloss.NewRenderer(os.Stdout)

	var (
		HeaderStyle  = re.NewStyle().Foreground(red).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(30)
		OddRowStyle  = CellStyle.Foreground(lightRed)
		EvenRowStyle = CellStyle.Foreground(lightRed)
		BorderStyle  = lipgloss.NewStyle().Foreground(red)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return HeaderStyle
			case row%2 == 0:
				return EvenRowStyle
			default:
				return OddRowStyle
			}
		}).
		Rows(rows...)

	fmt.Println(t)

}

func printDataTable(data []models.Data) error {
	var dataRows [][]string

	for _, d := range data {
		dataRow := []string{
			d.DataType,
			d.Content,
		}

		dataRows = append(dataRows, dataRow)
	}

	const (
		purple    = lipgloss.Color("#7E70FF")
		gray      = lipgloss.Color("#DBD7FF")
		lightGray = lipgloss.Color("#DBD7FF")
	)

	re := lipgloss.NewRenderer(os.Stdout)

	var (
		HeaderStyle  = re.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		CellStyle    = re.NewStyle().Padding(0, 1).Width(30)
		OddRowStyle  = CellStyle.Foreground(gray)
		EvenRowStyle = CellStyle.Foreground(lightGray)
		BorderStyle  = lipgloss.NewStyle().Foreground(purple)
	)

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == 0:
				return HeaderStyle
			case row%2 == 0:
				return EvenRowStyle
			default:
				return OddRowStyle
			}
		}).
		Headers("TYPE", "DATA").
		Rows(dataRows...)

	fmt.Println(t)

	return nil
}
