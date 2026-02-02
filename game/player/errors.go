package player

import "errors"

var (
	errPlayerNotFound       = errors.New("player not found")
	errPlayerAlreadyExists  = errors.New("player already exists")
	errTooManyPlayers       = errors.New("too many players online")
	errPlayerSessionInvalid = errors.New("invalid player session")
	errPlayerServiceClosed  = errors.New("player service is closed")
)

func IsPlayerNotFound(err error) bool {
	return errors.Is(err, errPlayerNotFound)
}

func IsPlayerAlreadyExists(err error) bool {
	return errors.Is(err, errPlayerAlreadyExists)
}

func IsTooManyPlayers(err error) bool {
	return errors.Is(err, errTooManyPlayers)
}

func IsPlayerSessionInvalid(err error) bool {
	return errors.Is(err, errPlayerSessionInvalid)
}

func IsPlayerServiceClosed(err error) bool {
	return errors.Is(err, errPlayerServiceClosed)
}