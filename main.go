package main

import (
	"net/http"
	"flag"
	"text/template"
	"log"
	"os"
	"time"
	"context"
  "github.com/aws/aws-sdk-go-v2/aws"
  "github.com/aws/aws-sdk-go-v2/config"
  "github.com/aws/aws-sdk-go-v2/service/s3"
  v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
  "crypto/rand"
  "fmt"
  "encoding/json"

)

var (
	addr     = flag.String("addr", ":8080", "http service address")
	indexTemplate = &template.Template{}
	cfg, err = config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	bucketName = "wormhole-uploads"
)

type Message struct {
	Type string `json:"type"`
	Body string `json:"body"`
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

func main() {
	// Parse the flags passed to program
	flag.Parse()
	// Read index.html from disk into memory, serve whenever anyone requests /
	indexHTML, err := os.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	indexTemplate = template.Must(template.New("indexTemplate").Parse(string(indexHTML)))

	// start HTTP server
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/upload", uploadHandler)
	log.Println("Serving requests at port", *addr)
	// log.Fatal(http.ListenAndServeTLS(*addr, "cert.pem", "key.pem", nil))
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Index route hit")

	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("File upload route hit")

	// Parse multipart form and extract file
  r.ParseMultipartForm(10 << 20)
  file, fileHeader, err := r.FormFile("file")
  if err != nil {
      log.Println("Error Retrieving the File")
      log.Println(err)
      return
  }
  defer file.Close()

  log.Printf("Uploaded File: %+v\n", fileHeader.Filename)
  log.Printf("Uploaded File: %+v\n", fileHeader.Header)
  log.Printf("File Size: %+v\n", fileHeader.Size)

  randomPrefix, err := generateUUIDV4()

  // Upload the file to S3.
	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)
	presigner := &Presigner{PresignClient: presignClient}

	log.Printf("Let's presign a request to upload a file to your bucket.")
	uploadKey := randomPrefix + fileHeader.Filename

	presignedPutRequest, err := presigner.PutObject(bucketName, uploadKey, 60)
	if err != nil {
		panic(err)
	}
	log.Printf("Got a presigned request", presignedPutRequest.URL)

	putRequest, err := http.NewRequest("PUT", presignedPutRequest.URL, file)
	if err != nil {
		log.Println(err)
	}
	putRequest.ContentLength = fileHeader.Size
	resp, err := http.DefaultClient.Do(putRequest)
	if err != nil {
		log.Println(err)
	}
	if resp.StatusCode !=200 {
		log.Println("Error %d", resp.StatusCode)
	}

	log.Printf("Let's presign a request to download the object.")
	presignedGetRequest, err := presigner.GetObject(bucketName, uploadKey, 3000)
	if err != nil {
		panic(err)
	}
	log.Printf("Got a presigned", presignedGetRequest.URL)
	log.Println("Using net/http to send the request...")

  message := Message{
		Type: "url",
		Body: presignedGetRequest.URL,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}
	
	w.Write(jsonData)
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
