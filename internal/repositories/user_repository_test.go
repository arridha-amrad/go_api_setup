package repositories

import (
	"context"
	"database/sql"
	"my-go-api/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	db   *sql.DB
	repo IUserRepository
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Connect to the test database
	db, err := sql.Open("pgx", "postgres://ari:123@localhost:5432/go_db_test?sslmode=disable")
	if err != nil {
		suite.T().Fatal(err)
	}
	suite.db = db
	suite.repo = NewUserRepository(db)

	suite.db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)
	suite.db.Exec(`CREATE TYPE providers AS ENUM ('credentials', 'google')`)
	suite.db.Exec(`CREATE TYPE user_roles AS ENUM ('user', 'admin')`)

	// Create the users table
	_, err = suite.db.Exec(`
		CREATE TABLE
			users (
				id UUID PRIMARY KEY DEFAULT uuid_generate_v4 (),
				username VARCHAR(50) UNIQUE NOT NULL,
				name VARCHAR(100) NOT NULL,
				email VARCHAR(100) UNIQUE NOT NULL,
				password TEXT,
				provider providers DEFAULT 'credentials',
				role user_roles DEFAULT 'user',
				created_at TIMESTAMP(0)
				WITH
					TIME ZONE NOT NULL DEFAULT NOW (),
					updated_at TIMESTAMP(0)
				WITH
					TIME ZONE NOT NULL DEFAULT NOW ()
			)
	`)
	if err != nil {
		suite.T().Fatal(err)
	}
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	// Drop the tokens table
	_, err := suite.db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		suite.T().Fatal(err)
	}
	_, err = suite.db.Exec(`DROP EXTENSION IF EXISTS "uuid-ossp"`)
	if err != nil {
		suite.T().Fatal(err)
	}
	_, err = suite.db.Exec("DROP TYPE IF EXISTS providers")
	if err != nil {
		suite.T().Fatal(err)
	}
	_, err = suite.db.Exec("DROP TYPE IF EXISTS user_roles")
	if err != nil {
		suite.T().Fatal(err)
	}

	// Close the database connection
	suite.db.Close()
}

func (suite *UserRepositoryTestSuite) localInsert() *models.User {
	name := "ari"
	username := "ari08"
	email := "ari@mail.com"
	password := "12345"

	newUser := &models.User{}
	err := suite.db.QueryRow(`
	INSERT INTO users (name, username, email, password)
	VALUES ($1, $2, $3, $4)
	RETURNING id, name, username, email, password, provider, role, created_at, updated_at
`, name, username, email, password).Scan(
		&newUser.ID,
		&newUser.Name,
		&newUser.Username,
		&newUser.Email,
		&newUser.Password,
		&newUser.Provider,
		&newUser.Role,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
	)
	if err != nil {
		suite.T().Fatal(err)
	}
	return newUser
}

func (suite *UserRepositoryTestSuite) TestCreate() {
	// insert params
	name := "ari"
	username := "ari08"
	email := "ari@mail.com"
	password := "12345"
	// insert action
	newUser, err := suite.repo.Create(context.Background(), name, username, email, password)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), newUser)

	// verify returned data
	assert.Equal(suite.T(), name, newUser.Name)
	assert.Equal(suite.T(), username, newUser.Username)
	assert.Equal(suite.T(), email, newUser.Email)
	assert.Equal(suite.T(), password, newUser.Password)
	assert.Equal(suite.T(), "user", newUser.Role)
	assert.Equal(suite.T(), "credentials", newUser.Provider)

	// verify the new user in inserted into database
	var dbUser models.User
	err = suite.db.QueryRow(`
SELECT id, username, name, email, password, provider, role, created_at, updated_at
FROM users
WHERE email = $1
	`, newUser.Email).Scan(&dbUser.ID, &dbUser.Username, &dbUser.Name, &dbUser.Email, &dbUser.Password, &dbUser.Provider, &dbUser.Role, &dbUser.CreatedAt, &dbUser.UpdatedAt)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newUser.Username, dbUser.Username)
	assert.Equal(suite.T(), password, dbUser.Password)
}

func (suite *UserRepositoryTestSuite) TestGetById() {
	newUser := suite.localInsert()
	// TEST
	user, err := suite.repo.GetById(context.Background(), newUser.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newUser, user)

}

func (suite *UserRepositoryTestSuite) TestGetByUsername() {
	testUser := suite.localInsert()
	user, err := suite.repo.GetByUsername(context.Background(), testUser.Username)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testUser, user)
}
func (suite *UserRepositoryTestSuite) TestGetByEmail() {
	testUser := suite.localInsert()
	user, err := suite.repo.GetByEmail(context.Background(), testUser.Email)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), testUser, user)
}
func (suite *UserRepositoryTestSuite) TestUpdate() {
	testUser := suite.localInsert()
	testUser.Name = "Fufufafa"

	// Add a small delay to ensure UpdatedAt is greater than CreatedAt
	time.Sleep(1 * time.Second)

	user, err := suite.repo.Update(context.Background(), testUser)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), user.Name, "Fufufafa")
	assert.Equal(suite.T(), testUser, user)
	u, _ := time.Parse(time.RFC3339, user.UpdatedAt)
	c, _ := time.Parse(time.RFC3339, user.CreatedAt)
	assert.Greater(suite.T(), u.UnixMilli(), c.UnixMilli())
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
