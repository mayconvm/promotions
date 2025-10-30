package supabase

import (
	"bot-telegram/src/internal/domain"
	"os"

	"github.com/go-faster/errors"
	"github.com/supabase-community/supabase-go"
)

func NewClient() (*supabase.Client, error) {
	API_URL := os.Getenv("SUPABASE_URL")
	API_KEY := os.Getenv("SUPABASE_KEY")
	USER := os.Getenv("SUPABASE_USER")
	PASSWORD := os.Getenv("SUPABASE_PASSWORD")

	client, err := supabase.NewClient(API_URL, API_KEY, &supabase.ClientOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "[SUPABASE] Failed to initalize the client: ")
	}

	session, err := client.SignInWithEmailPassword(USER, PASSWORD)
	if err != nil {
		return nil, errors.Wrap(err, "[SUPABASE] Sign in failed")
	}
	client.EnableTokenAutoRefresh(session)

	return client, nil
}

func GetAllSessions(client *supabase.Client) ([]domain.Session, error) {
	var sessions []domain.Session

	_, err := client.From("sessions").Select("*", "id", false).ExecuteTo(&sessions)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func GetAllProducts(client *supabase.Client, session *domain.Session) ([]domain.Product, error) {
	var products []domain.Product

	_, err := client.From("products").Select("*", "id", false).ExecuteTo(&products)
	if err != nil {
		return nil, err
	}

	return products, nil
}
