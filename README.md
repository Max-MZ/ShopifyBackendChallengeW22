# # ShopifyBackendChallengeW22

This is my submission to the Shopify Backend Challenge for Winter 2022.

Utilizing AWS S3, I have built an API in Golang.  The current functionality is uploading by single image or by bulk, zipped image, and also deletion. 

Each image has an associated author, which prevents other authors from deleting or overwriting images. This is stored in the metadata of each file on S3. 

## Getting Started
Requires Go 1.16 or greater.

Three easy steps: 

    git clone https://github.com/Max-MZ/ShopifyBackendChallengeW22.git
    cd ShopifyBackendChallengeW22
    go mod download
    go run main.go

# REST API Usage
 - Upload `POST /api/upload` 
  - Upload Bulk (zip) `POST /api/zipupload` 
 - Deletion `DELETE /api/delete` 
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
## Deletion

    {
    "filenames": <ARRAY_OF_FILENAMES>,
    "author": <AUTHOR_NAME>"
    }

## Testing and Examples

The code has been tested extensively thanks to Golang's very effective testing libraries. Currently about **70%** of code is tested by the tests written in `api_test.go`. 
