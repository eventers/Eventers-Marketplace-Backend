package response

import (
	"context"
	"encoding/json"
	"eventers-marketplace-backend/logger"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	StatusCode  int
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Status      string `json:"status"`
	Description string
}

func (r ErrorResponse) Error() string {
	return fmt.Sprintf("StatusCode: %d, Success: %t, Message: %s, Status: %s, Description: %s", r.StatusCode, r.Success, r.Message, r.Status, r.Description)
}

func (r ErrorResponse) Send(ctx context.Context, w http.ResponseWriter) {
	logger.Errorf(ctx, r.Error())
	w.WriteHeader(r.StatusCode)
	json.NewEncoder(w).Encode(r)
}

func BadRequest(message, description string) ErrorResponse {
	return ErrorResponse{
		StatusCode:  http.StatusBadRequest,
		Success:     false,
		Message:     message,
		Status:      "BAD REQUEST",
		Description: description,
	}
}

func ResourceNotFound(message, description string) ErrorResponse {
	return ErrorResponse{
		StatusCode:  http.StatusNotFound,
		Success:     false,
		Message:     message,
		Status:      "NOT FOUND",
		Description: description,
	}
}

func Unauthorized() ErrorResponse {
	return ErrorResponse{
		StatusCode: http.StatusUnauthorized,
		Success:    false,
		Message:    "No valid Auth Token",
		Status:     "UNAUTHORISED",
	}
}

func SomethingWrong() ErrorResponse {
	return ErrorResponse{
		StatusCode: http.StatusInternalServerError,
		Success:    false,
		Message:    "Sorry, Something went wrong",
		Status:     "SOMETHING_WRONG",
	}
}

func IncompleteData() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Not enough data to process this request",
		Status:  "INCOMPLETE_DATA",
	}
}

func IncompleteOrInvalidData() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Incomplete or invalid data passed",
		Status:  "INC_INV_DATA",
	}
}

func InvalidData(description string) ErrorResponse {
	return ErrorResponse{
		StatusCode:  http.StatusBadRequest,
		Success:     false,
		Message:     "Invalid data passed",
		Status:      "INVALID_DATA",
		Description: description,
	}
}

func UserExists() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "This User Already Exists",
		Status:  "USER_EXISTS",
	}
}

func UserExistsOrNotValid() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "This User Already Exists",
		Status:  "USER_EXISTS_NV",
	}
}

func InvalidFName() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Invalid First Name",
		Status:  "INVALID_FNAME",
	}
}

func InvalidLName() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Invalid Last Name",
		Status:  "INVALID_LNAME",
	}
}

func InvalidPass() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Invalid Password",
		Status:  "INVALID_PASS",
	}
}

func InvalidEmail() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Invalid Email Address",
		Status:  "INVALID_EMAIL",
	}
}

func DuplicateEntry() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Email already exists",
		Status:  "DUPLICATE_ENTRY",
	}
}

func UserNotExist() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "No such user exists",
		Status:  "USER_NOT_EXIST",
	}
}

func CanNotLogin() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Wrong Username or Password",
		Status:  "CANT_LOGIN",
	}
}

func EmailNotSent() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Email not sent",
		Status:  "EMAIL_NOT_SENT",
	}
}

func OTPExpired() ErrorResponse {
	return ErrorResponse{
		Success:    false,
		Message:    "OTP Expired, Please try again",
		Status:     "OTP_EXPIRED",
		StatusCode: http.StatusGone,
	}
}

func OTPMismatch() ErrorResponse {
	return ErrorResponse{
		Success:    false,
		Message:    "Wrong OTP entered",
		Status:     "OTP_MISMATCH",
		StatusCode: http.StatusBadRequest,
	}
}

func FirebaseUserNotFound() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "There is no user record corresponding to the provided identifier in Firebase.",
		Status:  "auth/user-not-found",
	}
}

func FirebaseInvalidUID() ErrorResponse {
	return ErrorResponse{
		StatusCode: http.StatusForbidden,
		Success:    false,
		Message:    "Failed to authenticate user",
		Status:     "FIREBASE_INVALID_UID",
	}
}

func NotFound() ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: "Requested Resource Not Found",
		Status:  "NOT_FOUND",
	}
}
