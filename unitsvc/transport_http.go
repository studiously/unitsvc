package unitsvc

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ory/hydra/oauth2"
	"github.com/studiously/introspector"
	"github.com/studiously/svcerror"
	"github.com/studiously/unitsvc/codes"
	"errors"
)

var (
	ErrBadRequest = errors.New( "request is malformed or invalid")
)

func MakeHTTPHandler(s Service, logger log.Logger, ti oauth2.Introspector) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
		// All endpoints are secured
		httptransport.ServerBefore(introspector.ToHTTPContext()),
	}

	r.Methods("GET").Path("/units/").Handler(httptransport.NewServer(
		introspector.New(ti, "units.list")(e.ListUnitsEndpoint),
		DecodeListUnitsRequest,
		encodeResponse,
		options...
	))
	r.Methods("GET").Path("/units/{unitID}").Handler(httptransport.NewServer(
		introspector.New(ti, "units.get")(e.GetUnitEndpoint),
		DecodeGetUnitRequest,
		encodeResponse,
		options...
	))
	r.Methods("POST").Path("/units/").Handler(httptransport.NewServer(
		introspector.New(ti, "units.create")(e.GetUnitEndpoint),
		DecodeCreateUnitRequest,
		encodeResponse,
		options...
	))
	r.Methods("PATCH").Path("/units/{id}").Handler(httptransport.NewServer(
		introspector.New(ti, "units.rename")(e.RenameUnitEndpoint),
		DecodeRenameUnitRequest,
		encodeResponse,
		options...
	))
	r.Methods("DELETE").Path("/units/{unitID}").Handler(httptransport.NewServer(
		introspector.New(ti, "units.delete")(e.DeleteUnitEndpoint),
		DecodeDeleteUnitRequest,
		encodeResponse,
		options...
	))

	return r
}

func EncodeListUnitsRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(listUnitsRequest)
	classID := url.QueryEscape(r.ClassID.String())
	req.Method, req.URL.Path = "GET", "/units?classID="+classID
	return encodeRequest(ctx, req, request)
}

func EncodeGetUnitRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(getUnitRequest)
	unitID := url.QueryEscape(r.UnitID.String())
	req.Method, req.URL.Path = "GET", "/units/"+unitID
	return encodeRequest(ctx, req, request)
}

func EncodeCreateUnitRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.Method, req.URL.Path = "POST", "/units/"
	return encodeRequest(ctx, req, request)
}

func EncodeRenameUnitRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(renameUnitRequest)
	unitID := url.QueryEscape(r.UnitID.String())
	req.Method, req.URL.Path = "PATCH", "/units/"+unitID
	return encodeRequest(ctx, req, request)
}

func EncodeDeleteUnitRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(deleteUnitRequest)
	unitID := url.QueryEscape(r.UnitID.String())
	req.Method, req.URL.Path = "DELETE", "/units/"+unitID
	return encodeRequest(ctx, req, request)
}

func DecodeListUnitsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	classID, err := uuid.Parse(r.URL.Query().Get("classID"))
	if err != nil {
		return nil, ErrBadRequest
	}
	return listUnitsRequest{ClassID: classID}, nil
}

func DecodeGetUnitRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	unitID, err := uuid.Parse(vars["unitID"])
	if err != nil {
		return nil, ErrBadRequest
	}
	return getUnitRequest{UnitID: unitID}, nil
}

func DecodeCreateUnitRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req createUnitRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func DecodeRenameUnitRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	unitID, err := uuid.Parse(vars["unitID"])
	if err != nil {
		return nil, ErrBadRequest
	}
	var req renameUnitRequest
	req.UnitID = unitID
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func DecodeDeleteUnitRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	unitID, err := uuid.Parse(vars["unitID"])
	if err != nil {
		return nil, ErrBadRequest
	}
	return deleteUnitRequest{unitID}, nil
}

func DecodeListUnitsResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response listUnitsResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeGetUnitResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response getUnitResponse
	err := json.NewDecoder(resp.Body).Decode(&response )
	return response, err
}

func DecodeCreateUnitResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response createUnitResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeRenameUnitResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response renameUnitResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeDeleteUnitResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response deleteUnitResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}


// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encodeRequest likewise JSON-encodes the request to the HTTP request body.
// Don't use it directly as a transport/http.Client EncodeRequestFunc:
// profilesvc endpoints require mutating the HTTP method and request path.
func encodeRequest(_ context.Context, req *http.Request, request interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	case ErrNotFound:
		return http.StatusNotFound
	case ErrBadRequest:
			return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
