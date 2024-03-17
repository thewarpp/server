package handler

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"server/internal/database"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/guregu/null/v5"
)

type Workspace struct {
	Id        string      `json:"id"`
	Name      string      `json:"name"`
	EditorId  null.String `json:"editor_id" db:"editor_id"`
	CreatorId null.String `json:"creator_id" db:"creator_id"`
	CreatedAt string      `json:"created_at" db:"created_at"`
	UpdatedAt string      `json:"updated_at" db:"updated_at"`
}

type CreateWorkspaceBody struct {
	Name      string `json:"name"`
	CreatorId string `json:"creator_id"`
}

type PatchWorkspace struct {
	Name     *string `json:"name,omitempty"`
	EditorId *string `json:"editor_id,omitempty"`
}

func HandleGetWorkspaces(w http.ResponseWriter, r *http.Request) {
	var workspaces []Workspace
	db := database.GetDB()

	query := "select * from workspace"
	filterMap := make(map[string]interface{})

	creatorID := r.URL.Query().Get("creator_id")
	if creatorID != "" {
		query += " where creator_id = :creator_id"
		filterMap["creator_id"] = creatorID
	}

	namedStatement, err := db.PrepareNamed(query)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	defer namedStatement.Close()

	err = namedStatement.Select(&workspaces, filterMap)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(workspaces)
	if err != nil {
		http.Error(w, "Error marshalling data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func HandleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	db := database.GetDB()

	var workspace Workspace
	query := "select * from workspace where id = $1"
	err := db.QueryRowx(query, id).StructScan(&workspace)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatal(err)
			http.Error(w, "workspace not found", http.StatusNotFound)
		}

		log.Fatal(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(workspace)
	if err != nil {
		http.Error(w, "Error marshalling data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func HandlePostWorkspace(w http.ResponseWriter, r *http.Request) {
	var workspace CreateWorkspaceBody
	err := json.NewDecoder(r.Body).Decode(&workspace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := database.GetDB()

	query := "insert into workspace (name, updated_at, creator_id) values ($1, $2, $3) returning id"
	row := db.QueryRowx(query, workspace.Name, time.Now(), workspace.CreatorId)

	var workspaceID string
	err = row.Scan(&workspaceID)

	if err != nil {
		log.Fatal(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		WorkspaceID string `json:"id"`
	}{
		WorkspaceID: workspaceID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error marshalling data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func HandlePatchWorkspace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")

	var patchWorkspace PatchWorkspace
	err := json.NewDecoder(r.Body).Decode(&patchWorkspace)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	updateMap := make(map[string]interface{})
	if patchWorkspace.Name != nil {
		updateMap["name"] = *patchWorkspace.Name
	}
	if patchWorkspace.EditorId != nil {
		updateMap["editor_id"] = *patchWorkspace.EditorId
	}

	if len(updateMap) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	var updateClause string
	for key := range updateMap {
		updateClause += key + " = :" + key + ","
	}

	updateQuery := "update workspace set "
	updateQuery += updateClause[:len(updateClause)-1] // Remove trailing comma
	updateQuery += "where id = :id"

	db := database.GetDB()
	stmt, err := db.PrepareNamed(updateQuery)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	allArgs := map[string]interface{}{}
	allArgs["id"] = id
	for key, value := range updateMap {
		allArgs[key] = value
	}

	_, err = stmt.Exec(allArgs)
	if err != nil {
		http.Error(w, "Error updating workspace", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func HandleDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "workspaceID")
	db := database.GetDB()

	query := "delete from workspace where id = $1 returning id"
	row := db.QueryRowx(query, id)

	var deletedID string
	err := row.Scan(&deletedID)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Fatal(err)
			http.Error(w, "workspace not found", http.StatusNotFound)
		}

		log.Fatal(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		DeletedID string `json:"deleted_id"`
	}{
		DeletedID: deletedID,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error marshalling data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
