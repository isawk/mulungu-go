package core

import (
	"net/http"
	"time"

	"google.golang.org/appengine"

	"github.com/edgedagency/mulungu/constant"
	"github.com/edgedagency/mulungu/logger"
	"github.com/edgedagency/mulungu/util"
	"golang.org/x/net/context"
)

//DatastoreQuery Datastore query
type DatastoreQuery struct {
	Filters []*DatastoreFilter `json:"filters,omitempty"`
}

//AddFilter adds filter to query
func (q *DatastoreQuery) AddFilter(datastoreFilter *DatastoreFilter) *DatastoreQuery {
	q.Filters = append(q.Filters, datastoreFilter)
	return q
}

//DatastoreFilter filter object
type DatastoreFilter struct {
	Key       string      `json:"key"`
	Operation string      `json:"operation,omitempty"`
	Value     interface{} `json:"value"`
}

//DatastoreModel struct representing a cloud function
type DatastoreModel struct {
	Namespace          string                 `json:"-"`
	Kind               string                 `json:"-"`
	Operation          string                 `json:"operation,omitempty"`
	Record             map[string]interface{} `json:"record,omitempty"`
	ExcludeFromIndexes []string               `json:"excludeFromIndexes,omitempty"`
	Query              *DatastoreQuery        `json:"query,omitempty"`
	Context            context.Context        `json:"-"`
}

//NewDatastoreModel returns a new datastore request struct pointer
func NewDatastoreModel(ctx context.Context, namespace, kind string, record map[string]interface{}, excludeFromIndexes []string) *DatastoreModel {
	datastoreRequest := &DatastoreModel{Context: ctx, Namespace: namespace, Kind: kind, Record: record, ExcludeFromIndexes: excludeFromIndexes}
	return datastoreRequest
}

//JSONBytes returns DatastoreCloudFunction as a transportable json byte
func (ds *DatastoreModel) JSONBytes() []byte {
	bytes, err := util.InterfaceToByte(ds)
	if err != nil {
		return nil
	}
	return bytes
}

//Timestamp timestamps record
func (ds *DatastoreModel) Timestamp() {
	timestamp := time.Now()
	//if no createdDate create it
	if _, ok := ds.Record["createdDate"]; !ok {
		ds.Record["createdDate"] = timestamp
	}
	ds.Record["modifiedDate"] = timestamp

	logger.Debugf(ds.Context, "datastore model", "record timestamped %#v", ds.Record)
}

//Save timestamps record
func (ds *DatastoreModel) Save() (map[string]interface{}, error) {
	ds.Timestamp()

	response, responseErr := ds.execute(http.MethodPost, nil)
	if responseErr != nil {
		return nil, responseErr
	}

	return util.ResponseToMap(response)
}

//Update update model by id
func (ds *DatastoreModel) Update(id string) (map[string]interface{}, error) {
	ds.Timestamp()

	response, responseErr := ds.execute(http.MethodPut, map[string]string{"id": id})
	if responseErr != nil {
		return nil, responseErr
	}

	return util.ResponseToMap(response)
}

//Delete delete model by id
func (ds *DatastoreModel) Delete(id string) (map[string]interface{}, error) {
	response, responseErr := ds.execute(http.MethodDelete, map[string]string{"id": id})
	if responseErr != nil {
		return nil, responseErr
	}

	return util.ResponseToMap(response)
}

//Get execute query on datastore
func (ds *DatastoreModel) Get(searchParams map[string]string) ([]interface{}, error) {

	//fixme:switch to query if we have query object
	ds.Operation = "query"

	response, responseErr := ds.execute(http.MethodPost, searchParams)
	if responseErr != nil {
		return nil, responseErr
	}

	reponseMap, responseMapErr := util.ResponseToMap(response)
	if responseMapErr != nil {
		return nil, responseMapErr
	}

	return reponseMap["entities"].([]interface{}), nil
}

func (ds *DatastoreModel) execute(method string, searchParams map[string]string) (*http.Response, error) {
	request, requestErr := util.HTTPNewRequest(ds.Context,
		method,
		util.CloudFunctionGetPath("us-central1", appengine.AppID(ds.Context),
			"dbdatastore"),
		map[string]string{constant.HeaderNamespace: ds.Namespace,
			constant.HeaderKind:        ds.Kind,
			constant.HeaderContentType: "application/json; charset=UTF-8"},
		ds.JSONBytes(), searchParams)

	if requestErr != nil {
		return nil, requestErr
	}

	response, responseErr := util.HTTPRequest(ds.Context, request)
	if responseErr != nil {
		return nil, responseErr
	}

	return response, nil
}