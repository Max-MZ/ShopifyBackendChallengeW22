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

var accesskey string      // access key
var secretkey string      // secret key
var awsRegion string      // region
var bucket string         // bucketname
var sess *session.Session // session created for s3 connection

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

	// unzip and iterate through
	for _, f := range archive.File {
		// filePath := filepath.Join(dst, f.Name)
		log.Printf("unzipping file " + f.Name)

		fileType := filepath.Ext(f.Name)

		if fileType != ".jpg" && fileType != ".jpeg" && fileType != ".png" {
			log.Printf("unrecognized file, skipping ")
			continue
		}

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

		fin, err := uploader.Upload(&s3manager.UploadInput{
			Bucket:   aws.String(bucket),
			ACL:      aws.String("public-read"),
			Key:      aws.String(file.Name()), // picture is prefixed by author name
			Body:     file,
			Metadata: meta,
		})

		log.Printf(fin.UploadID)

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
	uploader := s3manager.NewUploader(sess)

	log.Printf("uploaded.path: " + uploaded.Path)

	file, err := os.Open(uploaded.Path)

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("Unable to open file %q", err)
	}

	defer file.Close()

	meta := make(map[string]*string)

	meta["author"] = aws.String(uploaded.Author)

	fin, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:   aws.String(bucket),
		ACL:      aws.String("public-read"),
		Key:      aws.String(uploaded.Filename + "." + uploaded.Filetype), // picture is prefixed by author name
		Body:     file,
		Metadata: meta,
	})

	log.Printf(fin.UploadID)

	w.Write([]byte("Picture successfully uploaded!\n"))

}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/upload", uploadPicture).Methods("POST")
	router.HandleFunc("/api/zipupload", bulkUpload).Methods("POST")

	accesskey = os.Getenv("AWS_ACCESS_KEY")
	secretkey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsRegion = os.Getenv("AWS_REGION")
	bucket = os.Getenv("BUCKET")

	sess, err = session.NewSession(
		&aws.Config{
			Region: aws.String(awsRegion),
			Credentials: credentials.NewStaticCredentials(
				accesskey,
				secretkey,
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		panic(err)
	}

	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		log.Printf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets:")

	for _, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n",
			aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
	}

	// resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(bucket)})
	// if err != nil {
	// 	log.Printf("Unable to list items in bucket %q, %v", bucket, err)
	// }

	// for _, item := range resp.Contents {
	// 	fmt.Println("Name:         ", *item.Key)
	// 	fmt.Println("Last modified:", *item.LastModified)
	// 	fmt.Println("Size:         ", *item.Size)
	// 	fmt.Println("Storage class:", *item.StorageClass)
	// 	fmt.Println("")
	// }

	log.Fatal(http.ListenAndServe(":8080", router))

}
