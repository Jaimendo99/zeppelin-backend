package config_test

import (
	"context"
	"crypto/tls"
	"errors"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/option"
	"net/smtp"
	"testing"
	"zeppelin/internal/config"
	"zeppelin/internal/domain"
)

func TestGetSmtpConfig_BeforeInit(t *testing.T) {
	cfg := config.GetSmtpConfig()
	assert.Nil(t, cfg, "GetSmtpConfig should return nil before InitSmtp")
}

func TestInitSmtpAndGetSmtpConfig(t *testing.T) {
	testPassword := "test-password-123"
	config.InitSmtp(testPassword)
	retrievedConfig := config.GetSmtpConfig()
	assert.NotNil(t, retrievedConfig, "GetSmtpConfig should return a non-nil config after InitSmtp")
	assert.Equal(t, "smtp.gmail.com", retrievedConfig.Host, "Host should be set correctly")
	assert.Equal(t, "587", retrievedConfig.Port, "Port should be set correctly")
	assert.Equal(t, "zepppelin1.1@gmail.com", retrievedConfig.Username, "Username should be set correctly")
	assert.NotNil(t, retrievedConfig.Auth, "Auth should be initialized")

}

type MockFirebaseApp struct {
	mock.Mock
}

func (m *MockFirebaseApp) Messaging(ctx context.Context) (*messaging.Client, error) {
	args := m.Called(ctx)
	var client *messaging.Client
	if args.Get(0) != nil {
		client = args.Get(0).(*messaging.Client)
	}
	return client, args.Error(1)
}

// --- Test Cases ---

const validDummyCreds = `{"type": "service_account", "project_id": "test-project"}`
const invalidDummyCreds = `invalid json`

func TestGetFCMClient_BeforeInit(t *testing.T) {
	config.ResetFCMState()
	client := config.GetFCMClient()
	assert.Nil(t, client, "GetFCMClient should return nil before InitFCM")
}

func TestInitFCM_Success(t *testing.T) {
	config.ResetFCMState()

	mockApp := new(MockFirebaseApp)
	mockMessagingClient := &messaging.Client{} // Dummy client

	mockApp.On("Messaging", mock.Anything).Return(mockMessagingClient, nil)

	// Store the original function *value* that the pointer points to
	originalFunc := *config.FirebaseNewApp

	// Assign to the dereferenced pointer (*config.FirebaseNewApp)
	// The signature MUST match the underlying variable: func(...) (config.FirebaseApp, error)
	*config.FirebaseNewApp = func(ctx context.Context, fbConfig *firebase.Config, opts ...option.ClientOption) (config.FirebaseApp, error) {
		// Return the mockApp (type *MockFirebaseApp, satisfies config.FirebaseApp)
		// and nil error. This matches the required (config.FirebaseApp, error) signature.
		return mockApp, nil
	}

	// Use t.Cleanup to restore the original function *value*
	t.Cleanup(func() {
		*config.FirebaseNewApp = originalFunc
		// Optionally reset client state if needed after test
		// config.ResetFCMState()
	})

	// Call the function under test
	err := config.InitFCM(validDummyCreds)

	// Assertions
	assert.NoError(t, err, "InitFCM should succeed")
	retrievedClient := config.GetFCMClient()
	assert.NotNil(t, retrievedClient, "GetFCMClient should return non-nil after successful InitFCM")
	assert.Same(t, mockMessagingClient, retrievedClient, "GetFCMClient should return the client from Messaging")
	mockApp.AssertExpectations(t) // Verify that Messaging was called
}

func TestInitFCM_NewApp_Error(t *testing.T) {
	config.ResetFCMState()
	expectedErr := errors.New("firebase init failed")

	originalFunc := *config.FirebaseNewApp

	// Assign to the dereferenced pointer (*config.FirebaseNewApp)
	// Signature: func(...) (config.FirebaseApp, error)
	*config.FirebaseNewApp = func(ctx context.Context, fbConfig *firebase.Config, opts ...option.ClientOption) (config.FirebaseApp, error) {
		// Return nil for the interface type (explicitly typed) and the error
		var nilApp config.FirebaseApp = nil
		return nilApp, expectedErr
	}
	t.Cleanup(func() {
		*config.FirebaseNewApp = originalFunc
	})

	err := config.InitFCM(invalidDummyCreds)

	assert.Error(t, err, "InitFCM should return an error")
	assert.ErrorIs(t, err, expectedErr, "InitFCM should return the specific error from NewApp")
	assert.Nil(t, config.GetFCMClient(), "FCM client should be nil after NewApp failure")
}

func TestInitFCM_Messaging_Error(t *testing.T) {
	config.ResetFCMState()
	mockApp := new(MockFirebaseApp)
	expectedErr := errors.New("messaging client failed")
	mockApp.On("Messaging", mock.Anything).Return(nil, expectedErr)

	originalFunc := *config.FirebaseNewApp

	*config.FirebaseNewApp = func(ctx context.Context, fbConfig *firebase.Config, opts ...option.ClientOption) (config.FirebaseApp, error) {
		return mockApp, nil
	}
	t.Cleanup(func() {
		*config.FirebaseNewApp = originalFunc
	})

	err := config.InitFCM(validDummyCreds)

	assert.Error(t, err, "InitFCM should return an error")
	assert.ErrorIs(t, err, expectedErr, "InitFCM should return the specific error from Messaging")
	assert.Nil(t, config.GetFCMClient(), "FCM client should be nil after Messaging failure")
	mockApp.AssertExpectations(t)
}

// --- Mock AMQP Connection ---
type MockAmqpConnection struct {
	mock.Mock
	// *** UNCOMMENT THIS LINE ***
	// Embed the real connection type. This makes *MockAmqpConnection
	// also satisfy the *amqp091.Connection type for the assertion.
	*amqp091.Connection
}

// Implement the Channel method for the mock
// This method will OVERRIDE the embedded type's Channel method when called on the mock.
func (m *MockAmqpConnection) Channel() (*amqp091.Channel, error) {
	args := m.Called()
	var chanVal *amqp091.Channel
	if args.Get(0) != nil {
		chanVal = args.Get(0).(*amqp091.Channel)
	}
	return chanVal, args.Error(1)
} // --- Test Cases ---

const dummyMQConnString = "amqp://guest:guest@localhost:5672/"

func TestInitMQ_Success(t *testing.T) {
	config.ResetMQState()

	// 1. Create Mocks
	mockConnection := new(MockAmqpConnection)
	// Create dummy channels (just need non-nil pointers usually)
	mockConsumerChan := &amqp091.Channel{}
	mockProducerChan := &amqp091.Channel{}

	// 2. Set up Expectations
	// Expect Channel() to be called twice
	mockConnection.On("Channel").Return(mockConsumerChan, nil).Once() // First call
	mockConnection.On("Channel").Return(mockProducerChan, nil).Once() // Second call

	// 3. Patch the dialer
	originalDial := *config.AmqpDial
	*config.AmqpDial = func(url string) (config.AmqpConnection, error) {
		// Assert the connection string if needed: assert.Equal(t, dummyMQConnString, url)
		// Return the mock connection (satisfies interface) and no error
		return mockConnection, nil
	}
	t.Cleanup(func() {
		*config.AmqpDial = originalDial
		config.ResetMQState() // Ensure full reset
	})

	// 4. Call function under test
	err := config.InitMQ(dummyMQConnString)

	// 5. Assertions
	assert.NoError(t, err, "InitMQ should succeed")
	// Assert that package variables are set (adjust type if MQConn is interface)
	assert.NotNil(t, config.MQConn, "MQConn should be set") // Check concrete type if needed
	assert.Same(t, mockConsumerChan, config.ConsumerChannel, "ConsumerChannel should be the first mock channel")
	assert.Same(t, mockProducerChan, config.ProducerChannel, "ProducerChannel should be the second mock channel")
	mockConnection.AssertExpectations(t) // Verify Channel() was called twice
}

func TestInitMQ_Dial_Error(t *testing.T) {
	config.ResetMQState()
	expectedErr := errors.New("dial failed")

	originalDial := *config.AmqpDial
	*config.AmqpDial = func(url string) (config.AmqpConnection, error) {
		var nilConn config.AmqpConnection = nil
		return nilConn, expectedErr // Return nil interface and error
	}
	t.Cleanup(func() {
		*config.AmqpDial = originalDial
		config.ResetMQState()
	})

	err := config.InitMQ(dummyMQConnString)

	assert.Error(t, err, "InitMQ should return an error on dial failure")
	assert.ErrorIs(t, err, expectedErr, "Error should be the one from dial")
	assert.Nil(t, config.MQConn, "MQConn should be nil after dial failure")
	assert.Nil(t, config.ConsumerChannel, "ConsumerChannel should be nil after dial failure")
	assert.Nil(t, config.ProducerChannel, "ProducerChannel should be nil after dial failure")
}

func TestInitMQ_ConsumerChannel_Error(t *testing.T) {
	config.ResetMQState()

	mockConnection := new(MockAmqpConnection)
	expectedErr := errors.New("consumer channel failed")

	// Expect Channel() to be called once and return an error
	mockConnection.On("Channel").Return(nil, expectedErr).Once()

	originalDial := *config.AmqpDial
	*config.AmqpDial = func(url string) (config.AmqpConnection, error) {
		return mockConnection, nil // Dial succeeds, returns mock
	}
	t.Cleanup(func() {
		*config.AmqpDial = originalDial
		config.ResetMQState()
	})

	err := config.InitMQ(dummyMQConnString)

	assert.Error(t, err, "InitMQ should return an error on consumer channel failure")
	assert.ErrorIs(t, err, expectedErr, "Error should be the one from the first Channel() call")
	assert.NotNil(t, config.MQConn, "MQConn should be set because dial succeeded") // Check concrete type if needed
	assert.Nil(t, config.ConsumerChannel, "ConsumerChannel should be nil")
	assert.Nil(t, config.ProducerChannel, "ProducerChannel should be nil")
	mockConnection.AssertExpectations(t) // Verify Channel() was called once
}

func TestInitMQ_ProducerChannel_Error(t *testing.T) {
	config.ResetMQState()

	mockConnection := new(MockAmqpConnection)
	mockConsumerChan := &amqp091.Channel{} // Need this for the first successful call
	expectedErr := errors.New("producer channel failed")

	// Expect Channel() called twice: first succeeds, second fails
	mockConnection.On("Channel").Return(mockConsumerChan, nil).Once()
	mockConnection.On("Channel").Return(nil, expectedErr).Once()

	originalDial := *config.AmqpDial
	*config.AmqpDial = func(url string) (config.AmqpConnection, error) {
		return mockConnection, nil // Dial succeeds
	}
	t.Cleanup(func() {
		*config.AmqpDial = originalDial
		config.ResetMQState()
	})

	err := config.InitMQ(dummyMQConnString)

	assert.Error(t, err, "InitMQ should return an error on producer channel failure")
	assert.ErrorIs(t, err, expectedErr, "Error should be the one from the second Channel() call")
	assert.NotNil(t, config.MQConn, "MQConn should be set") // Check concrete type if needed
	assert.Same(t, mockConsumerChan, config.ConsumerChannel, "ConsumerChannel should be set")
	assert.Nil(t, config.ProducerChannel, "ProducerChannel should be nil")
	mockConnection.AssertExpectations(t) // Verify Channel() was called twice
}

type MockSmtpClient struct {
	mock.Mock
}

func (m *MockSmtpClient) StartTLS(config *tls.Config) error {
	args := m.Called(config)
	return args.Error(0)
}

func (m *MockSmtpClient) Auth(a smtp.Auth) error {
	args := m.Called(a)
	return args.Error(0)
}

func (m *MockSmtpClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// --- Helper to create a dummy SmtpConfig ---
func createDummySmtpConfig() *domain.SmtpConfig {
	return &domain.SmtpConfig{
		Host:     "mock.smtp.com",
		Port:     "587",
		Username: "user",
		Auth:     smtp.PlainAuth("", "user", "pass", "mock.smtp.com"),
	}
}

// --- Test Cases for CheckSmtpAuth ---

func TestCheckSmtpAuth_Success(t *testing.T) {
	config.ResetSmtpState() // Optional reset
	dummyCfg := createDummySmtpConfig()

	// 1. Create Mock
	mockClient := new(MockSmtpClient)

	// 2. Set Expectations
	// Use mock.AnythingOfType because comparing tls.Config or smtp.Auth directly is tricky
	mockClient.On("StartTLS", mock.AnythingOfType("*tls.Config")).Return(nil).Once()
	mockClient.On("Auth", mock.Anything).Return(nil).Once()
	mockClient.On("Close").Return(nil).Once() // Expect Close to be called by defer

	// 3. Patch Dialer
	originalDial := *config.SmtpDial

	*config.SmtpDial = func(addr string) (config.SmtpClient, error) {
		// Optional: Assert addr if needed: assert.Equal(t, "mock.smtp.com:587", addr)
		return mockClient, nil // Return mock client, no error
	}
	t.Cleanup(func() {
		*config.SmtpDial = originalDial
		// config.ResetSmtpState() // Or reset here
	})

	// 4. Call function
	err := config.CheckSmtpAuth(dummyCfg)

	// 5. Assertions
	assert.NoError(t, err, "CheckSmtpAuth should succeed")
	mockClient.AssertExpectations(t) // Verify all expected methods were called
}

func TestCheckSmtpAuth_Dial_Error(t *testing.T) {
	config.ResetSmtpState()
	dummyCfg := createDummySmtpConfig()
	expectedErr := errors.New("dial failed miserably")

	// 1. Patch Dialer to return error
	originalDial := *config.SmtpDial
	*config.SmtpDial = func(addr string) (config.SmtpClient, error) {
		var nilClient config.SmtpClient = nil
		return nilClient, expectedErr
	}
	t.Cleanup(func() {
		*config.SmtpDial = originalDial
	})

	// 2. Call function
	err := config.CheckSmtpAuth(dummyCfg)

	// 3. Assertions
	assert.Error(t, err, "CheckSmtpAuth should return an error")
	assert.ErrorContains(t, err, "failed to dial SMTP server") // Check wrapped error message
	assert.ErrorIs(t, err, expectedErr, "Underlying error should be the dial error")
}

func TestCheckSmtpAuth_StartTLS_Error(t *testing.T) {
	config.ResetSmtpState()
	dummyCfg := createDummySmtpConfig()
	expectedErr := errors.New("TLS handshake failed")

	// 1. Create Mock
	mockClient := new(MockSmtpClient)

	// 2. Set Expectations (StartTLS fails, Close still called)
	mockClient.On("StartTLS", mock.AnythingOfType("*tls.Config")).Return(expectedErr).Once()
	mockClient.On("Close").Return(nil).Once() // Defer still runs

	// 3. Patch Dialer (succeeds, returns mock)
	originalDial := *config.SmtpDial
	*config.SmtpDial = func(addr string) (config.SmtpClient, error) {
		return mockClient, nil
	}
	t.Cleanup(func() {
		*config.SmtpDial = originalDial
	})

	// 4. Call function
	err := config.CheckSmtpAuth(dummyCfg)

	// 5. Assertions
	assert.Error(t, err, "CheckSmtpAuth should return an error")
	assert.ErrorContains(t, err, "failed to start TLS")
	assert.ErrorIs(t, err, expectedErr)
	mockClient.AssertExpectations(t) // Verify StartTLS and Close were called
}

func TestCheckSmtpAuth_Auth_Error(t *testing.T) {
	config.ResetSmtpState()
	dummyCfg := createDummySmtpConfig()
	expectedErr := errors.New("invalid credentials")

	// 1. Create Mock
	mockClient := new(MockSmtpClient)

	// 2. Set Expectations (StartTLS succeeds, Auth fails, Close called)
	mockClient.On("StartTLS", mock.AnythingOfType("*tls.Config")).Return(nil).Once()
	mockClient.On("Auth", mock.Anything).Return(expectedErr).Once()

	mockClient.On("Close").Return(nil).Once()

	// 3. Patch Dialer (succeeds, returns mock)
	originalDial := *config.SmtpDial
	*config.SmtpDial = func(addr string) (config.SmtpClient, error) {
		return mockClient, nil
	}
	t.Cleanup(func() {
		*config.SmtpDial = originalDial
	})

	// 4. Call function
	err := config.CheckSmtpAuth(dummyCfg)

	// 5. Assertions
	assert.Error(t, err, "CheckSmtpAuth should return an error")
	assert.ErrorContains(t, err, "authentication failed")
	assert.ErrorIs(t, err, expectedErr)
	mockClient.AssertExpectations(t) // Verify StartTLS, Auth, and Close were called
}
