package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type TaskService struct {
	DB          *sql.DB
	TaskChannel chan Task
}

func (t *TaskService) AddTask(ts Task) error {
	query := "INSERT INTO tasks (title, description, completed, status, created_at) VALUES (?, ?, ?, ?, ?)"
	_, err := t.DB.Exec((query), ts.Title, ts.Description, ts.Completed, ts.Status, ts.CreatedAt)
	return err
}

func (t *TaskService) UpdateTask(ts Task) error {
	query := "UPDATE tasks SET status = ? WHERE id = ?"
	_, err := t.DB.Exec(query, ts.Status, ts.ID)
	return err
}

func (t *TaskService) ListTasks(ts Task) ([]Task, error) {
	rows, err := t.DB.Query("SELECT * FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.Status,
			&task.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (t *TaskService) ProcessTasks() {
	for task := range t.TaskChannel {
		log.Printf("processing task: %s", task.Title)
		time.Sleep(time.Second * 5)
		task.Status = "completed"
		t.UpdateTask(task)
		log.Printf("task processed: %s", task.Title)
	}
}

func (t *TaskService) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	task.Status = "pending"
	task.CreatedAt = time.Now()
	err = t.AddTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.TaskChannel <- task
	w.WriteHeader(http.StatusCreated)
}

func (t *TaskService) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := t.ListTasks(Task{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tasks)
}
func main() {

	db, err := sql.Open("sqlite3", "./tasks.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	taskService := TaskService{
		DB:          db,
		TaskChannel: make(chan Task),
	}
	go taskService.ProcessTasks()
	http.HandleFunc("POST /tasks", taskService.HandleCreateTask)
	http.HandleFunc("GET /tasks", taskService.HandleListTasks)
	log.Println("server started on port 8001")
	http.ListenAndServe(":8001", nil)

}
