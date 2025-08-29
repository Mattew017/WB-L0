package domain

import "errors"

var OrderNotFoundError = errors.New("order not found")
var OrderAlreadyExistsError = errors.New("order already exists")
