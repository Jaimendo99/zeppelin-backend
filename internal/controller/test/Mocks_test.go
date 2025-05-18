package controller_test

import (
	"errors"
	"net/http"
	"zeppelin/internal/domain"

	"github.com/clerkinc/clerk-sdk-go/clerk"
	"github.com/stretchr/testify/mock"
)

type MockAssignmentRepo struct {
	CreateA             func(userID string, courseID int) error
	VerifyA             func(assignmentID int) error
	GetAssignmentsByS   func(userID string) ([]domain.AssignmentWithCourse, error)
	GetStudentsByC      func(courseID int) ([]domain.AssignmentWithStudent, error)
	GetCourseIDByQR     func(qrCode string) (int, error)
	GetAssignmentsBySAC func(userID string, courseID int) (domain.AssignmentWithCourse, error)
}

func (m *MockAssignmentRepo) GetAssignmentsByStudentAndCourse(userID string, courseID int) (domain.AssignmentWithCourse, error) {
	if m.GetAssignmentsBySAC != nil {
		return m.GetAssignmentsBySAC(userID, courseID)
	}
	return domain.AssignmentWithCourse{}, errors.New("GetAssignmentsByStudentAndCourse function not implemented in mock")
}

func (m *MockAssignmentRepo) CreateAssignment(userID string, courseID int) error {
	if m.CreateA != nil {
		return m.CreateA(userID, courseID)
	}
	return errors.New("CreateA function not implemented in mock")
}

func (m *MockAssignmentRepo) VerifyAssignment(assignmentID int) error {
	if m.VerifyA != nil {
		return m.VerifyA(assignmentID)
	}
	return errors.New("VerifyA function not implemented in mock")
}

func (m *MockAssignmentRepo) GetAssignmentsByStudent(userID string) ([]domain.AssignmentWithCourse, error) {
	if m.GetAssignmentsByS != nil {
		return m.GetAssignmentsByS(userID)
	}
	return nil, errors.New("GetAssignmentsByS function not implemented in mock")
}

func (m *MockAssignmentRepo) GetStudentsByCourse(courseID int) ([]domain.AssignmentWithStudent, error) {
	if m.GetStudentsByC != nil {
		return m.GetStudentsByC(courseID)
	}
	return nil, errors.New("GetStudentsByC function not implemented in mock")
}

func (m *MockAssignmentRepo) GetCourseIDByQRCode(qrCode string) (int, error) {
	if m.GetCourseIDByQR != nil {
		return m.GetCourseIDByQR(qrCode)
	}
	return 0, errors.New("GetCourseIDByQR function not implemented in mock")
}

type MockNotificationRepo struct {
	SendToQ func(notification domain.NotificationQueue, queueName string) error
	// ConsumeFromQueue is not needed for testing SendNotification
}

func (m *MockNotificationRepo) SendToQueue(notification domain.NotificationQueue, queueName string) error {
	if m.SendToQ != nil {
		return m.SendToQ(notification, queueName)
	}
	return errors.New("SendToQ function not implemented in mock")
}

// ConsumeFromQueue needs to be implemented to satisfy the interface, but can be a no-op for these tests
func (m *MockNotificationRepo) ConsumeFromQueue(queueName string) error {
	// No implementation needed for SendNotification tests
	return errors.New("ConsumeFromQueue not implemented in mock for testing SendNotification")
}

type MockClerk struct {
	mock.Mock
}

func (m *MockClerk) VerifyToken(token string, opts ...clerk.VerifyTokenOption) (*clerk.SessionClaims, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) DecodeToken(token string) (*clerk.TokenClaims, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) CreateUser(params clerk.CreateUserParams) (*clerk.User, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) CreateOrganizationMembership(orgID string, params clerk.CreateOrganizationMembershipParams) (*clerk.OrganizationMembership, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MockClerk) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	// We pass body as separate args to Called for easier matching if needed
	allArgs := []interface{}{method, url}
	allArgs = append(allArgs, body...)
	args := m.Called(allArgs...)

	// Return value handling
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Request), args.Error(1)
}

func (m *MockClerk) Do(req *http.Request, v interface{}) (*http.Response, error) {
	args := m.Called(req, v) // Pass 'v' so mock.Run can access it

	// Return value handling
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

type MockCourseContentRepo struct {
	AddModuleT                    func(courseID int, module string, userID string) (int, error)
	GetContentByCourseT           func(courseID int) ([]domain.CourseContentWithDetails, error)
	GetContentByCourseForStudentT func(courseID int, userID string) ([]domain.CourseContentWithStudentDetails, error)
	AddSectionT                   func(input domain.AddSectionInput, userID string) (string, error)
	UpdateContentT                func(input domain.UpdateContentInput) error
	UpdateContentStatusT          func(contentID string, isActive bool) error
	UpdateModuleTitleT            func(courseContentID int, moduleTitle string) error
	UpdateUserContentStatusT      func(userID, contentID string, statusID int) error
	GetContentTypeIDT             func(contentID string) (int, error)
	GetUrlByContentIDT            func(contentID string) (string, error)
}

func (m MockCourseContentRepo) AddModule(courseID int, module string, userID string) (int, error) {
	if m.AddModuleT != nil {
		return m.AddModuleT(courseID, module, userID)
	}
	return 0, errors.New("AddModule not implemented")
}

func (m MockCourseContentRepo) GetContentByCourse(courseID int) ([]domain.CourseContentWithDetails, error) {
	if m.GetContentByCourseT != nil {
		return m.GetContentByCourseT(courseID)
	}
	return nil, errors.New("GetContentByCourse not implemented")
}

func (m MockCourseContentRepo) GetContentByCourseForStudent(courseID int, userID string) ([]domain.CourseContentWithStudentDetails, error) {
	if m.GetContentByCourseForStudentT != nil {
		return m.GetContentByCourseForStudentT(courseID, userID)
	}
	return nil, errors.New("GetContentByCourseForStudent not implemented")
}

func (m MockCourseContentRepo) AddSection(input domain.AddSectionInput, userID string) (string, error) {
	if m.AddSectionT != nil {
		return m.AddSectionT(input, userID)
	}
	return "", errors.New("AddSection not implemented")
}

func (m MockCourseContentRepo) UpdateContent(input domain.UpdateContentInput) error {
	if m.UpdateContentT != nil {
		return m.UpdateContentT(input)
	}
	return errors.New("UpdateContent not implemented")
}

func (m MockCourseContentRepo) UpdateContentStatus(contentID string, isActive bool) error {
	if m.UpdateContentStatusT != nil {
		return m.UpdateContentStatusT(contentID, isActive)
	}
	return errors.New("UpdateContentStatus not implemented")
}

func (m MockCourseContentRepo) UpdateModuleTitle(courseContentID int, moduleTitle string) error {
	if m.UpdateModuleTitleT != nil {
		return m.UpdateModuleTitleT(courseContentID, moduleTitle)
	}
	return errors.New("UpdateModuleTitle not implemented")
}

func (m MockCourseContentRepo) UpdateUserContentStatus(userID, contentID string, statusID int) error {
	if m.UpdateUserContentStatusT != nil {
		return m.UpdateUserContentStatusT(userID, contentID, statusID)
	}
	return errors.New("UpdateUserContentStatus not implemented")
}

func (m MockCourseContentRepo) GetContentTypeID(contentID string) (int, error) {
	if m.GetContentTypeIDT != nil {
		return m.GetContentTypeIDT(contentID)
	}
	return 0, errors.New("GetContentTypeID not implemented")
}

func (m MockCourseContentRepo) VerifyModuleOwnership(courseContentID int, userID string) error {
	return errors.New("VerifyModuleOwnership not implemented")
}

func (m MockCourseContentRepo) CreateContent(input domain.AddSectionInput) (string, error) {
	return "", errors.New("CreateContent not implemented")
}

func (m MockCourseContentRepo) GetUrlByContentID(contentID string) (string, error) {
	if m.GetUrlByContentIDT != nil {
		return m.GetUrlByContentIDT(contentID)
	}
	return "", errors.New("GetUrlByContentID not implemented")
}

// MockCourseRepo mocks CourseRepo
type MockCourseRepo struct {
	CreateC                        func(course domain.CourseDB) error
	GetCoursesByT                  func(teacherID string) ([]domain.CourseDB, error)
	GetCoursesByS                  func(studentID string) ([]domain.CourseDB, error)
	GetCourseByS2                  func(studentID string) ([]domain.CourseDbRelation, error)
	GetCourseByTeacherAndCourseIDT func(teacherID string, courseID int) (domain.CourseDB, error)
	GetCoursesByS2T                func(studentID, courseID string) (*domain.CourseDbRelation, error)
}

func (m MockCourseRepo) GetCourseByTeacherAndCourseID(teacherID string, courseID int) (domain.CourseDB, error) {
	if m.GetCourseByTeacherAndCourseIDT != nil {
		return m.GetCourseByTeacherAndCourseIDT(teacherID, courseID)
	}
	return domain.CourseDB{}, errors.New("GetCourseByTeacherAndCourseID function not implemented in mock")
}

func (m MockCourseRepo) CreateCourse(course domain.CourseDB) error {
	if m.CreateC != nil {
		return m.CreateC(course)
	}
	return errors.New("CreateC function not implemented in mock")
}

func (m MockCourseRepo) GetCoursesByTeacher(teacherID string) ([]domain.CourseDB, error) {
	if m.GetCoursesByT != nil {
		return m.GetCoursesByT(teacherID)
	}
	return nil, errors.New("GetCoursesByT function not implemented in mock")
}

func (m MockCourseRepo) GetCoursesByStudent(studentID string) ([]domain.CourseDB, error) {
	if m.GetCoursesByS != nil {
		return m.GetCoursesByS(studentID)
	}
	return nil, errors.New("GetCoursesByS function not implemented in mock")
}
func (m MockCourseRepo) GetCoursesByStudent2(studentID string) ([]domain.CourseDbRelation, error) {
	if m.GetCourseByS2 != nil {
		return m.GetCourseByS2(studentID)
	}
	return nil, errors.New("GetCourseByS2 function not implemented in mock")
}

func (m MockCourseRepo) GetCourse(studentID, courseID string) (*domain.CourseDbRelation, error) {
	if m.GetCoursesByS2T != nil {
		return m.GetCoursesByS2T(studentID, courseID)
	}
	return nil, errors.New("GetCourses function not implemented in mock")
}

// MockQuizRepo mocks QuizRepository
type MockQuizRepo struct {
	SaveQuizAttemptFn         func(attempt domain.QuizAnswer) error
	GetQuizAnswersByStudentFn func(studentID string) ([]domain.QuizSummary, error)
}

func (m *MockQuizRepo) SaveQuizAttempt(attempt domain.QuizAnswer) error {
	if m.SaveQuizAttemptFn != nil {
		return m.SaveQuizAttemptFn(attempt)
	}
	return errors.New("SaveQuizAttempt function not implemented in mock")
}

func (m *MockQuizRepo) GetQuizAnswersByStudent(studentID string) ([]domain.QuizSummary, error) {
	if m.GetQuizAnswersByStudentFn != nil {
		return m.GetQuizAnswersByStudentFn(studentID)
	}
	return nil, errors.New("GetQuizAnswersByStudent function not implemented in mock")
}

type MockUserFcmTokenRepo struct {
	CreateUserFcmTokenFn        func(token domain.UserFcmTokenDb) error
	GetUserFcmTokensByUserIDFn  func(userID string) ([]domain.UserFcmTokenDb, error)
	DeleteUserFcmTokenByTokenFn func(firebaseToken string) error
	UpdateDeviceInfoFn          func(firebaseToken string, deviceInfo string) error
	UpdateFirebaseTokenFn       func(userID, deviceType, newToken string) error
}

func (m MockUserFcmTokenRepo) CreateUserFcmToken(token domain.UserFcmTokenDb) error {
	if m.CreateUserFcmTokenFn != nil {
		return m.CreateUserFcmTokenFn(token)
	}
	return errors.New("CreateUserFcmToken not implemented")
}

func (m MockUserFcmTokenRepo) GetUserFcmTokensByUserID(userID string) ([]domain.UserFcmTokenDb, error) {
	if m.GetUserFcmTokensByUserIDFn != nil {
		return m.GetUserFcmTokensByUserIDFn(userID)
	}
	return nil, errors.New("GetUserFcmTokensByUserID not implemented")
}

func (m MockUserFcmTokenRepo) DeleteUserFcmTokenByToken(firebaseToken string) error {
	if m.DeleteUserFcmTokenByTokenFn != nil {
		return m.DeleteUserFcmTokenByTokenFn(firebaseToken)
	}
	return errors.New("DeleteUserFcmTokenByToken not implemented")
}

func (m MockUserFcmTokenRepo) UpdateDeviceInfo(firebaseToken string, deviceInfo string) error {
	if m.UpdateDeviceInfoFn != nil {
		return m.UpdateDeviceInfoFn(firebaseToken, deviceInfo)
	}
	return errors.New("UpdateDeviceInfo not implemented")
}

func (m MockUserFcmTokenRepo) UpdateFirebaseToken(userID string, firebaseToken string, deviceInfo string) error {
	if m.UpdateFirebaseTokenFn != nil {
		return m.UpdateFirebaseTokenFn(userID, firebaseToken, deviceInfo)
	}
	return errors.New("UpdateDeviceInfo not implemented")
}

// MockUserPomodoroRepo defines a mock for the UserPomodoroRepo interface
type MockUserPomodoroRepo struct {
	GetByUserIDFn    func(userID string) (*domain.UserPomodoro, error)
	UpdateByUserIDFn func(userID string, input domain.UpdatePomodoroInput) error
}

func (m MockUserPomodoroRepo) GetByUserID(userID string) (*domain.UserPomodoro, error) {
	if m.GetByUserIDFn != nil {
		return m.GetByUserIDFn(userID)
	}
	return nil, errors.New("GetByUserID not implemented")
}

func (m MockUserPomodoroRepo) UpdateByUserID(userID string, input domain.UpdatePomodoroInput) error {
	if m.UpdateByUserIDFn != nil {
		return m.UpdateByUserIDFn(userID, input)
	}
	return errors.New("UpdateByUserID not implemented")
}
