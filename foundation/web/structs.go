package web

type VigieResponseAPI struct {
	Meta VigieReponseMetaAPI `json:"meta"`
	Data interface{}         `json:"data"`
}

type VigieReponseMetaAPI struct {
	Success    bool   `json:"success"`
	HTTPStatus int    `json:"httpStatus"`
	Message    string `json:"message"`
	ErrorType  string `json:"errorType,omitempty"`
	ErrorCode  int    `json:"errorCode,omitempty"`
	ErrorTrace string `json:"errorTrace,omitempty"`
}

// NewVigieResponseAPI returns a new VigieResponseAPI
func NewVigieResponseAPI(data any, meta VigieReponseMetaAPI) *VigieResponseAPI {

	succ := false
	if meta.HTTPStatus >= 200 && meta.HTTPStatus < 300 {
		succ = true
	}

	return &VigieResponseAPI{
		Meta: VigieReponseMetaAPI{
			Success:    succ,
			HTTPStatus: meta.HTTPStatus,
			Message:    meta.Message,
			ErrorType:  meta.ErrorType,
			ErrorCode:  meta.ErrorCode,
			ErrorTrace: meta.ErrorTrace,
		},
		Data: data,
	}
}
