package classsvc

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
	"github.com/ory/hydra/sdk"
	"github.com/studiously/classsvc/codes"
	"github.com/studiously/introspector"
	"github.com/studiously/svcerror"
)

var (
	ErrBadRequest = svcerror.New(codes.BadRequest, "the request is malformed or invalid")
)

func MakeHTTPHandler(s Service, logger log.Logger, client *sdk.Client) http.Handler {
	r := mux.NewRouter()
	e := MakeServerEndpoints(s)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(EncodeError),
		httptransport.ServerBefore(introspector.ToHTTPContext()),
	}

	// GET /classes/
	// Get a list of classes the user has access to.
	r.Methods("GET").Path("/classes/{class}").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.get")(e.GetClassEndpoint),
		DecodeGetClassRequest,
		EncodeResponse,
		append(options, httptransport.ServerBefore(introspector.ToHTTPContext()))...
	))

	r.Methods("POST").Path("/classes/").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.new")(e.CreateClassEndpoint),
		DecodeCreateClassRequest,
		EncodeResponse,
		options...
	))

	r.Methods("GET").Path("/classes/").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.list")(e.ListClassesEndpoint),
		DecodeListClassesRequest,
		EncodeResponse,
		options...
	))

	r.Methods("PATCH").Path("/classes/{class}").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.update")(e.UpdateClassEndpoint),
		DecodeUpdateClassRequest,
		EncodeResponse,
		options...
	))

	r.Methods("DELETE").Path("/classes/{class}").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.delete")(e.DeleteClassEndpoint),
		DecodeDeleteClassRequest,
		EncodeResponse,
		options...
	))

	r.Methods("GET").Path("/classes/{class}/members").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.list_members")(e.ListMembersEndpoint),
		DecodeListMembersRequest,
		EncodeResponse,
		options...
	))

	r.Methods("POST").Path("/classes/{class}/join").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.join")(e.JoinClassEndpoint),
		DecodeJoinClassRequest,
		EncodeResponse,
		options...
	))

	leaveClassServer := httptransport.NewServer(
		introspector.New(client.Introspection, "classes.leave")(e.LeaveClassEndpoint),
		DecodeLeaveClassRequest,
		EncodeResponse,
		options...
	)
	r.Methods("DELETE").Path("/classes/{class}/leave").Handler(leaveClassServer)
	r.Methods("DELETE").Path("/classes/{class}/leave/{user}").Handler(leaveClassServer)

	r.Methods("PATCH").Path("/classes/{class}/members/{user}").Handler(httptransport.NewServer(
		introspector.New(client.Introspection, "classes.members:update")(e.SetRoleEndpoint),
		DecodeSetRoleRequest,
		EncodeResponse,
		options...
	))

	return r
}

func EncodeGetClassRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(getClassRequest)
	classID := url.QueryEscape(r.ClassID.String())
	req.Method, req.URL.Path = "GET", "/classes/"+classID
	return EncodeRequest(ctx, req, request)
}

func DecodeGetClassResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response getClassResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeGetClassRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	class, err := uuid.Parse(vars["class"])
	if err != nil {
		return getClassRequest{}, ErrBadRequest
	}
	return getClassRequest{class}, nil
}

func EncodeCreateClassRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.Method, req.URL.Path = "POST", "/classes/"
	return EncodeRequest(ctx, req, request)
}

func DecodeCreateClassResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response createClassResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeCreateClassRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req createClassRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func EncodeListClassesRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.Method, req.URL.Path = "GET", "/classes/"
	return EncodeRequest(ctx, req, request)
}

func DecodeListClassesResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response listClassesResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeListClassesRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func EncodeUpdateClassRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(updateClassRequest)
	classID := url.QueryEscape(r.ClassID.String())
	req.Method, req.URL.Path = "PATCH", "/classes/"+classID
	return EncodeRequest(ctx, req, request)
}

func DecodeUpdateClassResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response updateClassResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeUpdateClassRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req updateClassRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}

func EncodeDeleteClassRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(deleteClassRequest)
	classID := url.QueryEscape(r.ClassID.String())
	req.Method, req.URL.Path = "DELETE", "/classes/"+classID
	return EncodeRequest(ctx, req, request)
}

func DecodeDeleteClassResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response deleteClassResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeDeleteClassRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	class, err := uuid.Parse(vars["class"])
	if err != nil {
		return deleteClassRequest{}, ErrBadRequest
	}
	return deleteClassRequest{class}, nil
}

func EncodeListMembersRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(listMembersRequest)
	classID := url.QueryEscape(r.ClassID.String())
	req.Method, req.URL.Path = "GET", "/classes/"+classID+"/members"
	return EncodeRequest(ctx, req, request)
}

func DecodeListMembersResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response listMembersResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeListMembersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	class, err := uuid.Parse(vars["class"])
	if err != nil {
		return listMembersRequest{}, ErrBadRequest
	}
	return listMembersRequest{class}, nil
}

func EncodeJoinClassRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(joinClassRequest)
	classID := url.QueryEscape(r.ClassID.String())
	req.Method, req.URL.Path = "POST", "/classes/"+classID+"/join"
	return EncodeRequest(ctx, req, request)
}

func DecodeJoinClassResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response joinClassResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeJoinClassRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	class, err := uuid.Parse(vars["class"])
	if err != nil {
		return joinClassRequest{}, ErrBadRequest
	}
	return joinClassRequest{
		ClassID: class,
	}, nil
}

func EncodeLeaveClassRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(leaveClassRequest)
	classID := url.QueryEscape(r.ClassID.String())
	req.Method = "DELETE"
	if r.UserID == uuid.Nil {
		req.URL.Path = "/classes/" + classID + "/leave"
	} else {
		userID := url.QueryEscape(r.UserID.String())
		req.URL.Path = "/classes/" + classID + "/leave/" + userID
	}
	return EncodeRequest(ctx, req, request)
}

func DecodeLeaveClassResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response leaveClassResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeLeaveClassRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req leaveClassRequest
	vars := mux.Vars(r)
	class, err := uuid.Parse(vars["class"])
	if err != nil {
		return leaveClassRequest{}, ErrBadRequest
	}
	req.ClassID = class
	userS, ok := vars["user"]
	user := uuid.Nil
	if ok {
		user, err = uuid.Parse(userS)
		if err != nil {
			return nil, ErrBadRequest
		}
	}
	req.UserID = user
	return req, nil
}

func EncodeSetRoleRequest(ctx context.Context, req *http.Request, request interface{}) error {
	r := request.(setRoleRequest)
	classID := url.QueryEscape(r.ClassID.String())
	userID := url.QueryEscape(r.UserID.String())
	req.Method, req.URL.Path = "PATCH", "/classes/"+classID+"/members/"+userID
	return EncodeRequest(ctx, req, request)
}

func DecodeSetRoleResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response setRoleResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func DecodeSetRoleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req setRoleRequest
	vars := mux.Vars(r)

	class, err := uuid.Parse(vars["class"])
	if err != nil {
		return setRoleRequest{}, ErrBadRequest
	}
	req.ClassID = class
	user, err := uuid.Parse(vars["user"])
	if err != nil {
		return nil, ErrBadRequest
	}
	req.UserID = user
	if e := json.NewDecoder(r.Body).Decode(&req.Role); e != nil {
		return nil, e
	}
	return req, nil
}

// errorer is implemented by all concrete response types that may contain
// errors. It allows us to change the HTTP response code without needing to
// trigger an endpoint (transport-level) error. For more information, read the
// big comment in endpoints.go.
type errorer interface {
	error() error
}

// EncodeResponse is the common method to Encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func EncodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a Go kit transport error, but a business-logic error.
		// Provide those as HTTP errors.
		EncodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// EncodeRequest likewise JSON-Encodes the request to the HTTP request body.
// Don't use it directly as a transport/http.Client EncodeRequestFunc:
// profilesvc endpoints require mutating the HTTP method and request path.
func EncodeRequest(_ context.Context, req *http.Request, request interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(&buf)
	return nil
}

func EncodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("EncodeError with nil error")
	}
	if err, ok := err.(svcerror.Error); !ok {
		err = svcerror.Wrap(codes.Nil, err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(httpStatusFrom(err.(svcerror.Error).Status()))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error_code": err.(svcerror.Error).Status(),
		"error":      err.Error(),
	})
}

func httpStatusFrom(code int) int {
	switch code {
	case codes.MustSetOwner:
		return http.StatusBadRequest
	case codes.UserEnrolled:
		return http.StatusBadRequest
	case codes.Forbidden:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
