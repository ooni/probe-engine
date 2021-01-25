package ooniapimodel

// APIType is the type of an API.
type APIType interface {
	apiType()
}

// MethodType is the type of a method.
type MethodType interface {
	methodType() string
}

// RequestType is the type of a request.
type RequestType interface {
	requestType()
}

// ResponseType is the type of a response.
type ResponseType interface {
	responseType()
}

// URLPathType is the URL path type.
type URLPathType interface {
	urlPathType()
}
