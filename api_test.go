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

// handle new router
func createRouter() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/upload", uploadPicture).Methods("POST")
	router.HandleFunc("/api/zipupload", bulkUpload).Methods("POST")
	router.HandleFunc("/api/delete", deletion).Methods("DELETE")
	return router
}

// basic upload and delete test
// upload a file, check it exists
// delete the file, ensure it is gone
func TestUploadAndDeleteBasic(t *testing.T) {

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

	/*
		DELETION ATTEMPT
	*/

	var filesToDelete []string
	filesToDelete = append(filesToDelete, "uploadtest.jpeg")

	deleteTest := &DeletePictures{
		Filenames: filesToDelete,
		Author:    "testuser",
	}

	jsonBody, _ = json.Marshal(deleteTest)

	request, _ = http.NewRequest("DELETE", "/api/delete", bytes.NewBuffer(jsonBody))

	response = httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back, attempting to delete something that exists fails otherwise
	assert.Equal(t, 200, response.Code, "OK response is expected")

	// assert file doesn't exist anymore
	assert.Equal(t, false, checkExisting(filename.String()))
}

// test uploading a zip
// zip contains photos Birds.jpg and Cat03.jpg
func TestZipUpload(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initSession()

	if err != nil {
		panic(err)
	}

	uploadTest := &UploadZip{
		Filename: "ziptest",
		Path:     "./ziptest.zip",
		Author:   "ziptest_user",
	}

	jsonBody, _ := json.Marshal(uploadTest)

	request, _ := http.NewRequest("POST", "/api/zipupload", bytes.NewBuffer(jsonBody))

	response := httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back
	assert.Equal(t, 200, response.Code, "OK response is expected")

	// assert we have created files
	assert.Equal(t, true, checkExisting("Cat03.jpg"))
	assert.Equal(t, true, checkExisting("Birds.jpg"))

	/*
		DELETION ATTEMPT
	*/

	var filesToDelete []string
	filesToDelete = append(filesToDelete, "Cat03.jpg")
	filesToDelete = append(filesToDelete, "Birds.jpg")

	deleteTest := &DeletePictures{
		Filenames: filesToDelete,
		Author:    "ziptest_user",
	}

	jsonBody, _ = json.Marshal(deleteTest)

	request, _ = http.NewRequest("DELETE", "/api/delete", bytes.NewBuffer(jsonBody))

	response = httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back, attempting to delete something that exists fails otherwise
	assert.Equal(t, 200, response.Code, "OK response is expected")

	// assert files don't exist anymore
	assert.Equal(t, false, checkExisting("Cat03.jpg"))
	assert.Equal(t, false, checkExisting("Birds.jpg"))
}

// given file that is not .jpg, .jpeg, etc., program should respond
func TestInvalidFileType(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initSession()

	if err != nil {
		panic(err)
	}

	uploadTest := &UploadSingle{
		Filename: "filename",
		Filetype: "invalid",
		Path:     "./uploadtest.jpeg",
		Author:   "invalidfile_user",
	}

	jsonBody, _ := json.Marshal(uploadTest)

	request, _ := http.NewRequest("POST", "/api/upload", bytes.NewBuffer(jsonBody))

	response := httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back
	assert.Equal(t, 406, response.Code, "not accpetable response is expected")

	// assert we have not created file
	assert.Equal(t, false, checkExisting("filename.invalid"))

}

// test if a user can delete anothers' stuff
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

// test if a user can overwrite anothers' things
func TestUserOverwrite(t *testing.T) {

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
		Author:   "original_user",
	}

	jsonBody, _ := json.Marshal(uploadTest)

	request, _ := http.NewRequest("POST", "/api/upload", bytes.NewBuffer(jsonBody))

	response := httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back
	assert.Equal(t, 200, response.Code, "OK response is expected")

	// assert we have created file
	assert.Equal(t, true, checkExisting(filename.String()+"."+uploadTest.Filetype))

	// edit username, emulating a different user
	uploadTest = &UploadSingle{
		Filename: filename.String(),
		Filetype: "jpeg",
		Path:     "./uploadtest.jpeg",
		Author:   "overwriting_user",
	}

	jsonBody, _ = json.Marshal(uploadTest)

	request, _ = http.NewRequest("POST", "/api/upload", bytes.NewBuffer(jsonBody))

	response = httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 403 response back
	assert.Equal(t, 403, response.Code, "Forbidden response is expected")

	// assert original user still has ownership
	assert.Equal(t, true, checkAuthor(filename.String()+"."+uploadTest.Filetype, "original_user"), "Expect same user")

}

// test we successfully handle trying to delete something that doesn't exist
func TestDeleteNotExist(t *testing.T) {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initSession()

	var filesToDelete []string
	filesToDelete = append(filesToDelete, "thisfiledoesnotexist.jpeg")

	deleteTest := &DeletePictures{
		Filenames: filesToDelete,
		Author:    "testuser",
	}

	jsonBody, _ := json.Marshal(deleteTest)

	request, _ := http.NewRequest("DELETE", "/api/delete", bytes.NewBuffer(jsonBody))

	response := httptest.NewRecorder()
	createRouter().ServeHTTP(response, request)

	// assert get 200 response back, attempting to delete something that exists fails otherwise
	assert.Equal(t, 200, response.Code, "OK response is expected")

}
