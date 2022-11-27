package response

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"repo-scanner/internal/utils/serror"
	"repo-scanner/internal/utils/utint"
	"repo-scanner/internal/utils/utstring"
)

const (
	statusServerError   = http.StatusInternalServerError
	statusValidatorFail = http.StatusBadRequest
	statusOperationFail = http.StatusInternalServerError
	statusAllOk         = http.StatusOK
	statusCreated       = http.StatusCreated
	statusUpdated       = http.StatusOK
	statusDeleted       = http.StatusOK
	statusDataFound     = http.StatusOK
	statusDataNotFound  = http.StatusNotFound
	statusDuplicate     = http.StatusConflict
	statusUndefined     = http.StatusBadRequest
)

const (
	ErrorServer = iota + 1
	SuccessGetDataOk
	SuccessCreated
	SuccessUpdated
	SuccessDeleted
	ErrorParamValidationFail
	ErrorQueryValidationFail
	ErrorPayloadValidationFail
	ErrorUrlValidationFail
	ErrorOperationFail
)

type ResponseBody struct {
	Status  int               `json:"status"`
	Message map[string]string `json:"message"`
	Data    interface{}       `json:"data"`
	Meta    interface{}       `json:"meta"`
}

type ReturningError struct {
	UserMessage     string
	InternalMessage string
	Code            int
	MoreInfo        string
}

type ReturningValue struct {
	Status  int
	Message map[string]string
	Err     []ReturningError
}

func add(status int, eng string, vne string) ReturningValue {
	return ReturningValue{
		Status: status,
		Message: map[string]string{
			"en": eng,
			"vn": vne,
		},
		Err: nil,
	}
}

var mapping = map[int]ReturningValue{
	ErrorServer: add(statusServerError,
		"The server encountered an internal error or misconfiguration and was unable to complete your request",
		"The server encountered an internal error or misconfiguration and was unable to complete your request"),
	SuccessGetDataOk:           add(statusDataFound, "Success", "Success"),
	SuccessCreated:             add(statusCreated, "Success", "Success"),
	SuccessUpdated:             add(statusUpdated, "Success", "Success"),
	SuccessDeleted:             add(statusDeleted, "Success", "Success"),
	ErrorParamValidationFail:   add(statusValidatorFail, "Invalid param provided", "Invalid param provided"),
	ErrorQueryValidationFail:   add(statusValidatorFail, "Invalid query provided", "Invalid query provided"),
	ErrorPayloadValidationFail: add(statusValidatorFail, "Invalid payload provided", "Invalid payload provided"),
	ErrorUrlValidationFail:     add(statusValidatorFail, "Invalid url provided", "Invalid url provided"),
	ErrorOperationFail:         add(statusOperationFail, "Operation fail", "Operation fail"),
}

func New(code int) serror.SError {
	return NewError(code, errors.New("Something when wrong"))
}

func NewError(code int, err error) serror.SError {
	return serror.NewFromErrork(strconv.Itoa(code), err)
}

func NewSError(code int, serr serror.SError) serror.SError {
	if serr != nil && serr.Key() == "raw" {
		return serr
	}

	serr.SetKey(strconv.Itoa(code))
	return serr
}

func ResolveSError(errx serror.SError) serror.SError {
	var code int
	switch {
	case errx.Key() != "-":
		code = int(utint.StringToInt(errx.Key(), ErrorServer))

	case errx.Code() > 0:
		code = errx.Code()
	}

	if code <= 0 {
		return errx
	}

	result := mapping[code]
	errx.AddComments(utstring.Chains(result.Message["en"], result.Message["id"], "Something when wrong"))

	return errx
}

func Result(ctx *gin.Context, code int) {
	ResultWithData(ctx, code, nil)
}

func ResultWithData(ctx *gin.Context, code int, data interface{}) {
	ResultWithMeta(ctx, code, data, nil)
}

func ResultWithMeta(ctx *gin.Context, code int, data interface{}, meta interface{}) {
	result := mapping[code]
	body := ResponseBody{
		Status:  result.Status,
		Message: result.Message,
	}

	if data != nil {
		body.Data = data
	}
	if meta != nil {
		body.Meta = meta
	}

	ctx.JSON(result.Status, body)
}

func ResultSError(ctx *gin.Context, serr serror.SError) {
	if serr == nil {
		ctx.JSON(http.StatusOK, ResponseBody{
			Status: http.StatusOK,
			Message: map[string]string{
				"en": "Success",
				"vn": "Success",
			},
		})
		return
	}

	var (
		code   = int(ErrorServer)
		result ReturningValue
		ok     bool
	)

	switch {
	case serr.Key() == "raw" || serr.Code() == -1:
		result = ReturningValue{
			Status: http.StatusInternalServerError,
			Message: map[string]string{
				"en": serr.Title(),
				"vn": serr.Title(),
			},
		}
		ok = true

	case serr.Key() != "-":
		code = int(utint.StringToInt(serr.Key(), ErrorServer))

	case serr.Code() > 0:
		code = serr.Code()

	default:
		result = ReturningValue{
			Status: http.StatusInternalServerError,
			Message: map[string]string{
				"en": "The server encountered an internal error or misconfiguration and was unable to complete your request",
				"vn": "The server encountered an internal error or misconfiguration and was unable to complete your request",
			},
		}
	}

	if !ok {
		result, ok = mapping[code]
		if !ok {
			if strings.Contains(serr.Error(), "|") {
				result = ReturningValue{
					Status: code,
					Message: map[string]string{
						"en": strings.Split(serr.Error(), "|")[0],
						"vn": strings.Split(serr.Error(), "|")[1],
					},
				}
			} else {
				result = ReturningValue{
					Status: code,
					Message: map[string]string{
						"en": http.StatusText(code),
						"vn": http.StatusText(code),
					},
				}
			}

			/*if serr != nil {
				ctx.AbortWithError(serr.Code(), serr)
			}*/
		}
	}

	body := ResponseBody{
		Status:  result.Status,
		Message: result.Message,
	}

	/*if serr != nil && ok {
		if result.Status != http.StatusBadRequest && result.Status != http.StatusUnauthorized {
			ctx.AbortWithError(serr.Code(), serr)
		} else {
			log.Warn(serr)
		}

		ctx.JSON(result.Status, body)
		return
	}*/

	ctx.JSON(result.Status, body)
}

func ResultError(ctx *gin.Context, code int, err error) {
	result := mapping[code]
	/*if result.Status != http.StatusBadRequest && result.Status != http.StatusUnauthorized {
		if err != nil {
			ctx.AbortWithError(500, err)
		}
	}*/

	if result.Message["en"] == "" || result.Message["vn"] == "" {
		if err != nil {
			result.Message = map[string]string{
				"en": err.Error(),
				"vn": err.Error(),
			}
		} else {
			result.Message = map[string]string{
				"en": "undefined error",
				"vn": "undefined error",
			}
		}

	}

	body := ResponseBody{
		Status:  result.Status,
		Message: result.Message,
	}

	ctx.JSON(result.Status, body)
	return
}
