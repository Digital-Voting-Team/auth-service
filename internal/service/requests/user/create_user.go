package requests

import (
	"encoding/json"
	"github.com/Digital-Voting-Team/auth-service/internal/service/helpers"
	"github.com/Digital-Voting-Team/auth-service/resources"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type CreateUserRequest struct {
	Data resources.User
}

func NewCreateUserRequest(r *http.Request) (CreateUserRequest, error) {
	var request CreateUserRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return request, errors.Wrap(err, "failed to unmarshal")
	}

	return request, request.validate()
}

func (r *CreateUserRequest) validate() error {
	return helpers.MergeErrors(validation.Errors{
		"/data/attributes/username": validation.Validate(&r.Data.Attributes.Username, validation.Required,
			validation.Length(3, 30)),
		"/data/attributes/password": validation.Validate(&r.Data.Attributes.Password, validation.Required,
			validation.Length(3, 120)),
	}).Filter()
}
