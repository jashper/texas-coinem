package Server

import (
	"crypto/sha512"
	"github.com/pmylund/go-hmaccrypt"
)

const pepperStr string = "7EglJe24mzWjT96Vb98KtPT1wvA6p72C2EBxD9uuVtQ05sDM90r2OU1MFfFKtOm!"

func RegisterUser(user, password string, context ServerContext) (err error) {
	pepper := []byte(pepperStr)
	pass := []byte(password)

	crypt := hmaccrypt.New(sha512.New, pepper)

	//var digest string
	if _, err = crypt.Bcrypt(pass, 10); err != nil {
		return
	}

	//_ = string(digest)

	// TODO: Write user-digest pair to Server DB

	return
}

func VerifyPassword(user, password string, context ServerContext) (valid bool, err error) {
	digestStr := "" // TODO: lookup digest from Server DB (with user)
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
