package test_test

import (
	"regexp"
	"testing"
	"time"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuizAnswerRepository provides tests for the Quiz Answer repository implementation

func TestSaveQuizAttempt(t *testing.T) {
	t.Run("successful save", func(t *testing.T) {
		// Setup
		gormDb, mock := setupMockDb(t)
		repo := data.NewQuizRepository(gormDb)

		grade := 85.5
		reviewedAt := time.Now()
		totalPoints := 100

		quizAnswer := domain.QuizAnswer{
			ContentID:     "content123",
			UserID:        "user456",
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(time.Hour),
			Grade:         &grade,
			ReviewedAt:    &reviewedAt,
			QuizURL:       "http://example.com/quiz",
			QuizAnswerURL: "http://example.com/answer",
			TotalPoints:   &totalPoints,
		}

		// Expect the insert query
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "quiz_answer"`)).
			WithArgs(
				quizAnswer.ContentID,
				quizAnswer.UserID,
				quizAnswer.StartTime,
				quizAnswer.EndTime,
				quizAnswer.Grade,
				quizAnswer.ReviewedAt,
				quizAnswer.QuizURL,
				quizAnswer.QuizAnswerURL,
				quizAnswer.TotalPoints,
			).
			WillReturnRows(sqlmock.NewRows([]string{"quiz_answer_id"}).AddRow(1))
		mock.ExpectCommit()

		// Execute
		err := repo.SaveQuizAttempt(quizAnswer)

		// Assert
		assert.NoError(t, err)
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		gormDb, mock := setupMockDb(t)
		repo := data.NewQuizRepository(gormDb)

		quizAnswer := domain.QuizAnswer{
			ContentID:     "content456",
			UserID:        "user789",
			StartTime:     time.Now(),
			EndTime:       time.Now().Add(time.Hour),
			QuizURL:       "http://example.com/quiz2",
			QuizAnswerURL: "http://example.com/answer2",
		}

		// Expect the insert query with error
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "quiz_answer"`)).
			WithArgs(
				quizAnswer.ContentID,
				quizAnswer.UserID,
				quizAnswer.StartTime,
				quizAnswer.EndTime,
				nil, // Grade is nil
				nil, // ReviewedAt is nil
				quizAnswer.QuizURL,
				quizAnswer.QuizAnswerURL,
				nil, // TotalPoints is nil
			).
			WillReturnError(sqlmock.ErrCancelled)
		mock.ExpectRollback()

		// Execute
		err := repo.SaveQuizAttempt(quizAnswer)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error creating new quiz attempt")
		
		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}

func TestGetQuizAnswersByStudent(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		// Setup
		gormDb, mock := setupMockDb(t)
		repo := data.NewQuizRepository(gormDb)

		studentID := "student123"
		totalGrade := 90.5
		totalPoints := 100
		quizCount := 2

		// Now we expect a query that returns aggregated data
		lastQuizTime := time.Date(2023, 10, 15, 14, 30, 0, 0, time.UTC) // Fixed timestamp for testing
		expectedRows := sqlmock.NewRows([]string{
			"content_id",
			"quiz_count",
			"total_grade",
			"total_points",
			"last_quiz_time",
		}).AddRow(
			"content123",
			quizCount,
			totalGrade,
			totalPoints,
			lastQuizTime,
		)

		// Expect the select query with aggregation
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT content_id, COUNT(*) as quiz_count, SUM(grade) as total_grade, SUM(total_points) as total_points, max(end_time) as last_quiz_time FROM "quiz_answer" WHERE user_id = $1 GROUP BY "content_id"`)).
			WithArgs(studentID).
			WillReturnRows(expectedRows)

		// Execute
		results, err := repo.GetQuizAnswersByStudent(studentID)

		// Assert
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "content123", results[0].ContentID)
		assert.Equal(t, quizCount, results[0].QuizCount)
		assert.Equal(t, &totalGrade, results[0].TotalGrade)
		assert.Equal(t, &totalPoints, results[0].TotalPoints)
		assert.Equal(t, &lastQuizTime, results[0].LastQuizTime)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("database error", func(t *testing.T) {
		// Setup
		gormDb, mock := setupMockDb(t)
		repo := data.NewQuizRepository(gormDb)

		studentID := "student456"

		// Expect the select query with error
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT content_id, COUNT(*) as quiz_count, SUM(grade) as total_grade, SUM(total_points) as total_points, max(end_time) as last_quiz_time FROM "quiz_answer" WHERE user_id = $1 GROUP BY "content_id"`)).
			WithArgs(studentID).
			WillReturnError(sqlmock.ErrCancelled)

		// Execute
		results, err := repo.GetQuizAnswersByStudent(studentID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, results)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})

	t.Run("no quiz answers found", func(t *testing.T) {
		// Setup
		gormDb, mock := setupMockDb(t)
		repo := data.NewQuizRepository(gormDb)

		studentID := "student789"

		// Expect the select query with no results
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT content_id, COUNT(*) as quiz_count, SUM(grade) as total_grade, SUM(total_points) as total_points, max(end_time) as last_quiz_time FROM "quiz_answer" WHERE user_id = $1 GROUP BY "content_id"`)).
			WithArgs(studentID).
			WillReturnRows(sqlmock.NewRows([]string{
				"content_id", "quiz_count", "total_grade", "total_points", "last_quiz_time",
			}))

		// Execute
		results, err := repo.GetQuizAnswersByStudent(studentID)

		// Assert
		assert.NoError(t, err)
		assert.Empty(t, results)

		err = mock.ExpectationsWereMet()
		assert.NoError(t, err)
	})
}