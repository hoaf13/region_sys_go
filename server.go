package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

type data struct {
	id       string
	filename string
}

func serialize(msg data) ([]byte, error) {
	var b bytes.Buffer
	encoder := json.NewEncoder(&b)
	err := encoder.Encode(msg)
	return b.Bytes(), err
}

func main() {
	var ctx = context.Background()

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Server is running ...")
	})

	router.GET("apis/v1/model", func(c *gin.Context) {
		c.String(http.StatusOK, "Model is running ...")
	})

	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.POST("apis/v1/model", func(c *gin.Context) {

		uploaded_file, _ := c.FormFile("uploaded_file")
		c.SaveUploadedFile(uploaded_file, "upload/"+uploaded_file.Filename)
		uploaded_time := time.Now()
		body := uploaded_time.String() + "|" + uploaded_file.Filename

		// RabbitMQ Connection
		conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
		failOnError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		ch, err := conn.Channel()
		failOnError(err, "Failed to open a channel")
		defer ch.Close()

		q, err := ch.QueueDeclare(
			"task_queue", // name
			true,         // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
		)
		failOnError(err, "Failed to declare a queue")
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")

		// Redis Connection
		rdb := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		fmt.Print(uploaded_time.String())
		err = rdb.Set(ctx, uploaded_time.String(), "False", 0).Err()
		if err != nil {
			panic(err)
		}

		// Get answer from redis
		for {
			val, err := rdb.Get(ctx, uploaded_time.String()).Result()
			if err != nil {
				panic(err)
			}
			if val != "False" {
				fmt.Print("Label: ", val)
				break
			}
		}
	})
	router.Run(":8000")
}
