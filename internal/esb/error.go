package esb

import "errors"

var ErrNoStoresData = errors.New("got no stores data")
var ErrNoPageToFetch = errors.New("got no page to fetch")

var ErrUnexpectedStatus = errors.New("unexpected http status")
var ErrInvalidStoresCount = errors.New("invalid stores count")
