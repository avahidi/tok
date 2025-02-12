/*
 * This file contains default values and configurations
 */
package main

import (
	"encoding/binary"
)

const (
	DATABASE_VERSION   uint32 = 1
	DATABASE_FILENAME         = ".tokdb"
	PASSWORD_SALT_SIZE        = 32
	PBKDF2_ITERATIONS         = 45821

	DEFAULT_HASH   = "sha1"
	DEFAULT_DIGITS = 6
	DEFAULT_PERIOD = 30

	SHOWN_TIME = 30
)

var (
	DATABASE_MAGIC [4]byte = [4]byte{'t', 'o', 'k', 'B'}
	BYTE_ORDER             = binary.BigEndian
)
