// DO NOT EDIT THIS FILE. This file will be overwritten when re-running go-raml.
package types

import (
	"gopkg.in/validator.v2"
)

type TranscodeJob struct {
	UserID   string `json:"userID" validate:"nonzero"`
	VideoURL string `json:"videoURL" validate:"nonzero"`
}

func (s TranscodeJob) Validate() error {

	return validator.Validate(s)
}
