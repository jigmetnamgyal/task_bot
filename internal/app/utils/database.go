package utils

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var DB *sql.DB

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
}

func AddUser(telegramID int64) error {
	_, err := DB.Exec("INSERT INTO users (telegram_id) VALUES ($1) ON CONFLICT (telegram_id) DO NOTHING", telegramID)
	return err
}

func CompleteTask(telegramID int64, tID string, url string) error {
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	if err != nil {
		return err
	}

	_, err = DB.Exec("INSERT INTO user_tasks (user_id, task_id, completed, proof_file_url) VALUES ($1, $2, TRUE, $3)", userID, tID, url)
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
        SELECT st.name, st.points
        FROM sub_tasks st
        LEFT JOIN user_sub_tasks ut ON st.id = ut.sub_task_id
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

type Task struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Links        *string `json:"links"`
	Descriptions string  `json:"descriptions"`
	TotalTasks   int     `json:"total_tasks"`
}

func GetUnCompletedTasks(telegramID int64, offset int64) (*Task, error) {
	var task Task
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	if err != nil {
		return nil, err
	}

	queryString := `
		 WITH user_incomplete_tasks AS (
			SELECT t.id AS id, t.name, t.links, t.descriptions, COUNT(t.id) as total_tasks
			FROM tasks t
					 JOIN sub_tasks st ON t.id = st.task_id
					 LEFT JOIN user_sub_tasks ust ON st.id = ust.sub_task_id AND ust.user_id = $1
			WHERE ust.completed = FALSE OR ust.user_id IS NULL
			GROUP BY t.id, t.name, t.links, t.descriptions
		)
		SELECT *
		FROM user_incomplete_tasks
		ORDER BY id
		LIMIT 1 OFFSET $2;
	`

	prepare, err := DB.Prepare(queryString)
	if err != nil {
		return nil, err
	}

	err = prepare.QueryRow(userID, offset).Scan(&task.ID, &task.Name, &task.Links, &task.Descriptions, &task.TotalTasks)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func GetTotalNumberOfTasks(telegramID int64) (*int64, error) {
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	if err != nil {
		fmt.Println("here")
		return nil, err
	}

	queryString := `
	 SELECT COUNT(*)
		FROM (
         SELECT t.id
         FROM tasks t
                  JOIN sub_tasks st ON t.id = st.task_id
                  LEFT JOIN user_sub_tasks ust ON st.id = ust.sub_task_id AND ust.user_id = ($1)
         WHERE ust.completed = FALSE OR ust.user_id IS NULL
         GROUP BY t.id
     ) AS incomplete_tasks;
	`

	var totalTask int64
	err = DB.QueryRow(queryString, userID).Scan(&totalTask)
	if err != nil {
		return nil, err
	}

	return &totalTask, nil
}

type SubTask struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Links         *string `json:"links"`
	Descriptions  string  `json:"descriptions"`
	TotalSubTasks int     `json:"total_sub_tasks"`
	Points        int64   `json:"points"`
	TaskID        int     `json:"task_id"`
}

func GetUnCompletedSubTasks(telegramID int64, tID int, offset int64) (*SubTask, error) {
	var subTask SubTask
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	if err != nil {
		fmt.Println("here")
		return nil, err
	}

	queryString := `
		WITH user_incomplete_sub_tasks AS (
			SELECT st.id AS id, st.name, st.links, st.description, st.points, st.task_id, COUNT(st.id) as total_sub_tasks
			FROM sub_tasks st
				LEFT JOIN user_sub_tasks ust ON st.id = ust.sub_task_id AND ust.user_id = $1
			WHERE st.task_id = ($2) AND (ust.completed = FALSE OR ust.user_id IS NULL)
			GROUP BY st.id, st.name, st.links, st.description, st.points, st.task_id
		)
		SELECT *
		FROM user_incomplete_sub_tasks
		ORDER BY id
		LIMIT 1 OFFSET $3;
	`

	prepare, err := DB.Prepare(queryString)
	if err != nil {
		return nil, err
	}

	err = prepare.QueryRow(userID, tID, offset).Scan(&subTask.ID, &subTask.Name, &subTask.Links, &subTask.Descriptions, &subTask.Points, &subTask.TaskID, &subTask.TotalSubTasks)
	if err != nil {
		return nil, err
	}

	return &subTask, nil
}

func GetTotalNumberOfSubTasks(telegramID int64, tID int) (*int64, error) {
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE telegram_id = $1", telegramID).Scan(&userID)
	if err != nil {
		fmt.Println("here")
		return nil, err
	}

	queryString := `
	 SELECT COUNT(*)
		FROM (
         SELECT st.id
         FROM sub_tasks st
                  LEFT JOIN user_sub_tasks ust ON st.id = ust.sub_task_id AND ust.user_id = ($1)
         WHERE ust.completed = FALSE OR ust.user_id IS NULL AND st.task_id = $2
         GROUP BY st.id
     ) AS incomplete_tasks;
	`

	var totalSubTask int64
	err = DB.QueryRow(queryString, userID, tID).Scan(&totalSubTask)
	if err != nil {
		return nil, err
	}

	return &totalSubTask, nil
}
