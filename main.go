package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	addr          = flag.String("addr", ":8080", "http service address") // default port 8080
	indexTemplate = &template.Template{} // index page template
	cfg, err      = config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1")) // AWS config object for S3 client
	bucketName    = "wormhole-uploads" // AWS bucket name
)

// Message struct to send json responses to client
type Message struct {
	Type    string `json:"type"`
	URL     string `json:"url"`
	Name    string `json:"name"`
	Size    string  `json:"size"`
	Expires string `json:"expires"`
}

// Presigner encapsulates the Amazon Simple Storage Service (Amazon S3) presign actions
// used in the examples.
// It contains PresignClient, a client that is used to presign requests to Amazon S3.
// Presigned requests contain temporary credentials and can be made from any HTTP client.
type Presigner struct {
	PresignClient *s3.PresignClient
}

// GetObject makes a presigned request that can be used to get an object from a bucket.
// The presigned request is valid for the specified number of seconds.
func (presigner Presigner) GetObject(
	bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error) {
	request, err := presigner.PresignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to get %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request, err
}

// PutObject makes a presigned request that can be used to put an object in a bucket.
// The presigned request is valid for the specified number of seconds.
func (presigner Presigner) PutObject(
	bucketName string, objectKey string, lifetimeSecs int64) (*v4.PresignedHTTPRequest, error) {
	request, err := presigner.PresignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(lifetimeSecs * int64(time.Second))
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to put %v:%v. Here's why: %v\n",
			bucketName, objectKey, err)
	}
	return request, err
}

// DeleteObject makes a presigned request that can be used to delete an object from a bucket.
func (presigner Presigner) DeleteObject(bucketName string, objectKey string) (*v4.PresignedHTTPRequest, error) {
	request, err := presigner.PresignClient.PresignDeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		log.Printf("Couldn't get a presigned request to delete object %v. Here's why: %v\n", objectKey, err)
	}
	return request, err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form and extract file
	r.ParseMultipartForm(10 << 20) // File must be less than 10MB
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		log.Println("Error Retrieving the File")
		return
	}
	defer file.Close()

	// Upload the file to S3
	randomPrefix, err := generateUUIDV4()
	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)
	presigner := &Presigner{PresignClient: presignClient}
	uploadKey := randomPrefix + fileHeader.Filename

	presignedPutRequest, err := presigner.PutObject(bucketName, uploadKey, 60)
	if err != nil {
		panic(err)
	}

	putRequest, err := http.NewRequest("PUT", presignedPutRequest.URL, file)
	if err != nil {
		panic(err)
	}
	putRequest.ContentLength = fileHeader.Size

	_, err = http.DefaultClient.Do(putRequest)
	if err != nil {
		panic(err)
	}

	// Get presigned download link and send to client
	presignedGetRequest, err := presigner.GetObject(bucketName, uploadKey, 3700)
	if err != nil {
		panic(err)
	}

	// Create a Message struct
	expiration := time.Now().Add(time.Hour)
	message := Message{
		Type:    "links",
		URL:     presignedGetRequest.URL,
		Name:    fileHeader.Filename,
		Size:    fmt.Sprintf("%.2f", float32(fileHeader.Size)/(1024*1024)),
		Expires: expiration.Format(time.Kitchen), 
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		return
	}

	w.Write(jsonData)

	// Delete object after an hour
	ticker := time.NewTicker(time.Hour)
	go func() {
		select {
		case <-ticker.C:
			presignedDelRequest, err := presigner.DeleteObject(bucketName, uploadKey)
			if err != nil {
				panic(err)
			}
			delRequest, err := http.NewRequest("DELETE", presignedDelRequest.URL, file)
			if err != nil {
				panic(err)
			}
			_, err = http.DefaultClient.Do(delRequest)
			if err != nil {
				panic(err)
			}
			ticker.Stop()
		}
	}()

}

func generateUUIDV4() (string, error) {
	// Initialise an array of 16 bytes (32 hex digits)
	uuid := make([]byte, 16)
	// rand.Read() fills the array with random bytes instead of 0s
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}

	// Set the version (4) and variant bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	// Format the UUID in the standard string representation
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func main() {
	// Parse the flags passed to program
	flag.Parse()
	// Read index.html from disk into memory, serve whenever anyone requests /
	indexHTML, err := os.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	indexTemplate = template.Must(template.New("indexTemplate").Parse(string(indexHTML)))
	fs := http.FileServer(http.Dir("assets"))

	// start HTTP server
	http.HandleFunc("/", indexHandler)
  http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Serving requests at port", *addr)
	// log.Fatal(http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil))
	log.Fatal(http.ListenAndServe(*addr, nil))
}




