package repositories

import (
	"context"
	"database/sql"
	"my-go-api/internal/models"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib" // Import the pgx driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TokenRepositoryTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo ITokenRepository
}

func (suite *TokenRepositoryTestSuite) SetupTest() {
	// Connect to the test database
	db, err := sql.Open("pgx", "postgres://ari:123@localhost:5432/go_db_test?sslmode=disable")
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.db = db
	suite.repo = NewTokenRepository(db)

	// Create the tokens table
	_, err = suite.db.Exec(`
		CREATE TABLE IF NOT EXISTS tokens (
			id BIGSERIAL PRIMARY KEY,
			hash TEXT NOT NULL,
			is_revoked BOOLEAN NOT NULL DEFAULT false,
			device_id UUID NOT NULL,
			user_id UUID NOT NULL,
			expired_at TIMESTAMP(0)
    WITH
      TIME ZONE NOT NULL DEFAULT NOW () + INTERVAL '365 days'
		)
	`)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *TokenRepositoryTestSuite) TearDownTest() {
	// Drop the tokens table
	_, err := suite.db.Exec("DROP TABLE IF EXISTS tokens")
	if err != nil {
		suite.T().Fatal(err)
	}

	// Close the database connection
	suite.db.Close()
}

func (suite *TokenRepositoryTestSuite) TestInsert() {
	// Test data
	userId := uuid.New()
	deviceId := uuid.New()
	hash := "helloworld"

	// Call the Insert method
	token, err := suite.repo.Insert(context.Background(), userId, deviceId, hash)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), token)

	// Verify the returned token
	assert.Equal(suite.T(), hash, token.Hash)
	assert.Equal(suite.T(), userId, token.UserId)
	assert.Equal(suite.T(), deviceId, token.DeviceId)
	assert.False(suite.T(), token.IsRevoked)

	// Parse ExpiredAt string into time.Time
	expiredAt, err := time.Parse(time.RFC3339, token.ExpiredAt)
	assert.NoError(suite.T(), err)

	// Verify that ExpiredAt is within 1 hour of the current time
	assert.WithinDuration(suite.T(), time.Now().Add((365 * 24 * time.Hour)), expiredAt, time.Second)

	// Verify the token was inserted into the database
	var dbToken models.Token
	err = suite.db.QueryRow(`
		SELECT id, hash, is_revoked, device_id, user_id, expired_at
		FROM tokens
		WHERE user_id = $1 AND device_id = $2
	`, userId, deviceId).Scan(
		&dbToken.ID, &dbToken.Hash, &dbToken.IsRevoked, &dbToken.DeviceId, &dbToken.UserId, &dbToken.ExpiredAt,
	)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), token, &dbToken)
}

func (suite *TokenRepositoryTestSuite) TestGetToken() {
	// Insert a test token
	userId := uuid.New()
	deviceId := uuid.New()
	expiredAt := time.Now().Add(time.Hour)
	_, err := suite.db.Exec(`
		INSERT INTO tokens (id, hash, is_revoked, device_id, user_id, expired_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, 1, "helloworld", false, deviceId, userId, expiredAt)
	if err != nil {
		suite.T().Fatal(err)
	}

	// Test GetToken
	token, err := suite.repo.GetToken(context.Background(), userId, deviceId)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), token)
	assert.Equal(suite.T(), "helloworld", token.Hash)
}

func (suite *TokenRepositoryTestSuite) TestRemoveToken() {
	// Insert a test token
	userId := uuid.New()
	deviceId := uuid.New()
	expiredAt := time.Now().Add(time.Hour)
	_, err := suite.db.Exec(`
		INSERT INTO tokens (id, hash, is_revoked, device_id, user_id, expired_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, 1, "helloworld", false, deviceId, userId, expiredAt)
	if err != nil {
		suite.T().Fatal(err)
	}

	// Test Remove
	err = suite.repo.Remove(context.Background(), userId, deviceId)
	assert.NoError(suite.T(), err)

	// Verify that the token was removed
	token, err := suite.repo.GetToken(context.Background(), userId, deviceId)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), token)
}

func TestTokenRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(TokenRepositoryTestSuite))
}
