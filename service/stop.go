package service

import "context"

type Stopper interface {
	stop(context.Context)
}
