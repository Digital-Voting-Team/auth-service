package handlers

import (
	"errors"
	"github.com/Digital-Voting-Team/auth-serivce/internal/data"
	"github.com/Digital-Voting-Team/auth-serivce/internal/service/helpers"
	requests "github.com/Digital-Voting-Team/auth-serivce/internal/service/requests/login"
	"github.com/Digital-Voting-Team/auth-serivce/jwt"
	"github.com/Digital-Voting-Team/auth-serivce/resources"
	utils2 "github.com/Digital-Voting-Team/auth-serivce/utils"
	"gitlab.com/distributed_lab/ape"
	"gitlab.com/distributed_lab/ape/problems"
	"net/http"
)

func LoginUser(w http.ResponseWriter, r *http.Request) {
	request, err := requests.NewLoginUserRequest(r)
	if err != nil {
		helpers.Log(r).WithError(err).Info("failed to parse Login User Request")
		ape.Render(w, problems.BadRequest(err))
		return
	}

	checkHash := utils2.HashString(request.Data.Attributes.Username + request.Data.Attributes.Password + "CSCA")
	user := data.User{
		Username:         request.Data.Attributes.Username,
		PasswordHashHint: utils2.Hint(request.Data.Attributes.Password, 4),
		CheckHash:        checkHash,
	}

	foundUser, err := helpers.UsersQ(r).FilterByUsername(user.Username).Get()
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to find user by it's username")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if foundUser == nil {
		helpers.Log(r).WithError(err).Error("there is no such user with username: " + user.Username)
		ape.Render(w, problems.NotFound())
		return
	}

	if foundUser.PasswordHashHint != user.PasswordHashHint || foundUser.CheckHash != user.CheckHash {
		ape.Render(w, problems.BadRequest(errors.New("invalid credentials")))
		return
	}

	token, err := jwt.CreateToken(foundUser.CheckHash, foundUser.ID)
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to create token")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	jwtSample := data.JWT{
		UserID: foundUser.ID,
		JWT:    token,
	}

	var resultToken data.JWT
	checkUser, err := helpers.JWTsQ(r).FilterByUserID(foundUser.ID).Get()
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to get jwt by the user Id")
		ape.RenderErr(w, problems.InternalError())
		return
	}
	if checkUser != nil && checkUser.ID != 0 {
		resultToken, err = helpers.JWTsQ(r).FilterByUserID(foundUser.ID).Update(jwtSample)
	} else {
		resultToken, err = helpers.JWTsQ(r).Insert(jwtSample)
	}
	if err != nil {
		helpers.Log(r).WithError(err).Error("failed to insert/update new token")
		ape.RenderErr(w, problems.InternalError())
		return
	}

	result := resources.JwtResponse{
		Data: resources.Jwt{
			Key: resources.NewKeyInt64(resultToken.ID, resources.JWT),
			Attributes: resources.JwtAttributes{
				Jwt: resultToken.JWT,
			},
		},
	}

	ape.Render(w, result)
}
