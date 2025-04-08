package intrusearch

import (
	"github.com/intruderlabs/intrusearch/main/domain/errors"
	"github.com/intruderlabs/intrusearch/main/domain/helpers"
	dresponses "github.com/intruderlabs/intrusearch/main/domain/responses"
	"github.com/intruderlabs/intrusearch/main/infrastructure/requests"
	"github.com/intruderlabs/intrusearch/main/infrastructure/responses"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	logger "github.com/sirupsen/logrus"
	"io"
)

func (itself Client) ClientSearchRequest(
	queryPaginationRequest requests.OsSearchRequest,
) (responses.OsResponse, []errors.GenericError) {
	logger.Infoln("Initialize search request on OpenSearch")
	var ignoreUnavailableIndex = true
	wrapper, mapped := requests.DoRequest(itself.client, opensearchapi.SearchRequest{
		Size:              &queryPaginationRequest.Size,
		From:              &queryPaginationRequest.From,
		Query:             queryPaginationRequest.QueryString,
		Index:             queryPaginationRequest.Index,
		IgnoreUnavailable: &ignoreUnavailableIndex,
	})

	response := responses.OsResponse{}

	if wrapper.Success {
		helpers.NewSerializationHelper().FromBytes(wrapper.Body, &response)
	}
	return response, mapped
}

func (itself Client) ClientIdSearchRequest(index string, id string) (responses.OsResponse, []errors.GenericError) {
	logger.Infoln("Initialize search by ID request on OpenSearch")

	var mappedErrors []errors.GenericError
	indexResponse, err := itself.client.Get(index, id)
	if err != nil {
		mappedErrors := append(mappedErrors, errors.GenericError{Type: "osd_error", Reason: err.Error()})
		return responses.OsResponse{}, mappedErrors
	}

	response := responses.OsResponse{}
	indexResponseBytes, err := io.ReadAll(indexResponse.Body)

	if err != nil {
		mappedErrors = append(mappedErrors, errors.GenericError{Type: "response_error", Reason: err.Error()})
		return responses.OsResponse{}, mappedErrors
	}

	logger.Infoln("ClientIdSearchRequest HTTP status code:", indexResponse.StatusCode)

	helpers.NewSerializationHelper().FromBytes(indexResponseBytes, &response)

	response.Hits.Total.Value = indexResponse.StatusCode
	return response, nil
}

func (itself Client) ClientSearchRaw(request requests.OsSearchRequest) (dresponses.GenericResponse, []errors.GenericError) {
	logger.Infoln("Initialize search raw request on OpenSearch")

	var ignoreUnavailableIndex = true
	wrapper, mapped := requests.DoRequest(itself.client, opensearchapi.SearchRequest{
		Size:              &request.Size,
		From:              &request.From,
		Query:             request.QueryString,
		Index:             request.Index,
		IgnoreUnavailable: &ignoreUnavailableIndex,
	})

	if mapped != nil {
		return dresponses.GenericResponse{}, mapped
	}

	return wrapper, nil
}
