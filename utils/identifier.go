package utils

import (
	"crypto/rand"
	"fmt"
	"log"
)

func GenerateUUID4() string {
	uuidBytes := make([]byte, 16)
	_, err := rand.Read(uuidBytes)
	if err != nil {
		log.Fatal(":::utils::: migraine failed please check the docs ", err)
	}

	uuidBytes[6] = (uuidBytes[6] & 0x0F) | 0x40
	uuidBytes[8] = (uuidBytes[8] & 0x3F) | 0x80

	uuid := fmt.Sprintf(
		"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		uuidBytes[0], uuidBytes[1], uuidBytes[2], uuidBytes[3],
		uuidBytes[4], uuidBytes[5], uuidBytes[6], uuidBytes[7],
		uuidBytes[8], uuidBytes[9], uuidBytes[10], uuidBytes[11],
		uuidBytes[12], uuidBytes[13], uuidBytes[14], uuidBytes[15],
	)

	uuidChecksum := GenerateChecksum(uuid)

	return uuidChecksum
}
