package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type FinalChat struct {
	Message     interface{} `json:"message"`
	UserID      interface{} `json:"user_id"`
	SquadID     interface{} `json:"squad_id"`
	UserName    interface{} `json:"user_name"`
	CreatedAt   interface{} `json:"created_at"`
	MessageType interface{} `json:"message_type"`
}

type Squad struct {
	ID           int        `json:"squad_id"`
	Name         string     `json:"name"`
	SquadProfile string     `json:"squad_profile"`
	Members      int        `json:"members"`
	Category     string     `json:"category"`
	IsPrivate    int        `json:"is_private"`
	Admin        int        `json:"admin"`
	IsAdmin      int        `json:"is_admin"`
	NewMessages  int        `json:"new_messages"`
	FinalChat    FinalChat  `json:"final_chat"`
}

func connectToDB() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DOPAMINE_DB_USER"),
		os.Getenv("DOPAMINE_DB_PASS"),
		os.Getenv("DOPAMINE_DB_HOST"),
		os.Getenv("DOPAMINE_DB_NAME"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
func getSquadList(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key     string `json:"key"`
		Secret  string `json:"secret"`
		APIName string `json:"api_name"`
		UserID  int    `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.Key != "JWT" || request.Secret != "DOPAMINE" || request.APIName != "getSquadList" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	db, err := connectToDB()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
/*
    	err = db.AutoMigrate(&Squad{})
    		if err != nil {
        	log.Fatal(err)
    	}

*/
	var squads []Squad
	result := db.Table("squads").
		Select("squads.id AS squad_id, squads.name, squads.squad_profile, COUNT(sm.id) AS members, category.name AS category, squads.is_private, squads.admin, IF(squads.admin = ?, 1, 0) AS is_admin, 99 AS new_messages, ? AS final_chat", request.UserID, "{}").
		Joins("LEFT JOIN squad_members sm ON squads.id = sm.squad_id AND sm.is_active = 1").
		Joins("LEFT JOIN category ON category.id = squads.category").
		Where("squads.id IN (?)", db.Table("squad_members").Select("squad_id").Where("user_id = ? AND is_active = 1", request.UserID)).
		Group("squads.id").
		Preload("FinalChat").
		Find(&squads)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Success bool   `json:"success"`
		Msg     string `json:"msg"`
		Data    []Squad `json:"data"`
	}{
		Success: true,
		Msg:     "Squad List",
		Data:    squads,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}


func main() {
	router := mux.NewRouter()
	router.HandleFunc("/squad-list", getSquadList).Methods("POST")

	log.Fatal(http.ListenAndServe(":8080", router))
}


