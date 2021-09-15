
# # ShopifyBackendChallengeW22

This is my submission to the Shopify Backend Challenge for Winter 2022.

Utilizing AWS S3, I have built an API in Golang, using Gorilla/Mux.  The current functionality is uploading by single image or by bulk, zipped images, search by user uploaded, and also deletion. 

Each image has an associated author, which prevents other authors from deleting or overwriting images. This is stored in the metadata of each file on S3. 

 - [TESTING SECTION](#testing-and-examples)

## Getting Started
Requires Go 1.16 or greater.

Three easy steps: 

    git clone https://github.com/Max-MZ/ShopifyBackendChallengeW22.git
    cd ShopifyBackendChallengeW22
    go mod download
    go run main.go
### .env
In order to successfully connect to an S3 bucket, the environment variables must be set up in a .env

    AWS_REGION = <<AWS_REGION>> e.g us-east-1
    AWS_ACCESS_KEY = <<AWS_ACCESS_KEY>>
    AWS_SECRET_ACCESS_KEY = <<AWS_SECRET_KEY>>
    BUCKET = <<BUCKET_NAME>> e.g shopifychallengew22


# REST API Usage
 - Upload `POST /api/upload` 
  - Upload Bulk (zip) `POST /api/zipupload` 
 - Deletion `DELETE /api/delete` 
 - Search `GET /api/search/{author}` 
## Uploading Body
    {
    "filename": <FILENAME>,
    "filetype": <FILE_TYPE>,
    "path": <PATH_TO_FILE>,
    "author": <AUTHOR_NAME>"
    }

## Uploading Bulk Body

    {
    "filename": <FILENAME>,
    "path": <PATH_TO_FILE>,
    "author": <AUTHOR_NAME>"
    }
## Deletion Body

    {
    "filenames": <ARRAY_OF_FILENAMES>,
    "author": <AUTHOR_NAME>"
    }

## Testing and Examples

The code has been tested extensively thanks to Golang's very effective testing libraries. Currently about **70%** of code is tested by the tests written in `api_test.go`. The remaining 30 mostly consists of what is left in the `main()` function, which consists mainly of setup. 

![Coverage](https://i.imgur.com/8NlUvDD.png)

Running the tests is very simple: 

To run all tests: `go test`

To run a certain test: `go test -run {TestName}`

For example, `go test -run TestSearch` will net us:
![Coverage](https://i.imgur.com/BOLH7mD.png)

### There Are Seven Tests

#### TestUploadAndDeleteBasic
Basic test that takes a file, uploads it, and asserts that it has been created in S3.
Upon that assertion, it deletes and asserts it has been deleted.

####  TestZipUpload
Takes in a .zip and attempts to upload every file inside. Asserts files have been uploaded to S3 and deletes them afterwards.

####  TestInvalidFileType
Tests API handles invalid files and asserts that a proper response was received

####  TestUserPermission
Attempts to delete a file that does not belong to the author stated in the POST request. Expects the API to respond properly and asserts the file was not deleted and that the author was not changed.

####  TestUserOverwrite
Attempts to upload a file that another user has already uploaded/named. Asserts the author has not changed after attempting the upload. 

####  TestDeleteNotExist
Calls for deletion of a file that doesn't exist. Expects API to handle accordingly and respond. Typically when trying to delete with a key/filename that doesn't exist, socket hangs up. Asserts that socket is still open with a proper response. 
####  TestSearch
Attempts to search for files based on a username. In tests case, expects two files back from a search where  `author=User1`. Asserts that there are the two expected files.
