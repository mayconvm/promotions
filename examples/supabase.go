package examples

import (
	"fmt"
	"log"
	"os"

	"github.com/supabase-community/supabase-go"
)

func supabaseConnection() {
	API_URL := os.Getenv("SUPABASE_URL")
	API_KEY := os.Getenv("SUPABASE_KEY")

	client, err := supabase.NewClient(API_URL, API_KEY, &supabase.ClientOptions{})
	if err != nil {
		fmt.Println("Failed to initalize the client: ", err)
	}

	session, err := client.SignInWithEmailPassword("mayconvm@gmail.com", "Qaz@wsx3")
	if err != nil {
		log.Fatal("Sign in failed:", err)
	}
	client.EnableTokenAutoRefresh(session)

	data, _, err := client.From("products").Select("*", "id", false).ExecuteString()

	print(data)
}
