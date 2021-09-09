package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Upload struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Author   string `json:"author"`
}

var accesskey string      // access key
var secretkey string      // secret key
var awsRegion string      // region
var bucket string         // bucketname
var sess *session.Session // session created for s3 connection

func uploadPicture(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Methods", "POST")

	decoder := json.NewDecoder(r.Body) // read body of request
	var uploaded Upload

	log.Printf("Unable to decode")
	err := decoder.Decode(&uploaded)

	log.Printf(r.RequestURI)
	// data, err := ioutil.ReadAll(r.Body)

	if err != nil {
		log.Printf("Unable to decode")
	}
	uploader := s3manager.NewUploader(sess)

	log.Printf("uploaded.path: " + uploaded.Path)

	file, err := os.Open("./Birds.jpg")

	if err != nil {
		log.Printf("Unable to open file %q", err)
	}

	defer file.Close()

	fin, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("public-read"),
		Key:    aws.String(uploaded.Filename),
		Body:   file,
	})

	log.Printf(fin.UploadID)

}

func main() {
	fmt.Println("Hello, world.")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	router := mux.NewRouter()

	router.HandleFunc("/api/upload", uploadPicture).Methods("POST")

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

	log.Fatal(http.ListenAndServe(":8080", router))

}
