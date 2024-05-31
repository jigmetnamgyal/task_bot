package utils

import (
	"database/sql"
	_ "github.com/lib/pq"
	storagego "github.com/supabase-community/storage-go"
	"log"
	"os"
)

var DB *sql.DB

var SupabaseClient *storagego.Client

func ConnectToDb() {
	var connectionStr string

	log.Println(os.Getenv("ENVIRONMENT"))

	if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
		connectionStr = os.Getenv("DATABASE_URL")
	} else {
		connectionStr = os.Getenv("TEST_DATABASE_URL")
	}

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		log.Fatal("Error connecting to db")
	}

	if err = db.Ping(); err != nil {
		log.Fatal("Failed to create db connection:" + err.Error())
	}

	DB = db

	sbClient := storagego.NewClient(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_SECRET"), nil)
	SupabaseClient = sbClient
}

func AddUser(telegramID int64) error {
	_, err := DB.Exec("INSERT INTO users (telegram_id) VALUES ($1) ON CONFLICT (telegram_id) DO NOTHING", telegramID)
	return err
}

func CompleteTask(telegramID int64, tID string) error {
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = DB.Exec("INSERT INTO user_tasks (user_id, task_id, completed) VALUES ($1, $2, TRUE) ON CONFLICT (user_id, task_id) DO NOTHING", userID, tID)
	return err
}

func GetUserPoints(telegramID int64) (map[string]int, error) {
	var userID int
	points := make(map[string]int)

	err := DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	if err != nil {
		return nil, err
	}

	rows, err := DB.Query(`
        SELECT t.name, t.points
        FROM tasks t
        JOIN user_tasks ut ON t.id = ut.task_id
        WHERE ut.user_id = $1 AND ut.completed = TRUE
    `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskName string
		var taskPoints int
		if err = rows.Scan(&taskName, &taskPoints); err != nil {
			return nil, err
		}
		points[taskName] += taskPoints
	}

	return points, nil
}
