package test_test

import (
	"encoding/json"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"testing"
	"zeppelin/internal/data"
)

func TestCourseContentRepo_CreateVideo(t *testing.T) {
	url := "http://video.com"
	title := "Video Title"
	description := "Video Description"
	contentID := "vid_123"
	expectedSql := `INSERT INTO "video" \(.+?\) VALUES \(\$1,\$2,\$3,\$4\)`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		generateUID := func() string { return contentID }
		repo := data.NewCourseContentRepo(gormDb, generateUID)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(contentID, description, title, url).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		resultID, err := repo.CreateVideo(url, title, description)

		assert.NoError(t, err)
		assert.Equal(t, contentID, resultID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		generateUID := func() string { return contentID }
		repo := data.NewCourseContentRepo(gormDb, generateUID)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(contentID, description, title, url).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		resultID, err := repo.CreateVideo(url, title, description)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Empty(t, resultID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_CreateQuiz(t *testing.T) {
	title := "Quiz Title"
	description := "Quiz Description"
	jsonContent, _ := json.Marshal(map[string]string{"question": "Q1"})
	contentID := "quiz_123"
	expectedSql := `INSERT INTO "quiz" \(.+?\) VALUES \(\$1,\$2,\$3,\$4\)`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		generateUID := func() string { return contentID }
		repo := data.NewCourseContentRepo(gormDb, generateUID)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(contentID, description, jsonContent, title).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		resultID, err := repo.CreateQuiz(title, description, jsonContent)

		assert.NoError(t, err)
		assert.Equal(t, contentID, resultID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		generateUID := func() string { return contentID }
		repo := data.NewCourseContentRepo(gormDb, generateUID)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(contentID, description, jsonContent, title).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		resultID, err := repo.CreateQuiz(title, description, jsonContent)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Empty(t, resultID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_CreateText(t *testing.T) {
	title := "Text Title"
	url := "http://text.com"
	jsonContent, _ := json.Marshal(map[string]string{"text": "Sample"})
	contentID := "text_123"
	expectedSql := `INSERT INTO "text" \(.+?\) VALUES \(\$1,\$2,\$3,\$4\)`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		generateUID := func() string { return contentID }
		repo := data.NewCourseContentRepo(gormDb, generateUID)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(contentID, jsonContent, title, url).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		resultID, err := repo.CreateText(title, url, jsonContent)

		assert.NoError(t, err)
		assert.Equal(t, contentID, resultID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		generateUID := func() string { return contentID }
		repo := data.NewCourseContentRepo(gormDb, generateUID)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(contentID, jsonContent, title, url).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		resultID, err := repo.CreateText(title, url, jsonContent)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Empty(t, resultID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_AddVideoSection(t *testing.T) {
	courseID := 1
	contentID := "vid_123"
	module := "Module 1"
	sectionIndex := 1
	moduleIndex := 1
	expectedSql := `INSERT INTO "course_content" \(.+?\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\)`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectQuery(expectedSql).
			WithArgs(courseID, module, "video", contentID, sectionIndex, moduleIndex, true, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"course_content_id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.AddVideoSection(courseID, contentID, module, sectionIndex, moduleIndex)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectQuery(expectedSql).
			WithArgs(courseID, module, "video", contentID, sectionIndex, moduleIndex, true, sqlmock.AnyArg()).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.AddVideoSection(courseID, contentID, module, sectionIndex, moduleIndex)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_AddQuizSection(t *testing.T) {
	courseID := 1
	contentID := "quiz_123"
	module := "Module 1"
	sectionIndex := 1
	moduleIndex := 1
	expectedSql := `INSERT INTO "course_content" \(.+?\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\)`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectQuery(expectedSql).
			WithArgs(courseID, module, "quiz", contentID, sectionIndex, moduleIndex, true, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"course_content_id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.AddQuizSection(courseID, contentID, module, sectionIndex, moduleIndex)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectQuery(expectedSql).
			WithArgs(courseID, module, "quiz", contentID, sectionIndex, moduleIndex, true, sqlmock.AnyArg()).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.AddQuizSection(courseID, contentID, module, sectionIndex, moduleIndex)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_AddTextSection(t *testing.T) {
	courseID := 1
	contentID := "text_123"
	module := "Module 1"
	sectionIndex := 1
	moduleIndex := 1
	expectedSql := `INSERT INTO "course_content" \(.+?\) VALUES \(\$1,\$2,\$3,\$4,\$5,\$6,\$7,\$8\)`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectQuery(expectedSql).
			WithArgs(courseID, module, "text", contentID, sectionIndex, moduleIndex, true, sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"course_content_id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.AddTextSection(courseID, contentID, module, sectionIndex, moduleIndex)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db insert error")

		mock.ExpectBegin()
		mock.ExpectQuery(expectedSql).
			WithArgs(courseID, module, "text", contentID, sectionIndex, moduleIndex, true, sqlmock.AnyArg()).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.AddTextSection(courseID, contentID, module, sectionIndex, moduleIndex)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_UpdateVideo(t *testing.T) {
	contentID := "vid_123"
	title := "Updated Title"
	url := "http://updated.com"
	description := "Updated Description"
	expectedSql := `UPDATE "video" SET "description"=\$1,"title"=\$2,"url"=\$3 WHERE content_id = \$4`

	t.Run("Success with all fields", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(description, title, url, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateVideo(contentID, title, url, description)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No updates", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		err := repo.UpdateVideo(contentID, "", "", "")

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(description, title, url, contentID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateVideo(contentID, title, url, description)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_UpdateQuiz(t *testing.T) {
	contentID := "quiz_123"
	title := "Updated Title"
	description := "Updated Description"
	jsonContent, _ := json.Marshal(map[string]string{"question": "Updated Q1"})
	expectedSql := `UPDATE "quiz" SET "description"=\$1,"json_content"=\$2,"title"=\$3 WHERE content_id = \$4`

	t.Run("Success with all fields", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(description, jsonContent, title, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateQuiz(contentID, title, description, jsonContent)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No updates", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		err := repo.UpdateQuiz(contentID, "", "", nil)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(description, jsonContent, title, contentID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateQuiz(contentID, title, description, jsonContent)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_UpdateText(t *testing.T) {
	contentID := "text_123"
	title := "Updated Title"
	url := "http://updated.com"
	jsonContent, _ := json.Marshal(map[string]string{"text": "Updated Sample"})
	expectedSql := `UPDATE "text" SET "json_content"=\$1,"title"=\$2,"url"=\$3 WHERE content_id = \$4`

	t.Run("Success with all fields", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(jsonContent, title, url, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateText(contentID, title, url, jsonContent)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No updates", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		err := repo.UpdateText(contentID, "", "", nil)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(jsonContent, title, url, contentID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateText(contentID, title, url, jsonContent)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_UpdateContentStatus(t *testing.T) {
	contentID := "vid_123"
	isActive := false
	expectedSql := `UPDATE "course_content" SET "is_active"=\$1 WHERE content_id = \$2`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(isActive, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateContentStatus(contentID, isActive)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(isActive, contentID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateContentStatus(contentID, isActive)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_UpdateModuleTitle(t *testing.T) {
	courseContentID := 1
	moduleTitle := "Updated Module"
	expectedSql := `UPDATE "course_content" SET "module"=\$1 WHERE course_content_id = \$2`

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(moduleTitle, courseContentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateModuleTitle(courseContentID, moduleTitle)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(moduleTitle, courseContentID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateModuleTitle(courseContentID, moduleTitle)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
