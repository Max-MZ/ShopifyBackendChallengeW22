package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/stretchr/testify/assert"
)

func createRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/upload", uploadPicture).Methods("POST")
	router.HandleFunc("/api/zipupload", bulkUpload).Methods("POST")
	router.HandleFunc("/api/delete", deletion).Methods("DELETE")
	return router
}

func TestUploadBasic(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initSession()

	filename, _ := uuid.NewV4()

	if err != nil {
		panic(err)
	}

	uploadTest := &UploadSingle{
		Filename: filename.String(),
		Filetype: "jpeg",
		Path:     "./uploadtest.jpeg",
		Author:   "uploadtest_user",
	}

	jsonBody, _ := json.Marshal(uploadTest)

	request, _ := http.NewRequest("POST", "/api/upload", bytes.NewBuffer(jsonBody))

	response := httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back
	assert.Equal(t, 200, response.Code, "OK response is expected")

	// assert we have created file
	assert.Equal(t, true, checkExisting(filename.String()+"."+uploadTest.Filetype))

}

func TestUserPermission(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initSession()

	var filesToDelete []string
	filesToDelete = append(filesToDelete, "deletePermissionsTest.jpg")

	deleteTest := &DeletePictures{
		Filenames: filesToDelete,
		Author:    "not_the_correct_user",
	}

	jsonBody, _ := json.Marshal(deleteTest)

	request, _ := http.NewRequest("DELETE", "/api/delete", bytes.NewBuffer(jsonBody))

	response := httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back
	assert.Equal(t, 200, response.Code, "OK response is expected")

	// assert we have kept the file
	assert.Equal(t, true, checkExisting("deletePermissionsTest.jpg"))

}
