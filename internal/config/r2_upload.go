package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"time"
)

func UploadTeacherQuiz(courseID, contentID string, jsonBytes []byte) error {
	// Upload the teacher version (unchanged JSON)
	teacherKey := fmt.Sprintf("focused/%s/quiz/teacher/%s.json", courseID, contentID)
	if err := uploadJSONToR2(teacherKey, jsonBytes); err != nil {
		return err
	}

	// Parse the JSON to remove correctAnswer and correctAnswers for the student version
	var quiz map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &quiz); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	// Get the questions array
	questions, ok := quiz["questions"].([]interface{})
	if !ok {
		return fmt.Errorf("questions field is not an array")
	}

	// Remove correctAnswer and correctAnswers from each question
	for _, q := range questions {
		question, ok := q.(map[string]interface{})
		if !ok {
			return fmt.Errorf("question is not a map")
		}
		delete(question, "correctAnswer")
		delete(question, "correctAnswers")
	}

	// Marshal the modified JSON back to bytes
	studentJSON, err := json.Marshal(quiz)
	if err != nil {
		return fmt.Errorf("error marshaling student JSON: %w", err)
	}

	// Upload the student version (without correctAnswer/correctAnswers)
	studentKey := fmt.Sprintf("focused/%s/quiz/student/%s.json", courseID, contentID)
	return uploadJSONToR2(studentKey, studentJSON)
}

func UploadTeacherText(courseID, contentID string, jsonBytes []byte) error {
	key := fmt.Sprintf("focused/%s/text/teacher/%s.json", courseID, contentID)
	return uploadJSONToR2(key, jsonBytes)
}

func UploadStudentQuiz(courseID, contentID string, jsonBytes []byte) error {
	key := fmt.Sprintf("focused/%s/quiz/student/%s.json", courseID, contentID)
	return uploadJSONToR2(key, jsonBytes)
}

func GeneratePresignedURL(bucket, key string) (string, error) {
	// Creamos un cliente presignado con el cliente S3
	presignClient := s3.NewPresignClient(R2Client)

	// Definimos la expiración de la URL firmada
	expiration := 15 * time.Minute // Duración en que la URL será válida

	// Creamos la solicitud para obtener el archivo con la URL firmada
	req, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(p *s3.PresignOptions) {
		p.Expires = expiration
	})

	if err != nil {
		return "", fmt.Errorf("no se pudo generar la URL firmada: %w", err)
	}

	// Devolvemos la URL firmada
	return req.URL, nil
}

func uploadJSONToR2(key string, jsonBytes []byte) error {
	if R2Client == nil {
		return fmt.Errorf("R2 no inicializado")
	}

	_, err := R2Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String("zeppelin"),
		Key:         aws.String(key),
		Body:        bytes.NewReader(jsonBytes),
		ContentType: aws.String("application/json"), // Esto es importante
		ACL:         "private",
	})

	return err
}
