package intrusearch

import (
	"errors"
	"fmt"
	"github.com/intruderlabs/intrusearch/main/domain/entities"
	derrors "github.com/intruderlabs/intrusearch/main/domain/errors"
	"github.com/intruderlabs/intrusearch/main/domain/helpers"
	dresponses "github.com/intruderlabs/intrusearch/main/domain/responses"
	"github.com/intruderlabs/intrusearch/main/infrastructure/requests"
	iresponses "github.com/intruderlabs/intrusearch/main/infrastructure/responses"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	logger "github.com/sirupsen/logrus"
	"strings"
)

const documentActionCreate = "create"

func (itself Client) CreateDocuments(indexName string, documents []entities.Document) dresponses.CreateDocumentsResponse {
	logger.Infof("Creating a batch of documents inside the index '%s'...", indexName)

	request := ""
	for _, value := range documents {
		action := entities.DocumentAction{documentActionCreate: {Id: value.GetId()}}
		request += fmt.Sprintf("%s\n%s\n",
			helpers.NewSerializationHelper().ToString(action),
			helpers.NewSerializationHelper().ToString(value))
	}

	wrapper, mapped := requests.DoRequest(itself.client, opensearchapi.BulkRequest{
		Index: indexName,
		Body:  strings.NewReader(request),
	})

	return mapFromBulkResponse(len(documents), wrapper, mapped)
}

func mapFromBulkResponse(
	total int,
	wrapper dresponses.GenericResponse, mapped []derrors.GenericError,
) dresponses.CreateDocumentsResponse {
	result := dresponses.CreateDocumentsResponse{Total: total}

	if wrapper.Success { // if there's no HTTP error at all
		response := iresponses.BulkResponse{}
		helpers.NewSerializationHelper().FromBytes(wrapper.Body, &response)

		result.Total = len(response.Items)
		for _, item := range response.Items {
			value := item[documentActionCreate]
			if value.Error.Reason == "" {
				result.Successful += 1
			} else {
				result.Failed += 1
				mapped = append(mapped, derrors.GenericError{Type: value.Error.Type, Reason: value.Error.Reason})
			}
		}
	} else {
		result.Failed = total
	}

	if len(mapped) > 0 {
		result.Error = errors.New(derrors.SerializeErrors(mapped))
	}

	return result
}
