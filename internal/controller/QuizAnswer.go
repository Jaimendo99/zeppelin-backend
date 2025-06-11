package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

type QuizController struct {
	QuizRepo              domain.QuizRepository
	CourseContentRepo     domain.CourseContentRepo
	AssignmentRepo        domain.AssignmentRepo
	CourseRepo            domain.CourseRepo
	UserRepo              domain.UserRepo
	UploadStudentAnswers  func(key string, data []byte) error
	GetTeacherQuizContent func(bucket, key string) ([]byte, error)
	GeneratePresignedURL  func(bucket, key string) (string, error)
}

func (c *QuizController) SubmitQuiz() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		var input domain.StudentQuizAnswersInput // Usar el struct de domain
		if err := ValidateAndBind(e, &input); err != nil {
			return err
		}

		Url, err := c.CourseContentRepo.GetUrlByContentID(input.ContentID) // Necesitarás implementar GetContentByContentID
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("contenido con ID %s no encontrado", input.ContentID)), nil)
		}
		//Imprimer el URL
		parts := strings.SplitN(Url, "/focused/", 2)
		if len(parts) < 2 {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "URL malformada"), nil)
		}
		subparts := strings.Split(parts[1], "/")
		courseIDInt, err := strconv.Atoi(subparts[0])

		// 1. Verificar que el estudiante está asignado a este curso
		_, err = c.AssignmentRepo.GetAssignmentsByStudentAndCourse(userID, courseIDInt)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusForbidden, "Este estudiante no está asignado a este curso"), nil)
		}

		// 2. Verificar que el ContentID es realmente un quiz
		contentTypeID, err := c.CourseContentRepo.GetContentTypeID(input.ContentID)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al obtener tipo de contenido: %v", err)), nil)
		}
		if contentTypeID != 3 { // Asumiendo que 3 es el ID para Quiz
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "el content_id no corresponde a un quiz"), nil)
		}

		// 3. Serializar las respuestas del estudiante a JSON bytes
		studentAnswersJSONBytes, err := json.Marshal(input.Answers)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al serializar respuestas del estudiante: %s", err.Error())), nil)
		}

		// 4. Subir el JSON de respuestas del estudiante a R2
		// Ruta: focused/el accountID/quiz/answer/id de student/id de contentid.json
		// Ajustamos la key para seguir tu convención de ruta
		accountID := os.Getenv("R2_ACCOUNT_ID") // Obtener el AccountID
		studentAnswersKey := fmt.Sprintf("focused/%s/quiz/answer/%s/%s.json", accountID, userID, input.ContentID)
		err = c.UploadStudentAnswers(studentAnswersKey, studentAnswersJSONBytes) // Usa la función exportada
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al subir respuestas del estudiante a R2: %s", err.Error())), nil)
		}

		// Genera la URL del archivo subido
		studentAnswersURL := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", accountID, studentAnswersKey) // URL directa

		// Obtener el contenido del archivo del quiz del profesor desde R2
		teacherQuizBytes, err := c.GetTeacherQuizContent("zeppelin", strings.Replace(Url, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1)) // Eliminar el prefijo de la URL para obtener la key
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al obtener el quiz del profesor desde R2: %s", err.Error())), nil)
		}

		var teacherQuiz domain.TeacherQuiz // Usar el struct definido en domain
		if err := json.Unmarshal(teacherQuizBytes, &teacherQuiz); err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al parsear el quiz del profesor: %s", err.Error())), nil)
		}

		// 6. Calificar el quiz
		score, totalPoints := c.GradeQuiz(teacherQuiz, input.Answers)

		// 7. Determinar el estado de revisión (`reviewed_at`)
		needsReview := false
		for _, question := range teacherQuiz.Questions {
			if question.Type == "text" {
				needsReview = true
				break
			}
		}

		var reviewedAt *time.Time
		if !needsReview {
			now := time.Now()
			reviewedAt = &now
		}

		// 8. Crear el registro del intento de quiz en la base de datos
		quizAttempt := domain.QuizAnswer{ // Usar el struct definido en domain
			ContentID:     input.ContentID,
			UserID:        userID,
			StartTime:     input.StartTime,
			EndTime:       input.EndTime,
			Grade:         &score, // Usamos el puntero para el grade
			ReviewedAt:    reviewedAt,
			QuizURL:       Url,               // URL del quiz del profesor
			QuizAnswerURL: studentAnswersURL, // URL del archivo de respuestas del estudiante en R2
			TotalPoints:   &totalPoints,      // Usamos el puntero para total_points
		}

		// 9. Guardar el intento de quiz en la base de datos
		err = c.QuizRepo.SaveQuizAttempt(quizAttempt) // Llama al método del QuizRepository
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al guardar el intento del quiz: %s", err.Error())), nil)
		}

		// 11. Devolver la puntuación al estudiante
		return ReturnWriteResponse(e, nil, map[string]interface{}{
			"message":             "Quiz calificado exitosamente",
			"score":               score,
			"total_points":        totalPoints,
			"quiz_answer_id":      quizAttempt.QuizAnswerID, // Devolver el ID del intento guardado
			"student_answers_url": studentAnswersURL,        // Devolver la URL de las respuestas guardadas
			"reviewed_at":         reviewedAt,               // Incluir el timestamp de revisión
			"quizTeacherResponse": teacherQuiz,              // Devolver la URL del quiz del profesor
		})
	}
}

func (c *QuizController) ReviewTextAnswer() echo.HandlerFunc {
	return func(e echo.Context) error {
		var input domain.TextAnswerReviewInput
		if err := ValidateAndBind(e, &input); err != nil {
			log.Printf("Error binding input: %v", err)
			return err
		}

		log.Printf("Received input: %+v", input)

		// 1. Obtener el intento de quiz desde la base de datos
		quizAttempt, err := c.QuizRepo.FindQuizAttemptByID(input.QuizAnswerID)
		if err != nil {
			log.Printf("Error finding quiz attempt ID %d: %v", input.QuizAnswerID, err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("intento de quiz con ID %d no encontrado", input.QuizAnswerID)), nil)
		}

		// 2. Obtener el JSON de respuestas del estudiante desde el bucket
		accountID := os.Getenv("R2_ACCOUNT_ID")
		studentAnswersKey := strings.Replace(quizAttempt.QuizAnswerURL, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1)
		studentAnswersBytes, err := c.GetTeacherQuizContent("zeppelin", studentAnswersKey)
		if err != nil {
			log.Printf("Error fetching student answers from R2, key %s: %v", studentAnswersKey, err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al obtener respuestas del estudiante desde R2: %s", err.Error())), nil)
		}

		log.Printf("Raw student answers JSON: %s", string(studentAnswersBytes))

		// 3. Unmarshal the JSON into a map
		var studentAnswers map[string]interface{}
		if err := json.Unmarshal(studentAnswersBytes, &studentAnswers); err != nil {
			log.Printf("Error unmarshaling student answers: %v", err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al parsear respuestas del estudiante: %s", err.Error())), nil)
		}

		if studentAnswers == nil {
			studentAnswers = make(map[string]interface{})
			log.Printf("Initialized empty student answers map")
		}

		log.Printf("Parsed student answers keys: %v", reflect.ValueOf(studentAnswers).MapKeys())

		// 4. Validar que la pregunta existe y es de tipo texto
		teacherQuizBytes, err := c.GetTeacherQuizContent("zeppelin", strings.Replace(quizAttempt.QuizURL, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1))
		if err != nil {
			log.Printf("Error fetching teacher quiz from R2: %v", err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al obtener el quiz del profesor desde R2: %s", err.Error())), nil)
		}

		var teacherQuiz domain.TeacherQuiz
		if err := json.Unmarshal(teacherQuizBytes, &teacherQuiz); err != nil {
			log.Printf("Error unmarshaling teacher quiz: %v", err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al parsear el quiz del profesor: %s", err.Error())), nil)
		}

		var questionPoints int
		var questionFound bool
		for _, question := range teacherQuiz.Questions {
			if question.ID == input.QuestionID {
				if question.Type != "text" {
					log.Printf("Question %s is not of type text, found type: %s", input.QuestionID, question.Type)
					return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "la pregunta no es de tipo texto"), nil)
				}
				questionPoints = question.Points
				questionFound = true
				break
			}
		}
		if !questionFound {
			log.Printf("Question ID %s not found in teacher quiz", input.QuestionID)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("pregunta con ID %s no encontrada", input.QuestionID)), nil)
		}

		// 5. Validar que los puntos asignados no excedan los puntos máximos
		if input.PointsAwarded > float64(questionPoints) {
			log.Printf("Points awarded %f exceed max points %d for question %s", input.PointsAwarded, questionPoints, input.QuestionID)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("los puntos asignados (%f) exceden los puntos máximos (%d)", input.PointsAwarded, questionPoints)), nil)
		}

		// 6. Initialize or append to extra_review array
		var extraReview []map[string]interface{}
		if existing, ok := studentAnswers["extra_review"]; ok {
			if arr, isArray := existing.([]interface{}); isArray {
				for _, item := range arr {
					if m, isMap := item.(map[string]interface{}); isMap {
						extraReview = append(extraReview, m)
					}
				}
			}
		}

		// 7. Append new review
		answerValue := studentAnswers[input.QuestionID]
		if answerValue == nil {
			log.Printf("No answer provided for question ID %s, using empty string", input.QuestionID)
			answerValue = ""
		}

		reviewEntry := map[string]interface{}{
			input.QuestionID: map[string]interface{}{
				"value":          answerValue,
				"is_correct":     input.IsCorrect,
				"points_awarded": input.PointsAwarded,
			},
		}
		extraReview = append(extraReview, reviewEntry)
		studentAnswers["extra_review"] = extraReview

		// 8. Serializar y subir el JSON actualizado al bucket
		updatedAnswersJSONBytes, err := json.Marshal(studentAnswers)
		if err != nil {
			log.Printf("Error serializing updated answers: %v", err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al serializar respuestas actualizadas: %s", err.Error())), nil)
		}
		err = c.UploadStudentAnswers(studentAnswersKey, updatedAnswersJSONBytes)
		if err != nil {
			log.Printf("Error uploading updated answers to R2: %v", err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al subir respuestas actualizadas a R2: %s", err.Error())), nil)
		}

		// 9. Calcular totalPoints sumando los puntos de todas las preguntas
		totalPoints := 0
		for _, question := range teacherQuiz.Questions {
			totalPoints += question.Points
		}
		log.Printf("Calculated total points: %d", totalPoints)

		// 10. Actualizar el score sumando points_awarded al grade actual
		currentScore := 0.0
		if quizAttempt.Grade != nil {
			currentScore = *quizAttempt.Grade
		}
		newScore := currentScore + input.PointsAwarded
		log.Printf("Current score: %f, Adding %f points, New score: %f", currentScore, input.PointsAwarded, newScore)

		// 11. Actualizar el quiz attempt en la base de datos
		quizAttempt.Grade = &newScore
		now := time.Now()
		quizAttempt.ReviewedAt = &now
		quizAttempt.TotalPoints = &totalPoints
		err = c.QuizRepo.UpdateQuizAttempt(quizAttempt)
		if err != nil {
			log.Printf("Error updating quiz attempt: %v", err)
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al actualizar el intento del quiz: %s", err.Error())), nil)
		}

		// 12. Devolver respuesta exitosa
		log.Printf("Successfully reviewed question %s for quiz attempt %d, New score: %f", input.QuestionID, quizAttempt.QuizAnswerID, newScore)
		return ReturnWriteResponse(e, nil, map[string]interface{}{
			"message":        "Respuesta de texto revisada exitosamente",
			"quiz_answer_id": quizAttempt.QuizAnswerID,
			"question_id":    input.QuestionID,
			"score":          newScore,
			"total_points":   totalPoints,
			"reviewed_at":    now,
		})
	}
}

func (c *QuizController) GradeQuiz(teacherQuiz domain.TeacherQuiz, studentAnswers map[string]interface{}) (float64, int) {
	earnedPoints := 0.0
	totalPoints := 0

	for _, question := range teacherQuiz.Questions {
		totalPoints += question.Points
		studentAnswer := studentAnswers[question.ID]

		switch question.Type {
		case "text":
			// Para preguntas de texto, usamos points_awarded si existe
			if answerMap, ok := studentAnswer.(map[string]interface{}); ok {
				if points, ok := answerMap["points_awarded"].(float64); ok {
					earnedPoints += points
				}
			}
		case "multiple":
			correctAnswer := question.CorrectAnswer.(string)
			studentAnswerStr, ok := studentAnswer.(string)
			if ok && studentAnswerStr == correctAnswer {
				earnedPoints += float64(question.Points)
			}
		case "checkbox":
			correctAnswers := question.CorrectAnswers
			studentSelectedInterface, ok := studentAnswer.([]interface{})
			if !ok {
				continue
			}

			studentSelectedStrings := make([]string, 0, len(studentSelectedInterface))
			validStudentAnswers := true
			for i, v := range studentSelectedInterface {
				str, typeOk := v.(string)
				if !typeOk {
					fmt.Printf("DEBUG CHECKBOX: Elemento %d de studentAnswer no es string. Es: %T\n", i, v)
					validStudentAnswers = false
					break
				}
				studentSelectedStrings = append(studentSelectedStrings, str)
			}

			if !validStudentAnswers {
				continue
			}

			if len(studentSelectedStrings) == len(correctAnswers) {
				correctMap := make(map[string]bool)
				for _, ans := range correctAnswers {
					correctMap[ans] = true
				}
				studentMap := make(map[string]bool)
				for _, ans := range studentSelectedStrings {
					studentMap[ans] = true
				}

				isCorrect := true
				if len(correctMap) != len(studentMap) {
					isCorrect = false
				} else {
					for key := range correctMap {
						if _, found := studentMap[key]; !found {
							isCorrect = false
							break
						}
					}
				}

				if isCorrect {
					earnedPoints += float64(question.Points)
				}
			}
		case "boolean":
			var expectedBoolValue bool
			isCorrectAnswerParsable := true

			if val, ok := question.CorrectAnswer.(bool); ok {
				expectedBoolValue = val
			} else if valStr, ok := question.CorrectAnswer.(string); ok {
				lowerValStr := strings.ToLower(valStr)
				if lowerValStr == "true" || lowerValStr == "verdadero" {
					expectedBoolValue = true
				} else if lowerValStr == "false" || lowerValStr == "falso" {
					expectedBoolValue = false
				} else {
					isCorrectAnswerParsable = false
				}
			} else {
				isCorrectAnswerParsable = false
			}

			if !isCorrectAnswerParsable {
				continue
			}

			studentAnswerBool, typeOk := studentAnswer.(bool)
			if !typeOk {
				if studentAnswerStr, okStr := studentAnswer.(string); okStr {
					lowerStudentStr := strings.ToLower(studentAnswerStr)
					if lowerStudentStr == "true" || lowerStudentStr == "verdadero" {
						studentAnswerBool = true
						typeOk = true
					} else if lowerStudentStr == "false" || lowerStudentStr == "falso" {
						studentAnswerBool = false
						typeOk = true
					}
				}
			}
			if typeOk && studentAnswerBool == expectedBoolValue {
				earnedPoints += float64(question.Points)
			}
		}
	}

	fmt.Printf("\n--- FIN CALIFICACIÓN ---\n")
	fmt.Printf("Puntos Totales Ganados: %f, Puntos Totales Posibles: %d\n", earnedPoints, totalPoints)
	return earnedPoints, totalPoints
}

// GetQuizzesByCourse reestructurado
func (c *QuizController) GetQuizzesByCourse() echo.HandlerFunc {
	return func(e echo.Context) error {

		// Obtener courseID
		courseIDStr := e.Param("courseId")
		courseID, err := strconv.Atoi(courseIDStr)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusBadRequest, "ID de curso inválido"), nil)
		}

		// Obtener R2_ACCOUNT_ID
		accountID := os.Getenv("R2_ACCOUNT_ID")
		if accountID == "" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "R2_ACCOUNT_ID no configurado"), nil)
		}

		// Consultar la vista
		attempts, err := c.QuizRepo.GetQuizAttemptsByCourse(courseID)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al obtener intentos de quiz: %v", err)), nil)
		}

		if len(attempts) == 0 {
			return ReturnWriteResponse(e, nil, domain.CourseQuizResponse{
				CourseID:     courseID,
				CourseTitle:  "",
				TotalQuizzes: 0, // Agregar campo
				Quizzes:      []domain.Quiz{},
			})
		}

		// Agrupar por quiz
		quizMap := make(map[string]*domain.Quiz)
		var totalQuizzes int // Variable para almacenar el total

		for _, attempt := range attempts {
			// Capturar el total de quizzes del primer registro (será el mismo para todos)
			if totalQuizzes == 0 {
				totalQuizzes = attempt.TotalQuizzes
			}

			if _, exists := quizMap[attempt.ContentID]; !exists {
				quizMap[attempt.ContentID] = &domain.Quiz{
					ContentID:       attempt.ContentID,
					QuizURL:         attempt.QuizURL,
					QuizTitle:       attempt.QuizTitle,
					QuizDescription: attempt.QuizDescription,
					CourseContentID: attempt.CourseContentID,
					Module:          attempt.Module,
					ModuleIndex:     attempt.ModuleIndex,
					Attempts:        []domain.QuizAttempt{},
				}
			}
			quizMap[attempt.ContentID].Attempts = append(quizMap[attempt.ContentID].Attempts, domain.QuizAttempt{
				QuizAnswerID:    attempt.QuizAnswerID,
				UserID:          attempt.UserID,
				StudentName:     attempt.StudentName,
				StudentLastname: attempt.StudentLastname,
				StudentEmail:    attempt.StudentEmail,
				Grade:           attempt.Grade,
				TotalPoints:     attempt.TotalPoints,
				NeedsReview:     attempt.NeedsReview,
				ReviewedAt:      attempt.ReviewedAt,
				QuizAnswerURL:   attempt.QuizAnswerURL,
				StartTime:       attempt.StartTime,
				EndTime:         attempt.EndTime,
			})
		}

		// Crear lista de quizzes
		quizzes := make([]domain.Quiz, 0, len(quizMap))
		for _, quiz := range quizMap {
			quizzes = append(quizzes, *quiz)
		}

		// Crear respuesta
		response := domain.CourseQuizResponse{
			CourseID:          courseID,
			CourseTitle:       attempts[0].CourseTitle,
			CourseDescription: attempts[0].CourseDescription,
			TotalQuizzes:      totalQuizzes, // Agregar el total
			Quizzes:           quizzes,
		}

		// ... resto del código para firmar URLs permanece igual
		// Mapear claves únicas para firmar
		quizKeys := make(map[string]string) // ContentID -> quizKey
		answerKeys := make(map[int]string)  // QuizAnswerID -> answerKey
		for _, quiz := range response.Quizzes {
			quizKeys[quiz.ContentID] = strings.Replace(quiz.QuizURL, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1)
			for _, attempt := range quiz.Attempts {
				answerKeys[attempt.QuizAnswerID] = strings.Replace(attempt.QuizAnswerURL, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1)
			}
		}

		// Firmar QuizURLs
		for contentID, quizKey := range quizKeys {
			signedURL, err := c.GeneratePresignedURL("zeppelin", quizKey)
			if err != nil {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para quiz %s: %v", contentID, err)), nil)
			}
			for i, quiz := range response.Quizzes {
				if quiz.ContentID == contentID {
					response.Quizzes[i].QuizURL = signedURL
				}
			}
		}

		// Firmar QuizAnswerURLs
		for quizAnswerID, answerKey := range answerKeys {
			signedURL, err := c.GeneratePresignedURL("zeppelin", answerKey)
			if err != nil {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para respuestas %d: %v", quizAnswerID, err)), nil)
			}
			for i, quiz := range response.Quizzes {
				for j, attempt := range quiz.Attempts {
					if attempt.QuizAnswerID == quizAnswerID {
						response.Quizzes[i].Attempts[j].QuizAnswerURL = signedURL
					}
				}
			}
		}

		return ReturnReadResponse(e, nil, response)
	}
}

// GetQuizzesByStudent (sin cambios)
func (c *QuizController) GetQuizzesByStudent() echo.HandlerFunc {
	return func(e echo.Context) error {
		userID := e.Get("user_id").(string)

		// Obtener R2_ACCOUNT_ID
		accountID := os.Getenv("R2_ACCOUNT_ID")
		if accountID == "" {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, "R2_ACCOUNT_ID no configurado"), nil)
		}

		// Consultar la vista
		attempts, err := c.QuizRepo.GetQuizAttemptsByStudent(userID)
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al obtener intentos de quiz: %v", err)), nil)
		}

		if len(attempts) == 0 {
			return ReturnWriteResponse(e, nil, domain.StudentCoursesQuizResponse{
				UserID:   userID,
				Name:     "",
				Lastname: "",
				Email:    "",
				Courses:  []domain.CourseQuizResponse{},
			})
		}

		// Agrupar por curso y quiz
		courseMap := make(map[int]map[string]*domain.Quiz)
		courseInfo := make(map[int]struct {
			CourseTitle       string
			CourseDescription string
			TotalQuizzes      int // Agregar campo
		})

		for _, attempt := range attempts {
			// Inicializar mapa de quizzes para el curso si no existe
			if _, exists := courseMap[attempt.CourseID]; !exists {
				courseMap[attempt.CourseID] = make(map[string]*domain.Quiz)
				courseInfo[attempt.CourseID] = struct {
					CourseTitle       string
					CourseDescription string
					TotalQuizzes      int
				}{
					CourseTitle:       attempt.CourseTitle,
					CourseDescription: attempt.CourseDescription,
					TotalQuizzes:      attempt.TotalQuizzes, // Capturar el total
				}
			}

			// Inicializar quiz si no existe
			if _, exists := courseMap[attempt.CourseID][attempt.ContentID]; !exists {
				courseMap[attempt.CourseID][attempt.ContentID] = &domain.Quiz{
					ContentID:       attempt.ContentID,
					QuizURL:         attempt.QuizURL,
					QuizTitle:       attempt.QuizTitle,
					QuizDescription: attempt.QuizDescription,
					CourseContentID: attempt.CourseContentID,
					Module:          attempt.Module,
					ModuleIndex:     attempt.ModuleIndex,
					Attempts:        []domain.QuizAttempt{},
				}
			}

			// Agregar intento
			courseMap[attempt.CourseID][attempt.ContentID].Attempts = append(
				courseMap[attempt.CourseID][attempt.ContentID].Attempts,
				domain.QuizAttempt{
					QuizAnswerID:    attempt.QuizAnswerID,
					UserID:          attempt.UserID,
					StudentName:     attempt.StudentName,
					StudentLastname: attempt.StudentLastname,
					StudentEmail:    attempt.StudentEmail,
					Grade:           attempt.Grade,
					TotalPoints:     attempt.TotalPoints,
					NeedsReview:     attempt.NeedsReview,
					ReviewedAt:      attempt.ReviewedAt,
					QuizAnswerURL:   attempt.QuizAnswerURL,
					StartTime:       attempt.StartTime,
					EndTime:         attempt.EndTime,
				},
			)
		}

		// Crear lista de cursos
		courses := make([]domain.CourseQuizResponse, 0, len(courseMap))
		for courseID, quizMap := range courseMap {
			quizzes := make([]domain.Quiz, 0, len(quizMap))
			for _, quiz := range quizMap {
				quizzes = append(quizzes, *quiz)
			}
			courses = append(courses, domain.CourseQuizResponse{
				CourseID:          courseID,
				CourseTitle:       courseInfo[courseID].CourseTitle,
				CourseDescription: courseInfo[courseID].CourseDescription,
				TotalQuizzes:      courseInfo[courseID].TotalQuizzes, // Agregar el total
				Quizzes:           quizzes,
			})
		}

		// Crear respuesta
		response := domain.StudentCoursesQuizResponse{
			UserID:   userID,
			Name:     attempts[0].StudentName,
			Lastname: attempts[0].StudentLastname,
			Email:    attempts[0].StudentEmail,
			Courses:  courses,
		}

		// ... resto del código para firmar URLs permanece igual
		// Mapear claves únicas para firmar
		quizKeys := make(map[string]string) // ContentID -> quizKey
		answerKeys := make(map[int]string)  // QuizAnswerID -> answerKey
		for _, course := range response.Courses {
			for _, quiz := range course.Quizzes {
				quizKeys[quiz.ContentID] = strings.Replace(quiz.QuizURL, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1)
				for _, attempt := range quiz.Attempts {
					answerKeys[attempt.QuizAnswerID] = strings.Replace(attempt.QuizAnswerURL, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1)
				}
			}
		}

		// Firmar QuizURLs
		for contentID, quizKey := range quizKeys {
			signedURL, err := c.GeneratePresignedURL("zeppelin", quizKey)
			if err != nil {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para quiz %s: %v", contentID, err)), nil)
			}
			for i, course := range response.Courses {
				for j, quiz := range course.Quizzes {
					if quiz.ContentID == contentID {
						response.Courses[i].Quizzes[j].QuizURL = signedURL
					}
				}
			}
		}

		// Firmar QuizAnswerURLs
		for quizAnswerID, answerKey := range answerKeys {
			signedURL, err := c.GeneratePresignedURL("zeppelin", answerKey)
			if err != nil {
				return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al generar URL firmada para respuestas %d: %v", quizAnswerID, err)), nil)
			}
			for i, course := range response.Courses {
				for j, quiz := range course.Quizzes {
					for k, attempt := range quiz.Attempts {
						if attempt.QuizAnswerID == quizAnswerID {
							response.Courses[i].Quizzes[j].Attempts[k].QuizAnswerURL = signedURL
						}
					}
				}
			}
		}

		return ReturnReadResponse(e, nil, response)
	}
}
