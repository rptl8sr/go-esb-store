package app

import "errors"

var ErrInvalidStoreFactsNumber = errors.New("invalid store facts number")
var ErrEmptyStoreFactsNumber = errors.New("empty store facts number")
var ErrParseStoreFactsNumber = errors.New("unable to parse store facts number")
var ErrInvalidStoreName = errors.New("invalid store name alias")
var ErrInvalidStoreAddress = errors.New("invalid primary address")
