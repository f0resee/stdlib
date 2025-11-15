package extension

import "github.com/f0resee/stdlib/apollo/protocol/http/auth"

var authSign auth.HTTPAuth

func SetHTTPAuth(httpAuth auth.HTTPAuth) {
	authSign = httpAuth
}

func GetHTTPAuth() auth.HTTPAuth {
	return authSign
}
