package api

import (
	"cmd/task_bot/internal/app/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

type User struct {
	ID           int64  `json:"id"`
	TelegramID   int64  `json:"telegram_id"`
	Completed    bool   `json:"completed"`
	ProofFileUrl string `json:"proof_file_url"`
}

func ListUser(c *gin.Context) {
	queryString := `SELECT DISTINCT (users.id), users.telegram_id, ut.completed, ut.proof_file_url FROM users JOIN user_tasks ut on users.id = ut.user_id`
	rows, err := utils.DB.Query(queryString)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	// Iterate through the result set
	for rows.Next() {
		var user User
		if err = rows.Scan(&user.ID, &user.TelegramID, &user.Completed, &user.ProofFileUrl); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
