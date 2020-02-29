package repositories

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tajud99n/go-micro/src/api/domain/repositories"
	"github.com/tajud99n/go-micro/src/api/services"

	"github.com/tajud99n/go-micro/src/api/utils/errors"
	"github.com/tajud99n/go-micro/src/api/utils/test"

	"github.com/stretchr/testify/assert"
)

var (
	funcCreateRepo func(request repositories.CreateRepoRequest) (*repositories.CreateRepoResponse, errors.APIError)
	funcCreateRepos func(request []repositories.CreateRepoRequest) (repositories.CreateReposResponse, errors.APIError)
)

type repoServiceMock struct {}

func (s *repoServiceMock) CreateRepo(request repositories.CreateRepoRequest) (*repositories.CreateRepoResponse, errors.APIError) {
	return funcCreateRepo(request)
}

func (s *repoServiceMock) CreateRepos(request []repositories.CreateRepoRequest) (repositories.CreateReposResponse, errors.APIError) {
	return funcCreateRepos(request)
}

func TestCreateRepoNoErrorMockingTheEntireService(t *testing.T) {
	services.RepositoryService = &repoServiceMock{}

	funcCreateRepo = func (request repositories.CreateRepoRequest) (*repositories.CreateRepoResponse, errors.APIError) {
		return &repositories.CreateRepoResponse{
			Id: 321,
			Name: "jane doe repo",
			Owner: "jane",
		}, nil
	}
	
	request, _ := http.NewRequest(http.MethodPost, "/repositories", strings.NewReader(`{"name": "testing"}`))
	response := httptest.NewRecorder()
	c := test.GetMockedContext(request, response)

	CreateRepo(c)

	assert.EqualValues(t, http.StatusCreated, response.Code)

	var result repositories.CreateRepoResponse
	err := json.Unmarshal(response.Body.Bytes(), &result)
	assert.Nil(t, err)
	assert.EqualValues(t, 321, result.Id)
	assert.EqualValues(t, "jane doe repo", result.Name)
}

func TestCreateRepoErrorFromGithubMockingTheEntireService(t *testing.T) {
	services.RepositoryService = &repoServiceMock{}

	funcCreateRepo = func (request repositories.CreateRepoRequest) (*repositories.CreateRepoResponse, errors.APIError) {
		return nil, errors.NewBadRequestError("invalid repository name")
	}
	
	request, _ := http.NewRequest(http.MethodPost, "/repositories", strings.NewReader(`{"name": "testing"}`))
	response := httptest.NewRecorder()
	c := test.GetMockedContext(request, response)

	CreateRepo(c)

	assert.EqualValues(t, http.StatusBadRequest, response.Code)

	apiErr, err := errors.NewApiErrFromBytes(response.Body.Bytes())
	assert.Nil(t, err)
	assert.EqualValues(t, http.StatusBadRequest, apiErr.Status())
	assert.EqualValues(t, "invalid repository name", apiErr.Message())
}