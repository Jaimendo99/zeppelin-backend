package test_test

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"regexp"
	"testing"
	"time"
	"zeppelin/internal/data"
	"zeppelin/internal/domain"
)

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

func TestCourseContentRepo_UpdateUserContentStatus(t *testing.T) {
	gormDb, mock := setupMockDb(t)
	repo := data.NewCourseContentRepo(gormDb, nil)

	userID := "user_123"
	contentID := "content_456"
	statusID := 2
	expectedSql := `UPDATE "user_content" SET "status_id"=\$1 WHERE user_id = \$2 AND content_id = \$3`

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(statusID, userID, contentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateUserContentStatus(userID, contentID, statusID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DB Error", func(t *testing.T) {
		dbErr := errors.New("update failed")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(statusID, userID, contentID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateUserContentStatus(userID, contentID, statusID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_AddModule(t *testing.T) {
	courseID := 1
	module := "New Module"
	userID := "teacher_123"
	expectedCourseQuery := quoteSql(`SELECT * FROM "course" WHERE course_id = $1 AND teacher_id = $2 ORDER BY "course"."course_id" LIMIT $3`)
	expectedCourseContentQuery := quoteSql(`SELECT * FROM "course_content" WHERE course_id = $1 AND module = $2 ORDER BY "course_content"."course_content_id" LIMIT $3`)
	expectedInsertQuery := quoteSql(`INSERT INTO "course_content" ("course_id","module","module_index","created_at") VALUES ($1,$2,$3,$4) RETURNING "course_content_id"`)

	t.Run("Success - Module does not exist", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		courseRows := sqlmock.NewRows([]string{"course_id", "teacher_id"}).AddRow(courseID, userID)
		mock.ExpectQuery(expectedCourseQuery).WithArgs(courseID, userID, 1).WillReturnRows(courseRows)

		mock.ExpectQuery(expectedCourseContentQuery).WithArgs(courseID, module, 1).WillReturnError(gorm.ErrRecordNotFound)

		mock.ExpectBegin()
		mock.ExpectQuery(expectedInsertQuery).
			WithArgs(courseID, module, sqlmock.AnyArg(), sqlmock.AnyArg()). // Using AnyArg for time
			WillReturnRows(sqlmock.NewRows([]string{"course_content_id"}).AddRow(1))
		mock.ExpectCommit()

		courseContentID, err := repo.AddModule(courseID, module, userID)

		assert.NoError(t, err)
		assert.Equal(t, 1, courseContentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Course does not belong to the teacher", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectQuery(expectedCourseQuery).WithArgs(courseID, userID, 1).WillReturnError(gorm.ErrRecordNotFound)

		courseContentID, err := repo.AddModule(courseID, module, userID)

		assert.Error(t, err)
		assert.Equal(t, "course does not belong to the teacher", err.Error())
		assert.Equal(t, 0, courseContentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_CreateContent(t *testing.T) {
	input := domain.AddSectionInput{
		CourseContentID: 1,
		ContentTypeID:   1,
		Title:           "New Section",
		Description:     "Section description",
	}
	generatedContentID := "generated_uid_123"
	expectedCountQuery := quoteSql(`SELECT count(*) FROM "course_content" WHERE course_content_id = $1`)
	expectedInsertQuery := quoteSql(`INSERT INTO "content" ("content_id","course_content_id","content_type_id","title","url","description","section_index","is_active") VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, func() string { return generatedContentID })

		mock.ExpectQuery(expectedCountQuery).WithArgs(input.CourseContentID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectBegin()
		mock.ExpectExec(expectedInsertQuery).
			WithArgs(generatedContentID, input.CourseContentID, input.ContentTypeID, input.Title, "", input.Description, 0, false). // Assuming default values
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		contentID, err := repo.CreateContent(input)

		assert.NoError(t, err)
		assert.Equal(t, generatedContentID, contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - CourseContentID not found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, func() string { return generatedContentID })

		mock.ExpectQuery(expectedCountQuery).WithArgs(input.CourseContentID).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		contentID, err := repo.CreateContent(input)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, "", contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_GetContentByCourse(t *testing.T) {
	courseID := 1
	expectedCourseContentQuery := quoteSql(`SELECT * FROM "course_content" WHERE course_id = $1 ORDER BY module_index`)
	expectedContentQuery := quoteSql(`SELECT * FROM "content" WHERE "content"."course_content_id" IN ($1,$2) ORDER BY section_index`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		courseContentRows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at"}).
			AddRow(1, courseID, "Module 1", 1, time.Now()).
			AddRow(2, courseID, "Module 2", 2, time.Now())
		mock.ExpectQuery(expectedCourseContentQuery).WithArgs(courseID).WillReturnRows(courseContentRows)

		contentRows := sqlmock.NewRows([]string{"content_id", "course_content_id", "content_type_id", "title", "url", "description", "section_index", "is_active"}).
			AddRow("content_1", 1, 1, "Section 1.1", "url1", "desc1", 1, true).
			AddRow("content_2", 1, 2, "Section 1.2", "url2", "desc2", 2, true).
			AddRow("content_3", 2, 1, "Section 2.1", "url3", "desc3", 1, true)
		mock.ExpectQuery(expectedContentQuery).WithArgs(1, 2).WillReturnRows(contentRows)

		courseContents, err := repo.GetContentByCourse(courseID)

		assert.NoError(t, err)
		assert.Len(t, courseContents, 2)
		assert.Len(t, courseContents[0].Details, 2)
		assert.Len(t, courseContents[1].Details, 1)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db error")
		mock.ExpectQuery(expectedCourseContentQuery).WithArgs(courseID).WillReturnError(dbErr)

		courseContents, err := repo.GetContentByCourse(courseID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, courseContents)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_UpdateContent(t *testing.T) {
	input := domain.UpdateContentInput{
		ContentID:   "content_123",
		Title:       "Updated Title",
		Url:         "http://newurl.com",
		Description: "Updated description",
	}
	// Note: The order of columns in the UPDATE query might vary, use regexp if needed
	expectedSql := quoteSql(`UPDATE "content" SET "description"=$1,"title"=$2,"url"=$3 WHERE content_id = $4`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(input.Description, input.Title, input.Url, input.ContentID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.UpdateContent(input)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectExec(expectedSql).
			WithArgs(input.Description, input.Title, input.Url, input.ContentID).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		err := repo.UpdateContent(input)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_UpdateContentStatus(t *testing.T) {
	contentID := "content_123"
	isActive := true
	expectedSql := quoteSql(`UPDATE "content" SET "is_active"=$1 WHERE content_id = $2`)

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

	t.Run("Error - DB Error", func(t *testing.T) {
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

func TestCourseContentRepo_VerifyModuleOwnership(t *testing.T) {
	courseContentID := 1
	userID := "teacher_123"
	// Use the revised quoteSql and directly pass the regexp object
	expectedQueryRegex := regexp.MustCompile(quoteSql(`SELECT "course_content"."course_content_id","course_content"."course_id","course_content"."module","course_content"."module_index","course_content"."created_at" FROM "course_content" JOIN course ON course_content.course_id = course.course_id WHERE course_content.course_content_id = $1 AND course.teacher_id = $2 ORDER BY "course_content"."course_content_id" LIMIT $3`))

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		rows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at"}).AddRow(courseContentID, 1, "Module 1", 1, time.Now())
		mock.ExpectQuery(expectedQueryRegex.String()).WithArgs(courseContentID, userID, 1).WillReturnRows(rows)

		err := repo.VerifyModuleOwnership(courseContentID, userID)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Module does not belong to the teacher's course", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectQuery(expectedQueryRegex.String()).WithArgs(courseContentID, userID, 1).WillReturnError(gorm.ErrRecordNotFound)

		err := repo.VerifyModuleOwnership(courseContentID, userID)

		assert.Error(t, err)
		assert.Equal(t, "module does not belong to the teacher's course", err.Error())
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db error")
		mock.ExpectQuery(expectedQueryRegex.String()).WithArgs(courseContentID, userID, 1).WillReturnError(dbErr)

		err := repo.VerifyModuleOwnership(courseContentID, userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_GetContentTypeID(t *testing.T) {
	contentID := "content_123"
	// Refined regex for SELECT column
	expectedQueryRegex := regexp.MustCompile(quoteSql(`SELECT "content_type_id" FROM "content" WHERE content_id = $1 ORDER BY "content"."content_id" LIMIT $2`))

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		rows := sqlmock.NewRows([]string{"content_type_id"}).AddRow(1)
		mock.ExpectQuery(expectedQueryRegex.String()).WithArgs(contentID, 1).WillReturnRows(rows)

		contentTypeID, err := repo.GetContentTypeID(contentID)

		assert.NoError(t, err)
		assert.Equal(t, 1, contentTypeID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Content not found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectQuery(expectedQueryRegex.String()).WithArgs(contentID, 1).WillReturnError(gorm.ErrRecordNotFound)

		contentTypeID, err := repo.GetContentTypeID(contentID)

		assert.Error(t, err)
		assert.Equal(t, "content not found", err.Error())
		assert.Equal(t, 0, contentTypeID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db error")
		mock.ExpectQuery(expectedQueryRegex.String()).WithArgs(contentID, 1).WillReturnError(dbErr)

		contentTypeID, err := repo.GetContentTypeID(contentID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, 0, contentTypeID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_AddSection(t *testing.T) {
	input := domain.AddSectionInput{
		CourseContentID: 1,
		ContentTypeID:   1,
		Title:           "New Section",
		Description:     "Section description",
	}
	userID := "teacher_123"
	generatedContentID := "generated_uid_123"
	expectedOwnershipQuery := quoteSql(`SELECT "course_content"."course_content_id","course_content"."course_id","course_content"."module","course_content"."module_index","course_content"."created_at" FROM "course_content" JOIN course ON course_content.course_id = course.course_id WHERE course_content.course_content_id = $1 AND course.teacher_id = $2 ORDER BY "course_content"."course_content_id" LIMIT $3`)
	expectedCountQuery := quoteSql(`SELECT count(*) FROM "course_content" WHERE course_content_id = $1`)
	expectedInsertQuery := quoteSql(`INSERT INTO "content" ("content_id","course_content_id","content_type_id","title","url","description","section_index","is_active") VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`)

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, func() string { return generatedContentID })

		// Mock VerifyModuleOwnership
		rows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at"}).
			AddRow(input.CourseContentID, 1, "Module 1", 1, time.Now())
		mock.ExpectQuery(expectedOwnershipQuery).
			WithArgs(input.CourseContentID, userID, 1).
			WillReturnRows(rows)

		// Mock CreateContent
		mock.ExpectQuery(expectedCountQuery).
			WithArgs(input.CourseContentID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		mock.ExpectBegin()
		mock.ExpectExec(expectedInsertQuery).
			WithArgs(generatedContentID, input.CourseContentID, input.ContentTypeID, input.Title, "", input.Description, 0, false).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		contentID, err := repo.AddSection(input, userID)

		assert.NoError(t, err)
		assert.Equal(t, generatedContentID, contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid Module Ownership", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, func() string { return generatedContentID })

		mock.ExpectQuery(expectedOwnershipQuery).
			WithArgs(input.CourseContentID, userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		contentID, err := repo.AddSection(input, userID)

		assert.Error(t, err)
		assert.Equal(t, "module does not belong to the teacher's course", err.Error())
		assert.Equal(t, "", contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Invalid ContentTypeID", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, func() string { return generatedContentID })

		invalidInput := input
		invalidInput.ContentTypeID = 4 // Invalid content type ID

		// Mock VerifyModuleOwnership
		rows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at"}).
			AddRow(input.CourseContentID, 1, "Module 1", 1, time.Now())
		mock.ExpectQuery(expectedOwnershipQuery).
			WithArgs(input.CourseContentID, userID, 1).
			WillReturnRows(rows)

		contentID, err := repo.AddSection(invalidInput, userID)

		assert.Error(t, err)
		assert.Equal(t, "invalid content_type_id", err.Error())
		assert.Equal(t, "", contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - DB Error in CreateContent", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, func() string { return generatedContentID })

		// Mock VerifyModuleOwnership
		rows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at"}).
			AddRow(input.CourseContentID, 1, "Module 1", 1, time.Now())
		mock.ExpectQuery(expectedOwnershipQuery).
			WithArgs(input.CourseContentID, userID, 1).
			WillReturnRows(rows)

		// Mock CreateContent DB error
		mock.ExpectQuery(expectedCountQuery).
			WithArgs(input.CourseContentID).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		dbErr := errors.New("db insert error")
		mock.ExpectBegin()
		mock.ExpectExec(expectedInsertQuery).
			WithArgs(generatedContentID, input.CourseContentID, input.ContentTypeID, input.Title, "", input.Description, 0, false).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		contentID, err := repo.AddSection(input, userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, "", contentID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_GetContentByCourseForStudent(t *testing.T) {
	courseID := 1
	userID := "student_123"
	expectedCourseContentQuery := quoteSql(`SELECT * FROM "course_content" WHERE course_id = $1 ORDER BY module_index`)
	expectedContentQuery := quoteSql(`SELECT * FROM "content" WHERE is_active = $1 AND "content"."course_content_id" = $2 ORDER BY section_index`)
	expectedUserContentQuery := quoteSql(`SELECT * FROM "user_content" WHERE "user_content"."content_id" = $1 AND user_id = $2`)

	t.Run("Success - With Content and Status", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		// Mock course_content query
		courseContentRows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at"}).
			AddRow(1, courseID, "Module 1", 1, time.Now())
		mock.ExpectQuery(expectedCourseContentQuery).
			WithArgs(courseID).
			WillReturnRows(courseContentRows)

		// Mock content query for course_content_id = 1
		contentRows := sqlmock.NewRows([]string{"content_id", "course_content_id", "content_type_id", "title", "url", "description", "section_index", "is_active"}).
			AddRow("content_1", 1, 1, "Section 1", "url1", "desc1", 1, true)
		mock.ExpectQuery(expectedContentQuery).
			WithArgs(true, 1). // is_active = true, course_content_id = 1
			WillReturnRows(contentRows)

		// Mock UserContent preload
		userContentRows := sqlmock.NewRows([]string{"user_id", "content_id", "status_id"}).
			AddRow(userID, "content_1", 2)
		mock.ExpectQuery(expectedUserContentQuery).
			WithArgs("content_1", userID). // Note: content_id first, then user_id
			WillReturnRows(userContentRows)

		result, err := repo.GetContentByCourseForStudent(courseID, userID)

		assert.NoError(t, err)
		if assert.Len(t, result, 1) {
			assert.Len(t, result[0].Details, 1)
			assert.Equal(t, "Module 1", result[0].Module)
			assert.Equal(t, "content_1", result[0].Details[0].ContentID)
			if assert.NotNil(t, result[0].Details[0].StatusID) {
				assert.Equal(t, 2, *result[0].Details[0].StatusID)
			}
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Success - No Valid Content", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		// Mock course_content query
		courseContentRows := sqlmock.NewRows([]string{"course_content_id", "course_id", "module", "module_index", "created_at"}).
			AddRow(1, courseID, "Module 1", 1, time.Now())
		mock.ExpectQuery(expectedCourseContentQuery).
			WithArgs(courseID).
			WillReturnRows(courseContentRows)

		// Mock content query with no UserContent
		contentRows := sqlmock.NewRows([]string{"content_id", "course_content_id", "content_type_id", "title", "url", "description", "section_index", "is_active"}).
			AddRow("content_1", 1, 1, "Section 1", "url1", "desc1", 1, true)
		mock.ExpectQuery(expectedContentQuery).
			WithArgs(true, 1).
			WillReturnRows(contentRows)

		// Mock UserContent preload with no rows
		mock.ExpectQuery(expectedUserContentQuery).
			WithArgs("content_1", userID).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "content_id", "status_id"}))

		result, err := repo.GetContentByCourseForStudent(courseID, userID)

		assert.NoError(t, err)
		assert.Len(t, result, 0) // No content with UserContent, so empty result
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db error")
		mock.ExpectQuery(expectedCourseContentQuery).
			WithArgs(courseID).
			WillReturnError(dbErr)

		result, err := repo.GetContentByCourseForStudent(courseID, userID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Nil(t, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestCourseContentRepo_GetUrlByContentID(t *testing.T) {
	contentID := "content_123"
	expectedQuery := quoteSql(`SELECT "url" FROM "content" WHERE content_id = $1 ORDER BY "content"."content_id" LIMIT $2`)
	expectedURL := "http://example.com/video"

	t.Run("Success", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		rows := sqlmock.NewRows([]string{"url"}).AddRow(expectedURL)
		mock.ExpectQuery(expectedQuery).
			WithArgs(contentID, 1).
			WillReturnRows(rows)

		url, err := repo.GetUrlByContentID(contentID)

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - Content not found", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		mock.ExpectQuery(expectedQuery).
			WithArgs(contentID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		url, err := repo.GetUrlByContentID(contentID)

		assert.Error(t, err)
		assert.Equal(t, "content not found", err.Error())
		assert.Equal(t, "", url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Error - DB Error", func(t *testing.T) {
		gormDb, mock := setupMockDb(t)
		repo := data.NewCourseContentRepo(gormDb, nil)

		dbErr := errors.New("db error")
		mock.ExpectQuery(expectedQuery).
			WithArgs(contentID, 1).
			WillReturnError(dbErr)

		url, err := repo.GetUrlByContentID(contentID)

		assert.Error(t, err)
		assert.Equal(t, dbErr, err)
		assert.Equal(t, "", url)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
