package onepassword

import (
	"github.com/pkg/errors"
)

// ErrNotSignedIn is returned when an operation needs an authenticated
// session but none is available.
var ErrNotSignedIn = errors.New("you're not signed into your 1password account")
