package Server

import (
	"crypto/sha512"
	r "github.com/christopherhesse/rethinkgo"
	"github.com/pmylund/go-hmaccrypt"
)

const pepperStr string = "7EglJe24mzWjT96Vb98KtPT1wvA6p72C2EBxD9uuVtQ05sDM90r2OU1MFfFKtOm!"

type LoginDBEntry struct {
	Username string
	Password string
}

func RegisterUser(user, password string, context *ServerContext) (err error) {
	pepper := []byte(pepperStr)
	pass := []byte(password)

	crypt := hmaccrypt.New(sha512.New, pepper)

	var digest []byte
	if digest, err = crypt.Bcrypt(pass, 10); err != nil {
		return
	}

	digestStr := string(digest)

	row := LoginDBEntry{user, digestStr}
	request := r.Table("login").Insert(row)
	err = context.DB.MakeRequest(request).Err()

	return
}

func VerifyPassword(user, password string, context *ServerContext) (valid bool, err error) {
	request := r.Table("login").Filter(r.Row.Attr("Username").Eq(user))
	var result []LoginDBEntry
	err = context.DB.MakeRequest(request).All(&result)

	digestStr := result[0].Password
	digest := []byte(digestStr)
	pass := []byte(password)
	pepper := []byte(pepperStr)

	crypt := hmaccrypt.New(sha512.New, pepper)

	if err = crypt.BcryptCompare(digest, pass); err == nil {
		valid = true
	} else {
		valid = false
	}

	return
}
