package services

import (
	"net/http"
	"sync"

	"github.com/tajud99n/go-micro/src/api/config"
	"github.com/tajud99n/go-micro/src/api/domain/github"
	"github.com/tajud99n/go-micro/src/api/domain/repositories"
	providers "github.com/tajud99n/go-micro/src/api/providers/github"
	"github.com/tajud99n/go-micro/src/api/utils/errors"
)

type repoService struct {
}

type repoServiceInterface interface {
	CreateRepo(request repositories.CreateRepoRequest) (*repositories.CreateRepoResponse, errors.APIError)
	CreateRepos(request []repositories.CreateRepoRequest) (repositories.CreateReposResponse, errors.APIError)
}

var (
	RepositoryService repoServiceInterface
)

func init() {
	RepositoryService = &repoService{}
}

func (s *repoService) CreateRepo(input repositories.CreateRepoRequest) (*repositories.CreateRepoResponse, errors.APIError) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	request := github.CreateRepoRequest{
		Name:        input.Name,
		Description: input.Description,
		Private:     false,
	}

	response, err := providers.CreateRepo(config.GetGithubAccessToken(), request)
	if err != nil {
		return nil, errors.NewApiError(err.StatusCode, err.Message)
	}

	result := repositories.CreateRepoResponse{
		Id:    response.Id,
		Name:  response.Name,
		Owner: response.Owner.Login,
	}
	return &result, nil
}

func (s *repoService) CreateRepos(requests []repositories.CreateRepoRequest) (repositories.CreateReposResponse, errors.APIError) {
	inputChan := make(chan repositories.CreateRepositoriesResult)
	outputChan := make(chan repositories.CreateReposResponse)
	defer close(outputChan)

	var wg sync.WaitGroup

	go s.handleRepoResults(&wg, inputChan, outputChan)

	// n requests to process
	for _, current := range requests {
		wg.Add(1)
		go s.createRepoConcurrent(current, inputChan)
	}

	wg.Wait()
	close(inputChan)

	result := <-outputChan

	successCreations := 0
	for _, current := range result.Results {
		if current.Response != nil {
			successCreations++
		}
	}
	if successCreations == 0 {
		result.StatusCode = result.Results[0].Error.Status()
	} else if successCreations == len(requests) {
		result.StatusCode = http.StatusCreated
	} else {
		result.StatusCode = http.StatusPartialContent
	}

	return result, nil
}

func (s *repoService) handleRepoResults(wg *sync.WaitGroup, input chan repositories.CreateRepositoriesResult, output chan repositories.CreateReposResponse) {
	var results repositories.CreateReposResponse

	for incomingEvent := range input {
		repoResult := repositories.CreateRepositoriesResult{
			Response: incomingEvent.Response,
			Error:    incomingEvent.Error,
		}
		results.Results = append(results.Results, repoResult)
		wg.Done()
	}

	output <- results
}

func (s *repoService) createRepoConcurrent(input repositories.CreateRepoRequest, output chan repositories.CreateRepositoriesResult) {
	if err := input.Validate(); err != nil {
		output <- repositories.CreateRepositoriesResult{Error: err}
		return
	}

	result, err := s.CreateRepo(input)
	if err != nil {
		output <- repositories.CreateRepositoriesResult{
			Error: err,
		}
		return
	}

	output <- repositories.CreateRepositoriesResult{
		Response: result,
	}
}
