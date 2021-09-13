package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type UploadSingle struct {
	Filename string `json:"filename"`
	Filetype string `json:"filetype"`
	Path     string `json:"path"`
	Author   string `json:"author"`
}

type UploadZip struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Author   string `json:"author"`
}

type DeletePictures struct {
	Filenames []string `json:"filenames"`
	Author    string   `json:"author"`
}

var accesskey string      // access key
var secretkey string      // secret key
var awsRegion string      // region
var bucket string         // bucketname
var sess *session.Session // session created for s3 connection
var svc *s3.S3            // service client

// check if author of file in repo is the same as person requesting
func checkAuthor(filename string, requester string) bool {
	headObj := s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	}

	metadata, _ := svc.HeadObject(&headObj)

	return *metadata.Metadata["Author"] == requester

}

//check if file exists
func checkExisting(filename string) bool {

	headObj := s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(filename),
	}

	_, err := svc.HeadObject(&headObj)

	// if metadata cannot be found for some reason, means it does not exist
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "NotFound":
				return false
			default:
				return false
			}
		}
		return false
	}

	// return true if metadata is successfully retrieved
	return true
}

func deletion(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE")

	decoder := json.NewDecoder(r.Body) // read body of request
	var toDelete DeletePictures

	err := decoder.Decode(&toDelete)
	if err != nil {
		log.Printf("Unable to decode")
	}

	// iterate through array of names to delete
	for _, filename := range toDelete.Filenames {
		// headObj := s3.HeadObjectInput{
		// 	Bucket: aws.String(bucket),
		// 	Key:    aws.String(filename),
		// }

		// metadata, err := svc.HeadObject(&headObj)

		// if err != nil {
		// 	log.Printf("Problem with file, probably doesn't exist")
		// 	continue

		// }
		// // log.Printf(*metadata.Metadata["author"])

		// if the file doesn't exist, skip
		if !checkExisting(filename) {
			continue
		}

		// if the person trying to delete does not have permissions, skip
		if !checkAuthor(filename, toDelete.Author) {
			log.Printf("Not correct! ")
			w.Write([]byte("Not the author!\n"))
			continue
		} else { // safely delete
			_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(filename)})
			if err != nil {
				log.Printf("Unable to delete object %q from bucket %q, %v", filename, bucket, err)
				continue
			}

			err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(filename),
			})

			fmt.Printf("Deleted %q\n", filename)
		}

	}
}

func bulkUpload(w http.ResponseWriter, r *http.Request) { // upload a zip containing files

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	decoder := json.NewDecoder(r.Body) // read body of request
	var uploaded UploadZip

	err := decoder.Decode(&uploaded)

	log.Printf(r.RequestURI)

	if err != nil {
		log.Printf("Unable to decode")
	}
	uploader := s3manager.NewUploader(sess)

	log.Printf("uploaded.path: " + uploaded.Path)

	archive, err := zip.OpenReader(uploaded.Path)

	if err != nil {
		log.Fatalf("Can't Unzip")
	}

	defer archive.Close()

	// unzip and iterate through each file
	for _, f := range archive.File {
		log.Printf("unzipping file " + f.Name)

		fileType := filepath.Ext(f.Name)

		// accept only some filetypes
		if fileType != ".jpg" && fileType != ".jpeg" && fileType != ".png" {
			log.Printf("unrecognized file, skipping ")
			continue
		}

		// fail on directory
		if f.FileInfo().IsDir() {
			fmt.Println("found directory")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		file, err := os.Open(f.Name)

		if err != nil {
			log.Printf("Unable to open file %q", err)
			w.WriteHeader(http.StatusForbidden)
			return
		}

		meta := make(map[string]*string)

		meta["author"] = aws.String(uploaded.Author)

		_, _ = uploader.Upload(&s3manager.UploadInput{
			Bucket:   aws.String(bucket),
			ACL:      aws.String("public-read"),
			Key:      aws.String(file.Name()), // picture is prefixed by author name
			Body:     file,
			Metadata: meta,
		})

		file.Close()
	}

	w.WriteHeader(http.StatusOK)

}

func uploadPicture(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	decoder := json.NewDecoder(r.Body) // read body of request
	var uploaded UploadSingle

	err := decoder.Decode(&uploaded)

	if err != nil {
		log.Printf("Unable to decode")
	}

	if checkExisting(uploaded.Filename + "." + uploaded.Filetype) {

	}
	uploader := s3manager.NewUploader(sess)

	log.Printf("uploaded.path: " + uploaded.Path)

	file, err := os.Open(uploaded.Path)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Unable to open file %q", err)
		w.Write([]byte("Cannot open file!\n"))
		file.Close()
		return
	}

	defer file.Close()

	meta := make(map[string]*string)

	meta["author"] = aws.String(uploaded.Author)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:   aws.String(bucket),
		ACL:      aws.String("public-read"),
		Key:      aws.String(uploaded.Filename + "." + uploaded.Filetype), // picture is prefixed by author name
		Body:     file,
		Metadata: meta,
	})

	if err != nil {
		log.Fatal(err)
	}

	w.Write([]byte("Picture successfully uploaded!\n"))

}

func initSession() {

	accesskey = os.Getenv("AWS_ACCESS_KEY")
	secretkey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion = os.Getenv("AWS_REGION")
	bucket = os.Getenv("BUCKET")

	fmt.Printf(awsRegion + "\n")

	sess, _ = session.NewSession(
		&aws.Config{
			Region: aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(
				accesskey,
				secretkey,
				"", // a token will be created when the session it's used.
			),
		})

	svc = s3.New(sess)
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	initSession()

	router := mux.NewRouter()

	router.HandleFunc("/api/upload", uploadPicture).Methods("POST")
	router.HandleFunc("/api/zipupload", bulkUpload).Methods("POST")
	router.HandleFunc("/api/delete", deletion).Methods("DELETE")

	result, err := svc.ListBuckets(nil)
	if err != nil {
		log.Printf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}

	log.Fatal(http.ListenAndServe(":8080", router))

}
