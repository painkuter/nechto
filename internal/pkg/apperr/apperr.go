package apperr

import "nechto/internal/pkg/log"

func Check(err error) {
	if err != nil {
		//TODO: logging
		log.Error(err.Error())
		panic(err)
	}
}
