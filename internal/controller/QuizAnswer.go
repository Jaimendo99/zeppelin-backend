package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"zeppelin/internal/config"
	"zeppelin/internal/domain"

	"github.com/labstack/echo/v4"
)

// Struct para recibir las respuestas del estudiante - Movido a domain
// type StudentQuizAnswersInput struct {
// 	ContentID string                   `json:"content_id" validate:"required"`
// 	StartTime time.Time                `json:"start_time" validate:"required"`
// 	EndTime   time.Time                `json:"end_time" validate:"required"`
// 	Answers   map[string]interface{} `json:"answers" validate:"required"` // Mapa de pregunta_id a respuesta(s)
// }

// Controlador de Quizzes
type QuizController struct {
	QuizRepo          domain.QuizRepository
	CourseContentRepo domain.CourseContentRepo
	AssignmentRepo    domain.AssignmentRepo
}

// Variables for testing - these will be overriden in tests
var (
	ConfigUploadJSONToR2 = config.UploadJSONToR2
	ConfigGetR2Object    = config.GetR2Object
)

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
		courseID := strings.Split(Url, "/")[2]
		courseIDInt, err := strconv.Atoi(courseID)

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

		err = ConfigUploadJSONToR2(studentAnswersKey, studentAnswersJSONBytes) // Usa la función exportada
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al subir respuestas del estudiante a R2: %s", err.Error())), nil)
		}

		// Genera la URL del archivo subido
		studentAnswersURL := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s", accountID, studentAnswersKey) // URL directa

		// Obtener el contenido del archivo del quiz del profesor desde R2
		teacherQuizBytes, err := ConfigGetR2Object("zeppelin", strings.Replace(Url, fmt.Sprintf("https://%s.r2.cloudflarestorage.com/", accountID), "", 1)) // Eliminar el prefijo de la URL para obtener la key
		if err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al obtener el quiz del profesor desde R2: %s", err.Error())), nil)
		}

		var teacherQuiz domain.TeacherQuiz // Usar el struct definido en domain
		if err := json.Unmarshal(teacherQuizBytes, &teacherQuiz); err != nil {
			return ReturnWriteResponse(e, echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error al parsear el quiz del profesor: %s", err.Error())), nil)
		}

		// 6. Calificar el quiz
		score, totalPoints := c.gradeQuiz(teacherQuiz, input.Answers)

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

func (c *QuizController) gradeQuiz(teacherQuiz domain.TeacherQuiz, studentAnswers map[string]interface{}) (float64, int) {
	earnedPoints := 0.0
	totalPoints := 0

	fmt.Println("--- INICIO CALIFICACIÓN ---")
	fmt.Printf("Student Answers: %+v\n", studentAnswers)

	for _, question := range teacherQuiz.Questions {
		totalPoints += question.Points
		fmt.Printf("\n--- Pregunta ID: %s, Tipo: %s, Puntos: %d ---\n", question.ID, question.Type, question.Points)
		fmt.Printf("Question.CorrectAnswer: %+v (Tipo: %T)\n", question.CorrectAnswer, question.CorrectAnswer)
		if len(question.CorrectAnswers) > 0 {
			fmt.Printf("Question.CorrectAnswers: %+v\n", question.CorrectAnswers)
		}

		studentAnswer, ok := studentAnswers[question.ID]
		if !ok {
			fmt.Printf("DEBUG: Pregunta %s NO respondida por el estudiante.\n", question.ID)
			continue
		}
		if studentAnswer == nil {
			fmt.Printf("DEBUG: Pregunta %s tiene respuesta NULA del estudiante.\n", question.ID)
			continue
		}
		fmt.Printf("Student Answer para %s: %+v (Tipo: %T)\n", question.ID, studentAnswer, studentAnswer)

		switch question.Type {
		case "text":
			correctAnswer, typeOk := question.CorrectAnswer.(string)
			if !typeOk {
				fmt.Printf("DEBUG TEXT: question.CorrectAnswer no es string. Es: %T\n", question.CorrectAnswer)
				continue
			}
			studentAnswerStr, typeOk := studentAnswer.(string)
			if !typeOk {
				fmt.Printf("DEBUG TEXT: studentAnswer no es string. Es: %T\n", studentAnswer)
				continue
			}
			fmt.Printf("DEBUG TEXT: Comparando '%s' (estudiante) con '%s' (correcta) después de trim/lower.\n", strings.ToLower(strings.TrimSpace(studentAnswerStr)), strings.ToLower(strings.TrimSpace(correctAnswer)))
			if strings.ToLower(strings.TrimSpace(studentAnswerStr)) == strings.ToLower(strings.TrimSpace(correctAnswer)) {
				earnedPoints += float64(question.Points)
				fmt.Printf("DEBUG TEXT: ¡CORRECTO! Puntos ganados: %d\n", question.Points)
			} else {
				fmt.Printf("DEBUG TEXT: Incorrecto.\n")
			}

		case "multiple":
			correctAnswer, typeOk := question.CorrectAnswer.(string)
			if !typeOk {
				fmt.Printf("DEBUG MULTIPLE: question.CorrectAnswer no es string. Es: %T\n", question.CorrectAnswer)
				continue
			}
			studentAnswerStr, typeOk := studentAnswer.(string)
			if !typeOk {
				fmt.Printf("DEBUG MULTIPLE: studentAnswer no es string. Es: %T\n", studentAnswer)
				continue
			}
			fmt.Printf("DEBUG MULTIPLE: Comparando '%s' (estudiante) con '%s' (correcta).\n", studentAnswerStr, correctAnswer)
			if studentAnswerStr == correctAnswer {
				earnedPoints += float64(question.Points)
				fmt.Printf("DEBUG MULTIPLE: ¡CORRECTO! Puntos ganados: %d\n", question.Points)
			} else {
				fmt.Printf("DEBUG MULTIPLE: Incorrecto.\n")
			}

		case "checkbox":
			correctAnswers := question.CorrectAnswers
			if len(correctAnswers) == 0 {
				fmt.Printf("DEBUG CHECKBOX: No hay respuestas correctas definidas para la pregunta %s.\n", question.ID)
				continue
			}
			fmt.Printf("DEBUG CHECKBOX: CorrectAnswers: %+v\n", correctAnswers)

			studentSelectedInterface, typeOk := studentAnswer.([]interface{})
			if !typeOk {
				fmt.Printf("DEBUG CHECKBOX: studentAnswer no es []interface{}. Es: %T\n", studentAnswer)
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
			fmt.Printf("DEBUG CHECKBOX: StudentSelectedStrings: %+v\n", studentSelectedStrings)

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
				if len(correctMap) != len(studentMap) { // Debería ser redundante si el len de slices ya coincidió, pero doble check
					isCorrect = false
				} else {
					for key := range correctMap {
						if _, found := studentMap[key]; !found {
							isCorrect = false
							fmt.Printf("DEBUG CHECKBOX: Falta la respuesta correcta '%s' en las respuestas del estudiante.\n", key)
							break
						}
					}
					// Opcional: verificar que el estudiante no haya seleccionado opciones extra (ya cubierto por la comparación de len y el bucle anterior)
				}

				if isCorrect {
					earnedPoints += float64(question.Points)
					fmt.Printf("DEBUG CHECKBOX: ¡CORRECTO! Puntos ganados: %d\n", question.Points)
				} else {
					fmt.Printf("DEBUG CHECKBOX: Incorrecto (diferencias en selecciones o cantidad).\n")
				}
			} else {
				fmt.Printf("DEBUG CHECKBOX: Incorrecto (diferente cantidad de selecciones. Estudiante: %d, Correctas: %d).\n", len(studentSelectedStrings), len(correctAnswers))
			}

		case "boolean":
			var expectedBoolValue bool
			isCorrectAnswerParsable := true

			// Intentar interpretar question.CorrectAnswer
			if val, ok := question.CorrectAnswer.(bool); ok {
				expectedBoolValue = val
				fmt.Printf("DEBUG BOOLEAN: question.CorrectAnswer es bool: %t\n", val)
			} else if valStr, ok := question.CorrectAnswer.(string); ok {
				fmt.Printf("DEBUG BOOLEAN: question.CorrectAnswer es string: '%s'\n", valStr)
				lowerValStr := strings.ToLower(valStr)
				if lowerValStr == "true" || lowerValStr == "verdadero" {
					expectedBoolValue = true
				} else if lowerValStr == "false" || lowerValStr == "falso" {
					expectedBoolValue = false
				} else {
					fmt.Printf("DEBUG BOOLEAN: question.CorrectAnswer string ('%s') no reconocido como booleano.\n", valStr)
					isCorrectAnswerParsable = false
				}
			} else {
				fmt.Printf("DEBUG BOOLEAN: question.CorrectAnswer no es bool ni string. Es: %T\n", question.CorrectAnswer)
				isCorrectAnswerParsable = false
			}

			if !isCorrectAnswerParsable {
				continue
			}

			studentAnswerBool, typeOk := studentAnswer.(bool)
			if !typeOk {
				// Podrías también intentar convertir la respuesta del estudiante si viene como string "true"/"false"
				if studentAnswerStr, okStr := studentAnswer.(string); okStr {
					fmt.Printf("DEBUG BOOLEAN: studentAnswer es string: '%s', intentando convertir.\n", studentAnswerStr)
					lowerStudentStr := strings.ToLower(studentAnswerStr)
					if lowerStudentStr == "true" || lowerStudentStr == "verdadero" {
						studentAnswerBool = true
						typeOk = true
					} else if lowerStudentStr == "false" || lowerStudentStr == "falso" {
						studentAnswerBool = false
						typeOk = true
					} else {
						fmt.Printf("DEBUG BOOLEAN: studentAnswer string ('%s') no reconocido como booleano.\n", studentAnswerStr)
					}
				}
				if !typeOk {
					fmt.Printf("DEBUG BOOLEAN: studentAnswer no es bool (ni string convertible). Es: %T\n", studentAnswer)
					continue
				}
			}
			fmt.Printf("DEBUG BOOLEAN: Comparando %t (estudiante) con %t (correcta).\n", studentAnswerBool, expectedBoolValue)
			if studentAnswerBool == expectedBoolValue {
				earnedPoints += float64(question.Points)
				fmt.Printf("DEBUG BOOLEAN: ¡CORRECTO! Puntos ganados: %d\n", question.Points)
			} else {
				fmt.Printf("DEBUG BOOLEAN: Incorrecto.\n")
			}
		}
		fmt.Printf("Puntos acumulados: %f\n", earnedPoints)
	}

	fmt.Printf("\n--- FIN CALIFICACIÓN ---\n")
	fmt.Printf("Puntos Totales Ganados: %f, Puntos Totales Posibles: %d\n", earnedPoints, totalPoints)
	return earnedPoints, totalPoints
}

func (c *QuizController) GetQuizAnswersByStudent() echo.HandlerFunc {
	return func(e echo.Context) error {
		role := e.Get("user_role").(string)
		userID := e.Get("user_id").(string)

		if role != "org:student" {
			return ReturnReadResponse(e, echo.NewHTTPError(403, "Solo los estudiantes pueden ver sus cursos"), nil)
		}
		courses, err := c.QuizRepo.GetQuizAnswersByStudent(userID)
		return ReturnReadResponse(e, err, courses)
	}
}
