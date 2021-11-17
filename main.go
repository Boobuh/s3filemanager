package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	AccessKeyID     string
	SecretAccessKey string
	MyRegion        string
	MyBucket        string
	filepath        string
)

//GetEnvWithKey : get env value
func GetEnvWithKey(key string) string {
	return os.Getenv(key)
}
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
		os.Exit(1)
	}
}
func main() {
	LoadEnv()
	// awsAccessKeyID := GetEnvWithKey("AWS_ACCESS_KEY_ID")
	// fmt.Println("My access key ID is ", awsAccessKeyID)

	sess := ConnectAws()
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		c.Set("sess", sess)
		c.Next()
	})

	router.POST("/upload", uploadFile)
	router.GET("/download", downloadFile)

	err := router.Run(":4040")
	if err != nil {
		fmt.Printf("failed to listen at port 4040: %v", err)
	}

}

func ConnectAws() *session.Session {
	AccessKeyID = GetEnvWithKey("AWS_ACCESS_KEY_ID")
	SecretAccessKey = GetEnvWithKey("AWS_SECRET_ACCESS_KEY")
	MyRegion = GetEnvWithKey("AWS_REGION")
	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(MyRegion),
			Credentials: credentials.NewStaticCredentials(
				AccessKeyID,
				SecretAccessKey,
				"", // a token will be created when the session it's used.
			),
		})
	if err != nil {
		panic(err)
	}
	return sess
}

func uploadFile(c *gin.Context) {

	sess := c.MustGet("sess").(*session.Session)
	uploader := s3manager.NewUploader(sess)
	MyBucket = GetEnvWithKey("BUCKET_NAME")
	file, header, _ := c.Request.FormFile("file")
	filename := header.Filename

	fi, err := os.Stat("/home/user/Стільниця/Desktop/tests")
	if err != nil {
		fmt.Printf("can't get filesize: %v", err)
		return
	}
	// get the size

	size := float64(fi.Size())
	size = (size / float64(1024)) / float64(1024)
	fmt.Printf("filesize: %v mb", size)

	if size > float64(5) {
		fmt.Printf("filesize is larger than 5 mb: %v mb\n", size)
		return
	}

	//upload to the s3 bucket
	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(MyBucket),
		ACL:    aws.String("public-read"),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Failed to upload file",
			"uploader": up,
		})
		fmt.Printf("failed to upload to bucket: %v", err)
		return
	}
	filepath = "https://" + MyBucket + "." + "s3-" + MyRegion + ".amazonaws.com/" + filename
	c.JSON(http.StatusOK, gin.H{
		"filepath": filepath,
	})
}

func downloadFile(c *gin.Context) {

	item := "tests"

	sess := c.MustGet("sess").(*session.Session)
	downloader := s3manager.NewDownloader(sess)
	MyBucket = GetEnvWithKey("BUCKET_NAME")

	file, err := os.Create(item)
	if err != nil {
		fmt.Println(err)
	}

	numBytes, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(MyBucket),
		Key:    aws.String(item),
	})

	if err != nil {
		fmt.Println(err)
	}

	b, err := ioutil.ReadFile(file.Name())

	if err != nil {
		fmt.Println("can't read file", err)
	}

	fmt.Println("Downloaded", file.Name(), numBytes, "bytes")

	c.Data(200, "multipart/form-data", b)

}
