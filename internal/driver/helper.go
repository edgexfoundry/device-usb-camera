package driver

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
)

type EdgeXErrorWrapper struct{}

func (e EdgeXErrorWrapper) CommandError(command string, err error) errors.EdgeX {
	return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to execute %s command", command), err)
}
